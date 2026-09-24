// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestBackendModel_toAPIMapsHealthCheck(t *testing.T) {
	t.Parallel()

	model := BackendResourceModel{
		Enabled:            types.BoolValue(true),
		Name:               types.StringValue("application"),
		Description:        types.StringValue(""),
		Mode:               types.StringValue("http"),
		Algorithm:          types.StringValue("roundrobin"),
		LinkedServers:      types.SetValueMust(types.StringType, []attr.Value{types.StringValue("server-1")}),
		HealthCheck:        types.StringValue("health-check-1"),
		HealthCheckEnabled: types.BoolValue(true),
		Persistence:        types.StringValue("sticktable"),
		ForwardFor:         types.BoolValue(true),
	}

	req := model.toAPI(context.Background())
	if req.HealthCheck != "health-check-1" || req.HealthCheckEnabled != "1" {
		t.Fatalf("unexpected health-check API mapping: %#v", req)
	}
}

func TestBackendModel_fromAPIMapsHealthCheck(t *testing.T) {
	t.Parallel()

	var model BackendResourceModel
	model.fromAPI(context.Background(), &backendAPIResponse{
		Enabled:            "1",
		Name:               "application",
		Mode:               opnsense.SelectedMap("http"),
		Algorithm:          opnsense.SelectedMap("roundrobin"),
		LinkedServers:      opnsense.SelectedMapList{"server-1"},
		HealthCheck:        opnsense.SelectedMap("health-check-1"),
		HealthCheckEnabled: "1",
		Persistence:        opnsense.SelectedMap("sticktable"),
		ForwardFor:         "1",
	}, "backend-1")

	if model.HealthCheck.ValueString() != "health-check-1" || !model.HealthCheckEnabled.ValueBool() {
		t.Fatalf("unexpected health-check Terraform mapping: %#v", model)
	}
}

func TestBackendModelTimeoutsToAPI(t *testing.T) {
	t.Parallel()

	model := BackendResourceModel{
		TimeoutConnect: types.StringValue("5s"),
		TimeoutCheck:   types.StringValue("5s"),
		TimeoutServer:  types.StringValue("10m"),
		CustomOptions:  types.StringValue("timeout tunnel 1h"),
	}

	request := model.toAPI(context.Background())
	if request.TimeoutConnect != "5s" || request.TimeoutCheck != "5s" || request.TimeoutServer != "10m" {
		t.Fatalf("unexpected timeout request values: %#v", request)
	}
	if request.CustomOptions != "timeout tunnel 1h" {
		t.Fatalf("unexpected custom options: %q", request.CustomOptions)
	}
}

func TestBackendModelTimeoutsFromAPI(t *testing.T) {
	t.Parallel()

	response := backendAPIResponse{
		TimeoutConnect: "5s",
		TimeoutCheck:   "5s",
		TimeoutServer:  "10m",
		CustomOptions:  "timeout tunnel 1h",
	}
	var model BackendResourceModel
	model.fromAPI(context.Background(), &response, "backend-uuid")

	if model.TimeoutConnect.ValueString() != "5s" || model.TimeoutCheck.ValueString() != "5s" || model.TimeoutServer.ValueString() != "10m" {
		t.Fatalf("unexpected timeout model values: %#v", model)
	}
	if model.CustomOptions.ValueString() != "timeout tunnel 1h" {
		t.Fatalf("unexpected custom options: %q", model.CustomOptions.ValueString())
	}
}
