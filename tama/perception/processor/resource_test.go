// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package thought_processor_test

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
	"github.com/upmaru/terraform-provider-tama/internal/processor"
)

func TestCompletionConfigStructure(t *testing.T) {
	// Test role mappings structure
	roleMappings := []map[string]any{
		{"from": "user", "to": "human"},
		{"from": "assistant", "to": "ai"},
	}

	if len(roleMappings) != 2 {
		t.Errorf("Expected 2 role mappings, got %d", len(roleMappings))
	}

	if roleMappings[0]["from"] != "user" || roleMappings[0]["to"] != "human" {
		t.Errorf("First role mapping incorrect: %+v", roleMappings[0])
	}

	if roleMappings[1]["from"] != "assistant" || roleMappings[1]["to"] != "ai" {
		t.Errorf("Second role mapping incorrect: %+v", roleMappings[1])
	}
}

func TestEmbeddingConfigStructure(t *testing.T) {
	// Test templates structure
	templates := []map[string]any{
		{"type": "query", "content": "Query: {text}"},
		{"type": "document", "content": "Document: {text}"},
	}

	if len(templates) != 2 {
		t.Errorf("Expected 2 templates, got %d", len(templates))
	}

	if templates[0]["type"] != "query" || templates[0]["content"] != "Query: {text}" {
		t.Errorf("First template incorrect: %+v", templates[0])
	}

	if templates[1]["type"] != "document" || templates[1]["content"] != "Document: {text}" {
		t.Errorf("Second template incorrect: %+v", templates[1])
	}
}

func TestValidateConfiguration(t *testing.T) {
	// Test valid completion configuration
	data := processor.PerceptionProcessorModel{
		Completion: &processor.CompletionConfigModel{
			Temperature: types.Float64Value(0.8),
		},
	}

	processorType := processor.DetermineProcessorType(&data)
	if processorType != "completion" {
		t.Errorf("Expected completion type, got %s", processorType)
	}

	// Test invalid - no config
	data = processor.PerceptionProcessorModel{}

	processorType = processor.DetermineProcessorType(&data)
	if processorType != "" {
		t.Error("No config should return empty string")
	}

	// Test valid embedding config (type is auto-determined)
	data = processor.PerceptionProcessorModel{
		Embedding: &processor.EmbeddingConfigModel{
			MaxTokens: types.Int64Value(512),
		},
	}

	processorType = processor.DetermineProcessorType(&data)
	if processorType != "embedding" {
		t.Errorf("Expected embedding type, got %s", processorType)
	}

	// Test valid reranking config
	data = processor.PerceptionProcessorModel{
		Reranking: &processor.RerankingConfigModel{
			TopN: types.Int64Value(5),
		},
	}

	processorType = processor.DetermineProcessorType(&data)
	if processorType != "reranking" {
		t.Errorf("Expected reranking type, got %s", processorType)
	}
}

func TestConfigurationMapping(t *testing.T) {
	// Test that JSON marshaling/unmarshaling works for role mappings
	roleMappings := []any{
		map[string]any{"from": "user", "to": "human"},
		map[string]any{"from": "assistant", "to": "ai"},
	}

	jsonBytes, err := json.Marshal(roleMappings)
	if err != nil {
		t.Fatalf("Failed to marshal role mappings: %v", err)
	}

	var unmarshaled []any
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal role mappings: %v", err)
	}

	if len(unmarshaled) != 2 {
		t.Errorf("Expected 2 items after round-trip, got %d", len(unmarshaled))
	}

	// Test templates
	templates := []any{
		map[string]any{"type": "query", "content": "Query: {text}"},
		map[string]any{"type": "document", "content": "Document: {text}"},
	}

	jsonBytes, err = json.Marshal(templates)
	if err != nil {
		t.Fatalf("Failed to marshal templates: %v", err)
	}

	var unmarshaledTemplates []any
	err = json.Unmarshal(jsonBytes, &unmarshaledTemplates)
	if err != nil {
		t.Fatalf("Failed to unmarshal templates: %v", err)
	}

	if len(unmarshaledTemplates) != 2 {
		t.Errorf("Expected 2 templates after round-trip, got %d", len(unmarshaledTemplates))
	}

	// Test parameters JSON handling
	parameters := map[string]any{
		"reasoning_effort": "low",
		"max_tokens":       1000,
		"temperature":      0.5,
	}

	parametersJSON, err := json.Marshal(parameters)
	if err != nil {
		t.Fatalf("Failed to marshal parameters: %v", err)
	}

	var unmarshaledParameters map[string]any
	err = json.Unmarshal(parametersJSON, &unmarshaledParameters)
	if err != nil {
		t.Fatalf("Failed to unmarshal parameters: %v", err)
	}

	if len(unmarshaledParameters) != 3 {
		t.Errorf("Expected 3 parameters after round-trip, got %d", len(unmarshaledParameters))
	}

	if unmarshaledParameters["reasoning_effort"] != "low" {
		t.Errorf("Expected reasoning_effort 'low', got %v", unmarshaledParameters["reasoning_effort"])
	}
}

