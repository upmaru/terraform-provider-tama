// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package class_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccClassResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccClassResourceConfigWithBlock(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class.test", "name"),
					resource.TestCheckResourceAttrSet("tama_class.test", "description"),
					resource.TestCheckResourceAttrSet("tama_class.test", "schema.0.title"),
					resource.TestCheckResourceAttrSet("tama_class.test", "schema.0.description"),
					resource.TestCheckResourceAttrSet("tama_class.test", "schema.0.type"),
					resource.TestCheckResourceAttrSet("tama_class.test", "current_state"),
					resource.TestCheckResourceAttrSet("tama_class.test", "space_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_class.test",
				ImportState:       true,
				ImportStateVerify: false, // Skip verification due to schema format differences
			},
			// Update and Read testing
			{
				Config: testAccClassResourceConfigWithJSON(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class.test", "name"),
					resource.TestCheckResourceAttrSet("tama_class.test", "description"),
					resource.TestCheckResourceAttrSet("tama_class.test", "schema_json"),
					resource.TestCheckResourceAttrSet("tama_class.test", "current_state"),
					resource.TestCheckResourceAttrSet("tama_class.test", "space_id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccClassResource_InvalidSchemaJSON(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccClassResourceConfigWithInvalidJSON(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				ExpectError: regexp.MustCompile("Unable to parse schema JSON"),
			},
		},
	})
}

func TestAccClassResource_MissingTitleDescription(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccClassResourceConfigMissingFields(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				ExpectError: regexp.MustCompile("JSON schema must include 'title' field"),
			},
		},
	})
}

func TestAccClassResource_SchemaBlock(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassResourceConfigComplexBlock(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class.test", "name"),
					resource.TestCheckResourceAttrSet("tama_class.test", "description"),
					resource.TestCheckResourceAttr("tama_class.test", "schema.0.title", "entity-network"),
					resource.TestCheckResourceAttr("tama_class.test", "schema.0.type", "object"),
					resource.TestCheckResourceAttrSet("tama_class.test", "current_state"),
					resource.TestCheckResourceAttrSet("tama_class.test", "space_id"),
				),
			},
		},
	})
}

func TestAccClassResource_BothSchemaTypes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccClassResourceConfigBothSchemas(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				ExpectError: regexp.MustCompile("Cannot specify both schema block and schema_json attribute"),
			},
		},
	})
}

func TestAccClassResource_NoSchema(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccClassResourceConfigNoSchema(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				ExpectError: regexp.MustCompile("Either schema block or schema_json attribute must be provided"),
			},
		},
	})
}

func TestAccClassResource_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassResourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// First class
					resource.TestCheckResourceAttrSet("tama_class.test1", "id"),
					resource.TestCheckResourceAttrSet("tama_class.test1", "name"),
					resource.TestCheckResourceAttrSet("tama_class.test1", "description"),
					resource.TestCheckResourceAttrSet("tama_class.test1", "schema_json"),
					resource.TestCheckResourceAttrSet("tama_class.test1", "current_state"),
					resource.TestCheckResourceAttrSet("tama_class.test1", "space_id"),
					// Second class
					resource.TestCheckResourceAttrSet("tama_class.test2", "id"),
					resource.TestCheckResourceAttrSet("tama_class.test2", "name"),
					resource.TestCheckResourceAttrSet("tama_class.test2", "description"),
					resource.TestCheckResourceAttr("tama_class.test2", "schema.#", "1"),
					resource.TestCheckResourceAttrSet("tama_class.test2", "current_state"),
					resource.TestCheckResourceAttrSet("tama_class.test2", "space_id"),
				),
			},
		},
	})
}

func TestAccClassResource_SpaceIdChange(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassResourceConfigSpaceChange("initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class.test", "space_id"),
				),
			},
			{
				Config: testAccClassResourceConfigSpaceChange("updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class.test", "space_id"),
				),
			},
		},
	})
}

