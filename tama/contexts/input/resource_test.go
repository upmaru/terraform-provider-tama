// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package input_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccThoughtContextInputResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccThoughtContextInputResourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_context_input.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_context_input.test", "thought_context_id"),
					resource.TestCheckResourceAttr("tama_thought_context_input.test", "type", "metadata"),
					resource.TestCheckResourceAttrSet("tama_thought_context_input.test", "class_corpus_id"),
					resource.TestCheckResourceAttrSet("tama_thought_context_input.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_thought_context_input.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccThoughtContextInputResourceConfigUpdate(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_context_input.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_context_input.test", "thought_context_id"),
					resource.TestCheckResourceAttr("tama_thought_context_input.test", "type", "entity"),
					resource.TestCheckResourceAttrSet("tama_thought_context_input.test", "class_corpus_id"),
					resource.TestCheckResourceAttrSet("tama_thought_context_input.test", "provision_state"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccThoughtContextInputResource_InvalidType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccThoughtContextInputResourceConfigInvalidType(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				ExpectError: regexp.MustCompile("Attribute type value must be one of"),
			},
		},
	})
}

func TestAccThoughtContextInputResource_ConceptType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtContextInputResourceConfigConceptType(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_context_input.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_context_input.test", "thought_context_id"),
					resource.TestCheckResourceAttr("tama_thought_context_input.test", "type", "concept"),
					resource.TestCheckResourceAttrSet("tama_thought_context_input.test", "class_corpus_id"),
					resource.TestCheckResourceAttrSet("tama_thought_context_input.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccThoughtContextInputResource_ThoughtContextChange(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtContextInputResourceConfigThoughtContextChange("initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_context_input.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_context_input.test", "thought_context_id"),
				),
			},
			{
				Config: testAccThoughtContextInputResourceConfigThoughtContextChange("updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_context_input.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_context_input.test", "thought_context_id"),
				),
			},
		},
	})
}

func testAccThoughtContextInputResourceConfig(spaceName string) string {
	return fmt.Sprintf(`
provider "tama" {}

resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "test-input"
    description = "A test input for context processing."
    type = "object"
    properties = {
      content = {
        description = "The content of the input"
        type        = "string"
      }
      metadata = {
        description = "Additional metadata"
        type        = "object"
      }
    }
    required = ["content"]
  })
}

resource "tama_class_corpus" "test" {
  class_id = tama_class.test.id
  name     = "Test Corpus 2345"
  main     = true
  template = "{{ data.content }}"
}

resource "tama_prompt" "test" {
  space_id = tama_space.test.id
  name     = "test-prompt"
  content  = "Test prompt content"
  role     = "system"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "test-chain"
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

resource "tama_thought_context" "test" {
  thought_id = tama_modular_thought.test.id
  prompt_id  = tama_prompt.test.id
  layer      = 0
}

resource "tama_thought_context_input" "test" {
  thought_context_id = tama_thought_context.test.id
  type               = "metadata"
  class_corpus_id    = tama_class_corpus.test.id
}
`, spaceName)
}

func testAccThoughtContextInputResourceConfigUpdate(spaceName string) string {
	return fmt.Sprintf(`
provider "tama" {}

resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "test-input"
    description = "A test input for context processing."
    type = "object"
    properties = {
      content = {
        description = "The content of the input"
        type        = "string"
      }
      entity_data = {
        description = "Entity data"
        type        = "object"
      }
    }
    required = ["content"]
  })
}

resource "tama_class_corpus" "test" {
  class_id = tama_class.test.id
  name     = "Updated Corpus 2345"
  main     = true
  template = "{{ data.content }} - {{ data.entity_data }}"
}

resource "tama_prompt" "test" {
  space_id = tama_space.test.id
  name     = "test-prompt-updated"
  content  = "Updated test prompt content"
  role     = "user"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "test-chain"
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

resource "tama_thought_context" "test" {
  thought_id = tama_modular_thought.test.id
  prompt_id  = tama_prompt.test.id
  layer      = 0
}

resource "tama_thought_context_input" "test" {
  thought_context_id = tama_thought_context.test.id
  type               = "entity"
  class_corpus_id    = tama_class_corpus.test.id
}
`, spaceName)
}

