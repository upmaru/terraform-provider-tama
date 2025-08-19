// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package input_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccThoughtToolInputResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccThoughtToolInputResourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_tool_input.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_input.test", "thought_tool_id"),
					resource.TestCheckResourceAttr("tama_thought_tool_input.test", "type", "body"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_input.test", "class_corpus_id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_input.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_thought_tool_input.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update testing - change type from body to query
			{
				Config: testAccThoughtToolInputResourceConfigUpdate(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_tool_input.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_input.test", "thought_tool_id"),
					resource.TestCheckResourceAttr("tama_thought_tool_input.test", "type", "query"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_input.test", "class_corpus_id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_input.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccThoughtToolInputResource_AllTypes(t *testing.T) {
	types := []string{"path", "query", "header", "body"}

	for _, inputType := range types {
		t.Run(fmt.Sprintf("type_%s", inputType), func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccThoughtToolInputResourceConfigForType(fmt.Sprintf("test-space-%d", time.Now().UnixNano()), inputType),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttrSet("tama_thought_tool_input.test", "id"),
							resource.TestCheckResourceAttrSet("tama_thought_tool_input.test", "thought_tool_id"),
							resource.TestCheckResourceAttr("tama_thought_tool_input.test", "type", inputType),
							resource.TestCheckResourceAttrSet("tama_thought_tool_input.test", "class_corpus_id"),
							resource.TestCheckResourceAttrSet("tama_thought_tool_input.test", "provision_state"),
						),
					},
				},
			})
		})
	}
}

func testAccThoughtToolInputResourceConfig(spaceName string) string {
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
    title = "test-tool-input"
    description = "Test class for tool input"
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

resource "tama_thought_tool_input" "test" {
  thought_tool_id = tama_thought_tool.test.id
  type            = "body"
  class_corpus_id = tama_class_corpus.test.id
}
`, spaceName)
}

func testAccThoughtToolInputResourceConfigUpdate(spaceName string) string {
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
    title = "test-tool-input"
    description = "Test class for tool input"
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

resource "tama_thought_tool_input" "test" {
  thought_tool_id = tama_thought_tool.test.id
  type            = "query"
  class_corpus_id = tama_class_corpus.test.id
}
`, spaceName)
}

func testAccThoughtToolInputResourceConfigForType(spaceName, inputType string) string {
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
    title = "test-tool-input"
    description = "Test class for tool input"
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

resource "tama_thought_tool_input" "test" {
  thought_tool_id = tama_thought_tool.test.id
  type            = "%s"
  class_corpus_id = tama_class_corpus.test.id
}
`, spaceName, inputType)
}
