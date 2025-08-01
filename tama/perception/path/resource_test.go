// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package path_test

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

// testCheckJSONEqual creates a test function that checks if two JSON strings are semantically equal.
func testCheckJSONEqual(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["tama_thought_path.test"]
		if !ok {
			return fmt.Errorf("Not found: %s", "tama_thought_path.test")
		}

		actual := rs.Primary.Attributes["parameters"]

		// Normalize both JSON strings for comparison
		var expectedObj, actualObj any

		if err := json.Unmarshal([]byte(expected), &expectedObj); err != nil {
			return fmt.Errorf("Expected value is not valid JSON: %v", err)
		}

		if err := json.Unmarshal([]byte(actual), &actualObj); err != nil {
			return fmt.Errorf("Actual value is not valid JSON: %v", err)
		}

		// Compare the parsed objects
		expectedNorm, _ := json.Marshal(expectedObj)
		actualNorm, _ := json.Marshal(actualObj)

		if string(expectedNorm) != string(actualNorm) {
			return fmt.Errorf("JSON values are not equal.\nExpected: %s\nActual: %s", expected, actual)
		}

		return nil
	}
}

func TestAccThoughtPathResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccThoughtPathResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("tama_thought_path.test", "target_class_id", "tama_class.test_class", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path.test", "thought_id"),
					testCheckJSONEqual(`{"relation": "similarity"}`),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_thought_path.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccThoughtPathResourceConfigUpdate(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("tama_thought_path.test", "target_class_id", "tama_class.test_class", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path.test", "id"),
					testCheckJSONEqual(`{"relation": "updated"}`),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccThoughtPathResource_WithParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with parameters
			{
				Config: testAccThoughtPathResourceConfigWithParameters(`{"relation": "similarity", "similarity": {"threshold": 0.9}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("tama_thought_path.test", "target_class_id", "tama_class.test_class", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path.test", "parameters"),
					resource.TestCheckResourceAttrSet("tama_thought_path.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path.test", "thought_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "tama_thought_path.test",
				ImportState:             true,
				ImportStateVerify:       false, // Parameters might be normalized by API
				ImportStateVerifyIgnore: []string{"parameters"},
			},
			// Update parameters
			{
				Config: testAccThoughtPathResourceConfigWithParameters(`{"relation": "similarity", "similarity": {"threshold": 0.8}, "max_results": 10}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("tama_thought_path.test", "target_class_id", "tama_class.test_class", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path.test", "parameters"),
				),
			},
		},
	})
}

func TestAccThoughtPathResource_ComplexParameters(t *testing.T) {
	complexParams := `{
		"relation": "similarity",
		"similarity": {
			"threshold": 0.9,
			"algorithm": "cosine"
		},
		"filtering": {
			"enabled": true,
			"min_score": 0.5
		},
		"output": {
			"format": "json",
			"include_metadata": true
		}
	}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtPathResourceConfigWithParameters(complexParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("tama_thought_path.test", "target_class_id", "tama_class.test_class", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path.test", "parameters"),
					resource.TestCheckResourceAttrSet("tama_thought_path.test", "id"),
				),
			},
		},
	})
}

func TestAccThoughtPathResource_ParameterTypes(t *testing.T) {
	testCases := []struct {
		name       string
		parameters string
	}{
		{
			"Threshold parameters",
			`{"relation": "similarity", "similarity": {"threshold": 0.9}, "confidence": {"min": 0.8}}`,
		},
		{
			"Numeric parameters",
			`{"relation": "similarity", "max_results": 100, "timeout": 30, "batch_size": 50}`,
		},
		{
			"Boolean parameters",
			`{"relation": "similarity", "include_metadata": true, "cache_enabled": false, "debug": true}`,
		},
		{
			"Array parameters",
			`{"relation": "similarity", "filters": ["active", "verified"], "excluded_classes": ["spam", "test"]}`,
		},
		{
			"Mixed parameters",
			`{"relation": "similarity", "similarity": {"threshold": 0.9}, "enabled": true, "tags": ["prod", "v2"]}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccThoughtPathResourceConfigWithParameters(tc.parameters),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttrPair("tama_thought_path.test", "target_class_id", "tama_class.test_class", "id"),
							resource.TestCheckResourceAttrSet("tama_thought_path.test", "parameters"),
							resource.TestCheckResourceAttrSet("tama_thought_path.test", "id"),
						),
					},
				},
			})
		})
	}
}

