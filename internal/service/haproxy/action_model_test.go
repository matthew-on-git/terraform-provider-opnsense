// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestActionModel_toAPIMapsHTTPRequestTypes(t *testing.T) {
	t.Parallel()

	model := ActionResourceModel{
		Name:                 types.StringValue("redirect-to-https"),
		Description:          types.StringValue(""),
		TestType:             types.StringValue("if"),
		LinkedACLs:           types.SetValueMust(types.StringType, []attr.Value{types.StringValue("acl-1")}),
		Operator:             types.StringValue("and"),
		Type:                 types.StringValue(actionTypeHTTPRequestRedirect),
		Redirect:             types.StringValue("scheme https code 301"),
		UseBackend:           types.StringValue(""),
		MapUseBackendFile:    types.StringValue(""),
		MapUseBackendDefault: types.StringValue(""),
		HTTPRequestOption:    types.StringValue(""),
		DenyStatus:           types.Int64Null(),
		SetHeaderName:        types.StringValue(""),
		SetHeaderContent:     types.StringValue(""),
	}

	req := model.toAPI(context.Background())
	if req.Type != "http-request" || req.HTTPRequestAction != "redirect" {
		t.Fatalf("unexpected API action mapping: type=%q action=%q", req.Type, req.HTTPRequestAction)
	}
	if req.LinkedACLs != "acl-1" || req.HTTPRequestOption != "scheme https code 301" {
		t.Fatalf("unexpected API request: %#v", req)
	}
}

func TestActionModel_toAPIMapsHTTPRequestSetHeader(t *testing.T) {
	t.Parallel()

	model := ActionResourceModel{
		Name:                 types.StringValue("forwarded-proto"),
		Description:          types.StringValue(""),
		TestType:             types.StringValue("if"),
		LinkedACLs:           types.SetValueMust(types.StringType, []attr.Value{}),
		Operator:             types.StringValue("and"),
		Type:                 types.StringValue(actionTypeHTTPRequestSetHeader),
		UseBackend:           types.StringValue(""),
		MapUseBackendFile:    types.StringValue(""),
		MapUseBackendDefault: types.StringValue(""),
		HTTPRequestOption:    types.StringValue(""),
		DenyStatus:           types.Int64Null(),
		Redirect:             types.StringValue(""),
		SetHeaderName:        types.StringValue("X-Forwarded-Proto"),
		SetHeaderContent:     types.StringValue("https"),
	}

	req := model.toAPI(context.Background())
	if req.Type != "http-request" || req.HTTPRequestAction != "set-header" {
		t.Fatalf("unexpected API action mapping: type=%q action=%q", req.Type, req.HTTPRequestAction)
	}
	if req.HTTPRequestOption != "X-Forwarded-Proto https" {
		t.Fatalf("unexpected API request: %#v", req)
	}
}

func TestActionRequireString_allowsUnknownReferences(t *testing.T) {
	t.Parallel()

	var resp resource.ValidateConfigResponse
	requireString(&resp, types.StringUnknown(), "use_backend", "use_backend is required when type is use_backend.")
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics for unknown reference: %v", resp.Diagnostics)
	}
}

func TestActionRequireString_rejectsLiteralEmptyValue(t *testing.T) {
	t.Parallel()

	var resp resource.ValidateConfigResponse
	requireString(&resp, types.StringValue(""), "use_backend", "use_backend is required when type is use_backend.")
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic for empty required action field")
	}
}

func TestActionRequireString_rejectsNullValue(t *testing.T) {
	t.Parallel()

	var resp resource.ValidateConfigResponse
	requireString(&resp, types.StringNull(), "use_backend", "use_backend is required when type is use_backend.")
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic for null required action field")
	}
}

func TestActionModel_fromAPIMapsHTTPRequestTypes(t *testing.T) {
	t.Parallel()

	var model ActionResourceModel
	model.fromAPI(context.Background(), &actionAPIResponse{
		Name:              "deny-external",
		TestType:          opnsense.SelectedMap("unless"),
		Operator:          opnsense.SelectedMap("or"),
		Type:              opnsense.SelectedMap("http-request"),
		LinkedACLs:        opnsense.SelectedMapList{"acl-1", "acl-2"},
		HTTPRequestAction: opnsense.SelectedMap("deny"),
		HTTPRequestOption: "deny_status 403",
	}, "action-1")

	if model.Type.ValueString() != actionTypeHTTPRequestDeny {
		t.Fatalf("unexpected Terraform action type: %q", model.Type.ValueString())
	}
	if model.DenyStatus.ValueInt64() != 403 || model.LinkedACLs.Elements()[0].(types.String).ValueString() != "acl-1" {
		t.Fatalf("unexpected model: %#v", model)
	}
}

func TestActionModel_fromAPIMapsHTTPRequestSetHeader(t *testing.T) {
	t.Parallel()

	var model ActionResourceModel
	model.fromAPI(context.Background(), &actionAPIResponse{
		Name:              "forwarded-proto",
		TestType:          opnsense.SelectedMap("if"),
		Operator:          opnsense.SelectedMap("and"),
		Type:              opnsense.SelectedMap("http-request"),
		HTTPRequestAction: opnsense.SelectedMap("set-header"),
		HTTPRequestOption: "X-Forwarded-Proto https",
	}, "action-2")

	if model.Type.ValueString() != actionTypeHTTPRequestSetHeader {
		t.Fatalf("unexpected Terraform action type: %q", model.Type.ValueString())
	}
	if model.SetHeaderName.ValueString() != "X-Forwarded-Proto" || model.SetHeaderContent.ValueString() != "https" {
		t.Fatalf("unexpected model: %#v", model)
	}
	if model.HTTPRequestOption.ValueString() != "" {
		t.Fatalf("unexpected internal API option in Terraform state: %q", model.HTTPRequestOption.ValueString())
	}
}
