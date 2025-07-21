// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package space_processor_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccSpaceProcessorDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceProcessorDataSourceConfig("completion"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_space_processor.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_space_processor.test", "space_id"),
					resource.TestCheckResourceAttrSet("data.tama_space_processor.test", "model_id"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "type", "completion"),
					resource.TestCheckResourceAttrSet("data.tama_space_processor.test", "current_state"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "completion_config.#", "1"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "embedding_config.#", "0"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "reranking_config.#", "0"),
				),
			},
		},
	})
}

func TestAccSpaceProcessorDataSource_CompletionType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceProcessorDataSourceConfig("completion"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "type", "completion"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "completion_config.#", "1"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "embedding_config.#", "0"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "reranking_config.#", "0"),
				),
			},
		},
	})
}

func TestAccSpaceProcessorDataSource_EmbeddingType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceProcessorDataSourceConfig("embedding"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "type", "embedding"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "embedding_config.#", "1"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "completion_config.#", "0"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "reranking_config.#", "0"),
				),
			},
		},
	})
}

func TestAccSpaceProcessorDataSource_RerankingType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceProcessorDataSourceConfig("reranking"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "type", "reranking"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "reranking_config.#", "1"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "completion_config.#", "0"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "embedding_config.#", "0"),
				),
			},
		},
	})
}

func TestAccSpaceProcessorDataSource_CompletionWithRoleMappings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceProcessorDataSourceConfigWithRoleMappings(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "type", "completion"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "completion_config.#", "1"),
				),
			},
		},
	})
}

func TestAccSpaceProcessorDataSource_EmbeddingWithTemplates(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceProcessorDataSourceConfigWithTemplates(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "type", "embedding"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "embedding_config.#", "1"),
				),
			},
		},
	})
}

func TestAccSpaceProcessorDataSource_MultipleConfigurations(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test completion processor
			{
				Config: testAccSpaceProcessorDataSourceConfig("completion"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "type", "completion"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "completion_config.#", "1"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "embedding_config.#", "0"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "reranking_config.#", "0"),
				),
			},
			// Test embedding processor
			{
				Config: testAccSpaceProcessorDataSourceConfig("embedding"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "type", "embedding"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "completion_config.#", "0"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "embedding_config.#", "1"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "reranking_config.#", "0"),
				),
			},
			// Test reranking processor
			{
				Config: testAccSpaceProcessorDataSourceConfig("reranking"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "type", "reranking"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "completion_config.#", "0"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "embedding_config.#", "0"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "reranking_config.#", "1"),
				),
			},
		},
	})
}

func TestAccSpaceProcessorDataSource_DefaultValues(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceProcessorDataSourceConfigWithDefaults(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "type", "completion"),
					resource.TestCheckResourceAttr("data.tama_space_processor.test", "completion_config.#", "1"),
				),
			},
		},
	})
}

func testAccSpaceProcessorDataSourceConfig(processorType string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-processor-ds-%d"
  type = "root"
}

resource "tama_source" "test_source" {
  space_id = tama_space.test_space.id
  name     = "test-source-for-processor-ds"
  type     = "model"
  endpoint = "https://api.openai.com"
  api_key  = "test-api-key"
}

resource "tama_model" "test_model" {
  source_id  = tama_source.test_source.id
  identifier = "gpt-4"
  path       = "/chat/completions"
}

resource "tama_space_processor" "test" {
  space_id = tama_space.test_space.id
  model_id = tama_model.test_model.id

  dynamic "completion_config" {
    for_each = %[3]q == "completion" ? [1] : []
    content {
      temperature = 0.8
      tool_choice = "required"
    }
  }

  dynamic "embedding_config" {
    for_each = %[3]q == "embedding" ? [1] : []
    content {
      max_tokens = 512
    }
  }

  dynamic "reranking_config" {
    for_each = %[3]q == "reranking" ? [1] : []
    content {
      top_n = 3
    }
  }
}

data "tama_space_processor" "test" {
  space_id = tama_space_processor.test.space_id
  model_id = tama_space_processor.test.model_id
}
`, timestamp, timestamp, processorType)
}

func testAccSpaceProcessorDataSourceConfigWithRoleMappings() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-processor-ds-%d"
  type = "root"
}

resource "tama_source" "test_source" {
  space_id = tama_space.test_space.id
  name     = "test-source-for-processor-ds"
  type     = "model"
  endpoint = "https://api.openai.com"
  api_key  = "test-api-key"
}

resource "tama_model" "test_model" {
  source_id  = tama_source.test_source.id
  identifier = "gpt-4"
  path       = "/chat/completions"
}

resource "tama_space_processor" "test" {
  space_id = tama_space.test_space.id
  model_id = tama_model.test_model.id

  completion_config {
    temperature = 0.7
    tool_choice = "auto"

    role_mappings = [
      {
        from = "user"
        to   = "human"
      },
      {
        from = "assistant"
        to   = "assistant"
      }
    ]
  }
}

data "tama_space_processor" "test" {
  space_id = tama_space_processor.test.space_id
  model_id = tama_space_processor.test.model_id
}
`, timestamp)
}

func testAccSpaceProcessorDataSourceConfigWithTemplates() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-processor-ds-%d"
  type = "root"
}

resource "tama_source" "test_source" {
  space_id = tama_space.test_space.id
  name     = "test-source-for-processor-ds"
  type     = "model"
  endpoint = "https://api.openai.com"
  api_key  = "test-api-key"
}

resource "tama_model" "test_model" {
  source_id  = tama_source.test_source.id
  identifier = "text-embedding-ada-002"
  path       = "/embeddings"
}

resource "tama_space_processor" "test" {
  space_id = tama_space.test_space.id
  model_id = tama_model.test_model.id

  embedding_config {
    max_tokens = 1024

    templates = [
      {
        type    = "query"
        content = "Represent this sentence for searching relevant passages: {input}"
      },
      {
        type    = "document"
        content = "Represent this document for retrieval: {input}"
      }
    ]
  }
}

data "tama_space_processor" "test" {
  space_id = tama_space_processor.test.space_id
  model_id = tama_space_processor.test.model_id
}
`, timestamp)
}

func testAccSpaceProcessorDataSourceConfigWithDefaults() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-processor-ds-%d"
  type = "root"
}

resource "tama_source" "test_source" {
  space_id = tama_space.test_space.id
  name     = "test-source-for-processor-ds"
  type     = "model"
  endpoint = "https://api.openai.com"
  api_key  = "test-api-key"
}

resource "tama_model" "test_model" {
  source_id  = tama_source.test_source.id
  identifier = "gpt-4"
  path       = "/chat/completions"
}

resource "tama_space_processor" "test" {
  space_id = tama_space.test_space.id
  model_id = tama_model.test_model.id

  completion_config {
    temperature = 0.8
    tool_choice = "required"
  }
}

data "tama_space_processor" "test" {
  space_id = tama_space_processor.test.space_id
  model_id = tama_space_processor.test.model_id
}
`, timestamp)
}
