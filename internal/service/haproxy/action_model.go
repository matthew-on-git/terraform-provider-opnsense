// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

const (
	actionTypeUseBackend           = "use_backend"
	actionTypeMapUseBackend        = "map_use_backend"
	actionTypeHTTPRequestDeny      = "http-request_deny"
	actionTypeHTTPRequestRedirect  = "http-request_redirect"
	actionTypeHTTPRequestSetHeader = "http-request_set-header"
)

// ActionResourceModel is the Terraform state model for opnsense_haproxy_action.
type ActionResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	TestType             types.String `tfsdk:"test_type"`
	LinkedACLs           types.Set    `tfsdk:"linked_acls"`
	Operator             types.String `tfsdk:"operator"`
	Type                 types.String `tfsdk:"type"`
	UseBackend           types.String `tfsdk:"use_backend"`
	MapUseBackendFile    types.String `tfsdk:"mapfile"`
	MapUseBackendDefault types.String `tfsdk:"map_use_backend_default"`
	HTTPRequestOption    types.String `tfsdk:"http_request_option"`
	DenyStatus           types.Int64  `tfsdk:"deny_status"`
	Redirect             types.String `tfsdk:"redirect"`
	SetHeaderName        types.String `tfsdk:"set_header_name"`
	SetHeaderContent     types.String `tfsdk:"set_header_content"`
}

type actionAPIResponse struct {
	Name                 string                   `json:"name"`
	Description          string                   `json:"description"`
	TestType             opnsense.SelectedMap     `json:"testType"`
	LinkedACLs           opnsense.SelectedMapList `json:"linkedAcls"`
	Operator             opnsense.SelectedMap     `json:"operator"`
	Type                 opnsense.SelectedMap     `json:"type"`
	UseBackend           opnsense.SelectedMap     `json:"use_backend"`
	MapUseBackendFile    opnsense.SelectedMap     `json:"map_use_backend_file"`
	MapUseBackendDefault opnsense.SelectedMap     `json:"map_use_backend_default"`
	HTTPRequestOption    string                   `json:"http_request_option"`
	DenyStatus           string                   `json:"http_request_deny_status"`
	Redirect             string                   `json:"http_request_redirect"`
	SetHeaderName        string                   `json:"http_request_set_header_name"`
	SetHeaderContent     string                   `json:"http_request_set_header_content"`
}

type actionAPIRequest struct {
	Name                 string `json:"name"`
	Description          string `json:"description"`
	TestType             string `json:"testType"`
	LinkedACLs           string `json:"linkedAcls"`
	Operator             string `json:"operator"`
	Type                 string `json:"type"`
	UseBackend           string `json:"use_backend"`
	MapUseBackendFile    string `json:"map_use_backend_file"`
	MapUseBackendDefault string `json:"map_use_backend_default"`
	HTTPRequestOption    string `json:"http_request_option"`
	DenyStatus           string `json:"http_request_deny_status"`
	Redirect             string `json:"http_request_redirect"`
	SetHeaderName        string `json:"http_request_set_header_name"`
	SetHeaderContent     string `json:"http_request_set_header_content"`
}

func (m *ActionResourceModel) toAPI(ctx context.Context) *actionAPIRequest {
	var aclIDs []string
	if !m.LinkedACLs.IsNull() && !m.LinkedACLs.IsUnknown() {
		m.LinkedACLs.ElementsAs(ctx, &aclIDs, false)
	}

	req := &actionAPIRequest{
		Name:                 m.Name.ValueString(),
		Description:          m.Description.ValueString(),
		TestType:             m.TestType.ValueString(),
		LinkedACLs:           strings.Join(aclIDs, ","),
		Operator:             m.Operator.ValueString(),
		UseBackend:           m.UseBackend.ValueString(),
		MapUseBackendFile:    m.MapUseBackendFile.ValueString(),
		MapUseBackendDefault: m.MapUseBackendDefault.ValueString(),
		HTTPRequestOption:    m.HTTPRequestOption.ValueString(),
		Redirect:             m.Redirect.ValueString(),
		SetHeaderName:        m.SetHeaderName.ValueString(),
		SetHeaderContent:     m.SetHeaderContent.ValueString(),
	}
	if !m.DenyStatus.IsNull() && !m.DenyStatus.IsUnknown() {
		req.DenyStatus = opnsense.Int64ToString(m.DenyStatus.ValueInt64())
	}

	req.Type = m.Type.ValueString()

	return req
}

func (m *ActionResourceModel) fromAPI(_ context.Context, a *actionAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Name = types.StringValue(a.Name)
	m.Description = types.StringValue(a.Description)
	m.TestType = types.StringValue(string(a.TestType))
	m.Operator = types.StringValue(string(a.Operator))
	m.UseBackend = types.StringValue(string(a.UseBackend))
	m.MapUseBackendFile = types.StringValue(string(a.MapUseBackendFile))
	m.MapUseBackendDefault = types.StringValue(string(a.MapUseBackendDefault))
	m.HTTPRequestOption = types.StringValue(a.HTTPRequestOption)
	m.Redirect = types.StringValue(a.Redirect)
	m.SetHeaderName = types.StringValue(a.SetHeaderName)
	m.SetHeaderContent = types.StringValue(a.SetHeaderContent)

	if a.DenyStatus == "" {
		m.DenyStatus = types.Int64Null()
	} else if status, err := opnsense.StringToInt64(a.DenyStatus); err == nil {
		m.DenyStatus = types.Int64Value(status)
	} else {
		m.DenyStatus = types.Int64Null()
	}

	m.Type = types.StringValue(string(a.Type))
	m.LinkedACLs = selectedListToStringSet(a.LinkedACLs)
}

func selectedListToStringSet(values []string) types.Set {
	if len(values) == 0 {
		return types.SetValueMust(types.StringType, []attr.Value{})
	}
	attrs := make([]attr.Value, len(values))
	for i, value := range values {
		attrs[i] = types.StringValue(value)
	}
	return types.SetValueMust(types.StringType, attrs)
}
