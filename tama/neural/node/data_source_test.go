// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package node_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccNodeDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeDataSourceConfig("test-node"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_node.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "type"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "on"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "space_id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "class_id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "chain_id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "provision_state"),
					// Verify data source matches resource
					resource.TestCheckResourceAttrPair("data.tama_node.test", "id", "tama_node.test", "id"),
					resource.TestCheckResourceAttrPair("data.tama_node.test", "type", "tama_node.test", "type"),
					resource.TestCheckResourceAttrPair("data.tama_node.test", "on", "tama_node.test", "on"),
					resource.TestCheckResourceAttrPair("data.tama_node.test", "space_id", "tama_node.test", "space_id"),
					resource.TestCheckResourceAttrPair("data.tama_node.test", "class_id", "tama_node.test", "class_id"),
					resource.TestCheckResourceAttrPair("data.tama_node.test", "chain_id", "tama_node.test", "chain_id"),
				),
			},
		},
	})
}

func TestAccNodeDataSource_MultipleNodes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeDataSourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check first node data source
					resource.TestCheckResourceAttrSet("data.tama_node.test_reactive", "id"),
					resource.TestCheckResourceAttr("data.tama_node.test_reactive", "type", "reactive"),
					resource.TestCheckResourceAttr("data.tama_node.test_reactive", "on", "processing"),
					resource.TestCheckResourceAttrSet("data.tama_node.test_reactive", "space_id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test_reactive", "class_id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test_reactive", "chain_id"),

					// Check second node data source
					resource.TestCheckResourceAttrSet("data.tama_node.test_scheduled", "id"),
					resource.TestCheckResourceAttr("data.tama_node.test_scheduled", "type", "scheduled"),
					resource.TestCheckResourceAttr("data.tama_node.test_scheduled", "on", "processed"),
					resource.TestCheckResourceAttrSet("data.tama_node.test_scheduled", "space_id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test_scheduled", "class_id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test_scheduled", "chain_id"),

					// Check third node data source
					resource.TestCheckResourceAttrSet("data.tama_node.test_explicit", "id"),
					resource.TestCheckResourceAttr("data.tama_node.test_explicit", "type", "explicit"),
					resource.TestCheckResourceAttr("data.tama_node.test_explicit", "on", "processed"),
					resource.TestCheckResourceAttrSet("data.tama_node.test_explicit", "space_id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test_explicit", "class_id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test_explicit", "chain_id"),

					// Verify data sources match resources
					resource.TestCheckResourceAttrPair("data.tama_node.test_reactive", "id", "tama_node.reactive", "id"),
					resource.TestCheckResourceAttrPair("data.tama_node.test_scheduled", "id", "tama_node.scheduled", "id"),
					resource.TestCheckResourceAttrPair("data.tama_node.test_explicit", "id", "tama_node.explicit", "id"),
				),
			},
		},
	})
}

func TestAccNodeDataSource_VerifyAllAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeDataSourceConfig("verify-attrs"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify all required attributes are present
					resource.TestCheckResourceAttrSet("data.tama_node.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "type"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "on"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "space_id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "class_id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "chain_id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "provision_state"),

					// Verify that provision_state is not empty
					resource.TestCheckNoResourceAttr("data.tama_node.test", "provision_state.#"),
				),
			},
		},
	})
}

func TestAccNodeDataSource_StateVerification(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeDataSourceConfig("state-test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_node.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "type"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "on"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "space_id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "class_id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "chain_id"),
				),
			},
		},
	})
}

func TestAccNodeDataSource_DifferentTypes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeDataSourceConfigDifferentTypes(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check reactive node
					resource.TestCheckResourceAttrSet("data.tama_node.test_reactive", "id"),
					resource.TestCheckResourceAttr("data.tama_node.test_reactive", "type", "reactive"),
					resource.TestCheckResourceAttrSet("data.tama_node.test_reactive", "space_id"),

					// Check scheduled node
					resource.TestCheckResourceAttrSet("data.tama_node.test_scheduled", "id"),
					resource.TestCheckResourceAttr("data.tama_node.test_scheduled", "type", "scheduled"),
					resource.TestCheckResourceAttrSet("data.tama_node.test_scheduled", "space_id"),

					// Check explicit node
					resource.TestCheckResourceAttrSet("data.tama_node.test_explicit", "id"),
					resource.TestCheckResourceAttr("data.tama_node.test_explicit", "type", "explicit"),
					resource.TestCheckResourceAttrSet("data.tama_node.test_explicit", "space_id"),

					// Verify they have the same space_id
					resource.TestCheckResourceAttrPair("data.tama_node.test_reactive", "space_id", "tama_space.test", "id"),
					resource.TestCheckResourceAttrPair("data.tama_node.test_scheduled", "space_id", "tama_space.test", "id"),
					resource.TestCheckResourceAttrPair("data.tama_node.test_explicit", "space_id", "tama_space.test", "id"),
				),
			},
		},
	})
}

