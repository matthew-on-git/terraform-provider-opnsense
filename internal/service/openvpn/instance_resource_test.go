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
				Config: testAccOpenVPNInstanceConfig("vpn1.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_openvpn_instance.test", "id"),
					resource.TestCheckResourceAttrSet("opnsense_openvpn_instance.test", "vpnid"),
					resource.TestCheckResourceAttr("opnsense_openvpn_instance.test", "role", "client"),
					resource.TestCheckResourceAttr("opnsense_openvpn_instance.test", "remote", "vpn1.example.com"),
				),
			},
			{ResourceName: "opnsense_openvpn_instance.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccOpenVPNInstanceConfig("vpn2.example.com"),
				Check:  resource.TestCheckResourceAttr("opnsense_openvpn_instance.test", "remote", "vpn2.example.com"),
			},
		},
	})
}

func testAccOpenVPNInstanceConfig(remote string) string {
	return fmt.Sprintf(`
resource "opnsense_trust_ca" "test" {
  description = "tf-acc-ovpn-ca"
  common_name = "tf-acc-ovpn-ca"
  country     = "US"
}

resource "opnsense_openvpn_instance" "test" {
  role               = "client"
  description        = "tf-acc-test"
  protocol           = "udp"
  dev_type           = "tun"
  remote             = %[1]q
  ca                 = opnsense_trust_ca.test.refid
  verify_client_cert = "none"
  data_ciphers       = ["AES-256-GCM"]
}
`, remote)
}
