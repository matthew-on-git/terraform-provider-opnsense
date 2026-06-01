// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

func TestAccSystemInfoDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: `data "opnsense_system_info" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.opnsense_system_info.test", "id", "system_info"),
					// Firmware version looks like "25.7".
					resource.TestMatchResourceAttr("data.opnsense_system_info.test", "version", regexp.MustCompile(`^\d+\.\d+`)),
					resource.TestCheckResourceAttrSet("data.opnsense_system_info.test", "plugins.#"),
				),
			},
		},
	})
}
