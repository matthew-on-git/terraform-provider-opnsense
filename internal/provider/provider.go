// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package provider implements the OPNsense Terraform provider.
package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/acme"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/auth"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/cron"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/ddclient"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/dhcp"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/dnsmasq"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/firewall"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/haproxy"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/iface"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/ipsec"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/kea"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/monit"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/openvpn"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/quagga"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/syslog"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/system"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/trafficshaper"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/trust"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/unbound"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/wireguard"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// Ensure OpnsenseProvider satisfies the provider interface.
var (
	_ provider.Provider                       = &OpnsenseProvider{}
	_ provider.ProviderWithEphemeralResources = &OpnsenseProvider{}
)

// OpnsenseProvider defines the provider implementation.
type OpnsenseProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// maxRetrySetting bounds retry counts and backoff seconds to avoid int32/int
// overflow while staying far below the time.Duration limit.
const maxRetrySetting = 1<<31 - 1

// OpnsenseProviderModel describes the provider data model.
type OpnsenseProviderModel struct {
	URI        types.String `tfsdk:"uri"`
	APIKey     types.String `tfsdk:"api_key"`
	APISecret  types.String `tfsdk:"api_secret"`
	Insecure   types.Bool   `tfsdk:"insecure"`
	Retries    types.Int64  `tfsdk:"retries"`
	MinBackoff types.Int64  `tfsdk:"min_backoff"`
	MaxBackoff types.Int64  `tfsdk:"max_backoff"`
}

// resolvedConfig holds the resolved provider configuration after
// applying environment variable fallbacks.
type resolvedConfig struct {
	uri        string
	apiKey     string
	apiSecret  string
	insecure   bool
	retries    int
	minBackoff time.Duration
	maxBackoff time.Duration
}

// Metadata sets the provider type name and version.
func (p *OpnsenseProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "opnsense"
	resp.Version = p.version
}

// Schema defines the provider configuration schema.
func (p *OpnsenseProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The OPNsense provider enables management of OPNsense appliance configuration through Terraform.",
		Attributes: map[string]schema.Attribute{
			"uri": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The URI of the OPNsense appliance (e.g., `https://opnsense.example.com`). Can also be set with the `OPNSENSE_URI` environment variable.",
			},
			"api_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The API key for OPNsense authentication. Can also be set with the `OPNSENSE_API_KEY` environment variable.",
			},
			"api_secret": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The API secret for OPNsense authentication. Can also be set with the `OPNSENSE_API_SECRET` environment variable.",
			},
			"insecure": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether to disable TLS certificate verification. Required for self-signed certificates. Defaults to `false`. Can also be set with the `OPNSENSE_ALLOW_INSECURE` environment variable.",
			},
			"retries": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum number of retries for transient OPNsense API failures. Defaults to `3`. Can also be set with the `OPNSENSE_RETRIES` environment variable.",
				Validators:          []validator.Int64{int64validator.Between(1, maxRetrySetting)},
			},
			"min_backoff": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Minimum wait in seconds between retries. Defaults to `1`. Can also be set with the `OPNSENSE_MIN_BACKOFF` environment variable.",
				Validators:          []validator.Int64{int64validator.Between(1, maxRetrySetting)},
			},
			"max_backoff": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum wait in seconds between retries. Defaults to `30`. Can also be set with the `OPNSENSE_MAX_BACKOFF` environment variable.",
				Validators:          []validator.Int64{int64validator.Between(1, maxRetrySetting)},
			},
		},
	}
}

// Configure prepares the OPNsense API client for resources and data sources.
func (p *OpnsenseProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data OpnsenseProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve configuration values with environment variable fallback.
	cfg := resolveProviderConfig(data)

	// Validate required fields.
	validateRequiredConfig(cfg, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the OPNsense API client.
	client, err := opnsense.NewClient(cfg.toClientConfig())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create OPNsense API Client",
			fmt.Sprintf("An unexpected error occurred when creating the OPNsense API client: %s", err),
		)
		return
	}

	// Validate credentials with a low-privilege endpoint.
	validateCredentials(ctx, client, cfg.uri, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// Share the client with resources and data sources.
	resp.ResourceData = client
	resp.DataSourceData = client
	resp.EphemeralResourceData = client
}

