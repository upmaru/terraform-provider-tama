// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package space_processor_test

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
	data := processor.NeuralProcessorModel{
		Completion: &processor.CompletionConfigModel{
			Temperature: types.Float64Value(0.8),
		},
	}

	processorType := processor.DetermineProcessorType(&data)
	if processorType != "completion" {
		t.Errorf("Expected completion type, got %s", processorType)
	}

	// Test invalid - no config
	data = processor.NeuralProcessorModel{}

	processorType = processor.DetermineProcessorType(&data)
	if processorType != "" {
		t.Error("No config should return empty string")
	}

	// Test valid embedding config (type is auto-determined)
	data = processor.NeuralProcessorModel{
		Embedding: &processor.EmbeddingConfigModel{
			MaxTokens: types.Int64Value(512),
		},
	}

	processorType = processor.DetermineProcessorType(&data)
	if processorType != "embedding" {
		t.Errorf("Expected embedding type, got %s", processorType)
	}

	// Test valid reranking config
	data = processor.NeuralProcessorModel{
		Reranking: &processor.RerankingConfigModel{
			Parameters: types.StringValue("{\"top_n\":5}"),
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

func TestAccSpaceProcessorResource_Completion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSpaceProcessorResourceConfig_Completion(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "id"),
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "space_id"),
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "model_id"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "type", "completion"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "completion.temperature", "0.7"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "completion.tool_choice", "auto"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "completion.role_mappings.#", "2"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "completion.role_mappings.0.from", "user"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "completion.role_mappings.0.to", "human"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "completion.role_mappings.1.from", "assistant"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "completion.role_mappings.1.to", "ai"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_space_processor.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccSpaceProcessorImportStateIdFunc,
			},
			// Update and Read testing
			{
				Config: testAccSpaceProcessorResourceConfig_CompletionUpdated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "id"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "type", "completion"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "completion.temperature", "0.9"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "completion.tool_choice", "required"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "completion.role_mappings.#", "1"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "completion.role_mappings.0.from", "user"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "completion.role_mappings.0.to", "human"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccSpaceProcessorResource_CompletionWithDefaults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing - verify server defaults are synced
			{
				Config: testAccSpaceProcessorResourceConfig_CompletionWithDefaults(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "id"),
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "space_id"),
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "model_id"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "type", "completion"),
					// Verify server default for tool_choice is reflected in state
					resource.TestCheckResourceAttr("tama_space_processor.test", "completion.tool_choice", "required"),
				),
			},
		},
	})
}

func TestAccSpaceProcessorResource_EmbeddingWithDefaults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing - verify server defaults are synced
			{
				Config: testAccSpaceProcessorResourceConfig_EmbeddingWithDefaults(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "id"),
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "space_id"),
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "model_id"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "type", "embedding"),
					// Verify server default for max_tokens is reflected in state
					resource.TestCheckResourceAttr("tama_space_processor.test", "embedding.max_tokens", "512"),
					// Check that templates are included
					resource.TestCheckResourceAttr("tama_space_processor.test", "embedding.templates.#", "1"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "embedding.templates.0.type", "query"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "embedding.templates.0.content", "Query: {text}"),
				),
			},
		},
	})
}

func TestAccSpaceProcessorResource_Embedding(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSpaceProcessorResourceConfig_Embedding(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "id"),
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "space_id"),
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "model_id"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "type", "embedding"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "embedding.max_tokens", "1024"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "embedding.templates.#", "2"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "embedding.templates.0.type", "query"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "embedding.templates.0.content", "Query: {text}"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "embedding.templates.1.type", "document"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "embedding.templates.1.content", "Document: {text}"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_space_processor.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccSpaceProcessorImportStateIdFunc,
			},
			// Update and Read testing
			{
				Config: testAccSpaceProcessorResourceConfig_EmbeddingUpdated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "id"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "type", "embedding"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "embedding.max_tokens", "512"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "embedding.templates.#", "1"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "embedding.templates.0.type", "query"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "embedding.templates.0.content", "Search: {text}"),
				),
			},
		},
	})
}

