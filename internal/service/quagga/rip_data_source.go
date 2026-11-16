// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// ripLookupFields is the public GET allowlist. It is the resource refresh
// contract intersected with fields the stable RIP node returns. Released
// os-frr 26.7.8 and 26.7.9 also return passiveinterfaces from the whole-model
// GET. The resource never refreshes that field, so it is not lookup state.
// The node has no password or write-only secret. A missing required field is
// an error, not a synthetic model default. A disabled flag stays false. A
// blank optional metric stays null. Version and metric minima are 1, so
// literal 0 is not the resource unset sentinel and fails closed.
var ripLookupFields = []string{
	"id",
	"enabled",
	"version",
	"networks",
	"redistribute",
	"default_metric",
}

const (
	ripVersionLow  int64 = 1
	ripVersionHigh int64 = 2
	ripMetricLow   int64 = 1
	ripMetricHigh  int64 = 16
)

var ripRedistributeTokens = map[string]struct{}{
	"bgp":       {},
	"connected": {},
	"kernel":    {},
	"ospf":      {},
	"static":    {},
}

var _ datasource.DataSource = &ripDataSource{}

type ripDataSource struct{ client *opnsense.Client }

type ripLookupModel struct {
	ID            types.String `tfsdk:"id"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Version       types.Int64  `tfsdk:"version"`
	Networks      types.Set    `tfsdk:"networks"`
	Redistribute  types.Set    `tfsdk:"redistribute"`
	DefaultMetric types.Int64  `tfsdk:"default_metric"`
}

type ripLookupResponse struct {
	frrReadPresence `json:"-"`
	Enabled         string          `json:"enabled"`
	Version         string          `json:"version"`
	Networks        json.RawMessage `json:"networks"`
	Redistribute    json.RawMessage `json:"redistribute"`
	DefaultMetric   string          `json:"defaultmetric"`
}

func (a *ripLookupResponse) UnmarshalJSON(raw []byte) error {
	type plain ripLookupResponse
	if err := json.Unmarshal(raw, (*plain)(a)); err != nil {
		return err
	}
	return a.recordFields(raw, a)
}

func newRIPDataSource() datasource.DataSource { return &ripDataSource{} }

func (d *ripDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_rip"
}

func (d *ripDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads RIP settings from the stable Quagga GET. Never writes or reconfigures. The only accepted identity is `rip`. The public allowlist is the resource refresh contract that the released RIP node returns: enabled, version, networks, redistribute, and default metric. `passiveinterfaces` is returned by the whole-model GET but is not resource refresh state, so it is not lookup state. The node has no password or write-only secret. Missing required fields are not replaced with model defaults. A disabled flag stays false. A blank optional metric stays null. Literal version or metric 0 fails closed because both model minima are 1.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Synthetic singleton ID. Must be `rip`.",
			},
			"enabled": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Enable RIP. A disabled service stays false and is not replaced with the model default.",
			},
			"version": dsschema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "RIP version, 1 or 2. A missing or blank value fails closed and is not replaced with the model default of 2. Literal 0 fails closed.",
			},
			"networks": dsschema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "IPv4 networks in CIDR notation that RIP advertises. An empty stored list stays an empty set. Selected-map key order is not stored order. A value that is not a dotted-quad prefix fails closed.",
			},
			"redistribute": dsschema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Selected redistribution sources: `bgp`, `connected`, `kernel`, `ospf`, or `static`. An empty selection stays an empty set. An unknown selected token fails closed.",
			},
			"default_metric": dsschema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Default metric for redistributed routes, from 1 through 16. A blank or omitted value stays null. Literal 0 is not the resource unset sentinel and fails closed.",
			},
		},
	}
}

func (d *ripDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureQuaggaDataSource(req, resp, func(client *opnsense.Client) { d.client = client })
}

func (d *ripDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ripLookupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.ID.IsNull() || config.ID.IsUnknown() || config.ID.ValueString() != ripID {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid RIP settings ID",
			fmt.Sprintf("Expected %q for this singleton data source, got %q.", ripID, config.ID.ValueString()),
		)
		return
	}
	result, err := opnsense.GetSingleton[ripLookupResponse](ctx, d.client, ripReqOpts)
	if err != nil {
		var notFound *opnsense.NotFoundError
		if errors.As(err, &notFound) {
			resp.Diagnostics.AddError(
				"RIP settings not found",
				"The appliance returned an empty RIP settings object. An empty required object is not found and is not replaced with model defaults.",
			)
			return
		}
		resp.Diagnostics.AddError("Error reading RIP settings", frrDiagnostic(err))
		return
	}
	if field := ripLookupGap(result); field != "" {
		resp.Diagnostics.AddError(
			"Malformed RIP settings response",
			fmt.Sprintf("The appliance returned RIP settings without a readable %s. Missing required fields are not replaced with model defaults.", field),
		)
		return
	}
	applyRIPLookup(&config, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func ripLookupGap(result *ripLookupResponse) string {
	if result == nil {
		return "response"
	}
	if !result.hasField("enabled") || !ospfBoolToken(result.Enabled) {
		return "enabled"
	}
	if _, ok := ripRequiredInt(result.Version, ripVersionLow, ripVersionHigh); !ok {
		return "version"
	}
	if _, ok := ripNetworks(result); !ok {
		return "networks"
	}
	if _, ok := ripRedistribute(result); !ok {
		return "redistribute"
	}
	if _, ok := ospfOptionalInt(result.DefaultMetric, ripMetricLow, ripMetricHigh); !ok {
		return "default_metric"
	}
	return ""
}

func ripRequiredInt(raw string, low, high int64) (types.Int64, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return types.Int64Null(), false
	}
	return ospfOptionalInt(trimmed, low, high)
}

func ripNetworks(result *ripLookupResponse) (types.Set, bool) {
	if !result.hasField("networks") {
		return sliceToSet(nil), true
	}
	keys, ok := ripSelectedKeys(result.Networks)
	if !ok {
		return types.SetNull(types.StringType), false
	}
	var values []string
	for _, key := range keys {
		if key == "" {
			continue
		}
		if !ripIPv4CIDR(key) {
			return types.SetNull(types.StringType), false
		}
		values = append(values, key)
	}
	return sliceToSet(values), true
}

func ripRedistribute(result *ripLookupResponse) (types.Set, bool) {
	if !result.hasField("redistribute") {
		return sliceToSet(nil), true
	}
	keys, ok := ripSelectedKeys(result.Redistribute)
	if !ok {
		return types.SetNull(types.StringType), false
	}
	for _, key := range keys {
		if _, allowed := ripRedistributeTokens[key]; !allowed {
			return types.SetNull(types.StringType), false
		}
	}
	return sliceToSet(keys), true
}

func ripSelectedKeys(raw json.RawMessage) ([]string, bool) {
	trimmed := bytesTrim(raw)
	if bytesEqual(trimmed, []byte("null")) || len(trimmed) == 0 {
		return nil, false
	}
	if bytesEqual(trimmed, []byte("[]")) || bytesEqual(trimmed, []byte(`""`)) {
		return []string{}, true
	}
	if trimmed[0] != '{' {
		return nil, false
	}
	var entries map[string]struct {
		Selected any `json:"selected"`
	}
	if err := json.Unmarshal(trimmed, &entries); err != nil {
		return nil, false
	}
	var selected []string
	for key, entry := range entries {
		if !ripSelected(entry.Selected) {
			continue
		}
		selected = append(selected, key)
	}
	sort.Strings(selected)
	return selected, true
}

func ripSelected(value any) bool {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) == "1"
	case float64:
		return typed == 1
	default:
		return false
	}
}

func ripIPv4CIDR(raw string) bool {
	if strings.Contains(raw, ":") || strings.Count(raw, "/") != 1 {
		return false
	}
	address, prefix, ok := strings.Cut(raw, "/")
	if !ok || !ripDecimal(prefix, 0, 32) {
		return false
	}
	octets := strings.Split(address, ".")
	if len(octets) != 4 {
		return false
	}
	for _, octet := range octets {
		if !ripDecimal(octet, 0, 255) {
			return false
		}
	}
	return true
}

func ripDecimal(raw string, low, high int64) bool {
	if raw == "" {
		return false
	}
	for _, r := range raw {
		if r < '0' || r > '9' {
			return false
		}
	}
	if len(raw) > 1 && raw[0] == '0' {
		return false
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	return err == nil && value >= low && value <= high
}

func applyRIPLookup(model *ripLookupModel, result *ripLookupResponse) {
	model.ID = types.StringValue(ripID)
	model.Enabled = types.BoolValue(strings.TrimSpace(result.Enabled) == "1")
	model.Version, _ = ripRequiredInt(result.Version, ripVersionLow, ripVersionHigh)
	model.Networks, _ = ripNetworks(result)
	model.Redistribute, _ = ripRedistribute(result)
	model.DefaultMetric, _ = ospfOptionalInt(result.DefaultMetric, ripMetricLow, ripMetricHigh)
}

func bytesTrim(raw []byte) []byte {
	return bytes.TrimSpace(raw)
}

func bytesEqual(left, right []byte) bool {
	return bytes.Equal(left, right)
}
