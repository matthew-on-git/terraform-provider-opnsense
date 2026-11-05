// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

// Protocol state retention alone cannot detect Core taint. Exercise two actual
// applies against the released saved-UUID/status/neighbor response contracts.
func TestFRRCoreConfirmedCreateRecoveryDoesNotRecreate(t *testing.T) {
	binary, err := exec.LookPath("terraform")
	if err != nil {
		t.Fatal("container Terraform Core is required", err)
	}
	t.Setenv("TF_ACC_TERRAFORM_PATH", binary)
	for _, failure := range []string{"activation", "readback", "validation", "normalization"} {
		t.Run(failure, func(t *testing.T) {
			const id = "11111111-1111-4111-8111-111111111111"
			var mu sync.Mutex
			var object map[string]any
			adds, saves, deletes, activations, failures := 0, 0, 0, 0, 0
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				if frrFixture(w, r) {
					return
				}
				switch r.URL.Path {
				case "/api/quagga/bgp/search_neighbor":
					rows := []map[string]string{}
					if object != nil {
						rows = append(rows, map[string]string{"uuid": id})
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"rows": rows, "total": len(rows), "current": 1, "rowCount": len(rows)})
				case "/api/quagga/bgp/add_neighbor", "/api/quagga/bgp/set_neighbor/" + id:
					if r.URL.Path == "/api/quagga/bgp/add_neighbor" {
						adds++
					} else {
						saves++
					}
					var body map[string]map[string]string
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
						return
					}
					if object == nil {
						object = finalNeighbor()
					}
					for k, v := range body["neighbor"] {
						object[k] = v
					}
					_ = json.NewEncoder(w).Encode(map[string]string{"result": "saved", "uuid": id})
				case "/api/quagga/service/reconfigure":
					activations++
					if failure == "activation" && failures == 0 {
						failures++
						_, _ = fmt.Fprint(w, `{"status":"failed"}`)
						return
					}
					_, _ = fmt.Fprint(w, `{"status":"ok"}`)
				case "/api/quagga/bgp/get_neighbor/" + id:
					if object == nil {
						_, _ = fmt.Fprint(w, `{"neighbor":{}}`)
						return
					}
					response := map[string]any{}
					for k, v := range object {
						response[k] = v
					}
					if failures == 0 {
						switch failure {
						case "readback":
							failures++
							_, _ = fmt.Fprint(w, `{"neighbor":{}}`)
							return
						case "validation":
							failures++
							delete(response, "enabled")
						case "normalization":
							failures++
							response["peergroup"] = "stale-group"
						}
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"neighbor": response})
				case "/api/quagga/bgp/get_peergroup/stale-group":
					_, _ = fmt.Fprint(w, `{"peergroup":{}}`)
				case "/api/quagga/bgp/del_neighbor/" + id:
					deletes++
					object = nil
					_, _ = fmt.Fprint(w, `{"result":"deleted"}`)
				default:
					t.Errorf("unexpected recovery fixture API %s", r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer api.Close()
			config := fmt.Sprintf(`provider "opnsense" {
 uri = %q
 api_key = "fixture-key"
 api_secret = "fixture-secret"
}
resource "opnsense_quagga_bgp_neighbor" "peer" {
 address = "192.0.2.1"
 remote_as = 64512
 family = ["ipv4"]
}
`, api.URL)
			check := resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr("opnsense_quagga_bgp_neighbor.peer", "id", id), func(_ *terraform.State) error {
				mu.Lock()
				defer mu.Unlock()
				if failures != 1 || adds != 1 || saves != 0 || deletes != 0 {
					return fmt.Errorf("confirmed recovery: failures=%d adds=%d saves=%d deletes=%d activations=%d", failures, adds, saves, deletes, activations)
				}
				return nil
			})
			resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories, Steps: []resource.TestStep{
				{Config: config, Check: check},
				{Config: config, Check: resource.ComposeAggregateTestCheckFunc(check, func(_ *terraform.State) error {
					mu.Lock()
					defer mu.Unlock()
					if activations < 2 {
						return fmt.Errorf("pending activation was not retried: activations=%d", activations)
					}
					return nil
				})},
				{Config: config, PlanOnly: true, ExpectNonEmptyPlan: false},
			}})
			mu.Lock()
			defer mu.Unlock()
			if failures != 1 || adds != 1 || saves != 0 || deletes != 1 || object != nil {
				t.Fatalf("final ownership: failures=%d adds=%d saves=%d deletes=%d remaining=%t", failures, adds, saves, deletes, object != nil)
			}
			t.Logf("Core final ownership: failures=%d adds=%d saves=%d deletes=%d activations=%d", failures, adds, saves, deletes, activations)
		})
	}
}

