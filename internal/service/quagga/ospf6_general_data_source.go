// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// ospf6GeneralLookupFields is the public GET allowlist. It is the resource
// refresh contract intersected with fields the stable OSPFv3 node returns.
// Released os-frr 26.7.8 and 26.7.9 also return child arrays from the
// whole-model GET. Those are not lookup state. The node has no password or
// write-only secret. A missing required field is an error, not a synthetic
// model default. A disabled flag stays false. A blank optional router ID or
// metric stays null. Literal metric 0 is valid because the model minimum is 0.
var ospf6GeneralLookupFields = []string{
	"id",
	"enabled",
	"router_id",
	"originate_default",
	"originate_default_always",
	"originate_default_metric",
	"carp_demote",
}

const ospf6GeneralMetricMax int64 = 16777214

var _ datasource.DataSource = &ospf6GeneralDataSource{}

type ospf6GeneralDataSource struct{ client *opnsense.Client }

type ospf6GeneralLookupModel struct {
	ID              types.String `tfsdk:"id"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	RouterID        types.String `tfsdk:"router_id"`
	Originate       types.Bool   `tfsdk:"originate_default"`
	OriginateAlways types.Bool   `tfsdk:"originate_default_always"`
	OriginateMetric types.Int64  `tfsdk:"originate_default_metric"`
	CARPDemote      types.Bool   `tfsdk:"carp_demote"`
}

type ospf6GeneralLookupResponse struct {
	frrReadPresence `json:"-"`
	Enabled         string `json:"enabled"`
	RouterID        string `json:"routerid"`
	Originate       string `json:"originate"`
	OriginateAlways string `json:"originatealways"`
	OriginateMetric string `json:"originatemetric"`
	CARPDemote      string `json:"carp_demote"`
}

func (a *ospf6GeneralLookupResponse) UnmarshalJSON(raw []byte) error {
	type plain ospf6GeneralLookupResponse
	if err := json.Unmarshal(raw, (*plain)(a)); err != nil {
		return err
	}
	return a.recordFields(raw, a)
}

func newOSPF6GeneralDataSource() datasource.DataSource {
	return &ospf6GeneralDataSource{}
}

func (d *ospf6GeneralDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf6_general"
}

func (d *ospf6GeneralDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads OSPFv3 general settings from the stable Quagga GET. Never writes or reconfigures. The only accepted identity is `ospf6`. The public allowlist is the resource refresh contract that the released OSPFv3 node returns: enabled, router ID, default-route origination, origination metric, and CARP demote. Child arrays returned by the whole-model GET are not lookup state. The node has no password or write-only secret. Missing required fields are not replaced with model defaults. A disabled flag stays false. A blank optional router ID or metric stays null. Literal metric 0 is a real value.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Synthetic singleton ID. Must be `ospf6`.",
			},
			"enabled": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Enable OSPFv3. A disabled service stays false and is not replaced with the model default.",
			},
			"router_id": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "OSPFv3 router ID as an IPv4 identifier without a mask. A blank or omitted value stays null and is not replaced with an empty string.",
			},
			"originate_default": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Originate a default route. A disabled value stays false and is not replaced with the model default.",
			},
			"originate_default_always": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Always originate the default route. A disabled value stays false and is not replaced with the model default.",
			},
			"originate_default_metric": dsschema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Metric for the originated default route, from 0 through 16777214. A blank or omitted value stays null. Literal 0 is a real value, not an unset sentinel.",
			},
			"carp_demote": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Demote CARP if OSPFv3 is not converged. A disabled value stays false and is not replaced with the model default.",
			},
		},
	}
}

func (d *ospf6GeneralDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureQuaggaDataSource(req, resp, func(client *opnsense.Client) { d.client = client })
}

func (d *ospf6GeneralDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ospf6GeneralLookupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.ID.IsNull() || config.ID.IsUnknown() || config.ID.ValueString() != ospf6GeneralID {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid OSPFv3 general settings ID",
			fmt.Sprintf("Expected %q for this singleton data source, got %q.", ospf6GeneralID, config.ID.ValueString()),
		)
		return
	}
	result, err := opnsense.GetSingleton[ospf6GeneralLookupResponse](ctx, d.client, ospf6GeneralReqOpts)
	if err != nil {
		var notFound *opnsense.NotFoundError
		if errors.As(err, &notFound) {
			resp.Diagnostics.AddError(
				"OSPFv3 general settings not found",
				"The appliance returned an empty OSPFv3 general settings object. An empty required object is not found and is not replaced with model defaults.",
			)
			return
		}
		resp.Diagnostics.AddError("Error reading OSPFv3 general settings", frrDiagnostic(err))
		return
	}
	if field := ospf6GeneralLookupGap(result); field != "" {
		resp.Diagnostics.AddError(
			"Malformed OSPFv3 general settings response",
			fmt.Sprintf("The appliance returned OSPFv3 general settings without a readable %s. Missing required fields are not replaced with model defaults.", field),
		)
		return
	}
	applyOSPF6GeneralLookup(&config, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func ospf6GeneralLookupGap(result *ospf6GeneralLookupResponse) string {
	if result == nil {
		return "response"
	}
	if !result.hasField("enabled") || !ospf6BoolToken(result.Enabled) {
		return "enabled"
	}
	if _, ok := ospf6RouterID(result.RouterID); !ok {
		return "router_id"
	}
	if !result.hasField("originate") || !ospf6BoolToken(result.Originate) {
		return "originate_default"
	}
	if !result.hasField("originatealways") || !ospf6BoolToken(result.OriginateAlways) {
		return "originate_default_always"
	}
	if _, ok := ospf6Metric(result.OriginateMetric); !ok {
		return "originate_default_metric"
	}
	if !result.hasField("carp_demote") || !ospf6BoolToken(result.CARPDemote) {
		return "carp_demote"
	}
	return ""
}

func ospf6BoolToken(raw string) bool {
	switch strings.TrimSpace(raw) {
	case "0", "1":
		return true
	default:
		return false
	}
}

func ospf6RouterID(raw string) (types.String, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return types.StringNull(), true
	}
	if strings.Contains(trimmed, "/") {
		return types.StringNull(), false
	}
	parsed := net.ParseIP(trimmed)
	if parsed == nil || parsed.To4() == nil {
		return types.StringNull(), false
	}
	return types.StringValue(trimmed), true
}

func ospf6Metric(raw string) (types.Int64, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return types.Int64Null(), true
	}
	for _, r := range trimmed {
		if r < '0' || r > '9' {
			return types.Int64Null(), false
		}
	}
	value, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil || value > ospf6GeneralMetricMax {
		return types.Int64Null(), false
	}
	return types.Int64Value(value), true
}

func applyOSPF6GeneralLookup(model *ospf6GeneralLookupModel, result *ospf6GeneralLookupResponse) {
	model.ID = types.StringValue(ospf6GeneralID)
	model.Enabled = types.BoolValue(strings.TrimSpace(result.Enabled) == "1")
	model.RouterID, _ = ospf6RouterID(result.RouterID)
	model.Originate = types.BoolValue(strings.TrimSpace(result.Originate) == "1")
	model.OriginateAlways = types.BoolValue(strings.TrimSpace(result.OriginateAlways) == "1")
	model.OriginateMetric, _ = ospf6Metric(result.OriginateMetric)
	model.CARPDemote = types.BoolValue(strings.TrimSpace(result.CARPDemote) == "1")
}