func TestAccSpaceProcessorResource_Reranking(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSpaceProcessorResourceConfig_Reranking(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "id"),
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "space_id"),
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "model_id"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "type", "reranking"),
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "reranking.parameters"),
					resource.TestCheckResourceAttrWith("tama_space_processor.test", "reranking.parameters", func(value string) error {
						var params map[string]any
						if err := json.Unmarshal([]byte(value), &params); err != nil {
							return fmt.Errorf("parameters is not valid JSON: %v", err)
						}
						if params["top_n"] != 5.0 {
							return fmt.Errorf("expected top_n 5, got %v", params["top_n"])
						}
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_space_processor.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccSpaceProcessorImportStateIdFunc,
			},
			// Update and Read testing
			{
				Config: testAccSpaceProcessorResourceConfig_RerankingUpdated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "id"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "type", "reranking"),
					resource.TestCheckResourceAttrWith("tama_space_processor.test", "reranking.parameters", func(value string) error {
						var params map[string]any
						if err := json.Unmarshal([]byte(value), &params); err != nil {
							return fmt.Errorf("parameters is not valid JSON: %v", err)
						}
						if params["top_n"] != 10.0 {
							return fmt.Errorf("expected top_n 10, got %v", params["top_n"])
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccSpaceProcessorResource_NoConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSpaceProcessorResourceConfig_NoConfig(),
				ExpectError: regexp.MustCompile("Exactly one of these attributes must be configured"),
			},
		},
	})
}

func TestAccSpaceProcessorResource_MissingConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSpaceProcessorResourceConfig_MissingConfig(),
				ExpectError: regexp.MustCompile("Exactly one of these attributes must be configured"),
			},
		},
	})
}

func TestAccSpaceProcessorResource_MultipleConfigs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSpaceProcessorResourceConfig_MultipleConfigs(),
				ExpectError: regexp.MustCompile("Exactly one of these attributes must be configured"),
			},
		},
	})
}

func TestAccSpaceProcessorResource_CompletionEmbeddingConflict(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSpaceProcessorResourceConfig_CompletionEmbeddingConflict(),
				ExpectError: regexp.MustCompile("Exactly one of these attributes must be configured"),
			},
		},
	})
}

func TestAccSpaceProcessorResource_CompletionRerankingConflict(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSpaceProcessorResourceConfig_CompletionRerankingConflict(),
				ExpectError: regexp.MustCompile("Exactly one of these attributes must be configured"),
			},
		},
	})
}

func TestAccSpaceProcessorResource_EmbeddingRerankingConflict(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSpaceProcessorResourceConfig_EmbeddingRerankingConflict(),
				ExpectError: regexp.MustCompile("Exactly one of these attributes must be configured"),
			},
		},
	})
}

func TestAccSpaceProcessorResource_AutoTypeDetection(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceProcessorResourceConfig_AutoTypeDetection(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "id"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "type", "embedding"),
				),
			},
		},
	})
}

func TestAccSpaceProcessorResource_InvalidToolChoice(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSpaceProcessorResourceConfig_InvalidToolChoice(),
				ExpectError: regexp.MustCompile(`Attribute completion\.tool_choice value must be one of`),
			},
		},
	})
}

