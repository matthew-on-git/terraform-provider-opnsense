// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccTrustCA_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_trust_ca", opnsense.ReqOpts{GetEndpoint: "/api/trust/ca/get", Monad: "ca"}),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustCAConfig("tf-acc-ca"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_trust_ca.test", "id"),
					resource.TestCheckResourceAttrSet("opnsense_trust_ca.test", "refid"),
					resource.TestCheckResourceAttrSet("opnsense_trust_ca.test", "certificate"),
				),
			},
			{ResourceName: "opnsense_trust_ca.test", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"lifetime"}},
			{
				Config: testAccTrustCAConfig("tf-acc-ca-renamed"),
				Check:  resource.TestCheckResourceAttr("opnsense_trust_ca.test", "description", "tf-acc-ca-renamed"),
			},
		},
	})
}

func testAccTrustCAConfig(descr string) string {
	return fmt.Sprintf(`
resource "opnsense_trust_ca" "test" {
  description = %[1]q
  common_name = "tf-acc-ca"
  country     = "US"
}
`, descr)
}