func TestAccThoughtPathResource_EmptyParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtPathResourceConfigWithParameters(`{"relation": "similarity"}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("tama_thought_path.test", "target_class_id", "tama_class.test_class", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path.test", "parameters"),
					resource.TestCheckResourceAttrSet("tama_thought_path.test", "id"),
				),
			},
		},
	})
}

func TestAccThoughtPathResource_InvalidParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccThoughtPathResourceConfigWithParameters(`{"invalid": json}`),
				ExpectError: regexp.MustCompile("Invalid Parameters"),
			},
		},
	})
}

func TestAccThoughtPathResource_InvalidTargetClassId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccThoughtPathResourceConfigInvalidClassId(),
				ExpectError: regexp.MustCompile("Unable to create path"),
			},
		},
	})
}

func TestAccThoughtPathResource_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtPathResourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// First path
					resource.TestCheckResourceAttrPair("tama_thought_path.test1", "target_class_id", "tama_class.test_class1", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path.test1", "id"),
					// Second path
					resource.TestCheckResourceAttrPair("tama_thought_path.test2", "target_class_id", "tama_class.test_class2", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path.test2", "id"),
				),
			},
		},
	})
}

func TestAccThoughtPathResource_DifferentThoughts(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtPathResourceConfigDifferentThoughts(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Path from first thought
					resource.TestCheckResourceAttrPair("tama_thought_path.thought1", "target_class_id", "tama_class.test_class_a", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path.thought1", "id"),
					// Path from second thought
					resource.TestCheckResourceAttrPair("tama_thought_path.thought2", "target_class_id", "tama_class.test_class_b", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path.thought2", "id"),
				),
			},
		},
	})
}

func testAccThoughtPathResourceConfig() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-path-%d"
  type = "root"
}

resource "tama_class" "test_class" {
  space_id = tama_space.test_space.id
  schema_json = jsonencode({
    title       = "Test Path Target Schema"
    description = "Schema for path target"
    type        = "object"
    properties = {
      content = {
        type        = "string"
        description = "Content field"
      }
    }
    required = ["content"]
  })
}

resource "tama_chain" "test_chain" {
  space_id = tama_space.test_space.id
  name     = "test-chain-for-path"
}

resource "tama_modular_thought" "test" {
  chain_id = tama_chain.test_chain.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_thought_path" "test" {
  thought_id      = tama_modular_thought.test.id
  target_class_id = tama_class.test_class.id

  parameters = jsonencode({
    relation = "similarity"
  })
}
`, timestamp)
}

func testAccThoughtPathResourceConfigUpdate() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-path-%d"
  type = "root"
}

resource "tama_class" "test_class" {
  space_id = tama_space.test_space.id
  schema_json = jsonencode({
    title       = "Test Path Target Schema"
    description = "Schema for path target"
    type        = "object"
    properties = {
      content = {
        type        = "string"
        description = "Content field"
      }
    }
    required = ["content"]
  })
}

resource "tama_chain" "test_chain" {
  space_id = tama_space.test_space.id
  name     = "test-chain-for-path"
}

