// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package output_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccThoughtToolOutputResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccThoughtToolOutputResourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_tool_output.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_output.test", "thought_tool_id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_output.test", "class_corpus_id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_output.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_thought_tool_output.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update testing - change class_corpus_id to another corpus
			{
				Config: testAccThoughtToolOutputResourceConfigUpdate(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_tool_output.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_output.test", "thought_tool_id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_output.test", "class_corpus_id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_output.test", "provision_state"),
				),
			},
		},
	})
}

func testAccThoughtToolOutputResourceConfig(spaceName string) string {
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
    title = "test-tool-output"
    description = "Test class for tool output"
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
`, spaceName)
}

func testAccThoughtToolOutputResourceConfigUpdate(spaceName string) string {
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

resource "tama_class" "test1" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "test-tool-output-1"
    description = "Test class for tool output 1"
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

resource "tama_class_corpus" "test1" {
  class_id = tama_class.test1.id
  name     = "TestCorpus1"
  template = "{{ data.content }}"
}

resource "tama_class" "test2" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "test-tool-output-2"
    description = "Test class for tool output 2"
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

resource "tama_class_corpus" "test2" {
  class_id = tama_class.test2.id
  name     = "TestCorpus2"
  template = "{{ data.content }}"
}

resource "tama_thought_tool_output" "test" {
  thought_tool_id = tama_thought_tool.test.id
  class_corpus_id = tama_class_corpus.test2.id
}
`, spaceName)
}
