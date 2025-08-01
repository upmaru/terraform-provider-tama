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
	thought_processor "github.com/upmaru/terraform-provider-tama/tama/perception/processor"
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
	r := &thought_processor.Resource{}

	// Test valid completion configuration
	data := thought_processor.ResourceModel{
		CompletionConfig: []thought_processor.CompletionConfigModel{
			{
				Temperature: types.Float64Value(0.8),
			},
		},
	}

	err := r.ValidateConfiguration(data)
	if err != nil {
		t.Errorf("Valid completion config should not error: %v", err)
	}

	// Test invalid - multiple configs
	data = thought_processor.ResourceModel{
		CompletionConfig: []thought_processor.CompletionConfigModel{
			{Temperature: types.Float64Value(0.8)},
		},
		EmbeddingConfig: []thought_processor.EmbeddingConfigModel{
			{MaxTokens: types.Int64Value(512)},
		},
	}

	err = r.ValidateConfiguration(data)
	if err == nil {
		t.Error("Multiple configs should error")
	}

	// Test invalid - no config
	data = thought_processor.ResourceModel{}

	err = r.ValidateConfiguration(data)
	if err == nil {
		t.Error("No config should error")
	}

	// Test valid embedding config (type is auto-determined)
	data = thought_processor.ResourceModel{
		EmbeddingConfig: []thought_processor.EmbeddingConfigModel{
			{MaxTokens: types.Int64Value(512)},
		},
	}

	err = r.ValidateConfiguration(data)
	if err != nil {
		t.Errorf("Valid embedding config should not error: %v", err)
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
					resource.TestCheckResourceAttr("tama_thought_processor.test", "completion_config.0.temperature", "0.7"),
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
					resource.TestCheckResourceAttr("tama_thought_processor.test", "completion_config.0.temperature", "0.9"),
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
				ExpectError: regexp.MustCompile("exactly one configuration block must be provided"),
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
				ExpectError: regexp.MustCompile("only one configuration block can be provided"),
			},
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

  completion_config {
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

  completion_config {
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

  completion_config {
    temperature = 0.7
  }

  embedding_config {
    max_tokens = 512
  }
}
`, timestamp, timestamp)
}
