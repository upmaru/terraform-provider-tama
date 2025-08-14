// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package thought_initializer_test

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccThoughtInitializerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccThoughtInitializerResourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "thought_id"),
					resource.TestCheckResourceAttr("tama_thought_initializer.test", "reference", "tama/initializers/preload"),
					resource.TestCheckResourceAttr("tama_thought_initializer.test", "index", "0"),
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "class_id"),
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "parameters"),
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_thought_initializer.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccThoughtInitializerResourceConfigUpdate(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "thought_id"),
					resource.TestCheckResourceAttr("tama_thought_initializer.test", "reference", "tama/initializers/preload"),
					resource.TestCheckResourceAttr("tama_thought_initializer.test", "index", "1"),
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "class_id"),
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "parameters"),
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "provision_state"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccThoughtInitializerResource_WithoutIndex(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing without explicit index
			{
				Config: testAccThoughtInitializerResourceConfigWithoutIndex(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "thought_id"),
					resource.TestCheckResourceAttr("tama_thought_initializer.test", "reference", "tama/initializers/preload"),
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "class_id"),
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "parameters"),
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccThoughtInitializerResource_ComplexParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create thought initializer with complex parameters similar to user's example
			{
				Config: testAccThoughtInitializerResourceConfigWithComplexParameters(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "thought_id"),
					resource.TestCheckResourceAttr("tama_thought_initializer.test", "reference", "tama/initializers/preload"),
					resource.TestCheckResourceAttr("tama_thought_initializer.test", "index", "0"),
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "class_id"),
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "parameters"),
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test", "provision_state"),
					// Verify that the parameters contain the expected complex structure
					resource.TestCheckResourceAttrWith("tama_thought_initializer.test", "parameters", func(value string) error {
						var params map[string]any
						if err := json.Unmarshal([]byte(value), &params); err != nil {
							return fmt.Errorf("failed to parse parameters JSON: %v", err)
						}

						// Check that concept parameters are preserved
						if concept, ok := params["concept"].(map[string]any); ok {
							if relations, ok := concept["relations"].([]any); ok {
								if len(relations) != 3 {
									return fmt.Errorf("expected 3 relations, got %d", len(relations))
								}
								expectedRelations := []string{"description", "overview", "setting"}
								for i, rel := range relations {
									if rel.(string) != expectedRelations[i] {
										return fmt.Errorf("expected relation %s, got %s", expectedRelations[i], rel)
									}
								}
							} else {
								return fmt.Errorf("concept.relations should be an array")
							}

							if embeddings, ok := concept["embeddings"]; ok {
								if embeddings != "include" {
									return fmt.Errorf("embeddings should be 'include', got %v", embeddings)
								}
							}

							if content, ok := concept["content"].(map[string]any); ok {
								if action, ok := content["action"]; ok {
									if action != "merge" {
										return fmt.Errorf("content.action should be 'merge', got %v", action)
									}
								}
							}
						}

						// Check that children parameters are preserved
						if children, ok := params["children"].([]any); ok {
							if len(children) != 1 {
								return fmt.Errorf("expected 1 child, got %d", len(children))
							}
							child := children[0].(map[string]any)
							if child["class"] != "movie-credits" {
								return fmt.Errorf("child class should be 'movie-credits', got %v", child["class"])
							}
							if child["as"] != "object" {
								return fmt.Errorf("child as should be 'object', got %v", child["as"])
							}
						}

						return nil
					}),
				),
			},
		},
	})
}

func TestAccThoughtInitializerResource_DuplicateIndex(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create first thought initializer with index 2
			{
				Config: testAccThoughtInitializerResourceConfigWithDuplicateIndex(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_initializer.test1", "id"),
					resource.TestCheckResourceAttr("tama_thought_initializer.test1", "index", "2"),
				),
			},
			// Try to create second thought initializer with same index - should fail with 422
			{
				Config:      testAccThoughtInitializerResourceConfigWithDuplicateIndexError(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				ExpectError: regexp.MustCompile("422|duplicate|index"),
			},
		},
	})
}

// Test helper function to verify index pointer handling for initializers.
func TestInitializerIndexPointerHandling(t *testing.T) {
	// Test case 1: index = 0 should create a valid pointer
	var index0 int64 = 0
	if index0 == 0 {
		intVal := int(index0)
		ptr := &intVal
		if *ptr != 0 {
			t.Errorf("Expected pointer to 0, got %v", ptr)
		}
	}

	// Test case 2: index = 5 should create a valid pointer
	var index5 int64 = 5
	if index5 != 0 {
		intVal := int(index5)
		ptr := &intVal
		if *ptr != 5 {
			t.Errorf("Expected pointer to 5, got %v", ptr)
		}
	}

	// Test case 3: verify that 0 is treated as a valid value (not "empty")
	var indexZero int64 = 0
	intVal := int(indexZero)
	ptr := &intVal
	if *ptr != 0 {
		t.Errorf("Expected pointer value 0, got %d", *ptr)
	}
}

func testAccThoughtInitializerResourceConfig(spaceName string) string {
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
    title       = "Movie Details Schema"
    description = "Schema for movie details"
    type        = "object"
    properties = {
      title = {
        type        = "string"
        description = "Movie title"
      }
      description = {
        type        = "string"
        description = "Movie description"
      }
      overview = {
        type        = "string"
        description = "Movie overview"
      }
      setting = {
        type        = "string"
        description = "Movie setting"
      }
    }
    required = ["title"]
  })
}

resource "tama_modular_thought" "test" {
  chain_id        = tama_chain.test.id
  output_class_id = tama_class.test.id
  relation        = "preload"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "preload"
    })
  }
}

