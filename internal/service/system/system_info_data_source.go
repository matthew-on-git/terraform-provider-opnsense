// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &systemInfoDataSource{}

// systemInfoDataSource implements the opnsense_system_info data source, exposing
// the firmware version and the list of installed plugins.
type systemInfoDataSource struct{ client *opnsense.Client }

func newSystemInfoDataSource() datasource.DataSource { return &systemInfoDataSource{} }

// systemInfoModel is the state model for opnsense_system_info.
type systemInfoModel struct {
	ID      types.String `tfsdk:"id"`
	Version types.String `tfsdk:"version"`
	Plugins types.Set    `tfsdk:"plugins"`
}

// firmwareInfo is the subset of /api/core/firmware/info we consume.
type firmwareInfo struct {
	ProductVersion string `json:"product_version"`
	Plugin         []struct {
		Name      string `json:"name"`
		Installed string `json:"installed"`
	} `json:"plugin"`
}

func (d *systemInfoDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_info"
}

func (d *systemInfoDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Provides information about the OPNsense appliance: its firmware version and the list of installed plugins.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Always `system_info` (this is a singleton data source).",
			},
			"version": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Installed OPNsense firmware version (e.g. `25.7`).",
			},
			"plugins": dsschema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Names of installed plugins (e.g. `os-haproxy`).",
			},
		},
	}
}

func (d *systemInfoDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *systemInfoDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if err := d.client.AcquireRead(ctx); err != nil {
		resp.Diagnostics.AddError("Error reading system info", err.Error())
		return
	}
	defer d.client.ReleaseRead()

	url := d.client.BaseURL() + "/api/core/firmware/info"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading system info", err.Error())
		return
	}
	httpResp, err := d.client.HTTPClient().Do(httpReq) //nolint:gosec // URL from provider-configured client
	if err != nil {
		resp.Diagnostics.AddError("Error reading system info", err.Error())
		return
	}
	defer func() { _ = httpResp.Body.Close() }()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Error reading system info", err.Error())
		return
	}

	var info firmwareInfo
	if err := json.Unmarshal(body, &info); err != nil {
		resp.Diagnostics.AddError("Error parsing system info", fmt.Sprintf("could not parse firmware info: %s", err))
		return
	}

	plugins := make([]attr.Value, 0, len(info.Plugin))
	for _, p := range info.Plugin {
		if p.Installed == "1" {
			plugins = append(plugins, types.StringValue(p.Name))
		}
	}

	state := systemInfoModel{
		ID:      types.StringValue("system_info"),
		Version: types.StringValue(info.ProductVersion),
		Plugins: types.SetValueMust(types.StringType, plugins),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
