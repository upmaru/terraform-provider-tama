// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bridge_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccSpaceBridgeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSpaceBridgeResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("tama_space_bridge.test", "space_id", "tama_space.test_space", "id"),
					resource.TestCheckResourceAttrPair("tama_space_bridge.test", "target_space_id", "tama_space.test_target_space", "id"),
					resource.TestCheckResourceAttrSet("tama_space_bridge.test", "id"),
					resource.TestCheckResourceAttrSet("tama_space_bridge.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_space_bridge.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccSpaceBridgeResourceConfigUpdate(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("tama_space_bridge.test", "space_id", "tama_space.test_space", "id"),
					resource.TestCheckResourceAttrPair("tama_space_bridge.test", "target_space_id", "tama_space.test_new_target_space", "id"),
					resource.TestCheckResourceAttrSet("tama_space_bridge.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccSpaceBridgeResource_InvalidSpaceId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSpaceBridgeResourceConfigInvalidSpace(),
				ExpectError: regexp.MustCompile("Unable to create bridge"),
			},
		},
	})
}

func TestAccSpaceBridgeResource_SameSpaces(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceBridgeResourceConfigSameSpaces(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("tama_space_bridge.test", "space_id", "tama_space.test_space", "id"),
					resource.TestCheckResourceAttrPair("tama_space_bridge.test", "target_space_id", "tama_space.test_space", "id"),
					resource.TestCheckResourceAttrSet("tama_space_bridge.test", "id"),
					resource.TestCheckResourceAttrSet("tama_space_bridge.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccSpaceBridgeResource_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceBridgeResourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// First bridge
					resource.TestCheckResourceAttrPair("tama_space_bridge.test1", "space_id", "tama_space.test_space", "id"),
					resource.TestCheckResourceAttrPair("tama_space_bridge.test1", "target_space_id", "tama_space.test_target_space1", "id"),
					resource.TestCheckResourceAttrSet("tama_space_bridge.test1", "id"),
					// Second bridge
					resource.TestCheckResourceAttrPair("tama_space_bridge.test2", "space_id", "tama_space.test_space", "id"),
					resource.TestCheckResourceAttrPair("tama_space_bridge.test2", "target_space_id", "tama_space.test_target_space2", "id"),
					resource.TestCheckResourceAttrSet("tama_space_bridge.test2", "id"),
				),
			},
		},
	})
}

func TestAccSpaceBridgeResource_DisappearResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceBridgeResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("tama_space_bridge.test", "space_id", "tama_space.test_space", "id"),
					resource.TestCheckResourceAttrPair("tama_space_bridge.test", "target_space_id", "tama_space.test_target_space", "id"),
					resource.TestCheckResourceAttrSet("tama_space_bridge.test", "id"),
					resource.TestCheckResourceAttrSet("tama_space_bridge.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccSpaceBridgeResource_DifferentSpaceTypes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceBridgeResourceConfigDifferentTypes(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("tama_space_bridge.test", "space_id", "tama_space.root_space", "id"),
					resource.TestCheckResourceAttrPair("tama_space_bridge.test", "target_space_id", "tama_space.component_space", "id"),
					resource.TestCheckResourceAttrSet("tama_space_bridge.test", "id"),
					resource.TestCheckResourceAttrSet("tama_space_bridge.test", "provision_state"),
				),
			},
		},
	})
}

func testAccSpaceBridgeResourceConfig() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-bridge-%d"
  type = "root"
}

resource "tama_space" "test_target_space" {
  name = "test-target-space-for-bridge-%d"
  type = "component"
}

resource "tama_space_bridge" "test" {
  space_id        = tama_space.test_space.id
  target_space_id = tama_space.test_target_space.id
}
`, timestamp, timestamp)
}

func testAccSpaceBridgeResourceConfigUpdate() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-bridge-%d"
  type = "root"
}

resource "tama_space" "test_target_space" {
  name = "test-target-space-for-bridge-%d"
  type = "component"
}

resource "tama_space" "test_new_target_space" {
  name = "test-new-target-space-for-bridge-%d"
  type = "component"
}

resource "tama_space_bridge" "test" {
  space_id        = tama_space.test_space.id
  target_space_id = tama_space.test_new_target_space.id
}
`, timestamp, timestamp, timestamp)
}

func testAccSpaceBridgeResourceConfigInvalidSpace() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-bridge-%d"
  type = "root"
}

resource "tama_space_bridge" "test" {
  space_id        = tama_space.test_space.id
  target_space_id = "invalid-space-id"
}
`, timestamp)
}

func testAccSpaceBridgeResourceConfigSameSpaces() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-bridge-%d"
  type = "root"
}

resource "tama_space_bridge" "test" {
  space_id        = tama_space.test_space.id
  target_space_id = tama_space.test_space.id
}
`, timestamp)
}

func testAccSpaceBridgeResourceConfigMultiple() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-multiple-bridges-%d"
  type = "root"
}

resource "tama_space" "test_target_space1" {
  name = "test-target-space1-for-bridge-%d"
  type = "component"
}

resource "tama_space" "test_target_space2" {
  name = "test-target-space2-for-bridge-%d"
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
`, timestamp, timestamp, timestamp)
}

func testAccSpaceBridgeResourceConfigDifferentTypes() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "root_space" {
  name = "test-root-space-for-bridge-%d"
  type = "root"
}

resource "tama_space" "component_space" {
  name = "test-component-space-for-bridge-%d"
  type = "component"
}

resource "tama_space_bridge" "test" {
  space_id        = tama_space.root_space.id
  target_space_id = tama_space.component_space.id
}
`, timestamp, timestamp)
}
