// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

const (
	mapfileTypeBeg       = "beg"
	mapfileTypeDomain    = "dom"
	mapfileTypeEnd       = "end"
	mapfileTypeInt       = "int"
	mapfileTypeIP        = "ip"
	mapfileTypeReg       = "reg"
	mapfileTypeMapString = "str"
	mapfileTypeSub       = "sub"
)

// MapfileResourceModel is the Terraform state model for opnsense_haproxy_mapfile.
type MapfileResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Content     types.String `tfsdk:"content"`
}

type mapfileAPIResponse struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Type        opnsense.SelectedMap `json:"type"`
	Content     string               `json:"content"`
}

type mapfileAPIRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Content     string `json:"content"`
}

func (m *MapfileResourceModel) toAPI(_ context.Context) *mapfileAPIRequest {
	return &mapfileAPIRequest{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		Type:        m.Type.ValueString(),
		Content:     normalizeMapfileContent(m.Content.ValueString()),
	}
}

func (m *MapfileResourceModel) fromAPI(_ context.Context, a *mapfileAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Name = types.StringValue(a.Name)
	m.Description = types.StringValue(a.Description)
	if string(a.Type) != "" {
		m.Type = types.StringValue(string(a.Type))
	} else if m.Type.IsNull() || m.Type.IsUnknown() {
		m.Type = types.StringValue(mapfileTypeDomain)
	}
	m.Content = types.StringValue(normalizeMapfileContent(a.Content))
}

func normalizeMapfileContent(content string) string {
	return strings.TrimRight(content, " \t\r\n")
}
