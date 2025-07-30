// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package context_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccThoughtContextResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccThoughtContextResourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_context.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_context.test", "thought_id"),
					resource.TestCheckResourceAttrSet("tama_thought_context.test", "prompt_id"),
					resource.TestCheckResourceAttrSet("tama_thought_context.test", "layer"),
					resource.TestCheckResourceAttrSet("tama_thought_context.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_thought_context.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccThoughtContextResourceConfigUpdate(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_context.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_context.test", "thought_id"),
					resource.TestCheckResourceAttrSet("tama_thought_context.test", "prompt_id"),
					resource.TestCheckResourceAttrSet("tama_thought_context.test", "layer"),
					resource.TestCheckResourceAttrSet("tama_thought_context.test", "provision_state"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccThoughtContextResource_DuplicateLayer(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccThoughtContextResourceConfigDuplicateLayer(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				ExpectError: regexp.MustCompile("422|duplicate|layer|already exists"),
			},
		},
	})
}

func testAccThoughtContextResourceConfig(spaceName string) string {
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

resource "tama_thought" "test" {
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
  thought_id = tama_thought.test.id
  prompt_id  = tama_prompt.test.id
  layer      = 0
}
`, spaceName)
}

func testAccThoughtContextResourceConfigUpdate(spaceName string) string {
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

resource "tama_prompt" "test_update" {
  space_id = tama_space.test.id
  name     = "test-prompt-update"
  content  = "Updated test prompt content"
  role     = "user"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "test-chain"
}

resource "tama_thought" "test" {
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
  thought_id = tama_thought.test.id
  prompt_id  = tama_prompt.test_update.id
  layer      = 1
}
`, spaceName)
}

func testAccThoughtContextResourceConfigDuplicateLayer(spaceName string) string {
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

resource "tama_prompt" "test_second" {
  space_id = tama_space.test.id
  name     = "test-prompt-second"
  content  = "Second test prompt content"
  role     = "user"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "test-chain"
}

resource "tama_thought" "test" {
  chain_id = tama_chain.test.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_thought_context" "test_first" {
  thought_id = tama_thought.test.id
  prompt_id  = tama_prompt.test.id
  layer      = 0
}

resource "tama_thought_context" "test_second" {
  thought_id = tama_thought.test.id
  prompt_id  = tama_prompt.test_second.id
  layer      = 0
}
`, spaceName)
}
