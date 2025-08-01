// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package context_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccThoughtContextDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccThoughtContextDataSourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_thought_context.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_thought_context.test", "thought_id"),
					resource.TestCheckResourceAttrSet("data.tama_thought_context.test", "prompt_id"),
					resource.TestCheckResourceAttrSet("data.tama_thought_context.test", "layer"),
					resource.TestCheckResourceAttrSet("data.tama_thought_context.test", "provision_state"),
				),
			},
		},
	})
}

func testAccThoughtContextDataSourceConfig(spaceName string) string {
	return fmt.Sprintf(`
provider "tama" {}

resource "tama_space" "test" {
  name = "%s"
  type = "root"
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

data "tama_thought_context" "test" {
  id = tama_thought_context.test.id
}
`, spaceName)
}
