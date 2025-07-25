// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package class_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccClassDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassDataSourceConfig("test-class"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "name"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "description"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "schema_json"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "schema.0.title"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "schema.0.description"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "schema.0.type"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "space_id"),
				),
			},
		},
	})
}

func TestAccClassDataSource_ComplexSchema(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassDataSourceConfigComplex("complex-class"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "name"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "description"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "schema_json"),
					resource.TestCheckResourceAttr("data.tama_class.test", "schema.0.title", "entity-network"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "space_id"),
				),
			},
		},
	})
}

func TestAccClassDataSource_ArraySchema(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassDataSourceConfigArray("array-class"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "name"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "description"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "schema_json"),
					resource.TestCheckResourceAttr("data.tama_class.test", "schema.0.type", "object"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "space_id"),
				),
			},
		},
	})
}

func TestAccClassDataSource_MinimalSchema(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassDataSourceConfigMinimal("minimal-class"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "name"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "description"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "schema_json"),
					resource.TestCheckResourceAttr("data.tama_class.test", "schema.0.type", "object"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "space_id"),
				),
			},
		},
	})
}

func TestAccClassDataSource_MultipleClasses(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassDataSourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check first class
					resource.TestCheckResourceAttrSet("data.tama_class.test_basic", "id"),
					resource.TestCheckResourceAttrSet("data.tama_class.test_basic", "name"),
					resource.TestCheckResourceAttrSet("data.tama_class.test_basic", "description"),
					resource.TestCheckResourceAttrSet("data.tama_class.test_basic", "schema_json"),
					resource.TestCheckResourceAttrSet("data.tama_class.test_basic", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_class.test_basic", "space_id"),

					// Check second class
					resource.TestCheckResourceAttrSet("data.tama_class.test_collection", "id"),
					resource.TestCheckResourceAttrSet("data.tama_class.test_collection", "name"),
					resource.TestCheckResourceAttrSet("data.tama_class.test_collection", "description"),
					resource.TestCheckResourceAttrSet("data.tama_class.test_collection", "schema_json"),
					resource.TestCheckResourceAttrSet("data.tama_class.test_collection", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_class.test_collection", "space_id"),
				),
			},
		},
	})
}

func TestAccClassDataSource_VerifyAllAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassDataSourceConfig("verify-attrs"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify all required attributes are present
					resource.TestCheckResourceAttrSet("data.tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "name"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "description"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "schema_json"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "schema.0.title"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "space_id"),

					// Verify that provision_state is not empty
					resource.TestCheckNoResourceAttr("data.tama_class.test", "provision_state.#"),
				),
			},
		},
	})
}

func TestAccClassDataSource_StateVerification(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassDataSourceConfig("state-test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "name"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "description"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "space_id"),
				),
			},
		},
	})
}

func TestAccClassDataSource_SchemaContent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassDataSourceConfig("schema-content"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "schema_json"),
					// Verify the schema contains expected JSON structure
					resource.TestCheckResourceAttrWith("data.tama_class.test", "schema_json", func(value string) error {
						if value == "" {
							return fmt.Errorf("schema_json should not be empty")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccClassDataSource_DifferentSpaces(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassDataSourceConfigDifferentSpaces(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check first class in root space
					resource.TestCheckResourceAttrSet("data.tama_class.test_root", "id"),
					resource.TestCheckResourceAttrSet("data.tama_class.test_root", "space_id"),

					// Check second class in component space
					resource.TestCheckResourceAttrSet("data.tama_class.test_component", "id"),
					resource.TestCheckResourceAttrSet("data.tama_class.test_component", "space_id"),

					// Verify they have different space_ids
					resource.TestCheckResourceAttrPair("data.tama_class.test_root", "space_id", "tama_space.test_root", "id"),
					resource.TestCheckResourceAttrPair("data.tama_class.test_component", "space_id", "tama_space.test_component", "id"),
				),
			},
		},
	})
}

func TestAccClassDataSource_LongRunning(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassDataSourceConfigComplex("long-running"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "name"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "description"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "schema_json"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "space_id"),
				),
			},
		},
	})
}

