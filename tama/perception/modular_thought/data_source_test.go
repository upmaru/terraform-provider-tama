// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package modular_thought_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccModularThoughtDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModularThoughtDataSourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "chain_id"),
					resource.TestCheckResourceAttr("data.tama_modular_thought.test", "relation", "description"),
					resource.TestCheckResourceAttr("data.tama_modular_thought.test", "module.reference", "tama/agentic/generate"),
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "module.parameters"),
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "index"),
				),
			},
		},
	})
}

func TestAccModularThoughtDataSource_WithOutputClass(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModularThoughtDataSourceConfigWithOutputClass(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "chain_id"),
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "output_class_id"),
					resource.TestCheckResourceAttr("data.tama_modular_thought.test", "relation", "validation"),
					resource.TestCheckResourceAttr("data.tama_modular_thought.test", "module.reference", "tama/agentic/generate"),
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "module.parameters"),
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "index"),
				),
			},
		},
	})
}

func TestAccModularThoughtDataSource_MinimalParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModularThoughtDataSourceConfigMinimalParameters(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "chain_id"),
					resource.TestCheckResourceAttr("data.tama_modular_thought.test", "relation", "analysis"),
					resource.TestCheckResourceAttr("data.tama_modular_thought.test", "module.reference", "tama/agentic/generate"),
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "module.parameters"),
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "index"),
				),
			},
		},
	})
}

func TestAccModularThoughtDataSource_NoParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModularThoughtDataSourceConfigNoParameters(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "chain_id"),
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "output_class_id"),
					resource.TestCheckResourceAttr("data.tama_modular_thought.test", "relation", "validation"),
					resource.TestCheckResourceAttr("data.tama_modular_thought.test", "module.reference", "tama/identities/validate"),
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_modular_thought.test", "index"),
				),
			},
		},
	})
}

func testAccModularThoughtDataSourceConfig(spaceName string) string {
	return fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
}

resource "tama_modular_thought" "test" {
  chain_id = tama_chain.test.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

data "tama_modular_thought" "test" {
  id = tama_modular_thought.test.id
}
`, spaceName)
}

func testAccModularThoughtDataSourceConfigWithOutputClass(spaceName string) string {
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

resource "tama_modular_thought" "test" {
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

data "tama_modular_thought" "test" {
  id = tama_modular_thought.test.id
}
`, spaceName)
}

func testAccModularThoughtDataSourceConfigNoParameters(spaceName string) string {
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

resource "tama_modular_thought" "test" {
  chain_id        = tama_chain.test.id
  output_class_id = tama_class.test.id
  relation        = "validation"

  module {
    reference = "tama/identities/validate"
  }
}

data "tama_modular_thought" "test" {
  id = tama_modular_thought.test.id
}
`, spaceName)
}

func testAccModularThoughtDataSourceConfigMinimalParameters(spaceName string) string {
	return fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
}

resource "tama_modular_thought" "test" {
  chain_id = tama_chain.test.id
  relation = "analysis"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "analysis"
    })
  }
}

data "tama_modular_thought" "test" {
  id = tama_modular_thought.test.id
}
`, spaceName)
}
