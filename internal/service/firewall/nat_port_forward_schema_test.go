// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
)

func TestNatPortForwardSourceNetDefaultMatchesOPNsenseUnset(t *testing.T) {
	var resp resource.SchemaResponse
	(&natPortForwardResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)

	attr, ok := resp.Schema.Attributes["source_net"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("source_net attribute type = %T, want schema.StringAttribute", resp.Schema.Attributes["source_net"])
	}
	if attr.Default == nil {
		t.Fatal("source_net default is nil, want empty string default")
	}

	defaultResp := &defaults.StringResponse{}
	attr.Default.DefaultString(context.Background(), defaults.StringRequest{}, defaultResp)
	if defaultResp.Diagnostics.HasError() {
		t.Fatalf("source_net default diagnostics: %s", defaultResp.Diagnostics.Errors())
	}
	if got := defaultResp.PlanValue.ValueString(); got != "" {
		t.Fatalf("source_net default = %q, want empty string", got)
	}
}
