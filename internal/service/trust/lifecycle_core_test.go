// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust_test

import (
	"fmt"
	"net/http/httptest"
	"os/exec"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
)

// Terraform Core verifies planned-value consistency, persistent state, import,
// and empty repeated plans against only a loopback synthetic API, never live PKI.
func TestTrustLifecycleCoreLegacyCreateRefreshMetadataImport(t *testing.T) {
	binary, err := exec.LookPath("terraform")
	if err != nil {
		t.Fatal("container Terraform Core is required", err)
	}
	t.Setenv("TF_ACC_TERRAFORM_PATH", binary)
	cert, key := selfSignedPEM(t, "example.test")
	for _, kind := range []string{"ca", "cert"} {
		t.Run(kind, func(t *testing.T) {
			a := &lifecycleAPI{kind: kind, certificate: cert, key: key}
			h := httptest.NewServer(a)
			defer h.Close()
			factories := map[string]func() (tfprotov6.ProviderServer, error){"opnsense": providerserver.NewProtocol6WithError(provider.New("offline-trust-test")())}
			config := func(description string) string {
				fields := "common_name = \"example.test\"\ncountry = \"US\""
				if kind == "cert" {
					fields = fmt.Sprintf("certificate = %q\nprivate_key = %q", cert, key)
				}
				return fmt.Sprintf(`provider "opnsense" {
 uri = %q
 api_key = "fixture-key"
 api_secret = "fixture-secret"
}
resource "opnsense_trust_%s" "test" {
 description = %q
 %s
}`, h.URL, kind, description, fields)
			}
			ignored := []string{"operation_status"}
			if kind == "ca" {
				ignored = append(ignored, "lifetime")
			} else {
				ignored = append(ignored, "certificate", "private_key")
			}
			resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: factories, Steps: []resource.TestStep{
				{Config: config("fixture"), Check: resource.TestCheckResourceAttr("opnsense_trust_"+kind+".test", "id", "fixture-id")},
				{Config: config("fixture"), PlanOnly: true},
				{Config: config("fixture renamed"), Check: resource.TestCheckResourceAttr("opnsense_trust_"+kind+".test", "description", "fixture renamed")},
				{Config: config("fixture renamed"), PlanOnly: true},
				{ResourceName: "opnsense_trust_" + kind + ".test", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: ignored},
			}})
			a.mu.Lock()
			defer a.mu.Unlock()
			if len(a.posts) != 2 {
				t.Fatalf("Core lifecycle unexpectedly mutated: %d writes", len(a.posts))
			}
		})
	}
}

func TestTrustLifecycleCoreExplicitLocalAndUncertainIntent(t *testing.T) {
	binary, err := exec.LookPath("terraform")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TF_ACC_TERRAFORM_PATH", binary)
	cert, key := selfSignedPEM(t, "example.test")
	for _, mode := range []string{"explicit", "local", "uncertain"} {
		t.Run(mode, func(t *testing.T) {
			a := &lifecycleAPI{kind: "cert", certificate: cert, key: key}
			h := httptest.NewServer(a)
			defer h.Close()
			factories := map[string]func() (tfprotov6.ProviderServer, error){"opnsense": providerserver.NewProtocol6WithError(provider.New("offline-trust-test")())}
			config := func(action, token string) string {
				location := "firewall"
				if mode == "local" {
					location = "local"
				}
				return fmt.Sprintf(`provider "opnsense" {
 uri=%q
 api_key="fixture-key"
 api_secret="fixture-secret"
}
resource "opnsense_trust_cert" "test" {
 description="fixture"
 common_name="example.test"
 ca_ref="issuer"
 action=%q
 operation_token=%q
 private_key_location=%q
}`, h.URL, action, token, location)
			}
			steps := []resource.TestStep{{Config: config("internal", "initial")}, {Config: config("internal", "initial"), PlanOnly: true}}
			switch mode {
			case "local":
				steps = append(steps, resource.TestStep{Config: config("internal", "initial"), Check: resource.TestCheckResourceAttr("opnsense_trust_cert.test", "generated_private_key", "one-time-sentinel")})
			case "explicit":
				steps = append(steps, resource.TestStep{Config: config("reissue", "transition")}, resource.TestStep{Config: config("reissue", "repeat")}, resource.TestStep{Config: config("reissue", "repeat"), PlanOnly: true})
			case "uncertain":
				steps = append(steps, resource.TestStep{Config: config("reissue", "attempt"), PreConfig: func() { a.mu.Lock(); a.failWrite = 500; a.mu.Unlock() }, ExpectError: regexp.MustCompile("Trust update failed")}, resource.TestStep{Config: config("reissue", "attempt"), PreConfig: func() { a.mu.Lock(); a.failWrite = 0; a.mu.Unlock() }, Check: resource.TestCheckResourceAttr("opnsense_trust_cert.test", "operation_status", "unknown")}, resource.TestStep{Config: config("reissue", "attempt"), PlanOnly: true})
			}
			resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: factories, Steps: steps})
			a.mu.Lock()
			defer a.mu.Unlock()
			want := 1
			if mode == "explicit" {
				want = 3
			}
			if mode == "uncertain" {
				want = 2
			}
			if len(a.posts) != want {
				t.Fatalf("Core repeated lifecycle operation: %d writes, want %d", len(a.posts), want)
			}
		})
	}
}
