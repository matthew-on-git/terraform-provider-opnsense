// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust_test

import (
	"fmt"
	"net/http/httptest"
	"os/exec"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
	"github.com/zclconf/go-cty/cty"
)

// Exercise the actual generated acceptance fixture's settings graph in Core,
// not a separately maintained imitation of its dependency declaration. Only
// unrelated PKI blocks are excluded; this test cannot contact a live appliance.
func TestTrustSettingsAcceptanceFixtureCoreLookupAfterUpdate(t *testing.T) {
	binary, err := exec.LookPath("terraform")
	if err != nil {
		t.Fatal("container Terraform Core is required", err)
	}
	t.Setenv("TF_ACC_TERRAFORM_PATH", binary)
	fixture, err := exec.Command("python3", "-c", `import importlib.util
s=importlib.util.spec_from_file_location('acceptance','../../../test/scripts/trust_lifecycle_acceptance.py')
m=importlib.util.module_from_spec(s); s.loader.exec_module(m)
print(m.resources('fixture','fixture-certificate','fixture-private-key'))`).Output()
	if err != nil {
		t.Fatal("generate actual acceptance fixture", err)
	}
	parsed, diagnostics := hclsyntax.ParseConfig(fixture, "acceptance.tf", hcl.InitialPos)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics.Error())
	}
	var settings []byte
	for _, block := range parsed.Body.(*hclsyntax.Body).Blocks {
		if (block.Type == "resource" || block.Type == "data") && len(block.Labels) == 2 && block.Labels[0] == settingsType {
			settings = append(settings, block.Range().SliceBytes(fixture)...)
			settings = append(settings, '\n')
		}
	}
	actual, diagnostics := hclwrite.ParseConfig(settings, "settings.tf", hcl.InitialPos)
	if diagnostics.HasErrors() || len(actual.Body().Blocks()) != 2 {
		t.Fatalf("expected actual managed and lookup settings blocks: %s", diagnostics.Error())
	}
	// Empty single-line resource blocks need a newline before adding intent.
	for _, block := range actual.Body().Blocks() {
		if block.Type() == "resource" {
			block.Body().AppendNewline()
		}
	}
	a := &settingsAPI{}
	h := httptest.NewServer(a)
	defer h.Close()
	config := func(enabled *bool) string {
		for _, block := range actual.Body().Blocks() {
			if block.Type() == "resource" {
				if enabled == nil {
					block.Body().RemoveAttribute("store_intermediate_certs")
				} else {
					block.Body().SetAttributeValue("store_intermediate_certs", cty.BoolVal(*enabled))
				}
			}
		}
		return fmt.Sprintf(`provider "opnsense" {
 uri = %q
 api_key = "fixture-key"
 api_secret = "fixture-secret"
}
import {
 to = opnsense_trust_settings.observed
 id = "trust_settings"
}
%s`, h.URL, actual.Bytes())
	}
	factories := map[string]func() (tfprotov6.ProviderServer, error){"opnsense": providerserver.NewProtocol6WithError(provider.New("offline-trust-fixture-test")())}
	var steps []resource.TestStep
	for _, enabled := range []bool{false, true, false, true} {
		configuration := config(&enabled)
		steps = append(steps, resource.TestStep{
			Config: configuration,
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("opnsense_trust_settings.observed", "store_intermediate_certs", fmt.Sprint(enabled)),
				resource.TestCheckResourceAttrPair("opnsense_trust_settings.observed", "store_intermediate_certs", "data.opnsense_trust_settings.observed", "store_intermediate_certs"),
			),
		}, resource.TestStep{Config: configuration, PlanOnly: true})
	}
	omitted := config(nil)
	steps = append(steps, resource.TestStep{
		Config: omitted,
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("opnsense_trust_settings.observed", "store_intermediate_certs", "true"),
			resource.TestCheckResourceAttr("data.opnsense_trust_settings.observed", "store_intermediate_certs", "true"),
		),
	}, resource.TestStep{Config: omitted, PlanOnly: true})
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: factories, Steps: steps})
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.saves != 4 || a.activations != 4 {
		t.Fatalf("expected four real loopback settings transitions only, got %d saves / %d activations", a.saves, a.activations)
	}
}
