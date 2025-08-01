// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package thought_processor_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccThoughtProcessorDataSource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccThoughtProcessorDataSourceConfig("completion"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttrSet("data.tama_thought_processor.test", "id"),
                    resource.TestCheckResourceAttrSet("data.tama_thought_processor.test", "thought_id"),
                    resource.TestCheckResourceAttrSet("data.tama_thought_processor.test", "model_id"),
                    resource.TestCheckResourceAttr("data.tama_thought_processor.test", "type", "completion"),
                    resource.TestCheckResourceAttrSet("data.tama_thought_processor.test", "provision_state"),
                    resource.TestCheckResourceAttr("data.tama_thought_processor.test", "completion_config.0.temperature", "0.8"),
                    resource.TestCheckResourceAttr("data.tama_thought_processor.test", "embedding_config.#", "0"),
                    resource.TestCheckResourceAttr("data.tama_thought_processor.test", "reranking_config.#", "0"),
                ),
            },
        },
    })
}

func TestAccThoughtProcessorDataSource_CompletionType(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccThoughtProcessorDataSourceConfig("completion"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("data.tama_thought_processor.test", "type", "completion"),
                    resource.TestCheckResourceAttr("data.tama_thought_processor.test", "completion_config.0.temperature", "0.8"),
                    resource.TestCheckResourceAttr("data.tama_thought_processor.test", "embedding_config.#", "0"),
                    resource.TestCheckResourceAttr("data.tama_thought_processor.test", "reranking_config.#", "0"),
                ),
            },
        },
    })
}

func TestAccThoughtProcessorDataSource_EmbeddingType(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccThoughtProcessorDataSourceConfig("embedding"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("data.tama_thought_processor.test", "type", "embedding"),
                    resource.TestCheckResourceAttr("data.tama_thought_processor.test", "embedding_config.0.max_tokens", "512"),
                    resource.TestCheckResourceAttr("data.tama_thought_processor.test", "completion_config.#", "0"),
                    resource.TestCheckResourceAttr("data.tama_thought_processor.test", "reranking_config.#", "0"),
                ),
            },
        },
    })
}

func TestAccThoughtProcessorDataSource_RerankingType(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccThoughtProcessorDataSourceConfig("reranking"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("data.tama_thought_processor.test", "type", "reranking"),
                    resource.TestCheckResourceAttr("data.tama_thought_processor.test", "reranking_config.0.top_n", "3"),
                    resource.TestCheckResourceAttr("data.tama_thought_processor.test", "completion_config.#", "0"),
                    resource.TestCheckResourceAttr("data.tama_thought_processor.test", "embedding_config.#", "0"),
                ),
            },
        },
    })
}

func TestAccThoughtProcessorDataSource_CompletionWithRoleMappings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtProcessorDataSourceConfigWithRoleMappings(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_thought_processor.test", "type", "completion"),
					resource.TestCheckResourceAttr("data.tama_thought_processor.test", "completion_config.#", "1"),
				),
			},
		},
	})
}

func TestAccThoughtProcessorDataSource_EmbeddingWithTemplates(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtProcessorDataSourceConfigWithTemplates(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_thought_processor.test", "type", "embedding"),
					resource.TestCheckResourceAttr("data.tama_thought_processor.test", "embedding_config.#", "1"),
				),
			},
		},
	})
}

func TestAccThoughtProcessorDataSource_MultipleConfigurations(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test completion processor
			{
				Config: testAccThoughtProcessorDataSourceConfigMultiple("completion"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_thought_processor.test_completion", "type", "completion"),
					resource.TestCheckResourceAttr("data.tama_thought_processor.test_completion", "completion_config.#", "1"),
					resource.TestCheckResourceAttr("data.tama_thought_processor.test_completion", "embedding_config.#", "0"),
					resource.TestCheckResourceAttr("data.tama_thought_processor.test_completion", "reranking_config.#", "0"),
				),
			},
			// Test embedding processor
			{
				Config: testAccThoughtProcessorDataSourceConfigMultiple("embedding"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_thought_processor.test_embedding", "type", "embedding"),
					resource.TestCheckResourceAttr("data.tama_thought_processor.test_embedding", "completion_config.#", "0"),
					resource.TestCheckResourceAttr("data.tama_thought_processor.test_embedding", "embedding_config.#", "1"),
					resource.TestCheckResourceAttr("data.tama_thought_processor.test_embedding", "reranking_config.#", "0"),
				),
			},
			// Test reranking processor
			{
				Config: testAccThoughtProcessorDataSourceConfigMultiple("reranking"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_thought_processor.test_reranking", "type", "reranking"),
					resource.TestCheckResourceAttr("data.tama_thought_processor.test_reranking", "completion_config.#", "0"),
					resource.TestCheckResourceAttr("data.tama_thought_processor.test_reranking", "embedding_config.#", "0"),
					resource.TestCheckResourceAttr("data.tama_thought_processor.test_reranking", "reranking_config.#", "1"),
				),
			},
		},
	})
}

