// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package acme_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccAcmeChallenge_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_acme_challenge", opnsense.ReqOpts{GetEndpoint: "/api/acmeclient/validations/get", Monad: "validation"}),
		Steps: []resource.TestStep{
			{
				Config: testAccAcmeChallengeConfig("tf_test_challenge", "http01"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_acme_challenge.test", "id"),
					resource.TestCheckResourceAttr("opnsense_acme_challenge.test", "method", "http01"),
				),
			},
			{ResourceName: "opnsense_acme_challenge.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func testAccAcmeChallengeConfig(name, method string) string {
	return fmt.Sprintf(`
resource "opnsense_acme_challenge" "test" {
  name   = %[1]q
  method = %[2]q
}
`, name, method)
}
