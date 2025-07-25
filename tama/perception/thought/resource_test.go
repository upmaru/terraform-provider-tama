// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package thought_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccThoughtResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccThoughtResourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "chain_id"),
					resource.TestCheckResourceAttr("tama_thought.test", "relation", "description"),
					resource.TestCheckResourceAttr("tama_thought.test", "module.0.reference", "tama/agentic/generate"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "module.0.parameters"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "index"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_thought.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccThoughtResourceConfigUpdate(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "chain_id"),
					resource.TestCheckResourceAttr("tama_thought.test", "relation", "analysis"),
					resource.TestCheckResourceAttr("tama_thought.test", "module.0.reference", "tama/agentic/generate"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "module.0.parameters"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "index"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccThoughtResource_WithOutputClass(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with output class
			{
				Config: testAccThoughtResourceConfigWithOutputClass(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "chain_id"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "output_class_id"),
					resource.TestCheckResourceAttr("tama_thought.test", "relation", "validation"),
					resource.TestCheckResourceAttr("tama_thought.test", "module.0.reference", "tama/agentic/generate"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "module.0.parameters"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "index"),
				),
			},
		},
	})
}

func testAccThoughtResourceConfig(spaceName string) string {
	return fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
}

resource "tama_thought" "test" {
  chain_id = tama_chain.test.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}
`, spaceName)
}

func testAccThoughtResourceConfigUpdate(spaceName string) string {
	return fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
}

resource "tama_thought" "test" {
  chain_id = tama_chain.test.id
  relation = "analysis"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "analysis"
    })
  }
}
`, spaceName)
}

func testAccThoughtResourceConfigWithOutputClass(spaceName string) string {
	return fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title       = "Validation Output Schema"
    description = "Schema for validation output"
    type        = "object"
    properties = {
      valid = {
        type        = "boolean"
        description = "Whether the input is valid"
      }
      errors = {
        type        = "array"
        description = "List of validation errors"
        items = {
          type = "string"
        }
      }
    }
    required = ["valid"]
  })
}

resource "tama_thought" "test" {
  chain_id        = tama_chain.test.id
  output_class_id = tama_class.test.id
  relation        = "validation"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "validation"
    })
  }
}
`, spaceName)
}