resource "tama_modular_thought" "test_thought" {
  chain_id = tama_chain.test_chain.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_thought_path" "test" {
  thought_id      = tama_modular_thought.test_thought.id
  target_class_id = tama_class.test_class.id

  parameters = jsonencode({
    relation = "updated"
  })
}
`, timestamp)
}

func testAccThoughtPathResourceConfigMultiple() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-multiple-paths-%d"
  type = "root"
}

resource "tama_class" "test_class1" {
  space_id = tama_space.test_space.id
  schema_json = jsonencode({
    title       = "Test Path Target Schema 1"
    description = "Schema for path target 1"
    type        = "object"
    properties = {
      content = {
        type        = "string"
        description = "Content field"
      }
    }
    required = ["content"]
  })
}

resource "tama_class" "test_class2" {
  space_id = tama_space.test_space.id
  schema_json = jsonencode({
    title       = "Test Path Target Schema 2"
    description = "Schema for path target 2"
    type        = "object"
    properties = {
      content = {
        type        = "string"
        description = "Content field"
      }
    }
    required = ["content"]
  })
}

resource "tama_chain" "test_chain" {
  space_id = tama_space.test_space.id
  name     = "test-chain-for-multiple-paths"
}

resource "tama_modular_thought" "test_thought" {
  chain_id = tama_chain.test_chain.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_thought_path" "test1" {
  thought_id      = tama_modular_thought.test_thought.id
  target_class_id = tama_class.test_class1.id

  parameters = jsonencode({
    relation = "similarity"
  })
}

resource "tama_thought_path" "test2" {
  thought_id      = tama_modular_thought.test_thought.id
  target_class_id = tama_class.test_class2.id

  parameters = jsonencode({
    relation = "similarity"
  })
}
`, timestamp)
}

func testAccThoughtPathResourceConfigDifferentThoughts() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-different-thoughts-%d"
  type = "root"
}

resource "tama_class" "test_class_a" {
  space_id = tama_space.test_space.id
  schema_json = jsonencode({
    title       = "Test Path Target Schema A"
    description = "Schema for path target A"
    type        = "object"
    properties = {
      content = {
        type        = "string"
        description = "Content field"
      }
    }
    required = ["content"]
  })
}

resource "tama_class" "test_class_b" {
  space_id = tama_space.test_space.id
  schema_json = jsonencode({
    title       = "Test Path Target Schema B"
    description = "Schema for path target B"
    type        = "object"
    properties = {
      content = {
        type        = "string"
        description = "Content field"
      }
    }
    required = ["content"]
  })
}

resource "tama_chain" "test_chain1" {
  space_id = tama_space.test_space.id
  name     = "test-chain-for-thought1"
}

resource "tama_chain" "test_chain2" {
  space_id = tama_space.test_space.id
  name     = "test-chain-for-thought2"
}

resource "tama_modular_thought" "thought1" {
  chain_id = tama_chain.test_chain1.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_modular_thought" "test" {
  chain_id = tama_chain.test_chain2.id
  relation = "analysis"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "analysis"
    })
  }
}

resource "tama_thought_path" "thought1" {
  thought_id      = tama_modular_thought.thought1.id
  target_class_id = tama_class.test_class_a.id

  parameters = jsonencode({
    relation = "similarity"
  })
}

resource "tama_thought_path" "thought2" {
  thought_id      = tama_modular_thought.test.id
  target_class_id = tama_class.test_class_b.id

  parameters = jsonencode({
    relation = "similarity"
  })
}
`, timestamp)
}

func testAccThoughtPathResourceConfigInvalidClassId() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-path-%d"
  type = "root"
}

resource "tama_chain" "test_chain" {
  space_id = tama_space.test_space.id
  name     = "test-chain-for-path"
}

resource "tama_modular_thought" "test_thought" {
  chain_id = tama_chain.test_chain.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_thought_path" "test" {
  thought_id      = tama_modular_thought.test_thought.id
  target_class_id = "invalid-class-id"

  parameters = jsonencode({
    relation = "similarity"
  })
}
`, timestamp)
}

func testAccThoughtPathResourceConfigWithParameters(parameters string) string {
	timestamp := time.Now().UnixNano()
	config := acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-path-%d"
  type = "root"
}

resource "tama_class" "test_class" {
  space_id = tama_space.test_space.id
  schema_json = jsonencode({
    title       = "Test Path Target Schema"
    description = "Schema for path target"
    type        = "object"
    properties = {
      content = {
        type        = "string"
        description = "Content field"
      }
    }
    required = ["content"]
  })
}

resource "tama_chain" "test_chain" {
  space_id = tama_space.test_space.id
  name     = "test-chain-for-path"
}

resource "tama_modular_thought" "test_thought" {
  chain_id = tama_chain.test_chain.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_thought_path" "test" {
  thought_id      = tama_modular_thought.test_thought.id
  target_class_id = tama_class.test_class.id`, timestamp)

	if parameters != "" {
		config += fmt.Sprintf(`
  parameters = %[1]q`, parameters)
	}

	config += `
}`

	return config
}
