// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

const assignmentCompatibilityDetail = "This resource requires OPNsense 26.7.6 or newer with the complete expanded interfaces/assignment API (address, gateway, DHCP, set_item, and reconfigure support). Verify the firmware version and API privileges."

var (
	_ resource.Resource                     = &assignmentResource{}
	_ resource.ResourceWithImportState      = &assignmentResource{}
	_ resource.ResourceWithConfigValidators = &assignmentResource{}

	assignmentNamePattern = regexp.MustCompile(`^(wan|lan|opt[0-9]+)$`)
	assignmentReqOpts     = opnsense.ReqOpts{
		GetEndpoint:         "/api/interfaces/assignment/get_item",
		UpdateEndpoint:      "/api/interfaces/assignment/set_item",
		ReconfigureEndpoint: "/api/interfaces/assignment/reconfigure",
		Monad:               "interface",
	}
)

type assignmentResource struct{ client *opnsense.Client }

type assignmentCompatibilityError struct{ detail string }

func (e *assignmentCompatibilityError) Error() string {
	if e.detail == "" {
		return assignmentCompatibilityDetail
	}
	return e.detail + "\n\n" + assignmentCompatibilityDetail
}

func newAssignmentResource() resource.Resource { return &assignmentResource{} }

func (r *assignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interface_assignment"
}

func (r *assignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	optionalString := func(description string) schema.StringAttribute {
		return schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: description}
	}
	optionalBool := func(description string) schema.BoolAttribute {
		return schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: description}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Adopts and manages safe enablement and IP settings for an existing OPNsense base interface assignment. Requires OPNsense 26.7.6 or newer. Destroy removes only Terraform state and deliberately retains the appliance assignment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "Interface assignment name, identical to `name`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required: true, MarkdownDescription: "Existing immutable assignment name: `wan`, `lan`, or `optN`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{stringvalidator.RegexMatches(assignmentNamePattern, "must be wan, lan, or optN")},
			},
			"enabled":      optionalBool("Whether the assigned interface is enabled."),
			"ipv4_mode":    optionalString("IPv4 mode: `none`, `staticv4`, or `dhcp`."),
			"ipv4_address": optionalString("Static IPv4 address without a prefix. Valid only with `ipv4_mode = \"staticv4\"`."),
			"ipv4_prefix_length": schema.Int64Attribute{
				Optional: true, Computed: true, MarkdownDescription: "Static IPv4 prefix length (0-32). Valid only with `ipv4_mode = \"staticv4\"`.",
				Validators: []validator.Int64{int64validator.Between(0, 32)},
			},
			"ipv4_gateway": optionalString("Existing OPNsense IPv4 gateway name. Valid only with `ipv4_mode = \"staticv4\"`."),
			"ipv6_mode":    optionalString("IPv6 mode: `none`, `staticv6`, `dhcp6`, `linklocal`, `slaac`, `idassoc6`, or `track6`."),
			"ipv6_address": optionalString("Static IPv6 address without a prefix. Valid only with `ipv6_mode = \"staticv6\"`."),
			"ipv6_prefix_length": schema.Int64Attribute{
				Optional: true, Computed: true, MarkdownDescription: "Static IPv6 prefix length (0-128). Valid only with `ipv6_mode = \"staticv6\"`.",
				Validators: []validator.Int64{int64validator.Between(0, 128)},
			},
			"ipv6_gateway":  optionalString("Existing OPNsense IPv6 gateway name. Valid only with `ipv6_mode = \"staticv6\"`."),
			"dhcp_hostname": optionalString("Hostname sent by the DHCPv4 client."),
			"dhcp_reject_from": schema.SetAttribute{
				ElementType: types.StringType, Optional: true, Computed: true,
				MarkdownDescription: "IPv4 DHCP server addresses from which offers are rejected.",
			},
			"dhcp_honor_mtu":         optionalBool("Honor the MTU supplied by DHCPv4."),
			"dhcp_vlan_priority":     optionalString("DHCPv4 VLAN priority: empty (disabled) or `0` through `7`."),
			"dhcp_send_options":      optionalString("OPNsense DHCPv4 send-options value."),
			"dhcp_request_options":   optionalString("OPNsense DHCPv4 request-options value."),
			"dhcp6_vlan_priority":    optionalString("DHCPv6 VLAN priority: empty (disabled) or `0` through `7`."),
			"dhcp6_request_dns":      optionalBool("Request DNS information through DHCPv6."),
			"dhcp6_information_only": optionalBool("Use DHCPv6 information-only mode."),
			"dhcp6_rapid_commit":     optionalBool("Enable DHCPv6 rapid commit."),
			"dhcp6_send_options":     optionalString("OPNsense DHCPv6 send-options value."),
			"dhcp6_request_options":  optionalString("OPNsense DHCPv6 request-options value."),
		},
	}

	attrs := resp.Schema.Attributes
	for name, attribute := range assignmentControlAttributes() {
		attrs[name] = attribute
	}
	for name, attribute := range assignmentDelegationAttributes() {
		attrs[name] = attribute
	}
	attrs["ipv4_mode"] = schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "IPv4 mode: `none`, `staticv4`, or `dhcp`.", Validators: []validator.String{stringvalidator.OneOf("none", "staticv4", "dhcp")}}
	attrs["ipv6_mode"] = schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "IPv6 mode: `none`, `staticv6`, `dhcp6`, `linklocal`, `slaac`, `idassoc6`, or `track6`.", Validators: []validator.String{stringvalidator.OneOf("none", "staticv6", "dhcp6", "linklocal", "slaac", "idassoc6", "track6")}}
	priorityValidator := []validator.String{stringvalidator.OneOf("", "0", "1", "2", "3", "4", "5", "6", "7")}
	attrs["dhcp_vlan_priority"] = schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "DHCPv4 VLAN priority: empty (disabled) or `0` through `7`.", Validators: priorityValidator}
	attrs["dhcp6_vlan_priority"] = schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "DHCPv6 VLAN priority: empty (disabled) or `0` through `7`.", Validators: priorityValidator}
}

