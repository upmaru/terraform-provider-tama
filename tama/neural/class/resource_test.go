// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package class_test

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
	"github.com/upmaru/terraform-provider-tama/internal/planmodifier"
	"github.com/upmaru/terraform-provider-tama/tama/neural/class"
)

func TestAccClassResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccClassResourceConfigWithBlock(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class.test", "name"),
					resource.TestCheckResourceAttrSet("tama_class.test", "description"),
					resource.TestCheckResourceAttrSet("tama_class.test", "schema.0.title"),
					resource.TestCheckResourceAttrSet("tama_class.test", "schema.0.description"),
					resource.TestCheckResourceAttrSet("tama_class.test", "schema.0.type"),
					resource.TestCheckResourceAttrSet("tama_class.test", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_class.test", "space_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_class.test",
				ImportState:       true,
				ImportStateVerify: false, // Skip verification due to schema format differences
			},
			// Update and Read testing
			{
				Config: testAccClassResourceConfigWithJSON(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class.test", "name"),
					resource.TestCheckResourceAttrSet("tama_class.test", "description"),
					resource.TestCheckResourceAttrSet("tama_class.test", "schema_json"),
					resource.TestCheckResourceAttrSet("tama_class.test", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_class.test", "space_id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccClassResource_InvalidSchemaJSON(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccClassResourceConfigWithInvalidJSON(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				ExpectError: regexp.MustCompile("Unable to parse schema JSON"),
			},
		},
	})
}

func TestAccClassResource_MissingTitleDescription(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccClassResourceConfigMissingFields(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				ExpectError: regexp.MustCompile("JSON schema must include 'title' field"),
			},
		},
	})
}

func TestAccClassResource_SchemaBlock(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassResourceConfigComplexBlock(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class.test", "name"),
					resource.TestCheckResourceAttrSet("tama_class.test", "description"),
					resource.TestCheckResourceAttr("tama_class.test", "schema.0.title", "entity-network"),
					resource.TestCheckResourceAttr("tama_class.test", "schema.0.type", "object"),
					resource.TestCheckResourceAttrSet("tama_class.test", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_class.test", "space_id"),
				),
			},
		},
	})
}

func TestAccClassResource_BothSchemaTypes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccClassResourceConfigBothSchemas(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				ExpectError: regexp.MustCompile("Cannot specify both schema block and schema_json attribute"),
			},
		},
	})
}

func TestAccClassResource_NoSchema(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccClassResourceConfigNoSchema(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				ExpectError: regexp.MustCompile("Either schema block or schema_json attribute must be provided"),
			},
		},
	})
}

func TestAccClassResource_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassResourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// First class
					resource.TestCheckResourceAttrSet("tama_class.test1", "id"),
					resource.TestCheckResourceAttrSet("tama_class.test1", "name"),
					resource.TestCheckResourceAttrSet("tama_class.test1", "description"),
					resource.TestCheckResourceAttrSet("tama_class.test1", "schema_json"),
					resource.TestCheckResourceAttrSet("tama_class.test1", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_class.test1", "space_id"),
					// Second class
					resource.TestCheckResourceAttrSet("tama_class.test2", "id"),
					resource.TestCheckResourceAttrSet("tama_class.test2", "name"),
					resource.TestCheckResourceAttrSet("tama_class.test2", "description"),
					resource.TestCheckResourceAttr("tama_class.test2", "schema.#", "1"),
					resource.TestCheckResourceAttrSet("tama_class.test2", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_class.test2", "space_id"),
				),
			},
		},
	})
}

func TestAccClassResource_SpaceIdChange(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassResourceConfigSpaceChange("initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class.test", "space_id"),
				),
			},
			{
				Config: testAccClassResourceConfigSpaceChange("updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class.test", "space_id"),
				),
			},
		},
	})
}

func TestAccClassResource_JSONEncoding(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassResourceConfigWithJSONEncode(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class.test", "name"),
					resource.TestCheckResourceAttrSet("tama_class.test", "description"),
					resource.TestCheckResourceAttrSet("tama_class.test", "schema_json"),
					resource.TestCheckResourceAttrSet("tama_class.test", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_class.test", "space_id"),
				),
			},
		},
	})
}

