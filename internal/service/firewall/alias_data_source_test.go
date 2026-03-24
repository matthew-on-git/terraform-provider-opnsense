// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

func TestAccFirewallAliasDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallAliasDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.opnsense_firewall_alias.test", "id"),
					resource.TestCheckResourceAttr("data.opnsense_firewall_alias.test", "name", "tf_test_ds_alias"),
					resource.TestCheckResourceAttr("data.opnsense_firewall_alias.test", "type", "host"),
					resource.TestCheckResourceAttr("data.opnsense_firewall_alias.test", "enabled", "true"),
				),
			},
		},
	})
}

func testAccFirewallAliasDataSourceConfig() string {
	return `
resource "opnsense_firewall_alias" "test" {
  name    = "tf_test_ds_alias"
  type    = "host"
  content = ["10.0.0.1"]
}

data "opnsense_firewall_alias" "test" {
  id = opnsense_firewall_alias.test.id
}
`
}