func TestAccThoughtProcessorResource_Completion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccThoughtProcessorResourceConfig_Completion(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_processor.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_processor.test", "thought_id"),
					resource.TestCheckResourceAttrSet("tama_thought_processor.test", "model_id"),
					resource.TestCheckResourceAttr("tama_thought_processor.test", "type", "completion"),
					resource.TestCheckResourceAttr("tama_thought_processor.test", "completion.temperature", "0.7"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_thought_processor.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccThoughtProcessorImportStateIdFunc,
			},
			// Update and Read testing
			{
				Config: testAccThoughtProcessorResourceConfig_CompletionUpdated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_processor.test", "id"),
					resource.TestCheckResourceAttr("tama_thought_processor.test", "type", "completion"),
					resource.TestCheckResourceAttr("tama_thought_processor.test", "completion.temperature", "0.9"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccThoughtProcessorResource_NoConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccThoughtProcessorResourceConfig_NoConfig(),
				ExpectError: regexp.MustCompile("Exactly one of these attributes must be configured"),
			},
		},
	})
}

func TestAccThoughtProcessorResource_MultipleConfigs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccThoughtProcessorResourceConfig_MultipleConfigs(),
				ExpectError: regexp.MustCompile("Exactly one of these attributes must be configured"),
			},
		},
	})
}

func TestAccThoughtProcessorResource_CompletionEmbeddingConflict(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccThoughtProcessorResourceConfig_CompletionEmbeddingConflict(),
				ExpectError: regexp.MustCompile("Exactly one of these attributes must be configured"),
			},
		},
	})
}

func TestAccThoughtProcessorResource_CompletionRerankingConflict(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccThoughtProcessorResourceConfig_CompletionRerankingConflict(),
				ExpectError: regexp.MustCompile("Exactly one of these attributes must be configured"),
			},
		},
	})
}

func TestAccThoughtProcessorResource_EmbeddingRerankingConflict(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccThoughtProcessorResourceConfig_EmbeddingRerankingConflict(),
				ExpectError: regexp.MustCompile("Exactly one of these attributes must be configured"),
			},
		},
	})
}

