// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// assignmentLookupFields is the public assignment GET allowlist. It matches
// the resource contract and nothing else. The released NetworkInterface model
// has no password or write-only secret. Excluded on purpose:
//   - lock, spoofmac, media, promiscuous mode
//   - identifier, pending_action, icon, optgroup
//   - alias and 6rd/6to4 fields
//
// Unsupported tunnel modes fail closed. They are not rewritten into a safe mode.
var assignmentLookupFields = []string{
	"device",
	"description",
	"mtu",
	"mss",
	"block_private",
	"block_bogons",
	"override_hardware_settings",
	"disable_checksum_offloading",
	"disable_segmentation_offloading",
	"disable_large_receive_offloading",
	"vlan_hardware_filtering",
	"dhcp6_prefix_length",
	"dhcp6_prefix_only",
	"dhcp6_send_prefix_hint",
	"dhcp6_prefix_id",
	"dhcp6_interface_id",
	"dhcp6_identity_association_id",

	"id",
	"name",
	"enabled",
	"ipv4_mode",
	"ipv4_address",
	"ipv4_prefix_length",
	"ipv4_gateway",
	"ipv6_mode",
	"ipv6_address",
	"ipv6_prefix_length",
	"ipv6_gateway",
	"ipv6_track_interface",
	"ipv6_prefix_id",
	"ipv6_interface_id",
	"ipv6_prefix_range",
	"ipv6_identity_association_id",
	"ipv6_allow_manual_dhcpv6",
	"dhcp_hostname",
	"dhcp_reject_from",
	"dhcp_honor_mtu",
	"dhcp_vlan_priority",
	"dhcp_send_options",
	"dhcp_request_options",
	"dhcp6_vlan_priority",
	"dhcp6_request_dns",
	"dhcp6_information_only",
	"dhcp6_rapid_commit",
	"dhcp6_send_options",
	"dhcp6_request_options",
}

var _ datasource.DataSource = &assignmentDataSource{}

type assignmentDataSource struct{ client *opnsense.Client }

func newAssignmentDataSource() datasource.DataSource { return &assignmentDataSource{} }

func (d *assignmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interface_assignment"
}

func (d *assignmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs, err := assignmentLookupAttributes()
	if err != nil {
		resp.Diagnostics.AddError("Invalid interface assignment lookup schema", err.Error())
		return
	}
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads safe enablement, interface controls and IP settings for an existing OPNsense base interface assignment. Identity is the assignment name (`wan`, `lan`, or `optN`), not a UUID. This data source calls the stable assignment GET and never writes or reconfigures. Unmodeled hardware settings, tunnel credentials and response-only fields are not lookup state. Unsupported tunnel modes fail closed.",
		Attributes:          attrs,
	}
}

func (d *assignmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *opnsense.Client.")
		return
	}
	d.client = client
}

func (d *assignmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AssignmentResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	name := ""
	if !config.ID.IsNull() && !config.ID.IsUnknown() {
		name = config.ID.ValueString()
	}
	if !assignmentNamePattern.MatchString(name) {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid interface assignment ID",
			"Lookup ID must be an existing wan, lan, or optN assignment name.",
		)
		return
	}
	result, err := (&assignmentResource{client: d.client}).readCompatibleAssignment(ctx, name)
	if err != nil {
		addAssignmentDiagnostic(&resp.Diagnostics, "Error reading interface assignment", err)
		return
	}
	// Older responses may omit this metadata. If supplied, validate it before
	// fromAPI assigns the requested name to both Terraform identity fields.
	if raw, present := result.Raw["identifier"]; present {
		var identifier string
		if err := json.Unmarshal(raw, &identifier); err != nil || identifier != name {
			resp.Diagnostics.AddError("Interface assignment identity changed", "The lookup refused to publish a response whose identity did not match the requested assignment name.")
			return
		}
	}
	if err := config.fromAPI(ctx, result, name); err != nil {
		resp.Diagnostics.AddError("Error normalizing interface assignment", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func assignmentLookupAttributes() (map[string]dsschema.Attribute, error) {
	var response resource.SchemaResponse
	(&assignmentResource{}).Schema(context.Background(), resource.SchemaRequest{}, &response)
	if response.Diagnostics.HasError() {
		return nil, fmt.Errorf("resource schema: %v", response.Diagnostics)
	}
	if len(response.Schema.Attributes) != len(assignmentLookupFields) {
		return nil, fmt.Errorf("lookup allowlist has %d fields, resource schema has %d", len(assignmentLookupFields), len(response.Schema.Attributes))
	}
	attrs := make(map[string]dsschema.Attribute, len(assignmentLookupFields))
	for _, name := range assignmentLookupFields {
		attribute, ok := response.Schema.Attributes[name]
		if !ok {
			return nil, fmt.Errorf("resource schema is missing allowlisted field %q", name)
		}
		converted, err := assignmentLookupAttribute(name, attribute)
		if err != nil {
			return nil, err
		}
		attrs[name] = converted
	}
	return attrs, nil
}

func assignmentLookupAttribute(name string, attribute schema.Attribute) (dsschema.Attribute, error) {
	switch typed := attribute.(type) {
	case schema.StringAttribute:
		copied := dsschema.StringAttribute{Computed: true, MarkdownDescription: typed.MarkdownDescription}
		if name == "id" {
			copied.Required = true
			copied.Computed = false
			copied.MarkdownDescription = "Existing assignment name and lookup identity: `wan`, `lan`, or `optN`."
			copied.Validators = []validator.String{stringvalidator.RegexMatches(assignmentNamePattern, "must be wan, lan, or optN")}
		}
		return copied, nil
	case schema.BoolAttribute:
		return dsschema.BoolAttribute{Computed: true, MarkdownDescription: typed.MarkdownDescription}, nil
	case schema.Int64Attribute:
		return dsschema.Int64Attribute{Computed: true, MarkdownDescription: typed.MarkdownDescription}, nil
	case schema.SetAttribute:
		return dsschema.SetAttribute{ElementType: typed.ElementType, Computed: true, MarkdownDescription: typed.MarkdownDescription}, nil
	default:
		return nil, fmt.Errorf("allowlisted field %q has unsupported schema type %T", name, attribute)
	}
}
