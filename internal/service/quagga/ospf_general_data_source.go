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

// ospfGeneralLookupFields is the public GET allowlist. It is the resource
// refresh contract intersected with fields the stable OSPF node returns.
// Released os-frr 26.7.8 and 26.7.9 also return child arrays from the
// whole-model GET. Those are not lookup state. The resource schema includes
// passive_interfaces, but the released OSPF node has no such field, so it is
// not lookup state and is not replaced with an empty set. The node has no
// password or write-only secret. A missing required field is an error, not a
// synthetic model default. A disabled flag stays false. A blank optional
// router ID, cost reference, or metric stays null. Cost reference minimum is
// 1, so literal 0 is not the resource sentinel and fails closed. Metric
// minimum is 0, so literal 0 is a real value.
var ospfGeneralLookupFields = []string{
	"id",
	"enabled",
	"router_id",
	"cost_reference",
	"log_adjacency_changes",
	"originate_default",
	"originate_default_always",
	"originate_default_metric",
	"carp_demote",
}

const (
	ospfGeneralCostLow    int64 = 1
	ospfGeneralCostHigh   int64 = 4294967
	ospfGeneralMetricLow  int64 = 0
	ospfGeneralMetricHigh int64 = 16777214
)

var _ datasource.DataSource = &ospfGeneralDataSource{}

type ospfGeneralDataSource struct{ client *opnsense.Client }