func TestAccSpaceProcessorResource_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceProcessorResourceConfig_Multiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// First processor (completion)
					resource.TestCheckResourceAttrSet("tama_space_processor.completion", "id"),
					resource.TestCheckResourceAttr("tama_space_processor.completion", "type", "completion"),
					resource.TestCheckResourceAttr("tama_space_processor.completion", "completion.temperature", "0.8"),
					// Second processor (embedding)
					resource.TestCheckResourceAttrSet("tama_space_processor.embedding", "id"),
					resource.TestCheckResourceAttr("tama_space_processor.embedding", "type", "embedding"),
					resource.TestCheckResourceAttr("tama_space_processor.embedding", "embedding.max_tokens", "512"),
					// Third processor (reranking)
					resource.TestCheckResourceAttrSet("tama_space_processor.reranking", "id"),
					resource.TestCheckResourceAttr("tama_space_processor.reranking", "type", "reranking"),
					resource.TestCheckResourceAttrWith("tama_space_processor.reranking", "reranking.parameters", func(value string) error {
						var params map[string]any
						if err := json.Unmarshal([]byte(value), &params); err != nil {
							return fmt.Errorf("parameters is not valid JSON: %v", err)
						}
						if params["top_n"] != 3.0 {
							return fmt.Errorf("expected top_n 3, got %v", params["top_n"])
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccSpaceProcessorResource_CompletionWithParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSpaceProcessorResourceConfig_CompletionWithParameters(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "id"),
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "space_id"),
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "model_id"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "type", "completion"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "completion.temperature", "0.7"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "completion.tool_choice", "auto"),
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "completion.parameters"),
					// Check that the parameters contain the expected values
					resource.TestCheckResourceAttrWith("tama_space_processor.test", "completion.parameters", func(value string) error {
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
				ResourceName:      "tama_space_processor.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccSpaceProcessorImportStateIdFunc,
			},
			// Update and Read testing with different parameters
			{
				Config: testAccSpaceProcessorResourceConfig_CompletionWithParametersUpdated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_space_processor.test", "id"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "type", "completion"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "completion.temperature", "0.9"),
					resource.TestCheckResourceAttr("tama_space_processor.test", "completion.tool_choice", "required"),
					resource.TestCheckResourceAttrWith("tama_space_processor.test", "completion.parameters", func(value string) error {
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
func testAccSpaceProcessorImportStateIdFunc(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["tama_space_processor.test"]
	if !ok {
		return "", fmt.Errorf("not found: %s", "tama_space_processor.test")
	}

	spaceId := rs.Primary.Attributes["space_id"]
	processorType := rs.Primary.Attributes["type"]

	return fmt.Sprintf("%s/%s", spaceId, processorType), nil
}

// Test configuration functions.
func testAccSpaceProcessorResourceConfig_Completion() string {
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

resource "tama_space_processor" "test" {
  space_id = tama_space.test.id
  model_id = tama_model.test.id

  completion {
    temperature = 0.7
    tool_choice = "auto"
    role_mappings = [
      {
        from = "user"
        to   = "human"
      },
      {
        from = "assistant"
        to   = "ai"
      }
    ]
  }
}
`, timestamp, timestamp)
}

func testAccSpaceProcessorResourceConfig_CompletionWithDefaults() string {
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

resource "tama_space_processor" "test" {
  space_id = tama_space.test.id
  model_id = tama_model.test.id

  completion {
    temperature = 0.8
    # tool_choice intentionally omitted to test server default
  }
}
`, timestamp, timestamp)
}

func testAccSpaceProcessorResourceConfig_EmbeddingWithDefaults() string {
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
  identifier = "text-embedding-ada-002"
  path       = "/embeddings"
}

resource "tama_space_processor" "test" {
  space_id = tama_space.test.id
  model_id = tama_model.test.id

  embedding {
    # max_tokens intentionally omitted to test server default
    # Add minimal templates to ensure valid config
    templates = [
      {
        type    = "query"
        content = "Query: {text}"
      }
    ]
  }
}
`, timestamp, timestamp)
}

func testAccSpaceProcessorResourceConfig_CompletionUpdated() string {
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

resource "tama_space_processor" "test" {
  space_id = tama_space.test.id
  model_id = tama_model.test.id

  completion {
    temperature = 0.9
    tool_choice = "required"
    role_mappings = [
      {
        from = "user"
        to   = "human"
      }
    ]
  }
}
`, timestamp, timestamp)
}

func testAccSpaceProcessorResourceConfig_Embedding() string {
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
  identifier = "text-embedding-ada-002"
  path       = "/embeddings"
}

resource "tama_space_processor" "test" {
  space_id = tama_space.test.id
  model_id = tama_model.test.id

  embedding {
    max_tokens = 1024
    templates = [
      {
        type    = "query"
        content = "Query: {text}"
      },
      {
        type    = "document"
        content = "Document: {text}"
      }
    ]
  }
}
`, timestamp, timestamp)
}

func testAccSpaceProcessorResourceConfig_EmbeddingUpdated() string {
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
  identifier = "text-embedding-ada-002"
  path       = "/embeddings"
}

resource "tama_space_processor" "test" {
  space_id = tama_space.test.id
  model_id = tama_model.test.id

  embedding {
    max_tokens = 512
    templates = [
      {
        type    = "query"
        content = "Search: {text}"
      }
    ]
  }
}
`, timestamp, timestamp)
}

func testAccSpaceProcessorResourceConfig_Reranking() string {
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
  endpoint = "https://api.cohere.ai/v1"
  api_key  = "test-key"
}

resource "tama_model" "test" {
  source_id  = tama_source.test.id
  identifier = "rerank-english-v2.0"
  path       = "/rerank"
}

resource "tama_space_processor" "test" {
  space_id = tama_space.test.id
  model_id = tama_model.test.id

  reranking {
    parameters = jsonencode({
      top_n = 5
    })
  }
}
`, timestamp, timestamp)
}

func testAccSpaceProcessorResourceConfig_RerankingUpdated() string {
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
  endpoint = "https://api.cohere.ai/v1"
  api_key  = "test-key"
}

resource "tama_model" "test" {
  source_id  = tama_source.test.id
  identifier = "rerank-english-v2.0"
  path       = "/rerank"
}

resource "tama_space_processor" "test" {
  space_id = tama_space.test.id
  model_id = tama_model.test.id

  reranking {
    parameters = jsonencode({
      top_n = 10
    })
  }
}
`, timestamp, timestamp)
}