func (r *assignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *opnsense.Client.")
		return
	}
	r.client = client
}

func (r *assignmentResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{assignmentAddressValidator{}}
}

type assignmentAddressValidator struct{}

func (assignmentAddressValidator) Description(_ context.Context) string {
	return "validates static interface address, prefix, gateway, and mode combinations"
}

func (v assignmentAddressValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (assignmentAddressValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config AssignmentResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	validateAddressFamily(resp, config.IPv4Mode, "staticv4", config.IPv4Address, config.IPv4PrefixLength, config.IPv4Gateway, "ipv4", false)
	validateAddressFamily(resp, config.IPv6Mode, "staticv6", config.IPv6Address, config.IPv6PrefixLength, config.IPv6Gateway, "ipv6", true)
}

func validateAddressFamily(resp *resource.ValidateConfigResponse, mode types.String, staticMode string, address types.String, prefix types.Int64, gateway types.String, label string, ipv6 bool) {
	if mode.IsUnknown() {
		return
	}
	if mode.IsNull() {
		if configuredString(address) || configuredInt(prefix) || configuredString(gateway) {
			resp.Diagnostics.AddAttributeError(path.Root(label+"_mode"), "Missing Interface Address Mode", fmt.Sprintf("%s_mode must be explicitly configured when %s address, prefix length, or gateway is set.", label, strings.ToUpper(label)))
		}
		return
	}
	if mode.ValueString() != staticMode {
		if configuredString(address) || configuredInt(prefix) || configuredString(gateway) {
			resp.Diagnostics.AddAttributeError(path.Root(label+"_mode"), "Invalid Interface Address Configuration", fmt.Sprintf("%s address, prefix length, and gateway may be set only when %s_mode is %s.", strings.ToUpper(label), label, staticMode))
		}
		return
	}
	if !address.IsUnknown() && !configuredString(address) {
		resp.Diagnostics.AddAttributeError(path.Root(label+"_address"), "Missing Static Interface Address", fmt.Sprintf("%s_address is required when %s_mode is %s.", label, label, staticMode))
	}
	if !prefix.IsUnknown() && !configuredInt(prefix) {
		resp.Diagnostics.AddAttributeError(path.Root(label+"_prefix_length"), "Missing Static Interface Prefix", fmt.Sprintf("%s_prefix_length is required when %s_mode is %s.", label, label, staticMode))
	}
	if configuredString(address) {
		parsed, err := netip.ParseAddr(address.ValueString())
		if err != nil || parsed.Is6() != ipv6 {
			resp.Diagnostics.AddAttributeError(path.Root(label+"_address"), "Invalid Static Interface Address", fmt.Sprintf("%s_address must be a valid %s address without a prefix.", label, strings.ToUpper(label)))
		}
	}
}

func configuredString(value types.String) bool {
	return !value.IsNull() && !value.IsUnknown() && value.ValueString() != ""
}

func configuredInt(value types.Int64) bool { return !value.IsNull() && !value.IsUnknown() }

func (r *assignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AssignmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.apply(ctx, &plan, plan.Name.ValueString(), "adopting", &resp.Diagnostics, func(model *AssignmentResourceModel) {
		resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
	})
}

func (r *assignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AssignmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	name := state.ID.ValueString()
	result, err := r.readCompatibleAssignment(ctx, name)
	if err != nil {
		var notFound *opnsense.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		addAssignmentDiagnostic(&resp.Diagnostics, "Error reading interface assignment", err)
		return
	}
	if err := state.fromAPI(ctx, result, name); err != nil {
		resp.Diagnostics.AddError("Error normalizing interface assignment", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *assignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state AssignmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.apply(ctx, &plan, state.ID.ValueString(), "updating", &resp.Diagnostics, func(model *AssignmentResourceModel) {
		resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
	})
}

func (r *assignmentResource) apply(ctx context.Context, plan *AssignmentResourceModel, name, operation string, diags interface {
	AddError(string, string)
}, setState func(*AssignmentResourceModel),
) {
	current, err := r.readCompatibleAssignment(ctx, name)
	if err != nil {
		var notFound *opnsense.NotFoundError
		if errors.As(err, &notFound) {
			diags.AddError("Interface assignment does not exist", fmt.Sprintf("Assignment %q must already exist; this resource adopts assignments and never creates them.", name))
			return
		}
		addAssignmentDiagnostic(diags, "Error reading interface assignment before update", err)
		return
	}
	payload, err := plan.applyToAPI(ctx, current)
	if err != nil {
		diags.AddError("Error preparing interface assignment", err.Error())
		return
	}
	if err := opnsense.Update(ctx, r.client, assignmentReqOpts, payload, name); err != nil {
		var reconfigureErr *opnsense.MutationReconfigureError
		if errors.As(err, &reconfigureErr) {
			if saved, readErr := r.readCompatibleAssignment(ctx, name); readErr == nil {
				if modelErr := plan.fromAPI(ctx, saved, name); modelErr == nil {
					setState(plan)
				}
			}
			diags.AddError("Interface assignment saved but not activated", fmt.Sprintf("%s\n\nOPNsense persisted the assignment, but reconfigure failed and the change may not be active. Fix the appliance reconfigure failure and apply again.", err))
			return
		}
		var authErr *opnsense.AuthError
		var pluginErr *opnsense.PluginNotFoundError
		if errors.As(err, &authErr) || errors.As(err, &pluginErr) {
			diags.AddError("Incompatible OPNsense interface assignment API", (&assignmentCompatibilityError{detail: err.Error()}).Error())
			return
		}
		diags.AddError("Error "+operation+" interface assignment", err.Error())
		return
	}
	result, err := r.readCompatibleAssignment(ctx, name)
	if err != nil {
		recovery := AssignmentResourceModel{}
		if recoveryErr := recovery.fromAPI(ctx, current, name); recoveryErr == nil {
			setState(&recovery)
		}
		addAssignmentDiagnostic(diags, "Error reading interface assignment after update", err)
		return
	}
	if err := plan.fromAPI(ctx, result, name); err != nil {
		diags.AddError("Error normalizing interface assignment", err.Error())
		return
	}
	setState(plan)
}

// Delete intentionally performs no API operation. Removing this resource from
// configuration forgets it from Terraform while retaining the assignment and
// its active configuration on OPNsense.
func (r *assignmentResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *assignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if !assignmentNamePattern.MatchString(req.ID) {
		resp.Diagnostics.AddError("Invalid interface assignment import ID", "Import ID must be an existing wan, lan, or optN assignment name.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

func (r *assignmentResource) readCompatibleAssignment(ctx context.Context, name string) (*assignmentAPIResponse, error) {
	version, err := r.firmwareVersion(ctx)
	if err != nil {
		return nil, err
	}
	if !versionAtLeast(version, 26, 7, 6) {
		return nil, &assignmentCompatibilityError{detail: fmt.Sprintf("OPNsense firmware %q is older than 26.7.6.", version)}
	}
	if err := r.client.AcquireRead(ctx); err != nil {
		return nil, fmt.Errorf("read interface assignment: %w", err)
	}
	defer r.client.ReleaseRead()

	endpoint := assignmentReqOpts.GetEndpoint + "/" + name
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, r.client.BaseURL()+endpoint, nil)
	if err != nil {
		return nil, err
	}
	httpResp, err := r.client.HTTPClient().Do(httpReq) //nolint:gosec // URL from provider-configured client
	if err != nil {
		return nil, opnsense.NewServerError(endpoint, err)
	}
	defer func() { _ = httpResp.Body.Close() }()
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read assignment response: %w", err)
	}
	if httpResp.StatusCode == http.StatusUnauthorized || httpResp.StatusCode == http.StatusForbidden || httpResp.StatusCode == http.StatusNotFound || strings.Contains(strings.ToLower(string(body)), "endpoint not found") {
		return nil, &assignmentCompatibilityError{detail: fmt.Sprintf("Assignment API request failed with HTTP %d.", httpResp.StatusCode)}
	}
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("assignment API returned HTTP %d", httpResp.StatusCode)
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, &assignmentCompatibilityError{detail: fmt.Sprintf("Assignment API returned an invalid response: %s", err)}
	}
	raw, ok := envelope[assignmentReqOpts.Monad]
	trimmed := strings.TrimSpace(string(raw))
	if !ok || trimmed == "" || trimmed == "null" || trimmed == "{}" || trimmed == "[]" {
		return nil, &opnsense.NotFoundError{Message: fmt.Sprintf("interface assignment %q not found", name)}
	}
	var result assignmentAPIResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, &assignmentCompatibilityError{detail: fmt.Sprintf("Assignment API returned an invalid item shape: %s", err)}
	}
	if missing := result.missingRequiredFields(); len(missing) > 0 {
		return nil, &assignmentCompatibilityError{detail: "Assignment API response is missing required fields: " + strings.Join(missing, ", ") + "."}
	}
	if err := validateSafeAssignmentModes(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func validateSafeAssignmentModes(assignment *assignmentAPIResponse) error {
	ipv4Mode := assignmentModeFromAPI(assignment.stringValue("type4"))
	switch ipv4Mode {
	case "none", "staticv4", "dhcp":
	default:
		return fmt.Errorf("existing assignment uses unsupported IPv4 mode %q; supported modes are none, staticv4, and dhcp", ipv4Mode)
	}
	ipv6Mode := assignmentModeFromAPI(assignment.stringValue("type6"))
	switch ipv6Mode {
	case "none", "staticv6", "dhcp6", "linklocal", "slaac", "idassoc6", "track6":
	default:
		return fmt.Errorf("existing assignment uses unsupported IPv6 mode %q; supported modes are none, staticv6, dhcp6, linklocal, slaac, idassoc6, and track6", ipv6Mode)
	}
	if ipv6Mode == "idassoc6" || ipv6Mode == "track6" {
		for _, key := range assignmentDelegationAPIFields {
			if _, ok := assignment.Raw[key]; !ok {
				return &assignmentCompatibilityError{detail: "Assignment API response is missing delegation field " + key + "."}
			}
		}
	}
	return nil
}

func (r *assignmentResource) firmwareVersion(ctx context.Context) (string, error) {
	if err := r.client.AcquireRead(ctx); err != nil {
		return "", fmt.Errorf("read firmware version: %w", err)
	}
	defer r.client.ReleaseRead()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, r.client.BaseURL()+"/api/core/firmware/info", nil)
	if err != nil {
		return "", err
	}
	httpResp, err := r.client.HTTPClient().Do(httpReq) //nolint:gosec // URL from provider-configured client
	if err != nil {
		return "", opnsense.NewServerError("/api/core/firmware/info", err)
	}
	defer func() { _ = httpResp.Body.Close() }()
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", err
	}
	if httpResp.StatusCode != http.StatusOK {
		return "", &assignmentCompatibilityError{detail: fmt.Sprintf("Unable to verify OPNsense firmware (HTTP %d).", httpResp.StatusCode)}
	}
	var info struct {
		ProductVersion string `json:"product_version"`
	}
	if err := json.Unmarshal(body, &info); err != nil || info.ProductVersion == "" {
		return "", &assignmentCompatibilityError{detail: "The firmware info API did not return product_version."}
	}
	return info.ProductVersion, nil
}

func versionAtLeast(version string, required ...int) bool {
	numbers := regexp.MustCompile(`[0-9]+`).FindAllString(version, 3)
	if len(numbers) < 2 {
		return false
	}
	for i, want := range required {
		got := 0
		if i < len(numbers) {
			got, _ = strconv.Atoi(numbers[i])
		}
		if got != want {
			return got > want
		}
	}
	return true
}

func addAssignmentDiagnostic(diags interface{ AddError(string, string) }, title string, err error) {
	var compatibilityErr *assignmentCompatibilityError
	if errors.As(err, &compatibilityErr) {
		diags.AddError("Incompatible OPNsense interface assignment API", compatibilityErr.Error())
		return
	}
	diags.AddError(title, err.Error())
}