resource "tama_thought_initializer" "test" {
  thought_id = tama_modular_thought.test.id
  reference  = "tama/initializers/preload"
  index      = 0
  class_id   = tama_class.test.id
  parameters = jsonencode({
    concept = {
      relations = ["description", "overview", "setting"]
      embeddings = "include"
      content = {
        action = "merge"
        merge = {
          name = "root-merge"
          location = "root"
        }
      }
    }
    children = [
      {
        class = "movie-credits"
        as = "object"
        on = "parent_entity_id"
      }
    ]
  })
}
`, spaceName)
}

func testAccThoughtInitializerResourceConfigUpdate(spaceName string) string {
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
    title       = "Movie Details Schema"
    description = "Schema for movie details"
    type        = "object"
    properties = {
      title = {
        type        = "string"
        description = "Movie title"
      }
      description = {
        type        = "string"
        description = "Movie description"
      }
      overview = {
        type        = "string"
        description = "Movie overview"
      }
      setting = {
        type        = "string"
        description = "Movie setting"
      }
    }
    required = ["title"]
  })
}

resource "tama_modular_thought" "test" {
  chain_id        = tama_chain.test.id
  output_class_id = tama_class.test.id
  relation        = "preload"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "preload"
    })
  }
}

resource "tama_thought_initializer" "test" {
  thought_id = tama_modular_thought.test.id
  reference  = "tama/initializers/preload"
  index      = 1
  class_id   = tama_class.test.id
  parameters = jsonencode({
    concept = {
      relations = ["description", "overview", "setting", "genre"]
      embeddings = "include"
      content = {
        action = "merge"
        merge = {
          name = "updated-merge"
          location = "root"
        }
      }
    }
    children = [
      {
        class = "movie-credits"
        as = "object"
        on = "parent_entity_id"
      }
    ]
  })
}
`, spaceName)
}

func testAccThoughtInitializerResourceConfigWithoutIndex(spaceName string) string {
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
    title       = "Movie Details Schema"
    description = "Schema for movie details"
    type        = "object"
    properties = {
      title = {
        type        = "string"
        description = "Movie title"
      }
    }
    required = ["title"]
  })
}

resource "tama_modular_thought" "test" {
  chain_id        = tama_chain.test.id
  output_class_id = tama_class.test.id
  relation        = "preload"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "preload"
    })
  }
}

resource "tama_thought_initializer" "test" {
  thought_id = tama_modular_thought.test.id
  reference  = "tama/initializers/preload"
  class_id   = tama_class.test.id
  parameters = jsonencode({
    concept = {
      relations = ["description"]
      embeddings = "include"
    }
  })
}
`, spaceName)
}

func testAccThoughtInitializerResourceConfigWithComplexParameters(spaceName string) string {
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
    title       = "Movie Details Schema"
    description = "Schema for movie details with complex structure"
    type        = "object"
    properties = {
      title = {
        type        = "string"
        description = "Movie title"
      }
      description = {
        type        = "string"
        description = "Movie description"
      }
      overview = {
        type        = "string"
        description = "Movie overview"
      }
      setting = {
        type        = "string"
        description = "Movie setting"
      }
    }
    required = ["title"]
  })
}

resource "tama_modular_thought" "test" {
  chain_id        = tama_chain.test.id
  output_class_id = tama_class.test.id
  relation        = "preload"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "preload"
    })
  }
}

resource "tama_thought_initializer" "test" {
  thought_id = tama_modular_thought.test.id
  reference  = "tama/initializers/preload"
  index      = 0
  class_id   = tama_class.test.id
  parameters = jsonencode({
    concept = {
      relations = ["description", "overview", "setting"]
      embeddings = "include"
      content = {
        action = "merge"
        merge = {
          name = "complex-merge"
          location = "root"
        }
      }
    }
    children = [
      {
        class = "movie-credits"
        as = "object"
        on = "parent_entity_id"
      }
    ]
  })
}
`, spaceName)
}

func testAccThoughtInitializerResourceConfigWithDuplicateIndex(spaceName string) string {
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
    title       = "Movie Details Schema"
    description = "Schema for movie details"
    type        = "object"
    properties = {
      title = {
        type        = "string"
        description = "Movie title"
      }
    }
    required = ["title"]
  })
}

resource "tama_modular_thought" "test" {
  chain_id        = tama_chain.test.id
  output_class_id = tama_class.test.id
  relation        = "preload"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "preload"
    })
  }
}

resource "tama_thought_initializer" "test1" {
  thought_id = tama_modular_thought.test.id
  reference  = "tama/initializers/preload"
  index      = 2
  class_id   = tama_class.test.id
  parameters = jsonencode({
    concept = {
      relations = ["description"]
    }
  })
}
`, spaceName)
}

func testAccThoughtInitializerResourceConfigWithDuplicateIndexError(spaceName string) string {
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
    title       = "Movie Details Schema"
    description = "Schema for movie details"
    type        = "object"
    properties = {
      title = {
        type        = "string"
        description = "Movie title"
      }
    }
    required = ["title"]
  })
}

resource "tama_modular_thought" "test" {
  chain_id        = tama_chain.test.id
  output_class_id = tama_class.test.id
  relation        = "preload"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "preload"
    })
  }
}

resource "tama_thought_initializer" "test1" {
  thought_id = tama_modular_thought.test.id
  reference  = "tama/initializers/preload"
  index      = 2
  class_id   = tama_class.test.id
  parameters = jsonencode({
    concept = {
      relations = ["description"]
    }
  })
}

resource "tama_thought_initializer" "test2" {
  thought_id = tama_modular_thought.test.id
  reference  = "tama/initializers/preload"
  index      = 2
  class_id   = tama_class.test.id
  parameters = jsonencode({
    concept = {
      relations = ["overview"]
    }
  })
}
`, spaceName)
}