type ospfGeneralLookupModel struct {
	ID                  types.String `tfsdk:"id"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	RouterID            types.String `tfsdk:"router_id"`
	CostReference       types.Int64  `tfsdk:"cost_reference"`
	LogAdjacencyChanges types.Bool   `tfsdk:"log_adjacency_changes"`
	Originate           types.Bool   `tfsdk:"originate_default"`
	OriginateAlways     types.Bool   `tfsdk:"originate_default_always"`
	OriginateMetric     types.Int64  `tfsdk:"originate_default_metric"`
	CARPDemote          types.Bool   `tfsdk:"carp_demote"`
}

type ospfGeneralLookupResponse struct {
	frrReadPresence     `json:"-"`
	Enabled             string `json:"enabled"`
	RouterID            string `json:"routerid"`
	CostReference       string `json:"costreference"`
	LogAdjacencyChanges string `json:"logadjacencychanges"`
	Originate           string `json:"originate"`
	OriginateAlways     string `json:"originatealways"`
	OriginateMetric     string `json:"originatemetric"`
	CARPDemote          string `json:"carp_demote"`
}

func (a *ospfGeneralLookupResponse) UnmarshalJSON(raw []byte) error {
	type plain ospfGeneralLookupResponse
	if err := json.Unmarshal(raw, (*plain)(a)); err != nil {
		return err
	}
	return a.recordFields(raw, a)
}

func newOSPFGeneralDataSource() datasource.DataSource {
	return &ospfGeneralDataSource{}
}

func (d *ospfGeneralDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf_general"
}

func (d *ospfGeneralDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads OSPF general settings from the stable Quagga GET. Never writes or reconfigures. The only accepted identity is `ospf`. The public allowlist is the resource refresh contract that the released OSPF node returns: enabled, router ID, cost reference, adjacency logging, default-route origination, origination metric, and CARP demote. Child arrays returned by the whole-model GET are not lookup state. `passive_interfaces` is on the resource schema but absent from the released OSPF node, so it is not lookup state. The node has no password or write-only secret. Missing required fields are not replaced with model defaults. A disabled flag stays false. A blank optional router ID, cost reference, or metric stays null. Literal cost reference 0 fails closed because the model minimum is 1. Literal metric 0 is a real value.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Synthetic singleton ID. Must be `ospf`.",
			},
			"enabled": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Enable OSPF. A disabled service stays false and is not replaced with the model default.",
			},
			"router_id": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "OSPF router ID as an IPv4 identifier without a mask. A blank or omitted value stays null and is not replaced with an empty string.",
			},
			"cost_reference": dsschema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Reference bandwidth in Mbit, from 1 through 4294967. A blank or omitted value stays null. Literal 0 is not the resource unset sentinel and fails closed.",
			},
			"log_adjacency_changes": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Log neighbor adjacency changes. A disabled value stays false and is not replaced with the model default.",
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
				MarkdownDescription: "Demote CARP if OSPF is not converged. A disabled value stays false and is not replaced with the model default.",
			},
		},
	}
}

func (d *ospfGeneralDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureQuaggaDataSource(req, resp, func(client *opnsense.Client) { d.client = client })
}

func (d *ospfGeneralDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ospfGeneralLookupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.ID.IsNull() || config.ID.IsUnknown() || config.ID.ValueString() != ospfGeneralID {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid OSPF general settings ID",
			fmt.Sprintf("Expected %q for this singleton data source, got %q.", ospfGeneralID, config.ID.ValueString()),
		)
		return
	}
	result, err := opnsense.GetSingleton[ospfGeneralLookupResponse](ctx, d.client, ospfGeneralReqOpts)
	if err != nil {
		var notFound *opnsense.NotFoundError
		if errors.As(err, &notFound) {
			resp.Diagnostics.AddError(
				"OSPF general settings not found",
				"The appliance returned an empty OSPF general settings object. An empty required object is not found and is not replaced with model defaults.",
			)
			return
		}
		resp.Diagnostics.AddError("Error reading OSPF general settings", frrDiagnostic(err))
		return
	}
	if field := ospfGeneralLookupGap(result); field != "" {
		resp.Diagnostics.AddError(
			"Malformed OSPF general settings response",
			fmt.Sprintf("The appliance returned OSPF general settings without a readable %s. Missing required fields are not replaced with model defaults.", field),
		)
		return
	}
	applyOSPFGeneralLookup(&config, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func ospfGeneralLookupGap(result *ospfGeneralLookupResponse) string {
	if result == nil {
		return "response"
	}
	if !result.hasField("enabled") || !ospfBoolToken(result.Enabled) {
		return "enabled"
	}
	if _, ok := ospfRouterID(result.RouterID); !ok {
		return "router_id"
	}
	if _, ok := ospfOptionalInt(result.CostReference, ospfGeneralCostLow, ospfGeneralCostHigh); !ok {
		return "cost_reference"
	}
	if !result.hasField("logadjacencychanges") || !ospfBoolToken(result.LogAdjacencyChanges) {
		return "log_adjacency_changes"
	}
	if !result.hasField("originate") || !ospfBoolToken(result.Originate) {
		return "originate_default"
	}
	if !result.hasField("originatealways") || !ospfBoolToken(result.OriginateAlways) {
		return "originate_default_always"
	}
	if _, ok := ospfOptionalInt(result.OriginateMetric, ospfGeneralMetricLow, ospfGeneralMetricHigh); !ok {
		return "originate_default_metric"
	}
	if !result.hasField("carp_demote") || !ospfBoolToken(result.CARPDemote) {
		return "carp_demote"
	}
	return ""
}

func ospfBoolToken(raw string) bool {
	switch strings.TrimSpace(raw) {
	case "0", "1":
		return true
	default:
		return false
	}
}

func ospfRouterID(raw string) (types.String, bool) {
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

func ospfOptionalInt(raw string, low, high int64) (types.Int64, bool) {
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
	if err != nil || value < low || value > high {
		return types.Int64Null(), false
	}
	return types.Int64Value(value), true
}

func applyOSPFGeneralLookup(model *ospfGeneralLookupModel, result *ospfGeneralLookupResponse) {
	model.ID = types.StringValue(ospfGeneralID)
	model.Enabled = types.BoolValue(strings.TrimSpace(result.Enabled) == "1")
	model.RouterID, _ = ospfRouterID(result.RouterID)
	model.CostReference, _ = ospfOptionalInt(result.CostReference, ospfGeneralCostLow, ospfGeneralCostHigh)
	model.LogAdjacencyChanges = types.BoolValue(strings.TrimSpace(result.LogAdjacencyChanges) == "1")
	model.Originate = types.BoolValue(strings.TrimSpace(result.Originate) == "1")
	model.OriginateAlways = types.BoolValue(strings.TrimSpace(result.OriginateAlways) == "1")
	model.OriginateMetric, _ = ospfOptionalInt(result.OriginateMetric, ospfGeneralMetricLow, ospfGeneralMetricHigh)
	model.CARPDemote = types.BoolValue(strings.TrimSpace(result.CARPDemote) == "1")
}