// resolveProviderConfig resolves provider configuration from the HCL model
// with environment variable fallback. HCL values take priority.
func resolveProviderConfig(data OpnsenseProviderModel) resolvedConfig {
	return resolvedConfig{
		uri:        envOrValue(data.URI, "OPNSENSE_URI"),
		apiKey:     envOrValue(data.APIKey, "OPNSENSE_API_KEY"),
		apiSecret:  envOrValue(data.APISecret, "OPNSENSE_API_SECRET"),
		insecure:   envOrBoolValue(data.Insecure, "OPNSENSE_ALLOW_INSECURE"),
		retries:    int(envOrPositiveInt64Value(data.Retries, "OPNSENSE_RETRIES")),
		minBackoff: secondsToDuration(envOrPositiveInt64Value(data.MinBackoff, "OPNSENSE_MIN_BACKOFF")),
		maxBackoff: secondsToDuration(envOrPositiveInt64Value(data.MaxBackoff, "OPNSENSE_MAX_BACKOFF")),
	}
}

// toClientConfig maps resolved provider configuration into the API client
// configuration. Zero retry values preserve the client's documented defaults.
func (c resolvedConfig) toClientConfig() opnsense.ClientConfig {
	return opnsense.ClientConfig{
		BaseURL:      c.uri,
		APIKey:       c.apiKey,
		APISecret:    c.apiSecret,
		Insecure:     c.insecure,
		RetryMax:     c.retries,
		RetryWaitMin: c.minBackoff,
		RetryWaitMax: c.maxBackoff,
	}
}

// validateRequiredConfig checks that all required configuration values are
// present and appends diagnostics for any that are missing.
func validateRequiredConfig(cfg resolvedConfig, diags *diag.Diagnostics) {
	if cfg.uri == "" {
		diags.AddError(
			"Missing OPNsense URI",
			"The provider cannot create the OPNsense API client because the URI is missing. "+
				"Set the `uri` attribute in the provider configuration or the `OPNSENSE_URI` environment variable.",
		)
	}
	if cfg.apiKey == "" {
		diags.AddError(
			"Missing OPNsense API Key",
			"The provider cannot create the OPNsense API client because the API key is missing. "+
				"Set the `api_key` attribute in the provider configuration or the `OPNSENSE_API_KEY` environment variable.",
		)
	}
	if cfg.apiSecret == "" {
		diags.AddError(
			"Missing OPNsense API Secret",
			"The provider cannot create the OPNsense API client because the API secret is missing. "+
				"Set the `api_secret` attribute in the provider configuration or the `OPNSENSE_API_SECRET` environment variable.",
		)
	}
}

// validateCredentials checks that the configured credentials are valid.
func validateCredentials(ctx context.Context, client *opnsense.Client, uri string, resp *provider.ConfigureResponse) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.BaseURL()+"/api/core/menu/search", nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create OPNsense Request",
			fmt.Sprintf("Unable to create the OPNsense credential-validation request for %s: %s", uri, err),
		)
		return
	}
	httpResp, err := client.HTTPClient().Do(req)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Connect to OPNsense",
			fmt.Sprintf("Unable to connect to OPNsense at %s: %s", uri, err),
		)
		return
	}
	defer func() { _ = httpResp.Body.Close() }()

	switch httpResp.StatusCode {
	case http.StatusOK:
		tflog.Info(ctx, "OPNsense connection validated", map[string]interface{}{"url": uri})
	case http.StatusUnauthorized, http.StatusForbidden:
		resp.Diagnostics.AddError(
			"OPNsense Authentication Failed",
			"Authentication failed — verify API key and secret.",
		)
	default:
		resp.Diagnostics.AddError(
			"Unexpected OPNsense API Response",
			fmt.Sprintf("Received unexpected status code %d from OPNsense at %s.", httpResp.StatusCode, uri),
		)
	}
}

// envOrValue returns the HCL value if set, otherwise falls back to the
// environment variable. HCL takes priority over environment variables.
func envOrValue(val types.String, envVar string) string {
	if !val.IsNull() && !val.IsUnknown() {
		return val.ValueString()
	}
	return os.Getenv(envVar)
}

