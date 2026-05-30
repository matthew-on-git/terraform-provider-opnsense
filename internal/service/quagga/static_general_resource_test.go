// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

func TestAccQuaggaStatic_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: `
resource "opnsense_quagga_static" "test" {
  enabled = true
}
`,
				Check: resource.TestCheckResourceAttr("opnsense_quagga_static.test", "id", "static"),
			},
			{ResourceName: "opnsense_quagga_static.test", ImportState: true, ImportStateId: "static", ImportStateVerify: true},
		},
	})
}
