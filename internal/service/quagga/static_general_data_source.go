// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// staticLookupFields is the public GET allowlist. It is the resource refresh
// contract intersected with fields the stable staticd node returns. Released
// os-frr 26.7.8 and 26.7.9 also return the routes child array from the
// whole-model GET. The resource never refreshes that array, so it is not
// lookup state. The node has no password or write-only secret. A missing or
// blank required enabled flag is an error, not the model default of false.
// A disabled flag stays false.
var staticLookupFields = []string{
	"id",
	"enabled",
}

var _ datasource.DataSource = &staticGeneralDataSource{}

type staticGeneralDataSource struct{ client *opnsense.Client }

type staticLookupModel struct {
	ID      types.String `tfsdk:"id"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

type staticLookupResponse struct {
	frrReadPresence `json:"-"`
	Enabled         string `json:"enabled"`
}

func (a *staticLookupResponse) UnmarshalJSON(raw []byte) error {
	type plain staticLookupResponse
	if err := json.Unmarshal(raw, (*plain)(a)); err != nil {
		return err
	}
	return a.recordFields(raw, a)
}

func newStaticGeneralDataSource() datasource.DataSource { return &staticGeneralDataSource{} }

func (d *staticGeneralDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_static"
}

func (d *staticGeneralDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads FRR static routing settings from the stable Quagga GET. Never writes or reconfigures. The only accepted identity is `static`. The public allowlist is the resource refresh contract that the released staticd node returns: enabled. The routes child array is returned by the whole-model GET but is not resource refresh state, so it is not lookup state. The node has no password or write-only secret. A missing or blank enabled flag fails closed and is not replaced with the model default. A disabled flag stays false.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Synthetic singleton ID. Must be `static`.",
			},
			"enabled": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Enable FRR static routing. A disabled service stays false and is not replaced with the model default. A missing or blank value fails closed.",
			},
		},
	}
}

func (d *staticGeneralDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureQuaggaDataSource(req, resp, func(client *opnsense.Client) { d.client = client })
}

func (d *staticGeneralDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config staticLookupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.ID.IsNull() || config.ID.IsUnknown() || config.ID.ValueString() != staticGeneralID {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid static routing settings ID",
			fmt.Sprintf("Expected %q for this singleton data source, got %q.", staticGeneralID, config.ID.ValueString()),
		)
		return
	}
	result, err := opnsense.GetSingleton[staticLookupResponse](ctx, d.client, staticGeneralReqOpts)
	if err != nil {
		var notFound *opnsense.NotFoundError
		if errors.As(err, &notFound) {
			resp.Diagnostics.AddError(
				"Static routing settings not found",
				"The appliance returned an empty static routing settings object. An empty required object is not found and is not replaced with model defaults.",
			)
			return
		}
		resp.Diagnostics.AddError("Error reading static routing settings", frrDiagnostic(err))
		return
	}
	if field := staticLookupGap(result); field != "" {
		resp.Diagnostics.AddError(
			"Malformed static routing settings response",
			fmt.Sprintf("The appliance returned static routing settings without a readable %s. Missing required fields are not replaced with model defaults.", field),
		)
		return
	}
	applyStaticLookup(&config, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func staticLookupGap(result *staticLookupResponse) string {
	if result == nil {
		return "response"
	}
	if !result.hasField("enabled") || !ospfBoolToken(result.Enabled) {
		return "enabled"
	}
	return ""
}

func applyStaticLookup(model *staticLookupModel, result *staticLookupResponse) {
	model.ID = types.StringValue(staticGeneralID)
	model.Enabled = types.BoolValue(strings.TrimSpace(result.Enabled) == "1")
}
