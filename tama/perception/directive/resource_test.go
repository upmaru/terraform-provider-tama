// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directive_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccThoughtPathDirectiveResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccThoughtPathDirectiveResourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_path_directive.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_directive.test", "thought_path_id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_directive.test", "prompt_id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_directive.test", "target_thought_id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_directive.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_thought_path_directive.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccThoughtPathDirectiveResourceConfigUpdate(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_path_directive.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_directive.test", "thought_path_id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_directive.test", "prompt_id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_directive.test", "target_thought_id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_directive.test", "provision_state"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccThoughtPathDirectiveResource_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtPathDirectiveResourceConfigMultiple(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_path_directive.test1", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_directive.test1", "thought_path_id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_directive.test1", "prompt_id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_directive.test1", "target_thought_id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_directive.test2", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_directive.test2", "thought_path_id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_directive.test2", "prompt_id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_directive.test2", "target_thought_id"),
				),
			},
		},
	})
}

func testAccThoughtPathDirectiveResourceConfig(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "%s"
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
  name     = "test-chain-for-directive"
}

resource "tama_chain" "target_chain" {
  space_id = tama_space.test_space.id
  name     = "target-chain-for-directive"
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

resource "tama_modular_thought" "target_thought" {
  chain_id = tama_chain.target_chain.id
  relation = "analysis"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "analysis"
    })
  }
}

resource "tama_thought_path" "test_path" {
  thought_id      = tama_modular_thought.test_thought.id
  target_class_id = tama_class.test_class.id

  parameters = jsonencode({
    relation = "similarity"
  })
}

resource "tama_prompt" "test_prompt" {
  space_id = tama_space.test_space.id
  name     = "test-prompt"
  content  = "Test prompt content"
  role     = "system"
}

resource "tama_thought_path_directive" "test" {
  thought_path_id   = tama_thought_path.test_path.id
  prompt_id         = tama_prompt.test_prompt.id
  target_thought_id = tama_modular_thought.target_thought.id
}
`, spaceName)
}

func testAccThoughtPathDirectiveResourceConfigUpdate(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "%s"
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
  name     = "test-chain-for-directive"
}

resource "tama_chain" "target_chain" {
  space_id = tama_space.test_space.id
  name     = "target-chain-for-directive"
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

resource "tama_modular_thought" "target_thought" {
  chain_id = tama_chain.target_chain.id
  relation = "analysis"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "analysis"
    })
  }
}

resource "tama_thought_path" "test_path" {
  thought_id      = tama_modular_thought.test_thought.id
  target_class_id = tama_class.test_class.id

  parameters = jsonencode({
    relation = "similarity"
  })
}

resource "tama_prompt" "test_prompt1" {
  space_id = tama_space.test_space.id
  name     = "test-prompt-1"
  content  = "Test prompt content 1"
  role     = "system"
}

resource "tama_prompt" "test_prompt2" {
  space_id = tama_space.test_space.id
  name     = "test-prompt-2"
  content  = "Test prompt content 2"
  role     = "user"
}

resource "tama_thought_path_directive" "test" {
  thought_path_id   = tama_thought_path.test_path.id
  prompt_id         = tama_prompt.test_prompt2.id
  target_thought_id = tama_modular_thought.target_thought.id
}
`, spaceName)
}

func testAccThoughtPathDirectiveResourceConfigMultiple(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "%s"
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
  name     = "test-chain-for-directive"
}

resource "tama_chain" "target_chain" {
  space_id = tama_space.test_space.id
  name     = "target-chain-for-directive"
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

resource "tama_modular_thought" "target_thought" {
  chain_id = tama_chain.target_chain.id
  relation = "analysis"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "analysis"
    })
  }
}

resource "tama_thought_path" "test_path" {
  thought_id      = tama_modular_thought.test_thought.id
  target_class_id = tama_class.test_class.id

  parameters = jsonencode({
    relation = "similarity"
  })
}

resource "tama_prompt" "test_prompt1" {
  space_id = tama_space.test_space.id
  name     = "test-prompt-1"
  content  = "Test prompt content 1"
  role     = "system"
}

resource "tama_prompt" "test_prompt2" {
  space_id = tama_space.test_space.id
  name     = "test-prompt-2"
  content  = "Test prompt content 2"
  role     = "user"
}

resource "tama_thought_path_directive" "test1" {
  thought_path_id   = tama_thought_path.test_path.id
  prompt_id         = tama_prompt.test_prompt1.id
  target_thought_id = tama_modular_thought.target_thought.id
}

resource "tama_thought_path_directive" "test2" {
  thought_path_id   = tama_thought_path.test_path.id
  prompt_id         = tama_prompt.test_prompt2.id
  target_thought_id = tama_modular_thought.target_thought.id
}
`, spaceName)
}
