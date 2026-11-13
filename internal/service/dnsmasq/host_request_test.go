// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package dnsmasq

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestHostClientIDWriteIntent(t *testing.T) {
	for _, tc := range []struct {
		name    string
		value   types.String
		present bool
		want    string
	}{
		{"clear", types.StringValue(""), true, ""},
		{"set", types.StringValue("test-client"), true, "test-client"},
		{"null", types.StringNull(), false, ""},
		{"unknown", types.StringUnknown(), false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			model := HostResourceModel{Host: types.StringValue("fixture"), IP: types.StringValue("192.0.2.10"), ClientID: tc.value}
			payload, err := json.Marshal(model.toAPI(context.Background()))
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]string
			if err := json.Unmarshal(payload, &fields); err != nil {
				t.Fatal(err)
			}
			got, present := fields["client_id"]
			if present != tc.present || got != tc.want {
				t.Fatalf("client_id = %q, present=%v; want %q, present=%v", got, present, tc.want, tc.present)
			}
			if fields["host"] != "fixture" || fields["ip"] != "192.0.2.10" {
				t.Fatalf("changed unrelated fields: %s", payload)
			}
		})
	}
}
