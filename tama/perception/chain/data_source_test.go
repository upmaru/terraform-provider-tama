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

func TestAccChainDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccChainDataSourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_chain.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_chain.test", "space_id"),
					resource.TestCheckResourceAttr("data.tama_chain.test", "name", "Identity Validation"),
					resource.TestCheckResourceAttrSet("data.tama_chain.test", "slug"),
					resource.TestCheckResourceAttrSet("data.tama_chain.test", "current_state"),
				),
			},
		},
	})
}

func TestAccChainDataSource_BasicChain(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccChainDataSourceConfigBasic(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_chain.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_chain.test", "space_id"),
					resource.TestCheckResourceAttr("data.tama_chain.test", "name", "Basic Chain"),
					resource.TestCheckResourceAttrSet("data.tama_chain.test", "slug"),
					resource.TestCheckResourceAttrSet("data.tama_chain.test", "current_state"),
				),
			},
		},
	})
}

func testAccChainDataSourceConfig(spaceName string) string {
	return fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Identity Validation"
}

data "tama_chain" "test" {
  id = tama_chain.test.id
}
`, spaceName)
}

func testAccChainDataSourceConfigBasic(spaceName string) string {
	return fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Basic Chain"
}

data "tama_chain" "test" {
  id = tama_chain.test.id
}
`, spaceName)
}
