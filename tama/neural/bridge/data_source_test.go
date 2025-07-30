// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bridge_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccSpaceBridgeDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceBridgeDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tama_space_bridge.test", "space_id", "tama_space.test_space", "id"),
					resource.TestCheckResourceAttrPair("data.tama_space_bridge.test", "target_space_id", "tama_space.test_target_space", "id"),
					resource.TestCheckResourceAttrSet("data.tama_space_bridge.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_space_bridge.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccSpaceBridgeDataSource_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceBridgeDataSourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// First bridge data source
					resource.TestCheckResourceAttrPair("data.tama_space_bridge.test1", "space_id", "tama_space.test_space", "id"),
					resource.TestCheckResourceAttrPair("data.tama_space_bridge.test1", "target_space_id", "tama_space.test_target_space1", "id"),
					resource.TestCheckResourceAttrSet("data.tama_space_bridge.test1", "id"),
					// Second bridge data source
					resource.TestCheckResourceAttrPair("data.tama_space_bridge.test2", "space_id", "tama_space.test_space", "id"),
					resource.TestCheckResourceAttrPair("data.tama_space_bridge.test2", "target_space_id", "tama_space.test_target_space2", "id"),
					resource.TestCheckResourceAttrSet("data.tama_space_bridge.test2", "id"),
				),
			},
		},
	})
}

func TestAccSpaceBridgeDataSource_DifferentSpaceTypes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceBridgeDataSourceConfigDifferentTypes(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tama_space_bridge.test", "space_id", "tama_space.root_space", "id"),
					resource.TestCheckResourceAttrPair("data.tama_space_bridge.test", "target_space_id", "tama_space.component_space", "id"),
					resource.TestCheckResourceAttrSet("data.tama_space_bridge.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_space_bridge.test", "provision_state"),
				),
			},
		},
	})
}

func testAccSpaceBridgeDataSourceConfig() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-bridge-ds-%d"
  type = "root"
}

resource "tama_space" "test_target_space" {
  name = "test-target-space-for-bridge-ds-%d"
  type = "component"
}

resource "tama_space_bridge" "test" {
  space_id        = tama_space.test_space.id
  target_space_id = tama_space.test_target_space.id
}

data "tama_space_bridge" "test" {
  id = tama_space_bridge.test.id
}
`, timestamp, timestamp)
}

func testAccSpaceBridgeDataSourceConfigMultiple() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-bridge-ds-multiple-%d"
  type = "root"
}

resource "tama_space" "test_target_space1" {
  name = "test-target-space1-for-bridge-ds-%d"
  type = "component"
}

resource "tama_space" "test_target_space2" {
  name = "test-target-space2-for-bridge-ds-%d"
  type = "component"
}

resource "tama_space_bridge" "test1" {
  space_id        = tama_space.test_space.id
  target_space_id = tama_space.test_target_space1.id
}

resource "tama_space_bridge" "test2" {
  space_id        = tama_space.test_space.id
  target_space_id = tama_space.test_target_space2.id
}

data "tama_space_bridge" "test1" {
  id = tama_space_bridge.test1.id
}

data "tama_space_bridge" "test2" {
  id = tama_space_bridge.test2.id
}
`, timestamp, timestamp, timestamp)
}

func testAccSpaceBridgeDataSourceConfigDifferentTypes() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "root_space" {
  name = "test-root-space-for-bridge-ds-%d"
  type = "root"
}

resource "tama_space" "component_space" {
  name = "test-component-space-for-bridge-ds-%d"
  type = "component"
}

resource "tama_space_bridge" "test" {
  space_id        = tama_space.root_space.id
  target_space_id = tama_space.component_space.id
}

data "tama_space_bridge" "test" {
  id = tama_space_bridge.test.id
}
`, timestamp, timestamp)
}
