// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
)

// Exercise real Core recovery, not just ApplyResourceChange's returned state:
// an error after a confirmed add taints the object even if its UUID was saved.
func TestTrustLifecycleCoreConfirmedCreateRecoveryDoesNotRecreate(t *testing.T) {
	binary, err := exec.LookPath("terraform")
	if err != nil {
		t.Fatal("container Terraform Core is required", err)
	}
	t.Setenv("TF_ACC_TERRAFORM_PATH", binary)
	cert, key := selfSignedPEM(t, "example.test")
	for _, tc := range []struct {
		name, kind string
		failRead   bool
		local      bool
	}{
		{name: "ca-first-readback", kind: "ca", failRead: true},
		{name: "cert-first-readback", kind: "cert", failRead: true},
		{name: "local-confirmed-missing-key", kind: "cert", local: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := &lifecycleAPI{kind: tc.kind, certificate: cert, key: key}
			if tc.local {
				// The appliance confirms creation but the one-time response lacks
				// its key. A later GET cannot recover that response-only material.
				a.key = ""
				a.reply = `{"result":"saved","uuid":"fixture-id"}`
			}
			var mu sync.Mutex
			adds, sets, deletes, failedReads := 0, 0, 0, 0
			pendingReadFailure := false
			h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				if strings.Contains(r.URL.Path, "/get/") && pendingReadFailure {
					pendingReadFailure = false
					failedReads++
					// A single malformed successful response avoids HTTP retries,
					// isolating the confirmed-create follow-up failure boundary.
					_, _ = fmt.Fprint(w, `{}`)
					return
				}
				if r.Method == http.MethodPost {
					switch {
					case strings.HasSuffix(r.URL.Path, "/add"):
						adds++
						pendingReadFailure = tc.failRead && adds == 1
					case strings.Contains(r.URL.Path, "/set/"):
						sets++
					case strings.Contains(r.URL.Path, "/del/"):
						deletes++
					}
				}
				a.ServeHTTP(w, r)
			}))
			defer h.Close()
			factories := map[string]func() (tfprotov6.ProviderServer, error){
				"opnsense": providerserver.NewProtocol6WithError(provider.New("offline-trust-recovery-test")()),
			}
			extra := ""
			if tc.local {
				extra = `private_key_location = "local"`
			}
			config := fmt.Sprintf(`provider "opnsense" {
 uri = %q
 api_key = "fixture-key"
 api_secret = "fixture-secret"
}
resource "opnsense_trust_%s" "test" {
 description = "fixture"
 action = "internal"
 common_name = "example.test"
 country = "US"
 %s
}`, h.URL, tc.kind, extra)
			check := resource.TestCheckResourceAttr("opnsense_trust_"+tc.kind+".test", "id", "fixture-id")
			resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: factories, Steps: []resource.TestStep{
				{Config: config, Check: check},
				// Refresh + a second real apply must not destroy/recreate a
				// tainted confirmed object or replay the cryptographic action.
				{Config: config, Check: check},
				{Config: config, PlanOnly: true},
			}})
			mu.Lock()
			defer mu.Unlock()
			// The test harness's final destroy is the only permitted deletion.
			if adds != 1 || sets != 0 || deletes != 1 {
				t.Fatalf("confirmed create recovery replayed mutation: adds=%d sets=%d deletes=%d", adds, sets, deletes)
			}
			if tc.failRead && failedReads != 1 {
				t.Fatalf("first-readback failure was not exercised: %d", failedReads)
			}
		})
	}
}