// envOrBoolValue returns the HCL bool value if set, otherwise falls back to
// the environment variable. The env var is considered true if set to "true" or "1".
func envOrBoolValue(val types.Bool, envVar string) bool {
	if !val.IsNull() && !val.IsUnknown() {
		return val.ValueBool()
	}
	env := os.Getenv(envVar)
	return env == "true" || env == "1"
}

// envOrPositiveInt64Value returns the HCL value if set, otherwise falls back to
// a positive environment value. Unset, unparsable, or out-of-range environment
// values return 0 so the client defaults apply.
func envOrPositiveInt64Value(val types.Int64, envVar string) int64 {
	if !val.IsNull() && !val.IsUnknown() {
		return val.ValueInt64()
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(os.Getenv(envVar)), 10, 64)
	if err != nil || parsed < 1 || parsed > maxRetrySetting {
		return 0
	}
	return parsed
}

// secondsToDuration converts a validated seconds value to a duration. Zero
// values return zero so the client applies its documented defaults.
func secondsToDuration(seconds int64) time.Duration {
	return time.Duration(seconds) * time.Second
}

// Resources returns the list of resource types supported by this provider.
func (p *OpnsenseProvider) Resources(_ context.Context) []func() resource.Resource {
	var resources []func() resource.Resource
	resources = append(resources, acme.Resources()...)
	resources = append(resources, firewall.Resources()...)
	resources = append(resources, haproxy.Resources()...)
	resources = append(resources, quagga.Resources()...)
	resources = append(resources, system.Resources()...)
	resources = append(resources, unbound.Resources()...)
	resources = append(resources, wireguard.Resources()...)
	resources = append(resources, ipsec.Resources()...)
	resources = append(resources, kea.Resources()...)
	resources = append(resources, openvpn.Resources()...)
	resources = append(resources, ddclient.Resources()...)
	resources = append(resources, dhcp.Resources()...)
	resources = append(resources, dnsmasq.Resources()...)
	resources = append(resources, trafficshaper.Resources()...)
	resources = append(resources, syslog.Resources()...)
	resources = append(resources, cron.Resources()...)
	resources = append(resources, monit.Resources()...)
	resources = append(resources, trust.Resources()...)
	resources = append(resources, auth.Resources()...)
	resources = append(resources, iface.Resources()...)
	return resources
}

// DataSources returns the list of data source types supported by this provider.
func (p *OpnsenseProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	var dataSources []func() datasource.DataSource
	dataSources = append(dataSources, acme.DataSources()...)
	dataSources = append(dataSources, firewall.DataSources()...)
	dataSources = append(dataSources, haproxy.DataSources()...)
	dataSources = append(dataSources, quagga.DataSources()...)
	dataSources = append(dataSources, system.DataSources()...)
	dataSources = append(dataSources, unbound.DataSources()...)
	dataSources = append(dataSources, wireguard.DataSources()...)
	dataSources = append(dataSources, ipsec.DataSources()...)
	dataSources = append(dataSources, kea.DataSources()...)
	dataSources = append(dataSources, openvpn.DataSources()...)
	dataSources = append(dataSources, ddclient.DataSources()...)
	dataSources = append(dataSources, dhcp.DataSources()...)
	dataSources = append(dataSources, dnsmasq.DataSources()...)
	dataSources = append(dataSources, trafficshaper.DataSources()...)
	dataSources = append(dataSources, syslog.DataSources()...)
	dataSources = append(dataSources, cron.DataSources()...)
	dataSources = append(dataSources, monit.DataSources()...)
	dataSources = append(dataSources, trust.DataSources()...)
	dataSources = append(dataSources, auth.DataSources()...)
	dataSources = append(dataSources, iface.DataSources()...)
	return dataSources
}

// EphemeralResources returns non-persistent provider operations. Terraform
// 1.10 or newer is required by Terraform Core to use these blocks.
func (p *OpnsenseProvider) EphemeralResources(_ context.Context) []func() ephemeral.EphemeralResource {
	return openvpn.EphemeralResources()
}

// New returns a new provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OpnsenseProvider{
			version: version,
		}
	}
}
