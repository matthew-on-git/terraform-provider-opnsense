// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

// TestAccPoolDataSource_basic verifies a generated data source reads back the
// same state as its corresponding resource (representative of all generated
// item data sources).
func TestAccPoolDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: `
resource "opnsense_ipsec_pool" "test" {
  name      = "tfaccdspool"
  addresses = "10.30.40.0/24"
}

data "opnsense_ipsec_pool" "test" {
  id = opnsense_ipsec_pool.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.opnsense_ipsec_pool.test", "name", "opnsense_ipsec_pool.test", "name"),
					resource.TestCheckResourceAttrPair("data.opnsense_ipsec_pool.test", "addresses", "opnsense_ipsec_pool.test", "addresses"),
					resource.TestCheckResourceAttr("data.opnsense_ipsec_pool.test", "name", "tfaccdspool"),
				),
			},
		},
	})
}
