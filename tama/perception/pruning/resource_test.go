// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pruning_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccThoughtPruningResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccThoughtPruningResourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_pruning.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_pruning.test", "thought_id"),
					resource.TestCheckResourceAttr("tama_thought_pruning.test", "previous_versions_count", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_thought_pruning.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccThoughtPruningResourceConfigUpdate(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_pruning.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_pruning.test", "thought_id"),
					resource.TestCheckResourceAttr("tama_thought_pruning.test", "previous_versions_count", "2"),
				),
			},
		},
	})
}

func testAccThoughtPruningResourceConfig(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "thought-pruning-chain"
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

resource "tama_thought_pruning" "test" {
  thought_id              = tama_modular_thought.test.id
  previous_versions_count = 0
}
`, spaceName)
}

func testAccThoughtPruningResourceConfigUpdate(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "thought-pruning-chain"
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

resource "tama_thought_pruning" "test" {
  thought_id              = tama_modular_thought.test.id
  previous_versions_count = 2
}
`, spaceName)
}
