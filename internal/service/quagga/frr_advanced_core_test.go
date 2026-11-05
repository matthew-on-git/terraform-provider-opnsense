// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

// This is an actual Terraform Core lifecycle against a strict HTTP appliance
// fixture, independent of appliance acceptance. No TF_ACC flag or lab is needed.
func TestFRRAdvancedTerraformCore(t *testing.T) {
	var mu sync.Mutex
	var object map[string]any
	var group map[string]any
	writes := 0
	secretSaves := 0
	emptySecretSaves := 0
	const id = "11111111-1111-4111-8111-111111111111"
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if frrFixture(w, r) {
			return
		}
		switch r.URL.Path {
		case "/api/quagga/bgp/search_peergroup":
			rows := []map[string]string{}
			if group != nil {
				rows = append(rows, map[string]string{"uuid": id})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"rows": rows, "total": len(rows), "current": 1, "rowCount": len(rows)})
		case "/api/quagga/bgp/add_peergroup", "/api/quagga/bgp/set_peergroup/" + id:
			var body map[string]map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if group == nil {
				group = map[string]any{"enabled": "1", "name": "fixture", "remote_as_mode": "", "remoteas": "64512", "family": "ipv4", "updatesource": "", "nexthopself": "0", "defaultoriginate": "0", "listenranges": "", "linkedPrefixlistIn": "", "linkedPrefixlistOut": "", "linkedRoutemapIn": "", "linkedRoutemapOut": ""}
			}
			for k, v := range body["peergroup"] {
				group[k] = v
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"result": "saved", "uuid": id})
		case "/api/quagga/bgp/get_peergroup/" + id:
			_ = json.NewEncoder(w).Encode(map[string]any{"peergroup": group})
		case "/api/quagga/bgp/del_peergroup/" + id:
			group = nil
			_, _ = fmt.Fprint(w, `{"result":"deleted"}`)
		case "/api/quagga/bgp/search_neighbor":
			rows := []map[string]string{}
			if object != nil {
				rows = append(rows, map[string]string{"uuid": id})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"rows": rows, "total": len(rows), "current": 1, "rowCount": len(rows)})
		case "/api/quagga/bgp/add_neighbor", "/api/quagga/bgp/set_neighbor/" + id:
			var body map[string]map[string]string
			if json.NewDecoder(r.Body).Decode(&body) != nil {
				t.Error("invalid request")
				return
			}
			writes++
			if object == nil {
				object = finalNeighbor()
			}
			for k, v := range body["neighbor"] {
				if k == "password" {
					secretSaves++
					if v == "" {
						emptySecretSaves++
					}
					continue
				}
				object[k] = v
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"result": "saved", "uuid": id})
		case "/api/quagga/bgp/get_neighbor/" + id:
			if object == nil {
				_, _ = fmt.Fprint(w, `{"neighbor":{}}`)
				return
			}
			// The API can return a credential; the provider response type must discard it.
			response := map[string]any{}
			for k, v := range object {
				response[k] = v
			}
			response["password"] = "remote-unowned-input"
			_ = json.NewEncoder(w).Encode(map[string]any{"neighbor": response})
		case "/api/quagga/bgp/del_neighbor/" + id:
			object = nil
			_, _ = fmt.Fprint(w, `{"result":"deleted"}`)
		case "/api/quagga/service/reconfigure":
			_, _ = fmt.Fprint(w, `{"status":"ok"}`)
		default:
			t.Errorf("unexpected fixture API %s", r.URL.Path)
		}
	}))
	defer api.Close()
	base := fmt.Sprintf(`provider "opnsense" {
 uri = %q
 api_key = "fixture-key"
 api_secret = "fixture-secret"
 insecure = true
}
resource "opnsense_quagga_bgp_neighbor" "peer" {
 address = "192.0.2.1"
 family = ["ipv4","ipv6"]
 remote_as_mode = "external"
 local_as = 64512
 weight = 0
 local_ip = "192.0.2.2"
 link_local_interface = ""
 next_hop_self = true
 next_hop_self_all = true
 multihop = false
 route_reflector_client = true
 soft_reconfiguration_inbound = true
 bfd = true
 bfd_strict = true
 connect_timer = 30
 default_originate = false
 as_override = false
 remove_private_as = "remove-private-AS all replace-AS"
 allow_as_in = "origin"
 disable_connected_check = true
 attribute_unchanged = "as-path"
 capabilities = ["dynamic","fqdn"]
 %%s
}
data "opnsense_quagga_bgp_neighbor" "read" { id = opnsense_quagga_bgp_neighbor.peer.id }
resource "opnsense_quagga_bgp_peergroup" "group" {
 name = "fixture"
 remote_as = 64512
 listen_ranges = ["192.0.2.0/24", "2001:db8::/64"]
}
data "opnsense_quagga_bgp_peergroup" "read" { id = opnsense_quagga_bgp_peergroup.group.id }
`, api.URL)
	config := fmt.Sprintf(base, `password = "configured-test-input"`)
	omitted := fmt.Sprintf(base, "")
	adopted := regexp.MustCompile(`(?m)^ (?:remote_as_mode|local_as|weight|local_ip|link_local_interface|next_hop_self_all|multihop|route_reflector_client|soft_reconfiguration_inbound|bfd|bfd_strict|connect_timer|default_originate|as_override|remove_private_as|allow_as_in|disable_connected_check|attribute_unchanged|capabilities) = [^\n]*\n`).ReplaceAllString(omitted, "")
	if adopted == omitted || regexp.MustCompile(`(?m)^ (?:weight|multihop|attribute_unchanged|capabilities) =`).MatchString(adopted) {
		t.Fatal("adoption control retained advanced assignments")
	}
	rangesOmitted := strings.Replace(omitted, ` listen_ranges = ["192.0.2.0/24", "2001:db8::/64"]`, "", 1)
	rangesUpdated := strings.Replace(rangesOmitted, ` name = "fixture"`, ` name = "renamed"`, 1)
	checkRanges := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("opnsense_quagga_bgp_peergroup.group", "listen_ranges.#", "2"),
		resource.TestCheckTypeSetElemAttr("opnsense_quagga_bgp_peergroup.group", "listen_ranges.*", "192.0.2.0/24"),
		resource.TestCheckTypeSetElemAttr("opnsense_quagga_bgp_peergroup.group", "listen_ranges.*", "2001:db8::/64"),
		resource.TestCheckResourceAttr("data.opnsense_quagga_bgp_peergroup.read", "listen_ranges.#", "2"),
		resource.TestCheckTypeSetElemAttr("data.opnsense_quagga_bgp_peergroup.read", "listen_ranges.*", "192.0.2.0/24"),
		resource.TestCheckTypeSetElemAttr("data.opnsense_quagga_bgp_peergroup.read", "listen_ranges.*", "2001:db8::/64"),
	)
	updated := strings.ReplaceAll(omitted, `attribute_unchanged = "as-path"`, `attribute_unchanged = ""`)
	updated = strings.ReplaceAll(updated, `listen_ranges = ["192.0.2.0/24", "2001:db8::/64"]`, `listen_ranges = []`)
	updated = strings.Replace(updated, `capabilities = ["dynamic","fqdn"]`, `capabilities = []`, 1)
	checkCapabilitiesCleared := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("opnsense_quagga_bgp_neighbor.peer", "capabilities.#", "0"),
		resource.TestCheckResourceAttr("data.opnsense_quagga_bgp_neighbor.read", "capabilities.#", "0"),
	)
	cleared := strings.Replace(updated, " capabilities =", " password = \"\"\n capabilities =", 1)
	numeric := strings.Replace(updated, ` remote_as_mode = "external"`, " remote_as_mode = \"\"\n remote_as = 64514", 1)
	modeOnly := updated
	inherited := strings.Replace(updated, ` remote_as_mode = "external"`, " remote_as_mode = \"\"\n peer_group = opnsense_quagga_bgp_peergroup.group.id", 1)
	checkAS := func(mode, numericAS string) resource.TestCheckFunc {
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("opnsense_quagga_bgp_neighbor.peer", "remote_as_mode", mode),
			resource.TestCheckResourceAttr("data.opnsense_quagga_bgp_neighbor.read", "remote_as_mode", mode),
			resource.TestCheckResourceAttr("opnsense_quagga_bgp_neighbor.peer", "remote_as", numericAS),
			resource.TestCheckResourceAttr("data.opnsense_quagga_bgp_neighbor.read", "remote_as", numericAS),
		)
	}
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr("opnsense_quagga_bgp_neighbor.peer", "weight", "0"), resource.TestCheckResourceAttr("opnsense_quagga_bgp_neighbor.peer", "password", "configured-test-input"), resource.TestCheckResourceAttr("data.opnsense_quagga_bgp_neighbor.read", "attribute_unchanged", "as-path"))},
		{ResourceName: "opnsense_quagga_bgp_neighbor.peer", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"password"}},
		{Config: config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
		{Config: omitted, Check: func(_ *terraform.State) error {
			mu.Lock()
			defer mu.Unlock()
			if secretSaves != 1 {
				return fmt.Errorf("omitted authentication was written %d times", secretSaves)
			}
			return nil
		}},
		{Config: rangesOmitted, Check: checkRanges, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
		{Config: rangesUpdated, Check: resource.ComposeAggregateTestCheckFunc(checkRanges, resource.TestCheckResourceAttr("opnsense_quagga_bgp_peergroup.group", "name", "renamed"))},
		{Config: rangesUpdated, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
		{Config: adopted},
		{Config: adopted, Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr("opnsense_quagga_bgp_neighbor.peer", "weight", "0"), resource.TestCheckResourceAttr("opnsense_quagga_bgp_neighbor.peer", "multihop", "false"), resource.TestCheckResourceAttr("opnsense_quagga_bgp_neighbor.peer", "attribute_unchanged", "as-path"), resource.TestCheckResourceAttr("opnsense_quagga_bgp_neighbor.peer", "capabilities.#", "2")), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
		{PreConfig: func() { mu.Lock(); object["attributeunchanged"] = "med"; mu.Unlock() }, Config: updated, Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr("opnsense_quagga_bgp_neighbor.peer", "attribute_unchanged", ""), checkCapabilitiesCleared)},
		{Config: updated, Check: checkCapabilitiesCleared, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
		{Config: cleared, Check: func(_ *terraform.State) error {
			mu.Lock()
			defer mu.Unlock()
			if emptySecretSaves != 1 || secretSaves != 2 {
				return fmt.Errorf("explicit authentication clear did not save exactly once")
			}
			return nil
		}},
		{Config: cleared, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
		{Config: numeric, Check: checkAS("", "64514")},
		// Removing numeric configuration is not a numeric clear: the retained
		// stored AS remains public while the configured mode takes precedence.
		{Config: modeOnly, Check: checkAS("external", "64514")},
		{Config: modeOnly, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
		// Inheritance requires an actually unset appliance AS. Simulate that
		// external clear before refresh; do not invent a provider clear sentinel.
		{PreConfig: func() { mu.Lock(); object["remoteas"] = ""; mu.Unlock() }, Config: inherited, Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckNoResourceAttr("opnsense_quagga_bgp_neighbor.peer", "remote_as"),
			resource.TestCheckNoResourceAttr("data.opnsense_quagga_bgp_neighbor.read", "remote_as"),
			resource.TestCheckResourceAttr("opnsense_quagga_bgp_neighbor.peer", "remote_as_mode", ""),
			resource.TestCheckResourceAttr("opnsense_quagga_bgp_neighbor.peer", "peer_group", id),
			resource.TestCheckResourceAttr("data.opnsense_quagga_bgp_neighbor.read", "peer_group", id),
		)},
		{Config: inherited, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
	}})
	mu.Lock()
	defer mu.Unlock()
	if object != nil || group != nil || writes < 2 || emptySecretSaves != 1 || secretSaves != 2 {
		t.Fatal("Core lifecycle did not mutate and delete fixture")
	}
}
