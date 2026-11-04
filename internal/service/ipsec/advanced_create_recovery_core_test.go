// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec_test

import (
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

// A saved UUID alone does not prevent Core taint/replacement after a failed
// create. Exercise the next real apply, including Core's refresh and planning.
func TestIPsecCoreConfirmedCreateRecoveryDoesNotRecreate(t *testing.T) {
	binary, err := exec.LookPath("terraform")
	if err != nil {
		t.Fatal("container Terraform Core is required", err)
	}
	t.Setenv("TF_ACC_TERRAFORM_PATH", binary)
	for _, owner := range []string{"connection", "child"} {
		for _, failure := range []string{"activation", "readback"} {
			t.Run(owner+"/"+failure, func(t *testing.T) {
				api := &ipsecAPI{owner: owner}
				var mu sync.Mutex
				pending, failed := false, 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					mu.Lock()
					defer mu.Unlock()
					if r.URL.Path == "/api/ipsec/connections/add_"+owner {
						pending = failed == 0
					}
					activationFailure := failure == "activation" && r.URL.Path == "/api/ipsec/service/reconfigure"
					readFailure := failure == "readback" && r.URL.Path == "/api/ipsec/connections/get_"+owner+"/"+ipsecID
					if pending && (activationFailure || readFailure) {
						pending = false
						failed++
						api.mu.Lock()
						if activationFailure {
							api.activations++
						} else {
							api.reads++
						}
						api.mu.Unlock()
						// Successful HTTP responses avoid transport retry behavior:
						// fail precisely the post-save activation/readback boundary.
						if activationFailure {
							_, _ = fmt.Fprint(w, `{"status":"failed"}`)
						} else {
							_, _ = fmt.Fprint(w, `{}`)
						}
						return
					}
					api.ServeHTTP(w, r)
				}))
				defer server.Close()
				config := fmt.Sprintf("provider \"opnsense\" {\nuri=%q\napi_key=\"fixture-key\"\napi_secret=\"fixture-secret\"\n}\n", server.URL) + ipsecCoreHCL(owner, ipsecConfig(owner))
				name := "opnsense_ipsec_" + owner + ".test"
				check := resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "id", ipsecID),
					func(_ *terraform.State) error {
						mu.Lock()
						defer mu.Unlock()
						api.mu.Lock()
						defer api.mu.Unlock()
						if failed != 1 || api.adds != 1 || api.saves != 0 || api.deletes != 0 {
							return fmt.Errorf("recovery changed confirmed ownership: failures=%d adds=%d saves=%d interim deletes=%d", failed, api.adds, api.saves, api.deletes)
						}
						t.Logf("Core applied state: failed=%d adds=%d saves=%d deletes=%d activations=%d", failed, api.adds, api.saves, api.deletes, api.activations)
						return nil
					},
				)
				resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories, Steps: []resource.TestStep{
					{Config: config, Check: check},
					{Config: config, Check: check},
					{Config: config, PlanOnly: true, ExpectNonEmptyPlan: false},
				}})
				mu.Lock()
				defer mu.Unlock()
				api.mu.Lock()
				defer api.mu.Unlock()
				// Only the harness's final destroy may delete the saved object.
				if failed != 1 || api.adds != 1 || api.saves != 0 || api.deletes != 1 {
					t.Fatalf("unexpected final ownership: failures=%d adds=%d saves=%d deletes=%d", failed, api.adds, api.saves, api.deletes)
				}
				wantActivations := 2 // Successful create and final destroy.
				if failure == "activation" {
					// Read-time activation recovery can repeat across a discarded
					// speculative refresh. plugin-testing v1.15.0 runs a post-apply
					// refresh plan (testing_new_config.go:340), then the next step's
					// pre-apply plan (:112). The former is not applied, so both see
					// the saved marker. Neither may repeat an object write.
					wantActivations += 2
				}
				if api.activations != wantActivations {
					t.Fatalf("activation replay/loss: got %d, want %d", api.activations, wantActivations)
				}
				t.Logf("Core final cleanup: failed=%d adds=%d saves=%d deletes=%d activations=%d", failed, api.adds, api.saves, api.deletes, api.activations)
			})
		}
	}
}
