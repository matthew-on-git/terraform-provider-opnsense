// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// bgpGlobalLookupFields is the public GET allowlist. It is the resource refresh
// contract intersected with fields the stable BGP general node returns.
// Released os-frr 26.7.8 and 26.7.9 also return bestpath and child arrays from
// the whole-model GET. Those are not lookup state. The general node has no
// password or write-only secret. A missing required field is an error, not a
// synthetic model default. A present blank optional stays null.
var bgpGlobalLookupFields = []string{
	"id",
	"enabled",
	"as_number",
	"router_id",
	"distance",
	"graceful_restart",
	"network_import_check",
	"enforce_first_as",
	"log_neighbor_changes",
	"networks",
	"maximum_paths",
	"maximum_paths_ibgp",
}

var _ datasource.DataSource = &bgpGlobalDataSource{}

type bgpGlobalDataSource struct{ client *opnsense.Client }

type bgpGlobalLookupModel struct {
	ID                 types.String `tfsdk:"id"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	ASNumber           types.Int64  `tfsdk:"as_number"`
	RouterID           types.String `tfsdk:"router_id"`
	Distance           types.Int64  `tfsdk:"distance"`
	GracefulRestart    types.Bool   `tfsdk:"graceful_restart"`
	NetworkImportCheck types.Bool   `tfsdk:"network_import_check"`
	EnforceFirstAS     types.Bool   `tfsdk:"enforce_first_as"`
	LogNeighborChanges types.Bool   `tfsdk:"log_neighbor_changes"`
	Networks           types.Set    `tfsdk:"networks"`
	MaximumPaths       types.Int64  `tfsdk:"maximum_paths"`
	MaximumPathsIBGP   types.Int64  `tfsdk:"maximum_paths_ibgp"`
}

func newBGPGlobalDataSource() datasource.DataSource {
	return &bgpGlobalDataSource{}
}

func (d *bgpGlobalDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_bgp_global"
}

func (d *bgpGlobalDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads BGP global settings from the stable Quagga/FRR GET. Never writes or reconfigures. The only accepted identity is `bgp`. The public allowlist is the resource refresh contract that the released general node returns: enabled, AS number, router ID, distance, graceful restart, network import check, enforce first AS, log neighbor changes, networks, and maximum paths. Bestpath and child arrays are returned by the whole-model GET and are not lookup state. The general node has no password or write-only secret. Missing required fields are not replaced with model defaults. A present blank router ID, distance, or maximum-paths value stays null. A disabled service stays false. An empty networks list stays an empty set.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Synthetic singleton ID. Must be `bgp`.",
			},
			"enabled": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Enable BGP. A disabled service stays false.",
			},
			"as_number": dsschema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Local autonomous system number. A missing value is not replaced with the model default.",
			},
			"router_id": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "BGP router ID. A present blank value stays null.",
			},
			"distance": dsschema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Administrative distance. A present blank value stays null and is not stored as zero.",
			},
			"graceful_restart": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Enable graceful restart. A disabled value stays false.",
			},
			"network_import_check": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Check BGP network route existence in the IGP before announcement.",
			},
			"enforce_first_as": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Enforce the first AS in the AS-path of eBGP updates.",
			},
			"log_neighbor_changes": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Log neighbor up and down changes.",
			},
			"networks": dsschema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Networks in CIDR notation advertised into BGP. An empty list stays empty.",
			},
			"maximum_paths": dsschema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Maximum eBGP ECMP paths. A present blank value stays null and is not stored as zero.",
			},
			"maximum_paths_ibgp": dsschema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Maximum iBGP ECMP paths. A present blank value stays null and is not stored as zero.",
			},
		},
	}
}

func (d *bgpGlobalDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureQuaggaDataSource(req, resp, func(client *opnsense.Client) { d.client = client })
}

func (d *bgpGlobalDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config bgpGlobalLookupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.ID.IsNull() || config.ID.IsUnknown() || config.ID.ValueString() != bgpGlobalID {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid BGP global settings ID",
			fmt.Sprintf("Expected %q for this singleton data source, got %q.", bgpGlobalID, config.ID.ValueString()),
		)
		return
	}
	result, err := getFRRSingleton[bgpGlobalAPIResponse](ctx, d.client, bgpGlobalReqOpts)
	if err != nil {
		if strings.Contains(err.Error(), "missing or empty settings") {
			resp.Diagnostics.AddError(
				"BGP global settings not found",
				"The appliance returned an empty BGP settings object. An empty required object is not found and is not replaced with model defaults.",
			)
			return
		}
		resp.Diagnostics.AddError("Error reading BGP global settings", frrDiagnostic(err))
		return
	}
	if field := bgpGlobalLookupGap(result); field != "" {
		resp.Diagnostics.AddError(
			"Malformed BGP global settings response",
			fmt.Sprintf("The appliance returned BGP settings without a readable %s. Missing required fields are not replaced with model defaults.", field),
		)
		return
	}
	applyBGPGlobalLookup(&config, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func bgpGlobalLookupGap(result *bgpGlobalAPIResponse) string {
	if result == nil {
		return "response"
	}
	if !result.hasField("enabled") || !bgpBoolToken(result.Enabled) {
		return "enabled"
	}
	if !result.hasField("asnumber") || !bgpASToken(result.ASNumber) {
		return "as_number"
	}
	if !result.hasField("routerid") {
		return "router_id"
	}
	if !result.hasField("distance") || !bgpOptionalInt(result.Distance, 1, 255) {
		return "distance"
	}
	if !result.hasField("graceful") || !bgpBoolToken(result.GracefulRestart) {
		return "graceful_restart"
	}
	if !result.hasField("networkimportcheck") || !bgpBoolToken(result.NetworkImportCheck) {
		return "network_import_check"
	}
	if !result.hasField("enforce_first_as") || !bgpBoolToken(result.EnforceFirstAS) {
		return "enforce_first_as"
	}
	if !result.hasField("logneighborchanges") || !bgpBoolToken(result.LogNeighborChanges) {
		return "log_neighbor_changes"
	}
	if !result.hasField("networks") {
		return "networks"
	}
	if !result.hasField("maximumpaths") || !bgpOptionalInt(result.MaximumPaths, 1, 128) {
		return "maximum_paths"
	}
	if !result.hasField("maximumpathsibgp") || !bgpOptionalInt(result.MaximumPathsIBGP, 1, 128) {
		return "maximum_paths_ibgp"
	}
	return ""
}

func bgpBoolToken(raw string) bool {
	switch strings.TrimSpace(raw) {
	case "0", "1":
		return true
	default:
		return false
	}
}

func bgpASToken(raw string) bool {
	value, ok := bgpTrimmedInt(raw)
	return ok && value >= 1 && value <= 4294967295
}

func bgpOptionalInt(raw string, low, high int64) bool {
	if strings.TrimSpace(raw) == "" {
		return true
	}
	value, ok := bgpTrimmedInt(raw)
	return ok && value >= low && value <= high
}

func bgpTrimmedInt(raw string) (int64, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, false
	}
	value, err := strconv.ParseInt(trimmed, 10, 64)
	return value, err == nil
}

func applyBGPGlobalLookup(model *bgpGlobalLookupModel, result *bgpGlobalAPIResponse) {
	model.ID = types.StringValue(bgpGlobalID)
	model.Enabled = types.BoolValue(strings.TrimSpace(result.Enabled) == "1")
	asNumber, _ := bgpTrimmedInt(result.ASNumber)
	model.ASNumber = types.Int64Value(asNumber)
	model.RouterID = bgpOptionalString(result.RouterID)
	model.Distance = bgpOptionalIntValue(result.Distance)
	model.GracefulRestart = types.BoolValue(strings.TrimSpace(result.GracefulRestart) == "1")
	model.NetworkImportCheck = types.BoolValue(strings.TrimSpace(result.NetworkImportCheck) == "1")
	model.EnforceFirstAS = types.BoolValue(strings.TrimSpace(result.EnforceFirstAS) == "1")
	model.LogNeighborChanges = types.BoolValue(strings.TrimSpace(result.LogNeighborChanges) == "1")
	model.Networks = sliceToSet(result.Networks)
	model.MaximumPaths = bgpOptionalIntValue(result.MaximumPaths)
	model.MaximumPathsIBGP = bgpOptionalIntValue(result.MaximumPathsIBGP)
}

func bgpOptionalString(raw string) types.String {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return types.StringNull()
	}
	return types.StringValue(trimmed)
}

func bgpOptionalIntValue(raw string) types.Int64 {
	if strings.TrimSpace(raw) == "" {
		return types.Int64Null()
	}
	value, _ := bgpTrimmedInt(raw)
	return types.Int64Value(value)
}
