// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package initializer_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccThoughtToolInitializerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccThoughtToolInitializerResourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_tool_initializer.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_initializer.test", "thought_tool_id"),
					resource.TestCheckResourceAttr("tama_thought_tool_initializer.test", "reference", "tama/initializers/import"),
					resource.TestCheckResourceAttr("tama_thought_tool_initializer.test", "index", "0"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_initializer.test", "parameters"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_initializer.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_thought_tool_initializer.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update testing - change reference and parameters
			{
				Config: testAccThoughtToolInitializerResourceConfigUpdate(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_tool_initializer.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_initializer.test", "thought_tool_id"),
					resource.TestCheckResourceAttr("tama_thought_tool_initializer.test", "reference", "tama/initializers/preload"),
					resource.TestCheckResourceAttr("tama_thought_tool_initializer.test", "index", "1"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_initializer.test", "parameters"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_initializer.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccThoughtToolInitializerResource_WithoutParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtToolInitializerResourceConfigWithoutParameters(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_tool_initializer.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_initializer.test", "thought_tool_id"),
					resource.TestCheckResourceAttr("tama_thought_tool_initializer.test", "reference", "tama/initializers/preload"),
					resource.TestCheckResourceAttr("tama_thought_tool_initializer.test", "index", "0"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_initializer.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccThoughtToolInitializerResource_CustomIndex(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtToolInitializerResourceConfigCustomIndex(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_tool_initializer.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_initializer.test", "thought_tool_id"),
					resource.TestCheckResourceAttr("tama_thought_tool_initializer.test", "reference", "tama/initializers/import"),
					resource.TestCheckResourceAttr("tama_thought_tool_initializer.test", "index", "2"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_initializer.test", "parameters"),
					resource.TestCheckResourceAttrSet("tama_thought_tool_initializer.test", "provision_state"),
				),
			},
		},
	})
}

func testAccThoughtToolInitializerResourceConfig(spaceName string) string {
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

resource "tama_thought_tool_initializer" "test" {
  thought_tool_id = tama_thought_tool.test.id
  reference       = "tama/initializers/import"
  parameters = jsonencode({
    resources = [
      { type = "concept", relation = "some-relation", scope = "space" }
    ]
  })
}
`, spaceName)
}

func testAccThoughtToolInitializerResourceConfigUpdate(spaceName string) string {
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

resource "tama_thought_tool_initializer" "test" {
  thought_tool_id = tama_thought_tool.test.id
  reference       = "tama/initializers/preload"
  index           = 1
  parameters = jsonencode({
    record = {
      rejections = []
    }
    parents = []
    concept = {
      relations = ["description", "overview"]
      embeddings = "include"
      content = {
        action = "merge"
        merge = {
          name = "tool-merge"
          location = "root"
        }
      }
    }
    children = []
  })
}
`, spaceName)
}

func testAccThoughtToolInitializerResourceConfigWithoutParameters(spaceName string) string {
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

resource "tama_thought_tool_initializer" "test" {
  thought_tool_id = tama_thought_tool.test.id
  reference       = "tama/initializers/preload"
  parameters = jsonencode({
    record = {
      rejections = []
    }
    parents = []
    concept = {
      relations = []
      embeddings = "exclude"
      content = {
        action = "merge"
        merge = {
          name = "simple-merge"
          location = "root"
        }
      }
    }
    children = []
  })
}
`, spaceName)
}

func testAccThoughtToolInitializerResourceConfigCustomIndex(spaceName string) string {
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

resource "tama_thought_tool_initializer" "test" {
  thought_tool_id = tama_thought_tool.test.id
  reference       = "tama/initializers/import"
  index           = 2
  parameters = jsonencode({
    resources = [
      { type = "concept", relation = "primary-relation", scope = "space" }
    ]
  })
}
`, spaceName)
}
