// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system_test

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

func TestAccSystemTunable_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.SkipIfEndpointMissing(t, "/api/core/tunables/get_item")
			acctest.SkipIfEndpointMissing(t, "/api/core/tunables/add_item")
			acctest.SkipIfEndpointMissing(t, "/api/core/tunables/set_item/00000000-0000-0000-0000-000000000000")
			acctest.SkipIfEndpointMissing(t, "/api/core/tunables/del_item/00000000-0000-0000-0000-000000000000")
			acctest.SkipIfEndpointMissing(t, "/api/core/tunables/reconfigure")
		},
		CheckDestroy: checkSystemTunableDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: testAccSystemTunableConfig("Terraform acceptance test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_system_tunable.test", "id"),
					resource.TestCheckResourceAttr("opnsense_system_tunable.test", "tunable", "kern.msgbuf_show_timestamp"),
					resource.TestCheckResourceAttr("opnsense_system_tunable.test", "value", "1"),
					resource.TestCheckResourceAttr("opnsense_system_tunable.test", "description", "Terraform acceptance test"),
				),
			},
			{
				Config:             testAccSystemTunableConfig("Terraform acceptance test"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{ResourceName: "opnsense_system_tunable.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccSystemTunableConfig("Terraform acceptance test updated"),
				Check:  resource.TestCheckResourceAttr("opnsense_system_tunable.test", "description", "Terraform acceptance test updated"),
			},
		},
	})
}

func checkSystemTunableDestroyed(t *testing.T) func(*terraform.State) error {
	t.Helper()
	return func(s *terraform.State) error {
		client := acctest.TestClient(t)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "opnsense_system_tunable" {
				continue
			}
			resp, err := client.HTTPClient().Get(client.BaseURL() + "/api/core/tunables/get_item/" + rs.Primary.ID)
			if err != nil {
				continue
			}
			body, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if readErr != nil {
				return fmt.Errorf("read tunable destroy check response: %w", readErr)
			}
			if resp.StatusCode == 404 || strings.TrimSpace(string(body)) == "[]" {
				continue
			}
			if strings.Contains(string(body), "\"sysctl\"") {
				return fmt.Errorf("system tunable %s still exists after destroy", rs.Primary.ID)
			}
		}
		return nil
	}
}

func testAccSystemTunableConfig(description string) string {
	return fmt.Sprintf(`
resource "opnsense_system_tunable" "test" {
  tunable     = "kern.msgbuf_show_timestamp"
  value       = "1"
  description = %[1]q
}
`, description)
}
