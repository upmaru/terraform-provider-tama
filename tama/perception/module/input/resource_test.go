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

func TestAccThoughtModuleInputResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccThoughtModuleInputResourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_module_input.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_module_input.test", "thought_id"),
					resource.TestCheckResourceAttrSet("tama_thought_module_input.test", "thought_module_id"),
					resource.TestCheckResourceAttr("tama_thought_module_input.test", "type", "concept"),
					resource.TestCheckResourceAttrSet("tama_thought_module_input.test", "class_corpus_id"),
					resource.TestCheckResourceAttrSet("tama_thought_module_input.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_thought_module_input.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccThoughtModuleInputResourceConfigUpdate(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_module_input.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_module_input.test", "thought_id"),
					resource.TestCheckResourceAttrSet("tama_thought_module_input.test", "thought_module_id"),
					resource.TestCheckResourceAttr("tama_thought_module_input.test", "type", "entity"),
					resource.TestCheckResourceAttrSet("tama_thought_module_input.test", "class_corpus_id"),
					resource.TestCheckResourceAttrSet("tama_thought_module_input.test", "provision_state"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccThoughtModuleInputResource_EntityType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtModuleInputResourceConfigEntityType(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_module_input.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_module_input.test", "thought_id"),
					resource.TestCheckResourceAttrSet("tama_thought_module_input.test", "thought_module_id"),
					resource.TestCheckResourceAttr("tama_thought_module_input.test", "type", "entity"),
					resource.TestCheckResourceAttrSet("tama_thought_module_input.test", "class_corpus_id"),
					resource.TestCheckResourceAttrSet("tama_thought_module_input.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccThoughtModuleInputResource_InvalidType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccThoughtModuleInputResourceConfigInvalidType(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				ExpectError: regexp.MustCompile(`unsupported type: invalid`),
			},
		},
	})
}

func testAccThoughtModuleInputResourceConfig(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
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
    }
    required = ["tool_id", "parameters"]
  })
}

resource "tama_class_corpus" "test" {
  class_id = tama_class.test.id
  name     = "Test Corpus"
  main     = true
  template = "{{ data.something }}"
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

resource "tama_thought_module_input" "test" {
  thought_id        = tama_modular_thought.test.id
  type              = "concept"
  class_corpus_id   = tama_class_corpus.test.id
}
`, spaceName)
}

func testAccThoughtModuleInputResourceConfigUpdate(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
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
    }
    required = ["tool_id", "parameters"]
  })
}

resource "tama_class_corpus" "test" {
  class_id = tama_class.test.id
  name     = "Test Corpus Updated"
  main     = false
  template = "{{ data.updated }}"
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

resource "tama_thought_module_input" "test" {
  thought_id        = tama_modular_thought.test.id
  type              = "entity"
  class_corpus_id   = tama_class_corpus.test.id
}
`, spaceName)
}

func testAccThoughtModuleInputResourceConfigEntityType(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "entity-definition"
    description = "Entity definition for NER tasks."
    type = "object"
    properties = {
      name = {
        description = "The entity name"
        type        = "string"
      }
      type = {
        description = "The entity type"
        type        = "string"
      }
    }
    required = ["name", "type"]
  })
}

resource "tama_class_corpus" "test" {
  class_id = tama_class.test.id
  name     = "Entity Corpus"
  main     = true
  template = "{{ data.entity }}"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Entity Processing Chain"
}

resource "tama_modular_thought" "test" {
  chain_id = tama_chain.test.id
  relation = "extraction"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "extraction"
    })
  }
}

resource "tama_thought_module_input" "test" {
  thought_id        = tama_modular_thought.test.id
  type              = "entity"
  class_corpus_id   = tama_class_corpus.test.id
}
`, spaceName)
}

func testAccThoughtModuleInputResourceConfigInvalidType(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "test-schema"
    description = "Test schema"
    type = "object"
    properties = {
      name = {
        type = "string"
      }
    }
  })
}

resource "tama_class_corpus" "test" {
  class_id = tama_class.test.id
  name     = "Test Corpus"
  main     = true
  template = "{{ data.test }}"
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

resource "tama_thought_module_input" "test" {
  thought_id        = tama_modular_thought.test.id
  type              = "invalid"
  class_corpus_id   = tama_class_corpus.test.id
}
`, spaceName)
}
