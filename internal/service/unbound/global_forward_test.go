// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestGlobalForwardRoundTrip(t *testing.T) {
	for _, kind := range []string{"forward", "dot"} {
		for _, domain := range []string{"", "example.com"} {
			t.Run(kind+"/"+domain, func(t *testing.T) {
				api := domainOverrideAPIResponse{Enabled: "1", Domain: domain, Type: opnsense.SelectedMap(kind), Server: "192.0.2.1", Port: "853", Verify: "dns.example.com"}
				var model DomainOverrideResourceModel
				if err := model.fromAPI(context.Background(), &api, "fixture"); err != nil {
					t.Fatal(err)
				}
				got := model.toAPI(context.Background())
				if got.Domain != domain || got.Type != kind || got.Verify != api.Verify {
					t.Fatalf("forward changed: %#v", got)
				}
			})
		}
	}
}

func TestGlobalForwardRejectsMissingDomain(t *testing.T) {
	for _, data := range []string{`{"type":"dot","port":"853"}`, `{"domain":null,"type":"dot","port":"853"}`} {
		var api domainOverrideAPIResponse
		if err := json.Unmarshal([]byte(data), &api); err == nil {
			t.Fatal("accepted absent/null domain")
		}
	}
}
