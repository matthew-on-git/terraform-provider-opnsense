// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec_test

import (
	"fmt"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

// UnitTest runs actual local Terraform CLI/Core against only an HTTP fixture.
// Unlike protocol tests, Core constructs ProposedNewState and checks no-op plans.
// Run with the package's bounded Go test timeout; no TF_ACC/live credentials.
func TestIPsecCoreImportedOmissionAndCSVNoChange(t *testing.T) {
	binary, err := exec.LookPath("terraform")
	if err != nil {
		t.Fatal("local Terraform Core required", err)
	}
	t.Setenv("TF_ACC_TERRAFORM_PATH", binary)
	for _, owner := range []string{"connection", "child"} {
		t.Run(owner, func(t *testing.T) {
			api := &ipsecAPI{owner: owner, otherFields: ipsecFixture(owner), authFields: map[string]string{}}
			httpServer := httptest.NewServer(api)
			defer httpServer.Close()
			providerConfig := fmt.Sprintf("provider \"opnsense\" {\nuri=%q\napi_key=\"fixture-key\"\napi_secret=\"fixture-secret\"\n}\n", httpServer.URL)
			fields := ipsecConfig(owner)
			all := providerConfig + ipsecCoreHCL(owner, fields)
			importedFields := map[string]any{}
			for key, value := range fields {
				if isLegacyIPsec(owner, key) {
					if text, ok := value.(string); ok && strings.Contains(text, ",") {
						value = canonicalIPsecCSV(text)
					}
					importedFields[key] = value
				}
			}
			omitted := providerConfig + ipsecCoreHCL(owner, importedFields)
			changedFields := map[string]any{}
			for key, value := range importedFields {
				changedFields[key] = value
			}
			changedFields["description"] = "updated imported fixture"
			changed := providerConfig + ipsecCoreHCL(owner, changedFields)
			name := "opnsense_ipsec_" + owner + ".test"
			checkRetained := func(_ *terraform.State) error {
				api.mu.Lock()
				defer api.mu.Unlock()
				for key, value := range expectedIPsecAdvancedWire(owner) {
					if api.fields[key] != value {
						return fmt.Errorf("import/omission reset %s", key)
					}
				}
				for key := range api.last {
					if !isLegacyWire(owner, key) {
						return fmt.Errorf("omitted imported advanced key serialized: %s", key)
					}
				}
				return nil
			}
			resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories, Steps: []resource.TestStep{
				{Config: all, Check: func(_ *terraform.State) error {
					assertIPsecAdvancedWire(t, owner, api.last, expectedIPsecAdvancedWire(owner))
					return nil
				}},
				{Config: all, PlanOnly: true, ExpectNonEmptyPlan: false},
			}})
			// Separate Core state imports a pre-existing source-shaped object.
			api.mu.Lock()
			api.fields = ipsecFixture(owner)
			api.mu.Unlock()
			resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories, Steps: []resource.TestStep{
				{Config: omitted, ResourceName: name, ImportState: true, ImportStateId: ipsecID, ImportStatePersist: true, ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 || states[0].ID != ipsecID || states[0].Attributes["rekey_time"] != "3600" {
						return fmt.Errorf("import lost advanced state")
					}
					return nil
				}},
				{Config: omitted, PlanOnly: true, ExpectNonEmptyPlan: false},
				{Config: changed, Check: checkRetained},
				{Config: changed, PlanOnly: true, ExpectNonEmptyPlan: false, Check: checkRetained},
			}})
			if api.adds != 1 || api.saves != 1 || api.deletes != 2 {
				t.Fatalf("Core lifecycle counts: add=%d save=%d delete=%d", api.adds, api.saves, api.deletes)
			}
		})
	}
}

func ipsecCoreHCL(owner string, fields map[string]any) string {
	var hcl strings.Builder
	fmt.Fprintf(&hcl, "resource \"opnsense_ipsec_%s\" \"test\" {\n", owner)
	for key, value := range fields {
		if text, ok := value.(string); ok {
			fmt.Fprintf(&hcl, "%s=%q\n", key, text)
		} else {
			fmt.Fprintf(&hcl, "%s=%t\n", key, value)
		}
	}
	hcl.WriteString("}\n")
	return hcl.String()
}