func TestAccThoughtProcessorResource_CompletionWithParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccThoughtProcessorResourceConfig_CompletionWithParameters(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_processor.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_processor.test", "thought_id"),
					resource.TestCheckResourceAttrSet("tama_thought_processor.test", "model_id"),
					resource.TestCheckResourceAttr("tama_thought_processor.test", "type", "completion"),
					resource.TestCheckResourceAttr("tama_thought_processor.test", "completion.temperature", "0.7"),
					resource.TestCheckResourceAttrSet("tama_thought_processor.test", "completion.parameters"),
					// Check that the parameters contain the expected values
					resource.TestCheckResourceAttrWith("tama_thought_processor.test", "completion.parameters", func(value string) error {
						var params map[string]any
						if err := json.Unmarshal([]byte(value), &params); err != nil {
							return fmt.Errorf("parameters is not valid JSON: %v", err)
						}
						if params["reasoning_effort"] != "low" {
							return fmt.Errorf("expected reasoning_effort 'low', got %v", params["reasoning_effort"])
						}
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_thought_processor.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccThoughtProcessorImportStateIdFunc,
			},
			// Update and Read testing with different parameters
			{
				Config: testAccThoughtProcessorResourceConfig_CompletionWithParametersUpdated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_thought_processor.test", "id"),
					resource.TestCheckResourceAttr("tama_thought_processor.test", "type", "completion"),
					resource.TestCheckResourceAttr("tama_thought_processor.test", "completion.temperature", "0.9"),
					resource.TestCheckResourceAttrWith("tama_thought_processor.test", "completion.parameters", func(value string) error {
						var params map[string]any
						if err := json.Unmarshal([]byte(value), &params); err != nil {
							return fmt.Errorf("parameters is not valid JSON: %v", err)
						}
						if params["reasoning_effort"] != "high" {
							return fmt.Errorf("expected reasoning_effort 'high', got %v", params["reasoning_effort"])
						}
						if params["max_tokens"] != 2000.0 {
							return fmt.Errorf("expected max_tokens 2000, got %v", params["max_tokens"])
						}
						return nil
					}),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Helper function for import state ID.
func testAccThoughtProcessorImportStateIdFunc(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["tama_thought_processor.test"]
	if !ok {
		return "", fmt.Errorf("not found: %s", "tama_thought_processor.test")
	}

	thoughtID := rs.Primary.Attributes["thought_id"]
	processorType := rs.Primary.Attributes["type"]

	return fmt.Sprintf("%s/%s", thoughtID, processorType), nil
}

// Test configuration functions.
func testAccThoughtProcessorResourceConfig_Completion() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "test-space-%d"
  type = "root"
}

resource "tama_source" "test" {
  space_id = tama_space.test.id
  name     = "test-source-%d"
  type     = "model"
  endpoint = "https://api.openai.com/v1"
  api_key  = "test-key"
}

resource "tama_model" "test" {
  source_id  = tama_source.test.id
  identifier = "gpt-4"
  path       = "/chat/completions"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
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

resource "tama_thought_processor" "test" {
  thought_id = tama_modular_thought.test.id
  model_id   = tama_model.test.id

  completion {
    temperature = 0.7
  }
}
`, timestamp, timestamp)
}

func testAccThoughtProcessorResourceConfig_CompletionUpdated() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "test-space-%d"
  type = "root"
}

resource "tama_source" "test" {
  space_id = tama_space.test.id
  name     = "test-source-%d"
  type     = "model"
  endpoint = "https://api.openai.com/v1"
  api_key  = "test-key"
}

resource "tama_model" "test" {
  source_id  = tama_source.test.id
  identifier = "gpt-4"
  path       = "/chat/completions"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
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

resource "tama_thought_processor" "test" {
  thought_id = tama_modular_thought.test.id
  model_id   = tama_model.test.id

  completion {
    temperature = 0.9
  }
}
`, timestamp, timestamp)
}

func testAccThoughtProcessorResourceConfig_NoConfig() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "test-space-%d"
  type = "root"
}

resource "tama_source" "test" {
  space_id = tama_space.test.id
  name     = "test-source-%d"
  type     = "model"
  endpoint = "https://api.openai.com/v1"
  api_key  = "test-key"
}

resource "tama_model" "test" {
  source_id  = tama_source.test.id
  identifier = "gpt-4"
  path       = "/chat/completions"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
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

resource "tama_thought_processor" "test" {
  thought_id = tama_modular_thought.test.id
  model_id   = tama_model.test.id
}
`, timestamp, timestamp)
}

func testAccThoughtProcessorResourceConfig_CompletionEmbeddingConflict() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "test-space-%d"
  type = "root"
}

resource "tama_source" "test" {
  space_id = tama_space.test.id
  name     = "test-source-%d"
  type     = "model"
  endpoint = "https://api.openai.com/v1"
  api_key  = "test-key"
}

resource "tama_model" "test" {
  source_id  = tama_source.test.id
  identifier = "gpt-4"
  path       = "/chat/completions"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
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

resource "tama_thought_processor" "test" {
  thought_id = tama_modular_thought.test.id
  model_id   = tama_model.test.id

  completion {
    temperature = 0.7
  }

  embedding {
    max_tokens = 512
  }
}
`, timestamp, timestamp)
}

