// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

// Package provider implements the OPNsense Terraform provider.
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/acme"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/ddclient"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/dhcp"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/firewall"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/haproxy"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/ipsec"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/quagga"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/system"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/unbound"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/service/wireguard"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// Ensure OpnsenseProvider satisfies the provider interface.
var _ provider.Provider = &OpnsenseProvider{}

// OpnsenseProvider defines the provider implementation.
type OpnsenseProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// OpnsenseProviderModel describes the provider data model.
type OpnsenseProviderModel struct {
	URI       types.String `tfsdk:"uri"`
	APIKey    types.String `tfsdk:"api_key"`
	APISecret types.String `tfsdk:"api_secret"`
	Insecure  types.Bool   `tfsdk:"insecure"`
}

// resolvedConfig holds the resolved provider configuration after
// applying environment variable fallbacks.
type resolvedConfig struct {
	uri       string
	apiKey    string
	apiSecret string
	insecure  bool
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
				Required:            true,
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
	client, err := opnsense.NewClient(opnsense.ClientConfig{
		BaseURL:   cfg.uri,
		APIKey:    cfg.apiKey,
		APISecret: cfg.apiSecret,
		Insecure:  cfg.insecure,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create OPNsense API Client",
			fmt.Sprintf("An unexpected error occurred when creating the OPNsense API client: %s", err),
		)
		return
	}

	// Validate credentials by calling the firmware status endpoint.
	validateCredentials(ctx, client, cfg.uri, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// Share the client with resources and data sources.
	resp.ResourceData = client
	resp.DataSourceData = client
}

// resolveProviderConfig resolves provider configuration from the HCL model
// with environment variable fallback. HCL values take priority.
func resolveProviderConfig(data OpnsenseProviderModel) resolvedConfig {
	return resolvedConfig{
		uri:       envOrValue(data.URI, "OPNSENSE_URI"),
		apiKey:    envOrValue(data.APIKey, "OPNSENSE_API_KEY"),
		apiSecret: envOrValue(data.APISecret, "OPNSENSE_API_SECRET"),
		insecure:  envOrBoolValue(data.Insecure, "OPNSENSE_ALLOW_INSECURE"),
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

// validateCredentials checks that the configured credentials are valid by
// calling the OPNsense firmware status API endpoint.
func validateCredentials(ctx context.Context, client *opnsense.Client, uri string, resp *provider.ConfigureResponse) {
	httpResp, err := client.HTTPClient().Get(client.BaseURL() + "/api/core/firmware/status")
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
		logFields := map[string]interface{}{"url": uri}

		// Parse version from firmware status response if available.
		var result map[string]interface{}
		if err := json.NewDecoder(httpResp.Body).Decode(&result); err == nil {
			if version, ok := result["product_version"].(string); ok && version != "" {
				logFields["version"] = version
			}
		}

		tflog.Info(ctx, "OPNsense connection validated", logFields)
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
	resources = append(resources, ddclient.Resources()...)
	resources = append(resources, dhcp.Resources()...)
	return resources
}

// DataSources returns the list of data source types supported by this provider.
func (p *OpnsenseProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return firewall.DataSources()
}

// New returns a new provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OpnsenseProvider{
			version: version,
		}
	}
}
