// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package activation_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccThoughtPathActivationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccThoughtPathActivationResourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_path_activation.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_activation.test", "thought_path_id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_activation.test", "chain_id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_activation.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_thought_path_activation.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccThoughtPathActivationResourceConfigUpdate(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_path_activation.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_activation.test", "thought_path_id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_activation.test", "chain_id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_activation.test", "provision_state"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccThoughtPathActivationResource_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtPathActivationResourceConfigMultiple(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_path_activation.test1", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_activation.test1", "thought_path_id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_activation.test1", "chain_id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_activation.test2", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_activation.test2", "thought_path_id"),
					resource.TestCheckResourceAttrSet("tama_thought_path_activation.test2", "chain_id"),
				),
			},
		},
	})
}

func testAccThoughtPathActivationResourceConfig(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "%s"
  type = "root"
}

resource "tama_node" "test_node" {
  space_id = tama_space.test_space.id
  class_id = tama_class.test_class.id
  chain_id = tama_chain.test_chain1.id
  type     = "explicit"
}

resource "tama_class" "test_class" {
  space_id = tama_space.test_space.id
  schema_json = jsonencode({
    title       = "Test Activation Target Schema"
    description = "Schema for activation target"
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
  name     = "test-chain-for-activation-1"
}

resource "tama_chain" "test_chain2" {
  space_id = tama_space.test_space.id
  name     = "test-chain-for-activation-2"
}

resource "tama_modular_thought" "test" {
  chain_id = tama_chain.test_chain1.id
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

resource "tama_node" "test_node_chain2" {
  space_id = tama_space.test_space.id
  class_id = tama_class.test_class.id
  chain_id = tama_chain.test_chain2.id
  type     = "explicit"
}

resource "tama_thought_path_activation" "test" {
  thought_path_id = tama_thought_path.test.id
  chain_id        = tama_chain.test_chain2.id
  depends_on      = [tama_node.test_node_chain2]
}
`, spaceName)
}

func testAccThoughtPathActivationResourceConfigUpdate(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "%s"
  type = "root"
}

resource "tama_node" "test_node" {
  space_id = tama_space.test_space.id
  class_id = tama_class.test_class.id
  chain_id = tama_chain.test_chain1.id
  type     = "explicit"
}

resource "tama_class" "test_class" {
  space_id = tama_space.test_space.id
  schema_json = jsonencode({
    title       = "Test Activation Target Schema"
    description = "Schema for activation target"
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
  name     = "test-chain-for-activation-1"
}

resource "tama_chain" "test_chain2" {
  space_id = tama_space.test_space.id
  name     = "test-chain-for-activation-2"
}

resource "tama_chain" "test_chain3" {
  space_id = tama_space.test_space.id
  name     = "test-chain-for-activation-3"
}

resource "tama_modular_thought" "test" {
  chain_id = tama_chain.test_chain1.id
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

resource "tama_node" "test_node_chain3" {
  space_id = tama_space.test_space.id
  class_id = tama_class.test_class.id
  chain_id = tama_chain.test_chain3.id
  type     = "explicit"
}

resource "tama_thought_path_activation" "test" {
  thought_path_id = tama_thought_path.test.id
  chain_id        = tama_chain.test_chain3.id
  depends_on      = [tama_node.test_node_chain3]
}
`, spaceName)
}

func testAccThoughtPathActivationResourceConfigMultiple(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "%s"
  type = "root"
}

resource "tama_node" "test_node" {
  space_id = tama_space.test_space.id
  class_id = tama_class.test_class.id
  chain_id = tama_chain.test_chain1.id
  type     = "explicit"
}

resource "tama_class" "test_class" {
  space_id = tama_space.test_space.id
  schema_json = jsonencode({
    title       = "Test Activation Target Schema"
    description = "Schema for activation target"
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
  name     = "test-chain-for-activation-1"
}

resource "tama_chain" "test_chain2" {
  space_id = tama_space.test_space.id
  name     = "test-chain-for-activation-2"
}

resource "tama_chain" "test_chain3" {
  space_id = tama_space.test_space.id
  name     = "test-chain-for-activation-3"
}

resource "tama_modular_thought" "test1" {
  chain_id = tama_chain.test_chain1.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_modular_thought" "test2" {
  chain_id = tama_chain.test_chain2.id
  relation = "summary"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "summary"
    })
  }
}

resource "tama_thought_path" "test1" {
  thought_id      = tama_modular_thought.test1.id
  target_class_id = tama_class.test_class.id

  parameters = jsonencode({
    relation = "similarity"
  })
}

resource "tama_thought_path" "test2" {
  thought_id      = tama_modular_thought.test2.id
  target_class_id = tama_class.test_class.id

  parameters = jsonencode({
    relation = "similarity"
  })
}

resource "tama_node" "test_node_chain2" {
  space_id = tama_space.test_space.id
  class_id = tama_class.test_class.id
  chain_id = tama_chain.test_chain2.id
  type     = "explicit"
}

resource "tama_node" "test_node_chain3" {
  space_id = tama_space.test_space.id
  class_id = tama_class.test_class.id
  chain_id = tama_chain.test_chain3.id
  type     = "explicit"
}

resource "tama_thought_path_activation" "test1" {
  thought_path_id = tama_thought_path.test1.id
  chain_id        = tama_chain.test_chain2.id
  depends_on      = [tama_node.test_node_chain2]
}

resource "tama_thought_path_activation" "test2" {
  thought_path_id = tama_thought_path.test2.id
  chain_id        = tama_chain.test_chain3.id
  depends_on      = [tama_node.test_node_chain3]
}
`, spaceName)
}