func testAccThoughtContextInputResourceConfigInvalidType(spaceName string) string {
	return fmt.Sprintf(`
provider "tama" {}

resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "test-input"
    description = "A test input for context processing."
    type = "object"
    properties = {
      content = {
        description = "The content of the input"
        type        = "string"
      }
    }
    required = ["content"]
  })
}

resource "tama_class_corpus" "test" {
  class_id = tama_class.test.id
  name     = "Test Corpus"
  main     = true
  template = "{{ data.content }}"
}

resource "tama_prompt" "test" {
  space_id = tama_space.test.id
  name     = "test-prompt"
  content  = "Test prompt content"
  role     = "system"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "test-chain"
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

resource "tama_thought_context" "test" {
  thought_id = tama_modular_thought.test.id
  prompt_id  = tama_prompt.test.id
  layer      = 0
}

resource "tama_thought_context_input" "test" {
  thought_context_id = tama_thought_context.test.id
  type               = "invalid_type"
  class_corpus_id    = tama_class_corpus.test.id
}
`, spaceName)
}

func testAccThoughtContextInputResourceConfigConceptType(spaceName string) string {
	return fmt.Sprintf(`
provider "tama" {}

resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "test-concept"
    description = "A test concept for context processing."
    type = "object"
    properties = {
      concept_name = {
        description = "The name of the concept"
        type        = "string"
      }
      concept_definition = {
        description = "The definition of the concept"
        type        = "string"
      }
    }
    required = ["concept_name", "concept_definition"]
  })
}

resource "tama_class_corpus" "test" {
  class_id = tama_class.test.id
  name     = "Concept Corpus"
  main     = true
  template = "{{ data.concept_name }}: {{ data.concept_definition }}"
}

resource "tama_prompt" "test" {
  space_id = tama_space.test.id
  name     = "test-prompt"
  content  = "Test prompt content"
  role     = "system"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "test-chain"
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

resource "tama_thought_context" "test" {
  thought_id = tama_modular_thought.test.id
  prompt_id  = tama_prompt.test.id
  layer      = 0
}

resource "tama_thought_context_input" "test" {
  thought_context_id = tama_thought_context.test.id
  type               = "concept"
  class_corpus_id    = tama_class_corpus.test.id
}
`, spaceName)
}

func testAccThoughtContextInputResourceConfigThoughtContextChange(suffix string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf(`
provider "tama" {}

resource "tama_space" "test_%s" {
  name = "test-space-%s-%d"
  type = "root"
}

resource "tama_class" "test_%s" {
  space_id = tama_space.test_%s.id
  schema_json = jsonencode({
    title = "test-input-%s"
    description = "A test input for context processing."
    type = "object"
    properties = {
      content = {
        description = "The content of the input"
        type        = "string"
      }
    }
    required = ["content"]
  })
}

resource "tama_class_corpus" "test_%s" {
  class_id = tama_class.test_%s.id
  name     = "Test Corpus"
  main     = true
  template = "{{ data.content }}"
}

resource "tama_prompt" "test_%s" {
  space_id = tama_space.test_%s.id
  name     = "test-prompt"
  content  = "Test prompt content"
  role     = "system"
}

resource "tama_chain" "test_%s" {
  space_id = tama_space.test_%s.id
  name     = "test-chain"
}

resource "tama_modular_thought" "test_%s" {
  chain_id = tama_chain.test_%s.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_thought_context" "test_%s" {
  thought_id = tama_modular_thought.test_%s.id
  prompt_id  = tama_prompt.test_%s.id
  layer      = 0
}

resource "tama_thought_context_input" "test" {
  thought_context_id = tama_thought_context.test_%s.id
  type               = "metadata"
  class_corpus_id    = tama_class_corpus.test_%s.id
}
`, suffix, suffix, timestamp, suffix, suffix, suffix, suffix, suffix, suffix, suffix, suffix, suffix, suffix, suffix, suffix, suffix, suffix, suffix, suffix)
}
