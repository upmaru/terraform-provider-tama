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

func TestAccThoughtDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtDataSourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "chain_id"),
					resource.TestCheckResourceAttr("data.tama_thought.test", "relation", "description"),
					resource.TestCheckResourceAttr("data.tama_thought.test", "module.0.reference", "tama/agentic/generate"),
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "module.0.parameters"),
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "index"),
				),
			},
		},
	})
}

func TestAccThoughtDataSource_WithOutputClass(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtDataSourceConfigWithOutputClass(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "chain_id"),
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "output_class_id"),
					resource.TestCheckResourceAttr("data.tama_thought.test", "relation", "validation"),
					resource.TestCheckResourceAttr("data.tama_thought.test", "module.0.reference", "tama/agentic/generate"),
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "module.0.parameters"),
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "index"),
				),
			},
		},
	})
}

func TestAccThoughtDataSource_MinimalParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtDataSourceConfigMinimalParameters(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "chain_id"),
					resource.TestCheckResourceAttr("data.tama_thought.test", "relation", "analysis"),
					resource.TestCheckResourceAttr("data.tama_thought.test", "module.0.reference", "tama/agentic/generate"),
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "module.0.parameters"),
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "index"),
				),
			},
		},
	})
}

func TestAccThoughtDataSource_NoParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtDataSourceConfigNoParameters(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "chain_id"),
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "output_class_id"),
					resource.TestCheckResourceAttr("data.tama_thought.test", "relation", "validation"),
					resource.TestCheckResourceAttr("data.tama_thought.test", "module.0.reference", "tama/identities/validate"),
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_thought.test", "index"),
				),
			},
		},
	})
}

func testAccThoughtDataSourceConfig(spaceName string) string {
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

data "tama_thought" "test" {
  id = tama_thought.test.id
}
`, spaceName)
}

func testAccThoughtDataSourceConfigWithOutputClass(spaceName string) string {
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

data "tama_thought" "test" {
  id = tama_thought.test.id
}
`, spaceName)
}

func testAccThoughtDataSourceConfigNoParameters(spaceName string) string {
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
    reference = "tama/identities/validate"
  }
}

data "tama_thought" "test" {
  id = tama_thought.test.id
}
`, spaceName)
}

func testAccThoughtDataSourceConfigMinimalParameters(spaceName string) string {
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

data "tama_thought" "test" {
  id = tama_thought.test.id
}
`, spaceName)
}
