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

func TestFrontendModel_toAPIMapsCertificateRefIDs(t *testing.T) {
	t.Parallel()

	model := FrontendResourceModel{
		Enabled:            types.BoolValue(true),
		Name:               types.StringValue("https-in"),
		Description:        types.StringValue(""),
		Bind:               types.StringValue("0.0.0.0:443"),
		Mode:               types.StringValue("http"),
		DefaultBackend:     types.StringValue("backend-1"),
		SSLEnabled:         types.BoolValue(true),
		Certificates:       types.SetValueMust(types.StringType, []attr.Value{types.StringValue("cert-ref-2"), types.StringValue("cert-ref-1")}),
		DefaultCertificate: types.StringValue("cert-ref-1"),
		LinkedActions:      types.SetValueMust(types.StringType, []attr.Value{}),
		ForwardFor:         types.BoolValue(true),
	}

	req := model.toAPI(context.Background())
	if req.SSLCertificates != "cert-ref-1,cert-ref-2" {
		t.Fatalf("ssl_certificates = %q, want comma-joined refids", req.SSLCertificates)
	}
	if req.SSLDefaultCertificate != "cert-ref-1" {
		t.Fatalf("ssl_default_certificate = %q, want cert-ref-1", req.SSLDefaultCertificate)
	}
}

func TestFrontendModel_toAPIIgnoresCertificatesWhenSSLDisabled(t *testing.T) {
	t.Parallel()

	model := FrontendResourceModel{
		Enabled:            types.BoolValue(true),
		Name:               types.StringValue("http-in"),
		Description:        types.StringValue(""),
		Bind:               types.StringValue("0.0.0.0:80"),
		Mode:               types.StringValue("http"),
		DefaultBackend:     types.StringValue("backend-1"),
		SSLEnabled:         types.BoolValue(false),
		Certificates:       types.SetValueMust(types.StringType, []attr.Value{types.StringValue("cert-ref-1")}),
		DefaultCertificate: types.StringValue("cert-ref-1"),
		LinkedActions:      types.SetValueMust(types.StringType, []attr.Value{}),
		ForwardFor:         types.BoolValue(false),
	}

	req := model.toAPI(context.Background())
	if req.SSLCertificates != "" || req.SSLDefaultCertificate != "" {
		t.Fatalf("certificate fields were not ignored when ssl disabled: %#v", req)
	}
}

func TestFrontendModel_fromAPIMapsCertificateRefIDs(t *testing.T) {
	t.Parallel()

	var model FrontendResourceModel
	model.fromAPI(context.Background(), &frontendAPIResponse{
		Enabled:               "1",
		Name:                  "https-in",
		Bind:                  opnsense.SelectedMapList{"0.0.0.0:443"},
		Mode:                  opnsense.SelectedMap("http"),
		DefaultBackend:        opnsense.SelectedMap("backend-1"),
		SSLEnabled:            "1",
		SSLCertificates:       opnsense.SelectedMapList{"cert-ref-1", "cert-ref-2"},
		SSLDefaultCertificate: opnsense.SelectedMap("cert-ref-1"),
		LinkedActions:         opnsense.SelectedMapList{},
		ForwardFor:            "1",
	}, "frontend-1")

	if model.DefaultCertificate.ValueString() != "cert-ref-1" {
		t.Fatalf("default_certificate = %q, want cert-ref-1", model.DefaultCertificate.ValueString())
	}
	if len(model.Certificates.Elements()) != 2 {
		t.Fatalf("expected 2 certificate refids, got %#v", model.Certificates.Elements())
	}
}
