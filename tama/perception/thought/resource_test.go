// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package thought_test

import (
	"encoding/json"
	"fmt"
	"regexp"
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

func TestAccThoughtResource_WithIndex(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with explicit index
			{
				Config: testAccThoughtResourceConfigWithIndex(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "chain_id"),
					resource.TestCheckResourceAttr("tama_thought.test", "relation", "description"),
					resource.TestCheckResourceAttr("tama_thought.test", "index", "5"),
					resource.TestCheckResourceAttr("tama_thought.test", "module.0.reference", "tama/agentic/generate"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "module.0.parameters"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccThoughtResource_DuplicateIndex(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create first thought with index 3
			{
				Config: testAccThoughtResourceConfigWithDuplicateIndex(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought.test1", "id"),
					resource.TestCheckResourceAttr("tama_thought.test1", "index", "3"),
				),
			},
			// Try to create second thought with same index - should fail with 422
			{
				Config:      testAccThoughtResourceConfigWithDuplicateIndexError(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				ExpectError: regexp.MustCompile("422|duplicate|index"),
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

func testAccThoughtResourceConfigWithIndex(spaceName string) string {
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
  index    = 5

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}
`, spaceName)
}

func testAccThoughtResourceConfigWithDuplicateIndex(spaceName string) string {
	return fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
}

resource "tama_thought" "test1" {
  chain_id = tama_chain.test.id
  relation = "description"
  index    = 3

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}
`, spaceName)
}

func testAccThoughtResourceConfigWithDuplicateIndexError(spaceName string) string {
	return fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
}

resource "tama_thought" "test1" {
  chain_id = tama_chain.test.id
  relation = "description"
  index    = 3

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_thought" "test2" {
  chain_id = tama_chain.test.id
  relation = "analysis"
  index    = 3

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "analysis"
    })
  }
}
`, spaceName)
}

func TestAccThoughtResource_ParameterMerging(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create thought with complex parameters similar to user's routing example
			{
				Config: testAccThoughtResourceConfigWithComplexParameters(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "chain_id"),
					resource.TestCheckResourceAttr("tama_thought.test", "relation", "routing"),
					resource.TestCheckResourceAttr("tama_thought.test", "module.0.reference", "tama/agentic/router"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "module.0.parameters"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_thought.test", "index"),
					// Verify that the threshold remains a number (0.9) in the parameters
					resource.TestCheckResourceAttrWith("tama_thought.test", "module.0.parameters", func(value string) error {
						var params map[string]any
						if err := json.Unmarshal([]byte(value), &params); err != nil {
							return fmt.Errorf("failed to parse parameters JSON: %v", err)
						}

						// Check that similarity.threshold is preserved as a number
						if similarity, ok := params["similarity"].(map[string]any); ok {
							if threshold, exists := similarity["threshold"]; exists {
								// Should be a number (float64), not a string
								thresholdFloat, isFloat := threshold.(float64)
								if !isFloat {
									return fmt.Errorf("threshold should be preserved as number, got %T: %v", threshold, threshold)
								}
								if thresholdFloat != 0.9 {
									return fmt.Errorf("threshold should be 0.9, got %v", threshold)
								}
							}
						}

						// Check that classification parameters are preserved
						if classification, ok := params["classification"].(map[string]any); ok {
							if className, exists := classification["class_name"]; exists {
								if className != "class" {
									return fmt.Errorf("class_name should be 'class', got %v", className)
								}
							}
						}

						return nil
					}),
				),
			},
		},
	})
}

func testAccThoughtResourceConfigWithComplexParameters(spaceName string) string {
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
  relation = "routing"
  index    = 1

  module {
    reference = "tama/agentic/router"
    parameters = jsonencode({
      similarity = {
        limit     = 10
        threshold = 0.9
      }
      classification = {
        class_name = "class"
        properties = ["class", "confidence"]
        look_back_limit = 5
      }
    })
  }
}
`, spaceName)
}