func TestAccThoughtProcessorDataSource_DefaultValues(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtProcessorDataSourceConfigWithDefaults(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_thought_processor.test", "type", "completion"),
					resource.TestCheckResourceAttr("data.tama_thought_processor.test", "completion_config.#", "1"),
				),
			},
		},
	})
}

func testAccThoughtProcessorDataSourceConfig(processorType string) string {
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

resource "tama_chain" "test_chain" {
  space_id = tama_space.test_space.id
  name     = "Test Processing Chain"
}

resource "tama_thought" "test_thought" {
  chain_id = tama_chain.test_chain.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_thought_processor" "test" {
  thought_id = tama_thought.test_thought.id
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

data "tama_thought_processor" "test" {
  thought_id = tama_thought_processor.test.thought_id
  type     = tama_thought_processor.test.type
}
`, timestamp, timestamp, processorType)
}

func testAccThoughtProcessorDataSourceConfigWithRoleMappings() string {
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

resource "tama_chain" "test_chain" {
  space_id = tama_space.test_space.id
  name     = "Test Processing Chain"
}

resource "tama_thought" "test_thought" {
  chain_id = tama_chain.test_chain.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_thought_processor" "test" {
  thought_id = tama_thought.test_thought.id
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

data "tama_thought_processor" "test" {
  thought_id = tama_thought_processor.test.thought_id
  type     = tama_thought_processor.test.type
}
`, timestamp)
}

func testAccThoughtProcessorDataSourceConfigWithTemplates() string {
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

resource "tama_chain" "test_chain" {
  space_id = tama_space.test_space.id
  name     = "Test Processing Chain"
}

resource "tama_thought" "test_thought" {
  chain_id = tama_chain.test_chain.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_thought_processor" "test" {
  thought_id = tama_thought.test_thought.id
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

data "tama_thought_processor" "test" {
  thought_id = tama_thought_processor.test.thought_id
  type     = tama_thought_processor.test.type
}
`, timestamp)
}

func testAccThoughtProcessorDataSourceConfigWithDefaults() string {
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

resource "tama_chain" "test_chain" {
  space_id = tama_space.test_space.id
  name     = "Test Processing Chain"
}

resource "tama_thought" "test_thought" {
  chain_id = tama_chain.test_chain.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_thought_processor" "test" {
  thought_id = tama_thought.test_thought.id
  model_id = tama_model.test_model.id

  completion_config {
    temperature = 0.8
    tool_choice = "required"
  }
}

data "tama_thought_processor" "test" {
  thought_id = tama_thought_processor.test.thought_id
  type     = tama_thought_processor.test.type
}
`, timestamp)
}

// Helper functions for multiple configuration processor tests
func testAccThoughtProcessorDataSourceConfigMultiple(processorType string) string {
	timestamp := time.Now().UnixNano()

	// Helper to get the appropriate model identifier based on processor type
	modelIdentifier := "gpt-4"
	modelPath := "/chat/completions"

	switch processorType {
	case "embedding":
		modelIdentifier = "text-embedding-ada-002"
		modelPath = "/embeddings"
	case "reranking":
		modelIdentifier = "rerank-multilingual-v3.0"
		modelPath = "/rerank"
	}

	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space_%[1]s" {
  name = "test-space-for-processor-ds-%[1]s-%[2]d"
  type = "root"
}

resource "tama_source" "test_source_%[1]s" {
  space_id = tama_space.test_space_%[1]s.id
  name     = "test-source-for-processor-ds-%[1]s"
  type     = "model"
  endpoint = "https://api.openai.com"
  api_key  = "test-api-key"
}

resource "tama_model" "test_model_%[1]s" {
  source_id  = tama_source.test_source_%[1]s.id
  identifier = "%[3]s"
  path       = "%[4]s"
}

resource "tama_chain" "test_chain_%[1]s" {
  space_id = tama_space.test_space_%[1]s.id
  name     = "Test Processing Chain %[1]s"
}

resource "tama_thought" "test_thought_%[1]s" {
  chain_id = tama_chain.test_chain_%[1]s.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_thought_processor" "test_%[1]s" {
  thought_id = tama_thought.test_thought_%[1]s.id
  model_id = tama_model.test_model_%[1]s.id

  dynamic "completion_config" {
    for_each = "%[1]s" == "completion" ? [1] : []
    content {
      temperature = 0.8
      tool_choice = "required"
    }
  }

  dynamic "embedding_config" {
    for_each = "%[1]s" == "embedding" ? [1] : []
    content {
      max_tokens = 512
    }
  }

  dynamic "reranking_config" {
    for_each = "%[1]s" == "reranking" ? [1] : []
    content {
      top_n = 3
    }
  }
}

data "tama_thought_processor" "test_%[1]s" {
  thought_id = tama_thought_processor.test_%[1]s.thought_id
  type     = tama_thought_processor.test_%[1]s.type
}
`, processorType, timestamp, modelIdentifier, modelPath)
}
