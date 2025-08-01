// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package path_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccThoughtPathDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtPathDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tama_thought_path.test", "target_class_id", "tama_class.test_class", "id"),
					resource.TestCheckResourceAttrSet("data.tama_thought_path.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_thought_path.test", "thought_id"),
					resource.TestCheckResourceAttrSet("data.tama_thought_path.test", "parameters"),
				),
			},
		},
	})
}

func TestAccThoughtPathDataSource_WithParameters(t *testing.T) {
	parameters := `{"relation": "similarity", "similarity": {"threshold": 0.9}}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtPathDataSourceConfigWithParameters(parameters),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tama_thought_path.test", "target_class_id", "tama_class.test_class", "id"),
					resource.TestCheckResourceAttrSet("data.tama_thought_path.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_thought_path.test", "thought_id"),
					resource.TestCheckResourceAttrSet("data.tama_thought_path.test", "parameters"),
				),
			},
		},
	})
}

func TestAccThoughtPathDataSource_ComplexParameters(t *testing.T) {
	complexParams := `{
		"relation": "similarity",
		"similarity": {
			"threshold": 0.9,
			"algorithm": "cosine"
		},
		"filtering": {
			"enabled": true,
			"min_score": 0.5
		},
		"output": {
			"format": "json",
			"include_metadata": true
		}
	}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtPathDataSourceConfigWithParameters(complexParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tama_thought_path.test", "target_class_id", "tama_class.test_class", "id"),
					resource.TestCheckResourceAttrSet("data.tama_thought_path.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_thought_path.test", "thought_id"),
					resource.TestCheckResourceAttrSet("data.tama_thought_path.test", "parameters"),
				),
			},
		},
	})
}

func TestAccThoughtPathDataSource_EmptyParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtPathDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tama_thought_path.test", "target_class_id", "tama_class.test_class", "id"),
					resource.TestCheckResourceAttrSet("data.tama_thought_path.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_thought_path.test", "thought_id"),
					resource.TestCheckResourceAttrSet("data.tama_thought_path.test", "parameters"),
				),
			},
		},
	})
}

func TestAccThoughtPathDataSource_SimilarityThreshold(t *testing.T) {
	testCases := []struct {
		name       string
		parameters string
	}{
		{
			"Low threshold",
			`{"relation": "similarity", "similarity": {"threshold": 0.5}}`,
		},
		{
			"Medium threshold",
			`{"relation": "similarity", "similarity": {"threshold": 0.8}}`,
		},
		{
			"High threshold",
			`{"relation": "similarity", "similarity": {"threshold": 0.95}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccThoughtPathDataSourceConfigWithParameters(tc.parameters),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttrPair("data.tama_thought_path.test", "target_class_id", "tama_class.test_class", "id"),
							resource.TestCheckResourceAttrSet("data.tama_thought_path.test", "id"),
							resource.TestCheckResourceAttrSet("data.tama_thought_path.test", "thought_id"),
							resource.TestCheckResourceAttrSet("data.tama_thought_path.test", "parameters"),
						),
					},
				},
			})
		})
	}
}

func TestAccThoughtPathDataSource_FilteringOptions(t *testing.T) {
	filteringParams := `{
		"relation": "similarity",
		"filtering": {
			"enabled": true,
			"min_score": 0.7,
			"max_results": 50,
			"exclude_patterns": ["test", "debug"]
		}
	}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtPathDataSourceConfigWithParameters(filteringParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tama_thought_path.test", "target_class_id", "tama_class.test_class", "id"),
					resource.TestCheckResourceAttrSet("data.tama_thought_path.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_thought_path.test", "thought_id"),
					resource.TestCheckResourceAttrSet("data.tama_thought_path.test", "parameters"),
				),
			},
		},
	})
}

func testAccThoughtPathDataSourceConfig() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-path-ds-%d"
  type = "root"
}

resource "tama_class" "test_class" {
  space_id = tama_space.test_space.id
  schema_json = jsonencode({
    title       = "Test Path Target Schema"
    description = "Schema for path target"
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

resource "tama_chain" "test_chain" {
  space_id = tama_space.test_space.id
  name     = "test-chain-for-path-ds"
}

resource "tama_modular_thought" "test" {
  chain_id = tama_chain.test_chain.id
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

data "tama_thought_path" "test" {
  id = tama_thought_path.test.id
}
`, timestamp)
}

func testAccThoughtPathDataSourceConfigWithParameters(parameters string) string {
	timestamp := time.Now().UnixNano()
	config := acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-path-ds-%d"
  type = "root"
}

resource "tama_class" "test_class" {
  space_id = tama_space.test_space.id
  schema_json = jsonencode({
    title       = "Test Path Target Schema"
    description = "Schema for path target"
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

resource "tama_chain" "test_chain" {
  space_id = tama_space.test_space.id
  name     = "test-chain-for-path-ds"
}

resource "tama_modular_thought" "test_thought" {
  chain_id = tama_chain.test_chain.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_thought_path" "test" {
  thought_id      = tama_modular_thought.test_thought.id
  target_class_id = tama_class.test_class.id`, timestamp)

	if parameters != "" {
		config += fmt.Sprintf(`
  parameters = %[1]q`, parameters)
	}

	config += `
}

data "tama_thought_path" "test" {
  id = tama_thought_path.test.id
}
`

	return config
}
