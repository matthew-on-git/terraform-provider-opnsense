// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestMapfileModel_toAPI(t *testing.T) {
	t.Parallel()

	model := MapfileResourceModel{
		Name:        types.StringValue("domain-map"),
		Description: types.StringValue("Domain routing map"),
		Type:        types.StringValue(mapfileTypeMapString),
		Content:     types.StringValue("grafana.example.com grafana-backend\nargocd.example.com argocd-backend"),
	}

	req := model.toAPI(context.Background())
	if req.Name != "domain-map" || req.Type != mapfileTypeMapString {
		t.Fatalf("unexpected API request: %#v", req)
	}
	if req.Content != "grafana.example.com grafana-backend\nargocd.example.com argocd-backend" {
		t.Fatalf("content was not preserved: %q", req.Content)
	}
}

func TestMapfileModel_fromAPINormalizesTrailingWhitespace(t *testing.T) {
	t.Parallel()

	var model MapfileResourceModel
	model.fromAPI(context.Background(), &mapfileAPIResponse{
		Name:        "domain-map",
		Description: "Domain routing map",
		Type:        opnsense.SelectedMap(mapfileTypeMapString),
		Content:     "grafana.example.com grafana-backend\nargocd.example.com argocd-backend\n\n\t ",
	}, "mapfile-1")

	if model.ID.ValueString() != "mapfile-1" || model.Type.ValueString() != mapfileTypeMapString {
		t.Fatalf("unexpected model metadata: %#v", model)
	}
	const expected = "grafana.example.com grafana-backend\nargocd.example.com argocd-backend"
	if model.Content.ValueString() != expected {
		t.Fatalf("content = %q, want %q", model.Content.ValueString(), expected)
	}
}