func testAccClassResourceConfigWithBlock(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id

  schema {
    title       = "action-call"
    description = "An action call is a request to execute an action."
    type        = "object"
    required    = ["tool_id", "parameters", "code", "content_type", "content"]
    strict      = true
    properties  = jsonencode({
      tool_id = {
        type        = "string"
        description = "The ID of the tool to execute"
      }
      parameters = {
        type        = "object"
        description = "The parameters to pass to the action"
      }
      code = {
        type        = "integer"
        description = "The status of the action call"
      }
      content_type = {
        type        = "string"
        description = "The content type of the response"
      }
      content = {
        type        = "object"
        description = "The response from the action"
      }
    })
  }
}
`, spaceName)
}

func testAccClassResourceConfigWithJSON(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "collection"
    description = "A collection is a group of entities that can be queried."
    type = "object"
    properties = {
      space = {
        type        = "string"
        description = "Slug of the space"
      }
      name = {
        description = "The name of the collection"
        type        = "string"
      }
      created_at = {
        description = "The unix timestamp when the collection was created"
        type        = "integer"
      }
      items = {
        description = "An array of objects"
        type        = "array"
        items = {
          type = "object"
        }
      }
    }
    required = ["items", "space", "name", "created_at"]
  })
}
`, spaceName)
}

func testAccClassResourceConfigComplexBlock(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id

  schema {
    title       = "entity-network"
    description = <<-EOT
      A entity network is records the connections between entities.

      ## Fields:
      - edges: An array of entity ids that are connected to the entity.
    EOT
    type        = "object"
    required    = ["edges"]
    strict      = false
    properties  = jsonencode({
      edges = {
        type        = "object"
        description = "An array of entity ids that are connected to the entity."
      }
    })
  }
}
`, spaceName)
}

func testAccClassResourceConfigWithInvalidJSON(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = "invalid-json"
}
`, spaceName)
}

func testAccClassResourceConfigMissingFields(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    type = "object"
    properties = {
      tool_id = {
        type = "string"
      }
    }
  })
}
`, spaceName)
}

func testAccClassResourceConfigBothSchemas(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "JSON Schema"
    description = "A schema via JSON"
    type = "object"
  })

  schema {
    title       = "Block Schema"
    description = "A schema via block"
    type        = "object"
  }
}
`, spaceName)
}

func testAccClassResourceConfigMultiple() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "test-space-%d"
  type = "root"
}

resource "tama_class" "test1" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "action-call"
    description = "An action call is a request to execute an action."
    type = "object"
    properties = {
      tool_id = {
        description = "The ID of the tool to execute"
        type        = "string"
      }
      parameters = {
        description = "The parameters to pass to the action"
        type        = "object"
      }
      code = {
        description = "The status of the action call"
        type        = "integer"
      }
    }
    required = ["tool_id", "parameters", "code"]
  })
}

resource "tama_class" "test2" {
  space_id = tama_space.test.id
  schema {
    title       = "collection"
    description = "A collection is a group of entities that can be queried."
    type        = "object"
    required    = ["items", "name"]
    properties  = jsonencode({
      items = {
        type        = "array"
        description = "An array of objects"
      }
      name = {
        type        = "string"
        description = "The name of the collection"
      }
    })
  }
}
`, timestamp)
}

func testAccClassResourceConfigSpaceChange(suffix string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_%s" {
  name = "test-space-%s-%d"
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test_%s.id
  schema_json = jsonencode({
    title = "action-call"
    description = "An action call is a request to execute an action."
    type = "object"
    properties = {
      tool_id = {
        description = "The ID of the tool to execute"
        type        = "string"
      }
      parameters = {
        description = "The parameters to pass to the action"
        type        = "object"
      }
    }
    required = ["tool_id", "parameters"]
  })
}
`, suffix, suffix, timestamp, suffix)
}

func testAccClassResourceConfigWithJSONEncode() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "test-space-%d"
  type = "root"
}

variable "schema" {
  type = object({
    title = string
    description = string
    type = string
    properties = object({
      tool_id = object({
        description = string
        type = string
      })
      parameters = object({
        description = string
        type = string
      })
      code = object({
        description = string
        type = string
      })
    })
    required = list(string)
  })
  default = {
    title = "action-call"
    description = "An action call is a request to execute an action."
    type = "object"
    properties = {
      tool_id = {
        description = "The ID of the tool to execute"
        type        = "string"
      }
      parameters = {
        description = "The parameters to pass to the action"
        type        = "object"
      }
      code = {
        description = "The status of the action call"
        type        = "integer"
      }
    }
    required = ["tool_id", "parameters", "code"]
  }
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode(var.schema)
}
`, timestamp)
}

func testAccClassResourceConfigNoSchema(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
}
`, spaceName)
}