// Singleton Delete is intentionally state-only. A tainted confirmed create must
// still not replay its settings save on the next real Core apply.
func TestFRRCoreConfirmedSingletonCreateRecoveryDoesNotResave(t *testing.T) {
	binary, err := exec.LookPath("terraform")
	if err != nil {
		t.Fatal("container Terraform Core is required", err)
	}
	t.Setenv("TF_ACC_TERRAFORM_PATH", binary)
	for _, kind := range []string{"general", "bgp"} {
		for _, failure := range []string{"activation", "readback", "validation"} {
			t.Run(kind+"/"+failure, func(t *testing.T) {
				var mu sync.Mutex
				saves, activations, failures := 0, 0, 0
				fields := map[string]string{"enabled": "1", "profile": "traditional", "enablecarp": "0", "enablesyslog": "1", "enablesnmp": "0", "sysloglevel": "notifications", "fwrules": "0"}
				name := "opnsense_quagga_general"
				hcl := " enabled = true\n"
				if kind == "bgp" {
					name = "opnsense_quagga_bgp_global"
					hcl += " as_number = 65111\n router_id = \"198.18.31.1\"\n"
					fields = map[string]string{"enabled": "1", "asnumber": "65111", "routerid": "198.18.31.1", "distance": "", "graceful": "0", "networkimportcheck": "1", "enforce_first_as": "1", "logneighborchanges": "0", "networks": "", "maximumpaths": "", "maximumpathsibgp": ""}
				}
				api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					mu.Lock()
					defer mu.Unlock()
					if frrFixture(w, r) {
						return
					}
					switch r.URL.Path {
					case "/api/quagga/" + kind + "/set":
						saves++
						var body map[string]map[string]string
						if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
							t.Error(err)
							return
						}
						for k, v := range body[kind] {
							fields[k] = v
						}
						_, _ = fmt.Fprint(w, `{"result":"saved"}`)
					case "/api/quagga/" + kind + "/get":
						result := map[string]string{}
						for k, v := range fields {
							result[k] = v
						}
						if saves > 0 && failures == 0 {
							if failure == "readback" {
								failures++
								_, _ = fmt.Fprint(w, `{}`)
								return
							}
							if failure == "validation" {
								failures++
								delete(result, "enabled")
							}
						}
						_ = json.NewEncoder(w).Encode(map[string]any{kind: result})
					case "/api/quagga/service/reconfigure":
						activations++
						if failure == "activation" && failures == 0 {
							failures++
							_, _ = fmt.Fprint(w, `{"status":"failed"}`)
							return
						}
						_, _ = fmt.Fprint(w, `{"status":"ok"}`)
					default:
						t.Errorf("unexpected singleton recovery API %s", r.URL.Path)
						http.NotFound(w, r)
					}
				}))
				defer api.Close()
				config := fmt.Sprintf("provider \"opnsense\" {\n uri = %q\n api_key = \"fixture-key\"\n api_secret = \"fixture-secret\"\n}\nresource %q \"settings\" {\n%s}\n", api.URL, name, hcl)
				check := resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(name+".settings", "id", kind), func(_ *terraform.State) error {
					mu.Lock()
					defer mu.Unlock()
					if failures != 1 || saves != 1 {
						return fmt.Errorf("singleton recovery: failures=%d saves=%d activations=%d", failures, saves, activations)
					}
					return nil
				})
				resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories, Steps: []resource.TestStep{
					{Config: config, Check: check},
					{Config: config, Check: resource.ComposeAggregateTestCheckFunc(check, func(_ *terraform.State) error {
						mu.Lock()
						defer mu.Unlock()
						if activations < 2 {
							return fmt.Errorf("singleton recovery did not retry activation: %d", activations)
						}
						return nil
					})},
					{Config: config, PlanOnly: true, ExpectNonEmptyPlan: false},
				}})
				mu.Lock()
				defer mu.Unlock()
				if failures != 1 || saves != 1 {
					t.Fatalf("singleton final recovery: failures=%d saves=%d", failures, saves)
				}
				t.Logf("Core singleton final: failures=%d saves=%d activations=%d", failures, saves, activations)
			})
		}
	}
}