func testAccSpaceProcessorResourceConfig_NoConfig() string {
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

resource "tama_space_processor" "test" {
  space_id = tama_space.test.id
  model_id = tama_model.test.id
}
`, timestamp, timestamp)
}

func testAccSpaceProcessorResourceConfig_MissingConfig() string {
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

resource "tama_space_processor" "test" {
  space_id = tama_space.test.id
  model_id = tama_model.test.id
}
`, timestamp, timestamp)
}

func testAccSpaceProcessorResourceConfig_MultipleConfigs() string {
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

resource "tama_space_processor" "test" {
  space_id = tama_space.test.id
  model_id = tama_model.test.id

  completion {
    temperature = 0.7
  }

  embedding {
    max_tokens = 512
  }
}
`, timestamp, timestamp)
}

func testAccSpaceProcessorResourceConfig_AutoTypeDetection() string {
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

resource "tama_space_processor" "test" {
  space_id = tama_space.test.id
  model_id = tama_model.test.id

  embedding {
    max_tokens = 512
  }
}
`, timestamp, timestamp)
}

func testAccSpaceProcessorResourceConfig_InvalidToolChoice() string {
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

resource "tama_space_processor" "test" {
  space_id = tama_space.test.id
  model_id = tama_model.test.id

  completion {
    temperature = 0.7
    tool_choice = "invalid-choice"
  }
}
`, timestamp, timestamp)
}

func testAccSpaceProcessorResourceConfig_CompletionWithParameters() string {
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

resource "tama_space_processor" "test" {
  space_id = tama_space.test.id
  model_id = tama_model.test.id

  completion {
    temperature = 0.7
    tool_choice = "auto"
    parameters = jsonencode({
      reasoning_effort = "low"
    })
  }
}
`, timestamp, timestamp)
}

func testAccSpaceProcessorResourceConfig_CompletionWithParametersUpdated() string {
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

resource "tama_space_processor" "test" {
  space_id = tama_space.test.id
  model_id = tama_model.test.id

  completion {
    temperature = 0.9
    tool_choice = "required"
    parameters = jsonencode({
      reasoning_effort = "high"
      max_tokens = 2000
    })
  }
}
`, timestamp, timestamp)
}

func testAccSpaceProcessorResourceConfig_Multiple() string {
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

resource "tama_model" "completion_model" {
  source_id  = tama_source.test.id
  identifier = "gpt-4"
  path       = "/chat/completions"
}

resource "tama_model" "embedding_model" {
  source_id  = tama_source.test.id
  identifier = "text-embedding-ada-002"
  path       = "/embeddings"
}

resource "tama_model" "reranking_model" {
  source_id  = tama_source.test.id
  identifier = "rerank-english-v2.0"
  path       = "/rerank"
}

resource "tama_space_processor" "completion" {
  space_id = tama_space.test.id
  model_id = tama_model.completion_model.id

  completion {
    temperature = 0.8
    tool_choice = "auto"
  }
}

resource "tama_space_processor" "embedding" {
  space_id = tama_space.test.id
  model_id = tama_model.embedding_model.id

  embedding {
    max_tokens = 512
  }
}

resource "tama_space_processor" "reranking" {
  space_id = tama_space.test.id
  model_id = tama_model.reranking_model.id

  reranking {
    parameters = jsonencode({
      top_n = 3
    })
  }
}
`, timestamp, timestamp)
}

func testAccSpaceProcessorResourceConfig_CompletionEmbeddingConflict() string {
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

resource "tama_space_processor" "test" {
  space_id = tama_space.test.id
  model_id = tama_model.test.id

  completion {
    temperature = 0.7
    tool_choice = "auto"
  }

  embedding {
    max_tokens = 512
    templates = [
      {
        type    = "query"
        content = "Query: {text}"
      }
    ]
  }
}
`, timestamp, timestamp)
}

func testAccSpaceProcessorResourceConfig_CompletionRerankingConflict() string {
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

resource "tama_space_processor" "test" {
  space_id = tama_space.test.id
  model_id = tama_model.test.id

  completion {
    temperature = 0.8
  }

  reranking {
    parameters = jsonencode({
      top_n = 5
    })
  }
}
`, timestamp, timestamp)
}

func testAccSpaceProcessorResourceConfig_EmbeddingRerankingConflict() string {
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
  identifier = "text-embedding-ada-002"
  path       = "/embeddings"
}

resource "tama_space_processor" "test" {
  space_id = tama_space.test.id
  model_id = tama_model.test.id

  embedding {
    max_tokens = 1024
  }

  reranking {
    parameters = jsonencode({
      top_n = 3
    })
  }
}
`, timestamp, timestamp)
}