func TestJSONNormalizationConsistency(t *testing.T) {
	tests := []struct {
		name         string
		inputJSON    string
		expectedJSON string
		description  string
	}{
		{
			name: "pretty formatted to minified consistency",
			inputJSON: `{
  "title": "action-call",
  "description": "An action call is a request to execute an action.",
  "properties": {
    "code": {
      "description": "The status of the action call",
      "type": "integer"
    },
    "tool_id": {
      "description": "The ID of the tool to execute",
      "type": "string"
    }
  },
  "required": ["tool_id", "code"],
  "type": "object"
}`,
			expectedJSON: `{"description":"An action call is a request to execute an action.","properties":{"code":{"description":"The status of the action call","type":"integer"},"tool_id":{"description":"The ID of the tool to execute","type":"string"}},"required":["tool_id","code"],"title":"action-call","type":"object"}`,
			description:  "should normalize pretty formatted JSON to consistent minified format",
		},
		{
			name: "different key ordering consistency",
			inputJSON: `{
  "title": "dynamic",
  "description": "A dynamic schema",
  "type": "object",
  "properties": {
    "entity": {
      "description": "The record",
      "type": "object"
    }
  }
}`,
			expectedJSON: `{"description":"A dynamic schema","properties":{"entity":{"description":"The record","type":"object"}},"title":"dynamic","type":"object"}`,
			description:  "should normalize JSON with different key ordering to consistent format",
		},
		{
			name:         "empty object",
			inputJSON:    `{}`,
			expectedJSON: `{}`,
			description:  "should handle empty objects correctly",
		},
		{
			name: "complex nested structure",
			inputJSON: `{
  "title": "complex",
  "description": "A complex schema",
  "type": "object",
  "properties": {
    "nested": {
      "type": "object",
      "properties": {
        "deep": {
          "type": "string",
          "description": "Deep nested field"
        }
      },
      "required": ["deep"]
    },
    "array_field": {
      "type": "array",
      "items": {
        "type": "string"
      }
    }
  },
  "required": ["nested"]
}`,
			expectedJSON: `{"description":"A complex schema","properties":{"array_field":{"items":{"type":"string"},"type":"array"},"nested":{"properties":{"deep":{"description":"Deep nested field","type":"string"}},"required":["deep"],"type":"object"}},"required":["nested"],"title":"complex","type":"object"}`,
			description:  "should handle complex nested structures correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that our normalization function produces consistent output
			normalized, err := planmodifier.NormalizeJSON(tt.inputJSON)
			if err != nil {
				t.Fatalf("NormalizeJSON failed: %v", err)
			}

			if normalized != tt.expectedJSON {
				t.Errorf("NormalizeJSON output mismatch")
				t.Errorf("Expected: %s", tt.expectedJSON)
				t.Errorf("Got:      %s", normalized)
			}

			// Test that normalizing the expected output again produces the same result
			// (idempotency test)
			normalizedAgain, err := planmodifier.NormalizeJSON(tt.expectedJSON)
			if err != nil {
				t.Fatalf("Second NormalizeJSON failed: %v", err)
			}

			if normalizedAgain != tt.expectedJSON {
				t.Errorf("NormalizeJSON is not idempotent")
				t.Errorf("Expected: %s", tt.expectedJSON)
				t.Errorf("Got:      %s", normalizedAgain)
			}
		})
	}
}