func TestAccNodeDataSource_WithDefaultOn(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeDataSourceConfigWithDefaultOn(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_node.test", "id"),
					resource.TestCheckResourceAttr("data.tama_node.test", "type", "explicit"),
					resource.TestCheckResourceAttr("data.tama_node.test", "on", "processing"), // Should be default value
					resource.TestCheckResourceAttrSet("data.tama_node.test", "space_id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "class_id"),
					resource.TestCheckResourceAttrSet("data.tama_node.test", "chain_id"),
					// Verify data source matches resource
					resource.TestCheckResourceAttrPair("data.tama_node.test", "on", "tama_node.test", "on"),
				),
			},
		},
	})
}

func testAccNodeDataSourceConfig(name string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s-%d"
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "action-call"
    description = "An action call is a request to execute an action."
    type = "object"
    properties = {
      tool_id = {
        description = "The ID of the tool to execute"
        type        = "string"
      }
      parameters = {
        description = "The parameters to pass to the action"
        type        = "object"
      }
    }
    required = ["tool_id", "parameters"]
  })
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
}

resource "tama_node" "test" {
  space_id = tama_space.test.id
  class_id = tama_class.test.id
  chain_id = tama_chain.test.id

  type = "reactive"
  on   = "processing"
}

data "tama_node" "test" {
  id = tama_node.test.id
}
`, name, timestamp)
}

func testAccNodeDataSourceConfigMultiple() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "multi-node-data-test-%d"
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "universal-schema"
    description = "Universal schema for all node types"
    type = "object"
    properties = {
      data = {
        type = "object"
        description = "Processing data"
      }
      metadata = {
        type = "object"
        description = "Additional metadata"
      }
    }
    required = ["data"]
  })
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Universal Processing Chain"
}

resource "tama_node" "reactive" {
  space_id = tama_space.test.id
  class_id = tama_class.test.id
  chain_id = tama_chain.test.id

  type = "reactive"
  on   = "processing"
}

resource "tama_node" "scheduled" {
  space_id = tama_space.test.id
  class_id = tama_class.test.id
  chain_id = tama_chain.test.id

  type = "scheduled"
  on   = "processed"
}

resource "tama_node" "explicit" {
  space_id = tama_space.test.id
  class_id = tama_class.test.id
  chain_id = tama_chain.test.id

  type = "explicit"
  on   = "processed"
}

data "tama_node" "test_reactive" {
  id = tama_node.reactive.id
}

data "tama_node" "test_scheduled" {
  id = tama_node.scheduled.id
}

data "tama_node" "test_explicit" {
  id = tama_node.explicit.id
}
`, timestamp)
}

func testAccNodeDataSourceConfigDifferentTypes() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "different-types-test-%d"
  type = "root"
}

resource "tama_class" "test1" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "reactive-schema"
    description = "Schema for reactive processing"
    type = "object"
    properties = {
      trigger = {
        type = "string"
        description = "Trigger event"
      }
    }
    required = ["trigger"]
  })
}

resource "tama_class" "test2" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "scheduled-schema"
    description = "Schema for scheduled processing"
    type = "object"
    properties = {
      schedule = {
        type = "string"
        description = "Schedule pattern"
      }
    }
    required = ["schedule"]
  })
}

resource "tama_class" "test3" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "explicit-schema"
    description = "Schema for explicit processing"
    type = "object"
    properties = {
      action = {
        type = "string"
        description = "Explicit action"
      }
    }
    required = ["action"]
  })
}

resource "tama_chain" "test1" {
  space_id = tama_space.test.id
  name     = "Reactive Chain"
}

resource "tama_chain" "test2" {
  space_id = tama_space.test.id
  name     = "Scheduled Chain"
}

resource "tama_chain" "test3" {
  space_id = tama_space.test.id
  name     = "Explicit Chain"
}

resource "tama_node" "reactive" {
  space_id = tama_space.test.id
  class_id = tama_class.test1.id
  chain_id = tama_chain.test1.id

  type = "reactive"
  on   = "processing"
}

resource "tama_node" "scheduled" {
  space_id = tama_space.test.id
  class_id = tama_class.test2.id
  chain_id = tama_chain.test2.id

  type = "scheduled"
  on   = "processed"
}

resource "tama_node" "explicit" {
  space_id = tama_space.test.id
  class_id = tama_class.test3.id
  chain_id = tama_chain.test3.id

  type = "explicit"
  on   = "processed"
}

data "tama_node" "test_reactive" {
  id = tama_node.reactive.id
}

data "tama_node" "test_scheduled" {
  id = tama_node.scheduled.id
}

data "tama_node" "test_explicit" {
  id = tama_node.explicit.id
}
`, timestamp)
}

func testAccNodeDataSourceConfigWithDefaultOn() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "default-on-test-%d"
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "default-schema"
    description = "Schema for testing default 'on' value"
    type = "object"
    properties = {
      value = {
        type = "string"
        description = "Test value"
      }
    }
    required = ["value"]
  })
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Default On Chain"
}

resource "tama_node" "test" {
  space_id = tama_space.test.id
  class_id = tama_class.test.id
  chain_id = tama_chain.test.id

  type = "explicit"
  # on is not specified, should default to "processing"
}

data "tama_node" "test" {
  id = tama_node.test.id
}
`, timestamp)
}
