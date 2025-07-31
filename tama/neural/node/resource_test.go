// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package node_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccNodeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNodeResourceConfig(fmt.Sprintf("test-node-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_node.test", "id"),
					resource.TestCheckResourceAttr("tama_node.test", "type", "reactive"),
					resource.TestCheckResourceAttr("tama_node.test", "on", "processing"),
					resource.TestCheckResourceAttrSet("tama_node.test", "space_id"),
					resource.TestCheckResourceAttrSet("tama_node.test", "class_id"),
					resource.TestCheckResourceAttrSet("tama_node.test", "chain_id"),
					resource.TestCheckResourceAttrSet("tama_node.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_node.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccNodeResourceConfigUpdate(fmt.Sprintf("test-node-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_node.test", "id"),
					resource.TestCheckResourceAttr("tama_node.test", "type", "scheduled"),
					resource.TestCheckResourceAttr("tama_node.test", "on", "processing"),
					resource.TestCheckResourceAttrSet("tama_node.test", "space_id"),
					resource.TestCheckResourceAttrSet("tama_node.test", "class_id"),
					resource.TestCheckResourceAttrSet("tama_node.test", "chain_id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccNodeResource_DefaultOn(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeResourceConfigDefaultOn(fmt.Sprintf("test-node-default-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_node.test", "id"),
					resource.TestCheckResourceAttr("tama_node.test", "type", "explicit"),
					resource.TestCheckResourceAttr("tama_node.test", "on", "processing"), // Should default to processing
					resource.TestCheckResourceAttrSet("tama_node.test", "space_id"),
					resource.TestCheckResourceAttrSet("tama_node.test", "class_id"),
					resource.TestCheckResourceAttrSet("tama_node.test", "chain_id"),
				),
			},
		},
	})
}

func TestAccNodeResource_InvalidType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNodeResourceConfigInvalidType(fmt.Sprintf("test-node-%d", time.Now().UnixNano())),
				ExpectError: regexp.MustCompile("Attribute type value must be one of"),
			},
		},
	})
}

func TestAccNodeResource_InvalidOn(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNodeResourceConfigInvalidOn(fmt.Sprintf("test-node-%d", time.Now().UnixNano())),
				ExpectError: regexp.MustCompile("Attribute on value must be one of"),
			},
		},
	})
}

func TestAccNodeResource_MultipleNodes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeResourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// First node
					resource.TestCheckResourceAttrSet("tama_node.test1", "id"),
					resource.TestCheckResourceAttr("tama_node.test1", "type", "reactive"),
					resource.TestCheckResourceAttr("tama_node.test1", "on", "processing"),
					resource.TestCheckResourceAttrSet("tama_node.test1", "space_id"),
					resource.TestCheckResourceAttrSet("tama_node.test1", "class_id"),
					resource.TestCheckResourceAttrSet("tama_node.test1", "chain_id"),

					// Second node
					resource.TestCheckResourceAttrSet("tama_node.test2", "id"),
					resource.TestCheckResourceAttr("tama_node.test2", "type", "scheduled"),
					resource.TestCheckResourceAttr("tama_node.test2", "on", "processing"),
					resource.TestCheckResourceAttrSet("tama_node.test2", "space_id"),
					resource.TestCheckResourceAttrSet("tama_node.test2", "class_id"),
					resource.TestCheckResourceAttrSet("tama_node.test2", "chain_id"),

					// Third node
					resource.TestCheckResourceAttrSet("tama_node.test3", "id"),
					resource.TestCheckResourceAttr("tama_node.test3", "type", "explicit"),
					resource.TestCheckResourceAttr("tama_node.test3", "on", "processing"), // default
					resource.TestCheckResourceAttrSet("tama_node.test3", "space_id"),
					resource.TestCheckResourceAttrSet("tama_node.test3", "class_id"),
					resource.TestCheckResourceAttrSet("tama_node.test3", "chain_id"),
				),
			},
		},
	})
}

