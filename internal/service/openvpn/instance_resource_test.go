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

func TestAccOpenVPNInstance_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_openvpn_instance", opnsense.ReqOpts{GetEndpoint: "/api/openvpn/instances/get", Monad: "instance"}),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenVPNInstanceConfig("10.10.8.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_openvpn_instance.test", "id"),
					resource.TestCheckResourceAttr("opnsense_openvpn_instance.test", "role", "server"),
					resource.TestCheckResourceAttr("opnsense_openvpn_instance.test", "server", "10.10.8.0/24"),
				),
			},
			{ResourceName: "opnsense_openvpn_instance.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccOpenVPNInstanceConfig("10.10.9.0/24"),
				Check:  resource.TestCheckResourceAttr("opnsense_openvpn_instance.test", "server", "10.10.9.0/24"),
			},
		},
	})
}

func testAccOpenVPNInstanceConfig(network string) string {
	return fmt.Sprintf(`
resource "opnsense_openvpn_instance" "test" {
  role         = "server"
  description  = "tf-acc-test"
  protocol     = "udp"
  dev_type     = "tun"
  server       = %[1]q
  data_ciphers = ["AES-256-GCM"]
}
`, network)
}