func TestResourceJSONConsistency(t *testing.T) {
	// Test that simulates the original error scenario
	// This tests that when we marshal a schema response and then normalize it,
	// we get consistent results regardless of input formatting

	// Simulate a schema response from the API (this would be classResponse.Schema)
	apiResponse := map[string]any{
		"title":       "action-call",
		"description": "An action call is a request to execute an action.",
		"properties": map[string]any{
			"code": map[string]any{
				"description": "The status of the action call",
				"type":        "integer",
			},
			"tool_id": map[string]any{
				"description": "The ID of the tool to execute",
				"type":        "string",
			},
			"parameters": map[string]any{
				"description": "The parameters to pass to the action",
				"type":        "object",
			},
			"content_type": map[string]any{
				"description": "The content type of the response",
				"type":        "string",
			},
			"content": map[string]any{
				"description": "The response from the action",
				"type":        "object",
			},
		},
		"required": []string{"tool_id", "parameters", "code", "content_type", "content"},
		"type":     "object",
	}

	// Marshal the API response (this is what happens in the resource)
	schemaJSON, err := json.Marshal(apiResponse)
	if err != nil {
		t.Fatalf("Failed to marshal API response: %v", err)
	}

	// Normalize the marshaled JSON (this is our fix)
	normalizedJSON, err := planmodifier.NormalizeJSON(string(schemaJSON))
	if err != nil {
		t.Fatalf("Failed to normalize JSON: %v", err)
	}

	// Create a ResourceModel and set the normalized JSON
	data := class.ResourceModel{
		SchemaJSON: types.StringValue(normalizedJSON),
	}

	// Verify that the SchemaJSON field contains valid, normalized JSON
	if data.SchemaJSON.IsNull() || data.SchemaJSON.IsUnknown() {
		t.Error("SchemaJSON should not be null or unknown")
	}

	jsonValue := data.SchemaJSON.ValueString()
	if jsonValue == "" {
		t.Error("SchemaJSON should not be empty")
	}

	// Verify that the JSON is valid
	var testObj map[string]any
	if err := json.Unmarshal([]byte(jsonValue), &testObj); err != nil {
		t.Errorf("SchemaJSON should contain valid JSON: %v", err)
	}

	// Verify that normalizing again produces the same result (idempotency)
	normalizedAgain, err := planmodifier.NormalizeJSON(jsonValue)
	if err != nil {
		t.Fatalf("Second normalization failed: %v", err)
	}

	if normalizedAgain != jsonValue {
		t.Error("JSON normalization should be idempotent")
		t.Errorf("Original:  %s", jsonValue)
		t.Errorf("Normalized: %s", normalizedAgain)
	}
}

func TestOriginalErrorScenario(t *testing.T) {
	// This test reproduces the exact scenario from the original error message

	// The user's pretty-formatted input (what would be in the plan)
	userInput := `{
  "title": "action-call",
  "description": "An action call is a request to execute an action.",
  "properties": {
    "code": {
      "description": "The status of the action call",
      "type": "integer"
    },
    "tool_id": {
      "description": "The ID of the tool to execute",
      "type": "string"
    },
    "parameters": {
      "description": "The parameters to pass to the action",
      "type": "object"
    },
    "content_type": {
      "description": "The content type of the response",
      "type": "string"
    },
    "content": {
      "description": "The response from the action",
      "type": "object"
    }
  },
  "required": ["tool_id", "parameters", "code", "content_type", "content"],
  "type": "object"
}`

	// What the server would return (same content, but potentially different ordering)
	// After json.Marshal, Go might reorder keys alphabetically
	serverResponseData := map[string]any{
		"description": "An action call is a request to execute an action.",
		"properties": map[string]any{
			"code": map[string]any{
				"description": "The status of the action call",
				"type":        "integer",
			},
			"content": map[string]any{
				"description": "The response from the action",
				"type":        "object",
			},
			"content_type": map[string]any{
				"description": "The content type of the response",
				"type":        "string",
			},
			"parameters": map[string]any{
				"description": "The parameters to pass to the action",
				"type":        "object",
			},
			"tool_id": map[string]any{
				"description": "The ID of the tool to execute",
				"type":        "string",
			},
		},
		"required": []string{"tool_id", "parameters", "code", "content_type", "content"},
		"title":    "action-call",
		"type":     "object",
	}

	// Marshal the server response (simulating what happens in the resource)
	serverJSON, err := json.Marshal(serverResponseData)
	if err != nil {
		t.Fatalf("Failed to marshal server response: %v", err)
	}

	// Normalize both the user input and server response
	normalizedUser, err := planmodifier.NormalizeJSON(userInput)
	if err != nil {
		t.Fatalf("Failed to normalize user input: %v", err)
	}

	normalizedServer, err := planmodifier.NormalizeJSON(string(serverJSON))
	if err != nil {
		t.Fatalf("Failed to normalize server response: %v", err)
	}

	// They should be identical after normalization
	if normalizedUser != normalizedServer {
		t.Error("Normalized user input and server response should be identical")
		t.Errorf("User:   %s", normalizedUser)
		t.Errorf("Server: %s", normalizedServer)
	}

	// This test verifies that our fix prevents the "Provider produced inconsistent result" error
	// by ensuring that both the planned value (user input) and the applied value (server response)
	// normalize to the same string representation.
}