func TestAccNodeResource_AllTypes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeResourceConfigAllTypes(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Reactive node
					resource.TestCheckResourceAttrSet("tama_node.reactive", "id"),
					resource.TestCheckResourceAttr("tama_node.reactive", "type", "reactive"),
					resource.TestCheckResourceAttr("tama_node.reactive", "on", "processing"),

					// Scheduled node
					resource.TestCheckResourceAttrSet("tama_node.scheduled", "id"),
					resource.TestCheckResourceAttr("tama_node.scheduled", "type", "scheduled"),
					resource.TestCheckResourceAttr("tama_node.scheduled", "on", "processed"),

					// Explicit node
					resource.TestCheckResourceAttrSet("tama_node.explicit", "id"),
					resource.TestCheckResourceAttr("tama_node.explicit", "type", "explicit"),
					resource.TestCheckResourceAttr("tama_node.explicit", "on", "processed"),
				),
			},
		},
	})
}

func testAccNodeResourceConfig(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
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
`, spaceName)
}

func testAccNodeResourceConfigUpdate(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
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

  type = "scheduled"
  on   = "processing"
}
`, spaceName)
}

func testAccNodeResourceConfigDefaultOn(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "validation-schema"
    description = "A schema for validation results"
    type = "object"
    properties = {
      valid = {
        description = "Whether the input is valid"
        type        = "boolean"
      }
      confidence = {
        description = "Confidence score"
        type        = "number"
      }
    }
    required = ["valid"]
  })
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Default Processing Chain"
}

resource "tama_node" "test" {
  space_id = tama_space.test.id
  class_id = tama_class.test.id
  chain_id = tama_chain.test.id

  type = "explicit"
  # on is not specified, should default to "processing"
}
`, spaceName)
}

func testAccNodeResourceConfigInvalidType(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "test-schema"
    description = "A test schema"
    type = "object"
    properties = {
      test_field = {
        type = "string"
      }
    }
    required = ["test_field"]
  })
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Chain"
}

resource "tama_node" "test" {
  space_id = tama_space.test.id
  class_id = tama_class.test.id
  chain_id = tama_chain.test.id

  type = "invalid_type"
  on   = "processing"
}
`, spaceName)
}

func testAccNodeResourceConfigMultiple() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "multi-node-test-%d"
  type = "root"
}

resource "tama_class" "test1" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "processing-schema"
    description = "Schema for processing operations"
    type = "object"
    properties = {
      operation = {
        type = "string"
        description = "Operation to perform"
      }
    }
    required = ["operation"]
  })
}

resource "tama_class" "test2" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "validation-schema"
    description = "Schema for validation results"
    type = "object"
    properties = {
      valid = {
        type = "boolean"
        description = "Validation result"
      }
    }
    required = ["valid"]
  })
}

resource "tama_chain" "test1" {
  space_id = tama_space.test.id
  name     = "Processing Chain"
}

resource "tama_chain" "test2" {
  space_id = tama_space.test.id
  name     = "Validation Chain"
}

resource "tama_node" "test1" {
  space_id = tama_space.test.id
  class_id = tama_class.test1.id
  chain_id = tama_chain.test1.id

  type = "reactive"
  on   = "processing"
}

resource "tama_node" "test2" {
  space_id = tama_space.test.id
  class_id = tama_class.test2.id
  chain_id = tama_chain.test2.id

  type = "scheduled"
  on   = "processing"
}

resource "tama_node" "test3" {
  space_id = tama_space.test.id
  class_id = tama_class.test1.id  # reuse first class
  chain_id = tama_chain.test1.id  # reuse first chain

  type = "explicit"
  # on defaults to "processing"
}
`, timestamp)
}

func testAccNodeResourceConfigInvalidOn(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "test-schema"
    description = "A test schema"
    type = "object"
    properties = {
      test_field = {
        type = "string"
      }
    }
    required = ["test_field"]
  })
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Chain"
}

resource "tama_node" "test" {
  space_id = tama_space.test.id
  class_id = tama_class.test.id
  chain_id = tama_chain.test.id

  type = "reactive"
  on   = "invalid_on_value"
}
`, spaceName)
}

func testAccNodeResourceConfigAllTypes() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "all-types-test-%d"
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
`, timestamp)
}