func testAccThoughtProcessorResourceConfig_CompletionRerankingConflict() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "test-space-%d"
  type = "root"
}

resource "tama_source" "test" {
  space_id = tama_space.test.id
  name     = "test-source-%d"
  type     = "model"
  endpoint = "https://api.openai.com/v1"
  api_key  = "test-key"
}

resource "tama_model" "test" {
  source_id  = tama_source.test.id
  identifier = "gpt-4"
  path       = "/chat/completions"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
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

resource "tama_thought_processor" "test" {
  thought_id = tama_modular_thought.test.id
  model_id   = tama_model.test.id

  completion {
    temperature = 0.7
  }

  reranking {
    top_n = 5
  }
}
`, timestamp, timestamp)
}

func testAccThoughtProcessorResourceConfig_EmbeddingRerankingConflict() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "test-space-%d"
  type = "root"
}

resource "tama_source" "test" {
  space_id = tama_space.test.id
  name     = "test-source-%d"
  type     = "model"
  endpoint = "https://api.openai.com/v1"
  api_key  = "test-key"
}

resource "tama_model" "test" {
  source_id  = tama_source.test.id
  identifier = "gpt-4"
  path       = "/chat/completions"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
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

resource "tama_thought_processor" "test" {
  thought_id = tama_modular_thought.test.id
  model_id   = tama_model.test.id

  embedding {
    max_tokens = 512
  }

  reranking {
    top_n = 5
  }
}
`, timestamp, timestamp)
}

func testAccThoughtProcessorResourceConfig_MultipleConfigs() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "test-space-%d"
  type = "root"
}

resource "tama_source" "test" {
  space_id = tama_space.test.id
  name     = "test-source-%d"
  type     = "model"
  endpoint = "https://api.openai.com/v1"
  api_key  = "test-key"
}

resource "tama_model" "test" {
  source_id  = tama_source.test.id
  identifier = "gpt-4"
  path       = "/chat/completions"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
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

resource "tama_thought_processor" "test" {
  thought_id = tama_modular_thought.test.id
  model_id   = tama_model.test.id

  completion {
    temperature = 0.7
  }

  embedding {
    max_tokens = 512
  }
}
`, timestamp, timestamp)
}

func testAccThoughtProcessorResourceConfig_CompletionWithParameters() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "test-space-%d"
  type = "root"
}

resource "tama_source" "test" {
  space_id = tama_space.test.id
  name     = "test-source-%d"
  type     = "model"
  endpoint = "https://api.openai.com/v1"
  api_key  = "test-key"
}

resource "tama_model" "test" {
  source_id  = tama_source.test.id
  identifier = "gpt-4o"
  path       = "/chat/completions"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
}

resource "tama_modular_thought" "test" {
  chain_id = tama_chain.test.id
  relation = "reasoning"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "reasoning"
    })
  }
}

resource "tama_thought_processor" "test" {
  thought_id = tama_modular_thought.test.id
  model_id   = tama_model.test.id

  completion {
    temperature = 0.7
    parameters = jsonencode({
      reasoning_effort = "low"
    })
  }
}
`, timestamp, timestamp)
}

func testAccThoughtProcessorResourceConfig_CompletionWithParametersUpdated() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "test-space-%d"
  type = "root"
}

resource "tama_source" "test" {
  space_id = tama_space.test.id
  name     = "test-source-%d"
  type     = "model"
  endpoint = "https://api.openai.com/v1"
  api_key  = "test-key"
}

resource "tama_model" "test" {
  source_id  = tama_source.test.id
  identifier = "gpt-4o"
  path       = "/chat/completions"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
}

resource "tama_modular_thought" "test" {
  chain_id = tama_chain.test.id
  relation = "reasoning"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "reasoning"
    })
  }
}

resource "tama_thought_processor" "test" {
  thought_id = tama_modular_thought.test.id
  model_id   = tama_model.test.id

  completion {
    temperature = 0.9
    parameters = jsonencode({
      reasoning_effort = "high"
      max_tokens = 2000
    })
  }
}
`, timestamp, timestamp)
}
