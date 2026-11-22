// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/tfconv"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// dnsblLookupFields is the public allowlist. It is the resource refresh
// contract, not every field the stable blocklist item returns. Released core
// 26.7.8 and 26.7.9 keep dnsbl.blocklist as an ArrayField. get_dnsbl without
// a UUID calls Add and returns a new in-memory node, so it is not a read.
// The resource identity stays unbound-dnsbl. cache_ttl and description are
// required on the item and source_nets is returned, but the resource never
// refreshes them. They are not lookup state. The node has no password or
// write-only secret. A missing or blank required enabled flag fails closed
// and is not replaced with the model default of 1.
var dnsblLookupFields = []string{
	"id",
	"enabled",
	"safe_search",
	"type",
	"lists",
	"whitelists",
	"blocklists",
	"wildcards",
	"address",
	"nxdomain",
}

// dnsblTypeOptions is the released multiple OptionField list. A selected key
// outside this list, including a padded key, is not lookup state.
var dnsblTypeOptions = map[string]struct{}{
	"ag": {}, "el": {}, "ep": {}, "sb": {}, "yy": {},
	"hgz001": {}, "hgz002": {}, "hgz003": {}, "hgz004": {}, "hgz005": {},
	"hgz006": {}, "hgz007": {}, "hgz008": {}, "hgz009": {}, "hgz010": {},
	"hgz011": {}, "hgz012": {}, "hgz013": {}, "hgz014": {}, "hgz015": {},
	"hgz016": {}, "hgz017": {}, "hgz018": {}, "hgz019": {}, "hgz020": {},
	"hgz021": {}, "hgz022": {},
	"oisd0": {}, "oisd1": {}, "oisd2": {},
}

const dnsblNotFoundSummary = "DNSBL compatibility item not found"

var _ datasource.DataSource = &dnsblDataSource{}

type dnsblDataSource struct{ client *opnsense.Client }

