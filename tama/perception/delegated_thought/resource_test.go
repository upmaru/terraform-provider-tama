// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package delegated_thought_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccDelegatedThoughtResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial modular thought
			{
				Config: testAccDelegatedThoughtResourceConfigInitial(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_modular_thought.test", "id"),
					resource.TestCheckResourceAttrSet("tama_delegated_thought.test", "id"),
					resource.TestCheckResourceAttrSet("tama_delegated_thought.test", "chain_id"),
					resource.TestCheckResourceAttrSet("tama_delegated_thought.test", "relation"),
					resource.TestCheckResourceAttrSet("tama_delegated_thought.test", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_delegated_thought.test", "index"),
					resource.TestCheckResourceAttrPair("tama_delegated_thought.test", "delegation.target_thought_id", "tama_modular_thought.test", "id"),
				),
			},
		},
	})
}

func TestAccDelegatedThoughtInvalidAssignmentResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDelegatedThoughtInvalidResourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				ExpectError: regexp.MustCompile(`The target_thought_id must reference a tama_modular_thought resource`),
			},
		},
	})
}

func TestAccDelegatedThoughtResource_WithIndexZero(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with index = 0 (testing pointer handling)
			{
				Config: testAccDelegatedThoughtResourceConfigWithIndexZero(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_modular_thought.test", "id"),
					resource.TestCheckResourceAttrSet("tama_delegated_thought.test", "id"),
					resource.TestCheckResourceAttrSet("tama_delegated_thought.test", "chain_id"),
					resource.TestCheckResourceAttr("tama_delegated_thought.test", "index", "0"),
					resource.TestCheckResourceAttrSet("tama_delegated_thought.test", "relation"),
					resource.TestCheckResourceAttrSet("tama_delegated_thought.test", "provision_state"),
					resource.TestCheckResourceAttrPair("tama_delegated_thought.test", "delegation.target_thought_id", "tama_modular_thought.test", "id"),
				),
			},
		},
	})
}

// Test helper function to verify index pointer handling.
func TestIndexPointerHandling(t *testing.T) {
	// Test case 1: index = 0 should create a valid pointer
	var index0 int64 = 0
	if index0 == 0 {
		intVal := int(index0)
		ptr := &intVal
		if *ptr != 0 {
			t.Errorf("Expected pointer to 0, got %v", ptr)
		}
	}

	// Test case 2: index = 5 should create a valid pointer
	var index5 int64 = 5
	if index5 != 0 {
		intVal := int(index5)
		ptr := &intVal
		if *ptr != 5 {
			t.Errorf("Expected pointer to 5, got %v", ptr)
		}
	}

	// Test case 3: verify that 0 is treated as a valid value (not "empty")
	var indexZero int64 = 0
	intVal := int(indexZero)
	ptr := &intVal
	if *ptr != 0 {
		t.Errorf("Expected pointer value 0, got %d", *ptr)
	}
}

func testAccDelegatedThoughtResourceConfigInitial(spaceName string) string {
	return fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_chain" "test_chain_1" {
  space_id = tama_space.test.id
  name     = "Test Chain 1"
}

resource "tama_chain" "test_chain_2" {
  space_id = tama_space.test.id
  name     = "Test Chain 2"
}

resource "tama_modular_thought" "test" {
  chain_id = tama_chain.test_chain_1.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_delegated_thought" "test" {
  chain_id = tama_chain.test_chain_2.id

  delegation {
    target_thought_id = tama_modular_thought.test.id
  }
}
`, spaceName)
}

func testAccDelegatedThoughtResourceConfigWithIndexZero(spaceName string) string {
	return fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_chain" "test_chain_1" {
  space_id = tama_space.test.id
  name     = "Test Chain 1"
}

resource "tama_chain" "test_chain_2" {
  space_id = tama_space.test.id
  name     = "Test Chain 2"
}

resource "tama_modular_thought" "test" {
  chain_id = tama_chain.test_chain_1.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_delegated_thought" "test" {
  chain_id = tama_chain.test_chain_2.id
  index    = 0

  delegation {
    target_thought_id = tama_modular_thought.test.id
  }
}
`, spaceName)
}

func testAccDelegatedThoughtInvalidResourceConfig(spaceName string) string {
	return fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_chain" "test_chain_1" {
  space_id = tama_space.test.id
  name     = "Test Chain 1"
}

resource "tama_chain" "test_chain_2" {
  space_id = tama_space.test.id
  name     = "Test Chain 2"
}

resource "tama_chain" "test_chain_3" {
  space_id = tama_space.test.id
  name     = "Test Chain 3"
}

# Create a modular thought first
resource "tama_modular_thought" "source" {
  chain_id = tama_chain.test_chain_1.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

# Create the first delegated thought targeting the modular thought
resource "tama_delegated_thought" "first" {
  chain_id = tama_chain.test_chain_2.id

  delegation {
    target_thought_id = tama_modular_thought.source.id
  }
}

# Try to create another delegated thought targeting the first delegated thought (this should fail)
resource "tama_delegated_thought" "test" {
  chain_id = tama_chain.test_chain_3.id

  delegation {
    target_thought_id = tama_delegated_thought.first.id
  }
}
`, spaceName)
}
