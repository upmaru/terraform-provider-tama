// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tool_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccThoughtToolResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccThoughtToolResourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_tool.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool.test", "thought_id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool.test", "action_id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_thought_tool.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update testing - change action_id
			{
				Config: testAccThoughtToolResourceConfigUpdate(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_tool.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool.test", "thought_id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool.test", "action_id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool.test", "provision_state"),
				),
			},
		},
	})
}

func testAccThoughtToolResourceConfig(spaceName string) string {
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
`, spaceName)
}

func testAccThoughtToolResourceConfigUpdate(spaceName string) string {
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

data "tama_action" "test_update" {
  specification_id = tama_specification.test.id
  identifier       = "create-or-update-document-with-id"
}

resource "tama_thought_tool" "test" {
  thought_id = tama_modular_thought.test.id
  action_id  = data.tama_action.test_update.id
}
`, spaceName)
}