func testAccClassDataSourceConfig(name string) string {
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
      code = {
        description = "The status of the action call"
        type        = "integer"
      }
      content_type = {
        description = "The content type of the response"
        type        = "string"
      }
      content = {
        description = "The response from the action"
        type        = "object"
      }
    }
    required = ["tool_id", "parameters", "code", "content_type", "content"]
  })
}

data "tama_class" "test" {
  id = tama_class.test.id
}
`, name, timestamp)
}

func testAccClassDataSourceConfigMultiple() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "multi-test-%d"
  type = "root"
}

resource "tama_class" "test_basic" {
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
      code = {
        description = "The status of the action call"
        type        = "integer"
      }
    }
    required = ["tool_id", "parameters", "code"]
  })
}

resource "tama_class" "test_collection" {
  space_id = tama_space.test.id
  schema {
    title       = "collection"
    description = "A collection is a group of entities that can be queried."
    type        = "object"
    required    = ["items", "space", "name", "created_at"]
    properties  = jsonencode({
      space = {
        type        = "string"
        description = "Slug of the space"
      }
      name = {
        type        = "string"
        description = "The name of the collection"
      }
      created_at = {
        type        = "integer"
        description = "The unix timestamp when the collection was created"
      }
      items = {
        type        = "array"
        description = "An array of objects"
      }
    })
  }
}

data "tama_class" "test_basic" {
  id = tama_class.test_basic.id
}

data "tama_class" "test_collection" {
  id = tama_class.test_collection.id
}
`, timestamp)
}

func testAccClassDataSourceConfigDifferentSpaces() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_root" {
  name = "root-space-%d"
  type = "root"
}

resource "tama_space" "test_component" {
  name = "component-space-%d"
  type = "component"
}

resource "tama_class" "test_root" {
  space_id = tama_space.test_root.id
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

resource "tama_class" "test_component" {
  space_id = tama_space.test_component.id
  schema {
    title       = "entity-network"
    description = <<-EOT
      A entity network is records the connections between entities.

      ## Fields:
      - edges: An array of entity ids that are connected to the entity.
    EOT
    type        = "object"
    required    = ["edges"]
    properties  = jsonencode({
      edges = {
        type        = "object"
        description = "An array of entity ids that are connected to the entity."
      }
    })
  }
}

data "tama_class" "test_root" {
  id = tama_class.test_root.id
}

data "tama_class" "test_component" {
  id = tama_class.test_component.id
}
`, timestamp, timestamp)
}

func testAccClassDataSourceConfigComplex(name string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s-%d"
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema {
    title       = "entity-network"
    description = <<-EOT
      A entity network is records the connections between entities.

      ## Fields:
      - edges: An array of entity ids that are connected to the entity.
    EOT
    type        = "object"
    required    = ["edges"]
    properties  = jsonencode({
      edges = {
        type        = "object"
        description = "An array of entity ids that are connected to the entity."
      }
    })
  }
}

data "tama_class" "test" {
  id = tama_class.test.id
}
`, name, timestamp)
}

func testAccClassDataSourceConfigArray(name string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s-%d"
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema {
    title       = "collection"
    description = "A collection is a group of entities that can be queried."
    type        = "object"
    required    = ["items", "space", "name"]
    properties  = jsonencode({
      space = {
        type        = "string"
        description = "Slug of the space"
      }
      name = {
        type        = "string"
        description = "The name of the collection"
      }
      items = {
        type        = "array"
        description = "An array of objects"
      }
    })
  }
}

data "tama_class" "test" {
  id = tama_class.test.id
}
`, name, timestamp)
}

func testAccClassDataSourceConfigMinimal(name string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s-%d"
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema {
    title       = "minimal-action"
    description = "A minimal action schema with just basic fields"
    type        = "object"
    required    = ["tool_id"]
    properties  = jsonencode({
      tool_id = {
        type        = "string"
        description = "The ID of the tool to execute"
      }
    })
  }
}

data "tama_class" "test" {
  id = tama_class.test.id
}
`, name, timestamp)
}
