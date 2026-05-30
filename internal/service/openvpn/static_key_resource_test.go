// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package openvpn_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccOpenVPNStaticKey_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_openvpn_static_key", opnsense.ReqOpts{GetEndpoint: "/api/openvpn/instances/get_static_key", Monad: "static_key"}),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenVPNStaticKeyConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_openvpn_static_key.test", "id"),
					resource.TestCheckResourceAttr("opnsense_openvpn_static_key.test", "mode", "crypt"),
				),
			},
			{ResourceName: "opnsense_openvpn_static_key.test", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"key"}},
		},
	})
}

const testAccOpenVPNStaticKeyConfig = `
resource "opnsense_openvpn_static_key" "test" {
  mode        = "crypt"
  description = "tf-acc-test"
  key         = <<-EOT
-----BEGIN OpenVPN Static key V1-----
00000000000000000000000000000000
11111111111111111111111111111111
22222222222222222222222222222222
33333333333333333333333333333333
-----END OpenVPN Static key V1-----
EOT
}
`
