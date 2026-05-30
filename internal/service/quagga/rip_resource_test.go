// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

func TestAccQuaggaRIP_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: `
resource "opnsense_quagga_rip" "test" {
  enabled  = true
  version  = 2
  networks = ["10.0.0.0/24"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_quagga_rip.test", "id", "rip"),
					resource.TestCheckResourceAttr("opnsense_quagga_rip.test", "version", "2"),
				),
			},
			{ResourceName: "opnsense_quagga_rip.test", ImportState: true, ImportStateId: "rip", ImportStateVerify: true},
		},
	})
}