type dnsblLookupModel struct {
	ID         types.String `tfsdk:"id"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	SafeSearch types.Bool   `tfsdk:"safe_search"`
	Type       types.Set    `tfsdk:"type"`
	Lists      types.Set    `tfsdk:"lists"`
	Whitelists types.Set    `tfsdk:"whitelists"`
	Blocklists types.Set    `tfsdk:"blocklists"`
	Wildcards  types.Set    `tfsdk:"wildcards"`
	Address    types.String `tfsdk:"address"`
	NXDomain   types.Bool   `tfsdk:"nxdomain"`
}

func newDNSBLDataSource() datasource.DataSource { return &dnsblDataSource{} }

func (d *dnsblDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_unbound_dnsbl"
}

func (d *dnsblDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads the Unbound DNSBL compatibility owner. The only accepted identity is `unbound-dnsbl`. This data source never writes, deletes, or reconfigures. Released core 26.7.8 and 26.7.9 store blocklists as `dnsbl.blocklist` items. The lookup uses the existing search and `get_dnsbl` readers for the sole item, plus the settings GET for `general.safesearch`. It does not call `get_dnsbl` without a UUID, because that action adds an in-memory node instead of reading stored state. Zero items are not found. More than one item fails closed. The public allowlist is the resource refresh contract: enabled, safe_search, type, lists, whitelists, blocklists, wildcards, address, and nxdomain. `source_nets`, `cache_ttl`, and `description` are returned by the item GET and are not lookup state. The node has no password or write-only secret. A missing or blank enabled flag fails closed and is not replaced with the model default. Optional booleans stay null when blank. An empty address stays empty and is not replaced with 0.0.0.0.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Synthetic singleton ID. Must be `unbound-dnsbl`.",
			},
			"enabled": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Enable DNS-based blocklisting. A disabled flag stays false. A missing or blank value fails closed and is not replaced with the model default.",
			},
			"safe_search": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Force safe search. Read from `general.safesearch`, not from the blocklist item. A blank or omitted optional flag stays null.",
			},
			"type": dsschema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Selected built-in blocklist feed keys. Unselected option keys are not state.",
			},
			"lists": dsschema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Selected custom blocklist URL entries.",
			},
			"whitelists": dsschema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Selected allow-domain entries. The wire field is `allowlists`.",
			},
			"blocklists": dsschema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Selected deny-domain entries.",
			},
			"wildcards": dsschema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Selected wildcard domain entries.",
			},
			"address": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Destination address for blocked responses. An empty value stays empty and is not replaced with 0.0.0.0.",
			},
			"nxdomain": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Return NXDOMAIN for blocked entries. A blank optional flag stays null and is not replaced with false.",
			},
		},
	}
}

func (d *dnsblDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dnsblDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dnsblLookupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.ID.IsNull() || config.ID.IsUnknown() || config.ID.ValueString() != dnsblID {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid DNSBL ID",
			fmt.Sprintf("Expected %q for this singleton data source, got %q.", dnsblID, config.ID.ValueString()),
		)
		return
	}
	rows, err := opnsense.Search[dnsblSearchRow](ctx, d.client, dnsblReqOpts, opnsense.SearchParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error searching DNSBL blocklist items",
			fmt.Sprintf("The stable blocklist search failed: %s. The legacy singleton reader is not this lookup.", err.Error()),
		)
		return
	}
	switch len(rows) {
	case 0:
		resp.Diagnostics.AddError(
			dnsblNotFoundSummary,
			"The stable blocklist search returned no items. An empty result is not found and is not replaced with a new item or model defaults.",
		)
		return
	case 1:
	default:
		resp.Diagnostics.AddError(
			"Ambiguous DNSBL blocklist ownership",
			fmt.Sprintf("The stable blocklist search returned %d items. This lookup has no prior ownership and does not choose one.", len(rows)),
		)
		return
	}
	if rows[0].UUID == "" {
		resp.Diagnostics.AddError("Malformed DNSBL search response", "The sole blocklist item has no UUID.")
		return
	}
	item, err := opnsense.Get[dnsblItemAPIResponse](ctx, d.client, dnsblReqOpts, rows[0].UUID)
	if err != nil {
		var notFound *opnsense.NotFoundError
		if errors.As(err, &notFound) {
			resp.Diagnostics.AddError(
				"DNSBL blocklist item not found",
				fmt.Sprintf("The stable item read returned %s. A missing item is not replaced with model defaults.", notFound.Error()),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading DNSBL blocklist item", err.Error())
		return
	}
	safeSearch, err := readDNSBLLookupSafeSearch(ctx, d.client)
	if err != nil {
		var notFound *opnsense.NotFoundError
		if errors.As(err, &notFound) {
			resp.Diagnostics.AddError(
				"Error reading Unbound SafeSearch",
				fmt.Sprintf("The settings reader returned %s. That is not an empty DNSBL search.", notFound.Error()),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading Unbound SafeSearch", err.Error())
		return
	}
	state, err := dnsblLookupFromItem(item, safeSearch)
	if err != nil {
		resp.Diagnostics.AddError("Malformed DNSBL blocklist item", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func readDNSBLLookupSafeSearch(ctx context.Context, client *opnsense.Client) (types.Bool, error) {
	settings, err := opnsense.GetSingleton[unboundSettingsEnvelope](ctx, client, dnsblSettingsReqOpts)
	if err != nil {
		return types.BoolNull(), err
	}
	if len(bytesTrim(settings.General)) == 0 || string(bytesTrim(settings.General)) == "null" {
		return types.BoolNull(), nil
	}
	var general map[string]json.RawMessage
	if err := json.Unmarshal(settings.General, &general); err != nil {
		return types.BoolNull(), fmt.Errorf("decode general branch: %w", err)
	}
	raw, ok := general["safesearch"]
	if !ok || len(bytesTrim(raw)) == 0 || string(bytesTrim(raw)) == "null" {
		return types.BoolNull(), nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return types.BoolNull(), errors.New("general.safesearch is not a string token")
	}
	parsed, ok := dnsblOptionalBool(value)
	if !ok {
		return types.BoolNull(), fmt.Errorf("general.safesearch token %q is not 0 or 1", strings.TrimSpace(value))
	}
	return parsed, nil
}

func dnsblLookupFromItem(item *dnsblItemAPIResponse, safeSearch types.Bool) (dnsblLookupModel, error) {
	enabled, ok := dnsblRequiredBool(item.Enabled)
	if !ok {
		return dnsblLookupModel{}, errors.New("enabled is missing, blank, or not 0/1; the model default is not copied")
	}
	nxdomain, ok := dnsblOptionalBool(item.NXDomain)
	if !ok {
		return dnsblLookupModel{}, fmt.Errorf("nxdomain token %q is not blank, 0, or 1", strings.TrimSpace(item.NXDomain))
	}
	if !dnsblTypeAllowed(item.Type) {
		return dnsblLookupModel{}, errors.New("type contains a selected key outside the released option list")
	}
	address, ok := dnsblAddress(item.Address)
	if !ok {
		return dnsblLookupModel{}, errors.New("address is not an empty value or a dotted IPv4 address")
	}
	return dnsblLookupModel{
		ID:         types.StringValue(dnsblID),
		Enabled:    types.BoolValue(enabled),
		SafeSearch: safeSearch,
		Type:       tfconv.SliceToSet(item.Type),
		Lists:      tfconv.SliceToSet(item.Lists),
		Whitelists: tfconv.SliceToSet(item.Allowlists),
		Blocklists: tfconv.SliceToSet(item.Blocklists),
		Wildcards:  tfconv.SliceToSet(item.Wildcards),
		Address:    types.StringValue(address),
		NXDomain:   nxdomain,
	}, nil
}

func dnsblRequiredBool(raw string) (bool, bool) {
	switch strings.TrimSpace(raw) {
	case "0":
		return false, true
	case "1":
		return true, true
	default:
		return false, false
	}
}

func dnsblOptionalBool(raw string) (types.Bool, bool) {
	switch strings.TrimSpace(raw) {
	case "":
		return types.BoolNull(), true
	case "0":
		return types.BoolValue(false), true
	case "1":
		return types.BoolValue(true), true
	default:
		return types.BoolNull(), false
	}
}

func dnsblTypeAllowed(keys []string) bool {
	for _, key := range keys {
		if _, ok := dnsblTypeOptions[key]; !ok {
			return false
		}
	}
	return true
}

func dnsblAddress(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", true
	}
	if !dnsblDottedQuad(trimmed) {
		return "", false
	}
	return trimmed, true
}

func dnsblDottedQuad(value string) bool {
	if strings.ContainsAny(value, "/:") {
		return false
	}
	parts := strings.Split(value, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if part == "" || (len(part) > 1 && part[0] == '0') {
			return false
		}
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 || number > 255 {
			return false
		}
	}
	return true
}

func bytesTrim(raw []byte) []byte {
	return []byte(strings.TrimSpace(string(raw)))
}
