// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestIPsecAcceptanceBothImportsAndLookups(t *testing.T) {
	binary, err := exec.LookPath("terraform")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TF_ACC_TERRAFORM_PATH", binary)
	t.Setenv("TF_VAR_ipsec_connection_import_id", ipsecID)
	t.Setenv("TF_VAR_ipsec_child_import_id", ipsecID)
	connection := &ipsecAPI{owner: "connection"}
	child := &ipsecAPI{owner: "child"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api := connection
		if strings.Contains(r.URL.Path, "_child") {
			api = child
		}
		// The existing fixture's legacy version array is fixed to IKEv2.
		// This acceptance scenario deliberately uses IKEv1 aggressive mode.
		if strings.Contains(r.URL.Path, "get_connection/") {
			recorded := httptest.NewRecorder()
			api.ServeHTTP(recorded, r)
			var wrapped map[string]map[string]any
			if json.Unmarshal(recorded.Body.Bytes(), &wrapped) == nil && wrapped["connection"] != nil {
				wrapped["connection"]["version"] = []map[string]int{{"selected": 0}, {"selected": 1}, {"selected": 0}}
				_ = json.NewEncoder(w).Encode(wrapped)
				return
			}
			_, _ = w.Write(recorded.Body.Bytes())
			return
		}
		api.ServeHTTP(w, r)
	}))
	defer server.Close()
	provider := fmt.Sprintf("provider \"opnsense\" {\nuri=%q\napi_key=\"fixture-key\"\napi_secret=\"fixture-secret\"\n}\n", server.URL)
	seed := strings.ReplaceAll(strings.ReplaceAll(ipsecAcceptanceConfig(true, false), "\"advanced\"", "\"seed\""), ".advanced.id", ".seed.id")
	check := func(name string, cleared bool) resource.TestCheckFunc {
		return func(state *terraform.State) error {
			for _, owner := range []string{"connection", "child"} {
				item := state.RootModule().Resources["opnsense_ipsec_"+owner+"."+name]
				if item == nil || item.Primary.ID != ipsecID || item.Primary.Attributes["enabled"] != "false" {
					return fmt.Errorf("%s import lost disabled owned identity", owner)
				}
			}
			if err := verifyIPsecAcceptanceLookupsAt(state, name, cleared); err != nil {
				return err
			}
			// Each of the 21 assertions must reject both a wrong value and a
			// missing output; exercise the real Core-produced lookup state.
			fields := map[string][]string{
				"connection": {"aggressive", "dpd_delay", "dpd_timeout", "ike_lifetime", "ip_pools", "keying_tries", "local_addresses", "local_port", "mobike", "reauthentication_time", "rekey_time", "remote_port", "send_certificate", "send_certificate_request", "udp_encapsulation"},
				"child":      {"close_action", "dpd_action", "install_policies", "request_id", "sha256_96", "rekey_time"},
			}
			for owner, attributes := range fields {
				lookup := state.RootModule().Resources["data.opnsense_ipsec_"+owner+"."+name].Primary.Attributes
				for _, attribute := range attributes {
					original := lookup[attribute]
					lookup[attribute] = "wrong-lookup-output"
					wrongErr := verifyIPsecAcceptanceLookupsAt(state, name, cleared)
					delete(lookup, attribute)
					missingErr := verifyIPsecAcceptanceLookupsAt(state, name, cleared)
					lookup[attribute] = original
					if wrongErr == nil || missingErr == nil {
						return fmt.Errorf("%s.%s lookup assertion failed its corruption control", owner, attribute)
					}
				}
			}
			return nil
		}
	}
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: provider + seed, Check: check("seed", false)},
		{Config: provider + ipsecAcceptanceImportConfig(), Check: func(state *terraform.State) error {
			if connection.adds != 1 || child.adds != 1 || connection.saves+child.saves+connection.deletes+child.deletes != 0 {
				return fmt.Errorf("import transfer mutated owners")
			}
			if state.RootModule().Resources["opnsense_ipsec_connection.seed"] != nil || state.RootModule().Resources["opnsense_ipsec_child.seed"] != nil {
				return fmt.Errorf("seed owners remain after imports")
			}
			return check("advanced", false)(state)
		}},
		{Config: provider + ipsecAcceptanceOmittedConfig(), Check: check("advanced", false)},
		{Config: provider + ipsecAcceptanceOmittedConfig(), PlanOnly: true, ExpectNonEmptyPlan: false},
		{Config: provider + ipsecAcceptanceConfig(true, false), Check: check("advanced", false)},
		{Config: provider + ipsecAcceptanceConfig(true, true), Check: check("advanced", true)},
	}})
	if connection.adds != 1 || child.adds != 1 || connection.deletes != 1 || child.deletes != 1 {
		t.Fatal("Core lifecycle duplicated owners or failed cleanup")
	}
}

func TestIPsecAcceptanceAbsentStructural(t *testing.T) {
	for _, tc := range []struct {
		body   string
		absent bool
	}{
		{"[ \n\t ]", true},
		{"{\"child\":[ \n ]}", true},
		{"null", false},
		{"{}", false},
		{"{\"child\":{}}", false},
		{"{\"child\":null}", false},
		{"{\"child\":[{}]}", false},
	} {
		t.Run(tc.body, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = fmt.Fprint(w, tc.body) }))
			defer server.Close()
			client, err := opnsense.NewClient(opnsense.ClientConfig{BaseURL: server.URL, APIKey: "fixture-key", APISecret: "fixture-secret"})
			if err != nil {
				t.Fatal(err)
			}
			if got := ipsecAbsent(client, "child", ipsecID) == nil; got != tc.absent {
				t.Fatalf("absence=%t, want %t", got, tc.absent)
			}
		})
	}
}

func TestIPsecAcceptanceExactFirmwarePackage(t *testing.T) {
	for _, tc := range []struct {
		raw   string
		valid bool
	}{
		{`{"product":{"product_name":"OPNsense","product_version":"26.7.8_2"}}`, true},
		{`{"product":{"product_name":"OPNsense","product_version":"26.7.8"}}`, false},
		{`{"product":{"product_name":"OPNsense","product_version":"26.7.8_3"}}`, false},
		{`{"product":{"product_name":"other","product_version":"26.7.8_2"}}`, false},
		{`{}`, false},
	} {
		if got := validateIPsecAcceptanceFirmware([]byte(tc.raw)) == nil; got != tc.valid {
			t.Fatalf("firmware gate result=%t want=%t", got, tc.valid)
		}
	}
}
