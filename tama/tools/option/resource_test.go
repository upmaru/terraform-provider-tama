// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package option_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccToolOutputOptionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccToolOutputOptionResourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_tool_output_option.test", "id"),
					resource.TestCheckResourceAttrSet("tama_tool_output_option.test", "thought_tool_output_id"),
					resource.TestCheckResourceAttrSet("tama_tool_output_option.test", "action_modifier_id"),
					resource.TestCheckResourceAttrSet("tama_tool_output_option.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_tool_output_option.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update testing - swap to another action modifier
			{
				Config: testAccToolOutputOptionResourceConfigUpdate(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_tool_output_option.test", "id"),
					resource.TestCheckResourceAttrSet("tama_tool_output_option.test", "thought_tool_output_id"),
					resource.TestCheckResourceAttrSet("tama_tool_output_option.test", "action_modifier_id"),
					resource.TestCheckResourceAttrSet("tama_tool_output_option.test", "provision_state"),
				),
			},
		},
	})
}

func testAccToolOutputOptionResourceConfig(spaceName string) string {
	return fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_specification" "test" {
  space_id = tama_space.test.id
  version  = "1.0.0"
  endpoint = "https://elasticsearch.arrakis.upmaru.network"
  schema   = jsonencode(jsondecode(file("${path.module}/testdata/elasticsearch_schema.json")))

  wait_for {
    field {
      name = "current_state"
      in   = ["completed"]
    }
  }
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Tool Chain"
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

data "tama_action" "test" {
  specification_id = tama_specification.test.id
  identifier       = "create-index"
}

resource "tama_thought_tool" "test" {
  thought_id = tama_modular_thought.test.id
  action_id  = data.tama_action.test.id
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "test-tool-output-option"
    description = "Test class for tool output option"
    type = "object"
    properties = {
      name = {
        type = "string"
        description = "Name of the item"
      }
    }
    required = ["name"]
  })
}

resource "tama_class_corpus" "test" {
  class_id = tama_class.test.id
  name     = "TestCorpus"
  template = "{{ data.content }}"
}

resource "tama_thought_tool_output" "test" {
  thought_tool_id = tama_thought_tool.test.id
  class_corpus_id = tama_class_corpus.test.id
}

resource "tama_action_modifier" "test" {
  action_id = data.tama_action.test.id
  name      = "region"
  schema    = jsonencode({ type = "string", description = "the region the user is in" })
}

resource "tama_tool_output_option" "test" {
  thought_tool_output_id = tama_thought_tool_output.test.id
  action_modifier_id     = tama_action_modifier.test.id
}
`, spaceName)
}

func testAccToolOutputOptionResourceConfigUpdate(spaceName string) string {
	return fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_specification" "test" {
  space_id = tama_space.test.id
  version  = "1.0.0"
  endpoint = "https://elasticsearch.arrakis.upmaru.network"
  schema   = jsonencode(jsondecode(file("${path.module}/testdata/elasticsearch_schema.json")))

  wait_for {
    field {
      name = "current_state"
      in   = ["completed"]
    }
  }
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Tool Chain"
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

data "tama_action" "test" {
  specification_id = tama_specification.test.id
  identifier       = "create-index"
}

resource "tama_thought_tool" "test" {
  thought_id = tama_modular_thought.test.id
  action_id  = data.tama_action.test.id
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "test-tool-output-option"
    description = "Test class for tool output option"
    type = "object"
    properties = {
      name = {
        type = "string"
        description = "Name of the item"
      }
    }
    required = ["name"]
  })
}

resource "tama_class_corpus" "test" {
  class_id = tama_class.test.id
  name     = "TestCorpus"
  template = "{{ data.content }}"
}

resource "tama_thought_tool_output" "test" {
  thought_tool_id = tama_thought_tool.test.id
  class_corpus_id = tama_class_corpus.test.id
}

resource "tama_action_modifier" "test2" {
  action_id = data.tama_action.test.id
  name      = "user-region"
  schema    = jsonencode({ type = "string", description = "user selected region" })
}

resource "tama_tool_output_option" "test" {
  thought_tool_output_id = tama_thought_tool_output.test.id
  action_modifier_id     = tama_action_modifier.test2.id
}
`, spaceName)
}
