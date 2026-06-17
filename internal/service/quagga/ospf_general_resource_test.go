// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

func TestAccQuaggaOSPFGeneral_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { preCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: `
resource "opnsense_quagga_ospf_general" "test" {
  enabled   = true
  router_id = "10.0.0.1"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_quagga_ospf_general.test", "id", "ospf"),
					resource.TestCheckResourceAttr("opnsense_quagga_ospf_general.test", "router_id", "10.0.0.1"),
				),
			},
			{ResourceName: "opnsense_quagga_ospf_general.test", ImportState: true, ImportStateId: "ospf", ImportStateVerify: true},
		},
	})
}
