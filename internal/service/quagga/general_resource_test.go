// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

func TestAccQuaggaGeneral_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		// Singleton: no CheckDestroy — the appliance config persists after destroy.
		Steps: []resource.TestStep{
			{
				Config: testAccQuaggaGeneralConfig(true, "datacenter"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_quagga_general.test", "id", "general"),
					resource.TestCheckResourceAttr("opnsense_quagga_general.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_quagga_general.test", "profile", "datacenter"),
				),
			},
			{ResourceName: "opnsense_quagga_general.test", ImportState: true, ImportStateId: "general", ImportStateVerify: true},
			{
				Config: testAccQuaggaGeneralConfig(false, "traditional"),
				Check:  resource.TestCheckResourceAttr("opnsense_quagga_general.test", "profile", "traditional"),
			},
		},
	})
}

func testAccQuaggaGeneralConfig(enabled bool, profile string) string {
	return `
resource "opnsense_quagga_general" "test" {
  enabled = ` + boolStr(enabled) + `
  profile = "` + profile + `"
}
`
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