func TestAccClassResource_JSONEncoding(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassResourceConfigWithJSONEncode(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class.test", "name"),
					resource.TestCheckResourceAttrSet("tama_class.test", "description"),
					resource.TestCheckResourceAttrSet("tama_class.test", "schema_json"),
					resource.TestCheckResourceAttrSet("tama_class.test", "current_state"),
					resource.TestCheckResourceAttrSet("tama_class.test", "space_id"),
				),
			},
		},
	})
}

func testAccClassResourceConfigWithBlock(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id

  schema {
    title       = "action-call"
    description = "An action call is a request to execute an action."
    type        = "object"
    required    = ["tool_id", "parameters", "code", "content_type", "content"]
    strict      = true
    properties  = jsonencode({
      tool_id = {
        type        = "string"
        description = "The ID of the tool to execute"
      }
      parameters = {
        type        = "object"
        description = "The parameters to pass to the action"
      }
      code = {
        type        = "integer"
        description = "The status of the action call"
      }
      content_type = {
        type        = "string"
        description = "The content type of the response"
      }
      content = {
        type        = "object"
        description = "The response from the action"
      }
    })
  }
}
`, spaceName)
}

func testAccClassResourceConfigWithJSON(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "collection"
    description = "A collection is a group of entities that can be queried."
    type = "object"
    properties = {
      space = {
        type        = "string"
        description = "Slug of the space"
      }
      name = {
        description = "The name of the collection"
        type        = "string"
      }
      created_at = {
        description = "The unix timestamp when the collection was created"
        type        = "integer"
      }
      items = {
        description = "An array of objects"
        type        = "array"
        items = {
          type = "object"
        }
      }
    }
    required = ["items", "space", "name", "created_at"]
  })
}
`, spaceName)
}

func testAccClassResourceConfigComplexBlock(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
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
    strict      = false
    properties  = jsonencode({
      edges = {
        type        = "object"
        description = "An array of entity ids that are connected to the entity."
      }
    })
  }
}
`, spaceName)
}

func testAccClassResourceConfigWithInvalidJSON(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = "invalid-json"
}
`, spaceName)
}

func testAccClassResourceConfigMissingFields(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    type = "object"
    properties = {
      tool_id = {
        type = "string"
      }
    }
  })
}
`, spaceName)
}

func testAccClassResourceConfigBothSchemas(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "JSON Schema"
    description = "A schema via JSON"
    type = "object"
  })

  schema {
    title       = "Block Schema"
    description = "A schema via block"
    type        = "object"
  }
}
`, spaceName)
}

func testAccClassResourceConfigMultiple() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "test-space-%d"
  type = "root"
}

resource "tama_class" "test1" {
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

resource "tama_class" "test2" {
  space_id = tama_space.test.id
  schema {
    title       = "collection"
    description = "A collection is a group of entities that can be queried."
    type        = "object"
    required    = ["items", "name"]
    properties  = jsonencode({
      items = {
        type        = "array"
        description = "An array of objects"
      }
      name = {
        type        = "string"
        description = "The name of the collection"
      }
    })
  }
}
`, timestamp)
}

func testAccClassResourceConfigSpaceChange(suffix string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_%s" {
  name = "test-space-%s-%d"
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test_%s.id
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
`, suffix, suffix, timestamp, suffix)
}

func testAccClassResourceConfigWithJSONEncode() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "test-space-%d"
  type = "root"
}

variable "schema" {
  type = object({
    title = string
    description = string
    type = string
    properties = object({
      tool_id = object({
        description = string
        type = string
      })
      parameters = object({
        description = string
        type = string
      })
      code = object({
        description = string
        type = string
      })
    })
    required = list(string)
  })
  default = {
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
  }
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode(var.schema)
}
`, timestamp)
}

func testAccClassResourceConfigNoSchema(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
}
`, spaceName)
}
