// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package openvpn_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccOpenVPNClientOverwrite_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_openvpn_client_overwrite", opnsense.ReqOpts{GetEndpoint: "/api/openvpn/client_overwrites/get", Monad: "overwrite"}),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenVPNClientOverwriteConfig("client.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_openvpn_client_overwrite.test", "id"),
					resource.TestCheckResourceAttr("opnsense_openvpn_client_overwrite.test", "common_name", "client.example.com"),
				),
			},
			{ResourceName: "opnsense_openvpn_client_overwrite.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func testAccOpenVPNClientOverwriteConfig(cn string) string {
	return fmt.Sprintf(`
resource "opnsense_openvpn_client_overwrite" "test" {
  common_name     = %[1]q
  description     = "tf-acc-test"
  tunnel_network  = "10.10.8.100/32"
  remote_networks = ["192.168.50.0/24"]
}
`, cn)
}
