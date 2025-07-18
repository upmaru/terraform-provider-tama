// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package model_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccModelDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModelDataSourceConfig("test-model", "/chat/completions"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_model.test", "identifier", "test-model"),
					resource.TestCheckResourceAttrSet("data.tama_model.test", "id"),
					resource.TestCheckResourceAttr("data.tama_model.test", "parameters", ""),
				),
			},
		},
	})
}

func TestAccModelDataSource_WithParameters(t *testing.T) {
	parameters := `{"reasoning_effort": "low", "temperature": 0.8}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModelDataSourceConfigWithParameters("grok-3-mini", "/chat/completions", parameters),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_model.test", "identifier", "grok-3-mini"),
					resource.TestCheckResourceAttrSet("data.tama_model.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_model.test", "parameters"),
				),
			},
		},
	})
}

func TestAccModelDataSource_ComplexParameters(t *testing.T) {
	complexParams := `{
		"temperature": 0.7,
		"max_tokens": 2000,
		"top_p": 0.9,
		"frequency_penalty": 0.1,
		"presence_penalty": 0.1,
		"stream": true,
		"stop": ["\n", "###", "END"],
		"response_format": {
			"type": "json_object"
		}
	}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModelDataSourceConfigWithParameters("gpt-4-turbo", "/chat/completions", complexParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_model.test", "identifier", "gpt-4-turbo"),
					resource.TestCheckResourceAttrSet("data.tama_model.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_model.test", "parameters"),
				),
			},
		},
	})
}

func TestAccModelDataSource_EmptyParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModelDataSourceConfig("simple-model", "/chat/completions"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_model.test", "identifier", "simple-model"),
					resource.TestCheckResourceAttrSet("data.tama_model.test", "id"),
					resource.TestCheckResourceAttr("data.tama_model.test", "parameters", ""),
				),
			},
		},
	})
}

func TestAccModelDataSource_EmbeddingModel(t *testing.T) {
	embeddingParams := `{
		"dimensions": 1536,
		"encoding_format": "float",
		"batch_size": 100
	}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModelDataSourceConfigWithParameters("text-embedding-3-large", "/embeddings", embeddingParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_model.test", "identifier", "text-embedding-3-large"),
					resource.TestCheckResourceAttrSet("data.tama_model.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_model.test", "parameters"),
				),
			},
		},
	})
}

func testAccModelDataSourceConfig(identifier, path string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-model-ds-%d"
  type = "root"
}

resource "tama_source" "test_source" {
  space_id = tama_space.test_space.id
  name     = "test-source-for-model-ds"
  type     = "model"
  endpoint = "https://api.example.com"
  api_key  = "test-api-key"
}

resource "tama_model" "test" {
  source_id  = tama_source.test_source.id
  identifier = %[2]q
  path       = %[3]q
}

data "tama_model" "test" {
  id = tama_model.test.id
}
`, timestamp, identifier, path)
}

func testAccModelDataSourceConfigWithParameters(identifier, path, parameters string) string {
	timestamp := time.Now().UnixNano()
	config := acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-model-ds-%d"
  type = "root"
}

resource "tama_source" "test_source" {
  space_id = tama_space.test_space.id
  name     = "test-source-for-model-ds"
  type     = "model"
  endpoint = "https://api.example.com"
  api_key  = "test-api-key"
}

resource "tama_model" "test" {
  source_id  = tama_source.test_source.id
  identifier = %[2]q
  path       = %[3]q`, timestamp, identifier, path)

	if parameters != "" {
		config += fmt.Sprintf(`
  parameters = %[1]q`, parameters)
	}

	config += `
}

data "tama_model" "test" {
  id = tama_model.test.id
}
`

	return config
}
