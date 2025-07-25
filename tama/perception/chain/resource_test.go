// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chain_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccChainResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccChainResourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_chain.test", "id"),
					resource.TestCheckResourceAttrSet("tama_chain.test", "space_id"),
					resource.TestCheckResourceAttr("tama_chain.test", "name", "Identity Validation"),
					resource.TestCheckResourceAttrSet("tama_chain.test", "slug"),
					resource.TestCheckResourceAttrSet("tama_chain.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_chain.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccChainResourceConfigUpdate(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_chain.test", "id"),
					resource.TestCheckResourceAttrSet("tama_chain.test", "space_id"),
					resource.TestCheckResourceAttr("tama_chain.test", "name", "Updated Identity Validation"),
					resource.TestCheckResourceAttrSet("tama_chain.test", "slug"),
					resource.TestCheckResourceAttrSet("tama_chain.test", "provision_state"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccChainResourceConfig(spaceName string) string {
	return fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Identity Validation"
}
`, spaceName)
}

func testAccChainResourceConfigUpdate(spaceName string) string {
	return fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Updated Identity Validation"
}
`, spaceName)
}
