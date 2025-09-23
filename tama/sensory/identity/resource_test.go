// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package source_identity_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccSourceIdentityResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSourceIdentityResourceConfig("ApiKey", "test-api-key", "/health", "GET", "[200]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source_identity.test", "identifier", "ApiKey"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "api_key", "test-api-key"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.path", "/health"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.method", "GET"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.codes.#", "1"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.codes.0", "200"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "id"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "specification_id"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "current_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "tama_source_identity.test",
				ImportState:             true,
				ImportStateVerify:       false, // api_key cannot be retrieved from API
				ImportStateVerifyIgnore: []string{"api_key"},
			},
			// Update and Read testing
			{
				Config: testAccSourceIdentityResourceConfig("ApiKey", "updated-api-key", "/status", "POST", "[200, 201]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source_identity.test", "identifier", "ApiKey"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "api_key", "updated-api-key"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.path", "/status"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.method", "POST"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.codes.#", "2"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.codes.0", "200"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.codes.1", "201"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "id"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "current_state"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccSourceIdentityResource_ClientCredentials(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing using client credentials
			{
				Config: testAccSourceIdentityResourceConfigWithClientCredentials("oauth", "test-client-id", "test-client-secret", "/health", "GET", "[200]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source_identity.test", "identifier", "oauth"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "client_id", "test-client-id"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "client_secret", "test-client-secret"),
					resource.TestCheckNoResourceAttr("tama_source_identity.test", "api_key"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.path", "/health"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.method", "GET"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.codes.#", "1"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.codes.0", "200"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "id"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "specification_id"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "current_state"),
				),
			},
		},
	})
}

func TestAccSourceIdentityResource_InvalidValidationPath(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSourceIdentityResourceConfig("ApiKey", "test-api-key", "", "GET", "[200]"),
				ExpectError: regexp.MustCompile("Unable to create source identity"),
			},
		},
	})
}

func TestAccSourceIdentityResource_InvalidValidationMethod(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSourceIdentityResourceConfig("ApiKey", "test-api-key", "/health", "", "[200]"),
				ExpectError: regexp.MustCompile("Unable to create source identity"),
			},
		},
	})
}

func TestAccSourceIdentityResource_EmptyApiKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSourceIdentityResourceConfig("ApiKey", "", "/health", "GET", "[200]"),
				ExpectError: regexp.MustCompile("Unable to create source identity"),
			},
		},
	})
}

func TestAccSourceIdentityResource_EmptyCodes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSourceIdentityResourceConfig("ApiKey", "test-api-key", "/health", "GET", "[]"),
				ExpectError: regexp.MustCompile("Unable to create source identity"),
			},
		},
	})
}

func TestAccSourceIdentityResource_DifferentIdentifiers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceIdentityResourceConfig("ApiKey", "test-api-key", "/health", "GET", "[200]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source_identity.test", "identifier", "ApiKey"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "api_key", "test-api-key"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "current_state"),
				),
			},
			{
				Config: testAccSourceIdentityResourceConfig("ApiKey2", "test-api-key-2", "/status", "POST", "[200]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source_identity.test", "identifier", "ApiKey2"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "api_key", "test-api-key-2"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.path", "/status"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.method", "POST"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "current_state"),
				),
			},
		},
	})
}

func TestAccSourceIdentityResource_DifferentHttpMethods(t *testing.T) {
	testCases := []struct {
		name   string
		method string
	}{
		{"GET method", "GET"},
		{"POST method", "POST"},
		{"PUT method", "PUT"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccSourceIdentityResourceConfig("ApiKey", "test-api-key", "/health", tc.method, "[200]"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("tama_source_identity.test", "validation.method", tc.method),
							resource.TestCheckResourceAttr("tama_source_identity.test", "validation.path", "/health"),
							resource.TestCheckResourceAttrSet("tama_source_identity.test", "provision_state"),
						),
					},
				},
			})
		})
	}
}

func TestAccSourceIdentityResource_MultipleStatusCodes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceIdentityResourceConfig("ApiKey", "test-api-key", "/health", "GET", "[200, 201, 202, 204]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.codes.#", "4"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.codes.0", "200"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.codes.1", "201"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.codes.2", "202"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.codes.3", "204"),
				),
			},
		},
	})
}

func TestAccSourceIdentityResource_DifferentValidationPaths(t *testing.T) {
	testCases := []struct {
		name string
		path string
	}{
		{"simple path", "/health"},
		{"nested path", "/api/v1/health"},
		{"path with query", "/health?check=all"},
		{"path with special chars", "/health-check_endpoint.php"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccSourceIdentityResourceConfig("ApiKey", "test-api-key", tc.path, "GET", "[200]"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("tama_source_identity.test", "validation.path", tc.path),
							resource.TestCheckResourceAttr("tama_source_identity.test", "validation.method", "GET"),
							resource.TestCheckResourceAttrSet("tama_source_identity.test", "provision_state"),
						),
					},
				},
			})
		})
	}
}

func TestAccSourceIdentityResource_SensitiveApiKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceIdentityResourceConfig("ApiKey", "super-secret-api-key-123", "/health", "GET", "[200]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source_identity.test", "identifier", "ApiKey"),
					// Note: We can verify the api_key value in tests, but it's marked as sensitive in the schema
					resource.TestCheckResourceAttr("tama_source_identity.test", "api_key", "super-secret-api-key-123"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccSourceIdentityResource_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceIdentityResourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Test single identity
					resource.TestCheckResourceAttr("tama_source_identity.test1", "identifier", "ApiKey"),
					resource.TestCheckResourceAttr("tama_source_identity.test2", "identifier", "ApiKey2"),
					resource.TestCheckResourceAttr("tama_source_identity.test1", "api_key", "test-api-key-1"),
					resource.TestCheckResourceAttr("tama_source_identity.test1", "validation.path", "/health"),
					resource.TestCheckResourceAttr("tama_source_identity.test1", "validation.method", "GET"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test1", "id"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test1", "provision_state"),
					// Test second identity
					resource.TestCheckResourceAttr("tama_source_identity.test2", "api_key", "test-api-key-2"),
					resource.TestCheckResourceAttr("tama_source_identity.test2", "validation.path", "/status"),
					resource.TestCheckResourceAttr("tama_source_identity.test2", "validation.method", "POST"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test2", "id"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test2", "provision_state"),
				),
			},
		},
	})
}

func TestAccSourceIdentityResource_StateValues(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceIdentityResourceConfig("ApiKey", "test-api-key", "/health", "GET", "[200]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source_identity.test", "identifier", "ApiKey"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "id"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "specification_id"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "current_state"),
					// Verify that states are not empty
					resource.TestMatchResourceAttr("tama_source_identity.test", "provision_state", regexp.MustCompile(".+")),
					resource.TestMatchResourceAttr("tama_source_identity.test", "current_state", regexp.MustCompile(".+")),
				),
			},
		},
	})
}

func TestAccSourceIdentityResource_WaitFor(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceIdentityResourceConfigWaitFor("ApiKey", "test-api-key", "/health", "GET", "[200]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source_identity.test", "identifier", "ApiKey"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "api_key", "test-api-key"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.path", "/health"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.method", "GET"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "current_state"),
				),
			},
		},
	})
}

func TestAccSourceIdentityResource_WaitForMultipleConditions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceIdentityResourceConfigWaitForMultiple("ApiKey", "test-api-key", "/health", "GET", "[200]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source_identity.test", "identifier", "ApiKey"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "api_key", "test-api-key"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.path", "/health"),
					resource.TestCheckResourceAttr("tama_source_identity.test", "validation.method", "GET"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "provision_state"),
					resource.TestCheckResourceAttrSet("tama_source_identity.test", "current_state"),
				),
			},
		},
	})
}

func testAccSourceIdentityResourceConfig(identifier, apiKey, validationPath, validationMethod, validationCodes string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-identity-%d"
  type = "root"
}

resource "tama_specification" "test_spec" {
  space_id = tama_space.test_space.id
  version  = "1.0.0"
  endpoint = "https://elasticsearch.arrakis.upmaru.network"
  schema   = jsonencode(jsondecode(file("${path.module}/testdata/elasticsearch_schema.json")))

  wait_for {
    field {
      name = "current_state"
      in   = ["completed"]
    }
  }
}`, timestamp) + fmt.Sprintf(`

resource "tama_source_identity" "test" {
  specification_id = tama_specification.test_spec.id
  identifier       = %[1]q
  api_key          = %[2]q

  validation {
    path   = %[3]q
    method = %[4]q
    codes  = %[5]s
  }
}
`, identifier, apiKey, validationPath, validationMethod, validationCodes)
}

func testAccSourceIdentityResourceConfigMultiple() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-multiple-identities-%d"
  type = "root"
}

resource "tama_specification" "test_spec" {
  space_id = tama_space.test_space.id
  version  = "1.0.0"
  endpoint = "https://elasticsearch.arrakis.upmaru.network"
  schema   = jsonencode(jsondecode(file("${path.module}/testdata/elasticsearch_schema.json")))

  wait_for {
    field {
      name = "current_state"
      in   = ["completed"]
    }
  }
}`, timestamp) + `

resource "tama_source_identity" "test1" {
  specification_id = tama_specification.test_spec.id
  identifier       = "ApiKey"
  api_key          = "test-api-key-1"

  validation {
    path   = "/health"
    method = "GET"
    codes  = [200]
  }
}

resource "tama_source_identity" "test2" {
  specification_id = tama_specification.test_spec.id
  identifier       = "ApiKey2"
  api_key          = "test-api-key-2"

  validation {
    path   = "/status"
    method = "POST"
    codes  = [200, 201]
  }
}
`
}

func testAccSourceIdentityResourceConfigWithClientCredentials(identifier, clientID, clientSecret, validationPath, validationMethod, validationCodes string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-identity-client-creds-%d"
  type = "root"
}

locals {
  openapi_spec = jsondecode(<<JSON
{
  "components": {
    "responses": {},
    "schemas": {
      "Artifact": {
        "description": "An artifact displays information to the user.",
        "properties": {
          "index": {"description": "The index of the artifact. Lower index value renders first.", "type": "integer"},
          "properties": {"items": {"$ref": "#/components/schemas/ArtifactProperty"}, "type": "array"},
          "reference": {"description": "The tool_call_id from the search results to display.", "type": "string"},
          "type": {"description": "The type of artifact displayed to the user.", "enum": ["list", "table", "grid", "detail", "metric"], "type": "string"}
        },
        "required": ["type", "index", "reference", "properties"],
        "title": "Artifact",
        "type": "object"
      },
      "ArtifactProperty": {
        "description": "The property of the artifact",
        "properties": {
          "name": {"description": "The name of the property.", "type": "string"},
          "relevance": {"description": "The relevance of the property.", "type": "integer"}
        },
        "title": "ArtifactProperty",
        "type": "object"
      },
      "ArtifactRequest": {"description": "POST body for creating artifacts", "properties": {"artifact": {"$ref": "#/components/schemas/Artifact"}}, "title": "ArtifactRequest", "type": "object"},
      "ArtifactResponse": {"description": "The response from the server when creating an artifact", "properties": {"data": {"properties": {"id": {"description": "The unique identifier of the artifact.", "type": "string"}}, "type": "object"}}, "title": "ArtifactResponse", "type": "object"},
      "LanguagePreference": {"description": "A user language preference", "properties": {"locale": {"description": "The language locale", "example": "en-US", "type": "string"}}, "required": ["locale"], "title": "LanguagePreference", "type": "object"},
      "Preference": {"description": "A user preference", "properties": {"type": {"description": "The type of the preference", "enum": ["region", "theme", "language"], "type": "string"}, "value": {"description": "The value of the preference", "oneOf": [{"$ref": "#/components/schemas/RegionPreference"}, {"$ref": "#/components/schemas/ThemePreference"}, {"$ref": "#/components/schemas/LanguagePreference"}], "type": "object"}}, "required": ["type", "value"], "title": "Preference", "type": "object"},
      "PreferenceRequest": {"description": "POST body for creating a user preference", "properties": {"preference": {"$ref": "#/components/schemas/Preference"}}, "required": ["preference"], "title": "PreferenceRequest", "type": "object"},
      "PreferenceResponse": {"description": "Response for creating a user preference", "properties": {"data": {"$ref": "#/components/schemas/Preference"}}, "title": "PreferenceResponse", "type": "object"},
      "PreferencesResponse": {"description": "Response schema for multiple preferences", "properties": {"data": {"description": "List of preferences for a given user", "items": {"$ref": "#/components/schemas/Preference"}, "type": "array"}}, "title": "PreferencesResponse", "type": "object"},
      "RegionPreference": {"description": "A user region preference", "properties": {"iso_alpha2": {"description": "The ISO 3166-1 alpha-2 code of the region", "example": "TH", "type": "string"}, "name": {"description": "The full name of the region", "example": "Thailand", "type": "string"}}, "required": ["name", "iso_alpha2"], "title": "RegionPreference", "type": "object"},
      "ThemePreference": {"description": "A user theme preference", "properties": {"setting": {"description": "The theme setting", "enum": ["light", "dark", "system"], "example": "dark", "type": "string"}}, "required": ["setting"], "title": "ThemePreference", "type": "object"}
    },
    "securitySchemes": {"oauth": {"description": "Authenticate to get access to Tama specific endpoints.", "flows": {"clientCredentials": {"scopes": {"all": "Access to all tama endpoints"}, "tokenUrl": "/tama/auth/tokens"}}, "type": "oauth2"}}
  },
  "info": {"description": "This API provides endpoints for tama engine to interact with the Memovee application.", "title": "Memovee App Tama API", "version": "0.1.0"},
  "openapi": "3.0.0",
  "paths": {
    "/tama/accounts/users/{user_id}/preferences": {
      "get": {"callbacks": {}, "description": "List of preferences for a given user_id", "operationId": "list-user-preferences", "parameters": [{"description": "The ID of the user", "in": "path", "name": "user_id", "required": true, "schema": {"type": "string"}}], "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/PreferencesResponse"}}}, "description": "Preference list response"}}, "security": [{"oauth": ["all"]}], "summary": "List preferences", "tags": []},
      "post": {"callbacks": {}, "description": "Creates a new preference for the specified user", "operationId": "create-user-preference", "parameters": [{"description": "The ID of the user", "in": "path", "name": "user_id", "required": true, "schema": {"type": "string"}}], "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/PreferenceRequest"}}}, "description": "The attribute of the preference", "required": true}, "responses": {"201": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/PreferenceResponse"}}}, "description": "Preference"}}, "security": [{"oauth": ["all"]}], "summary": "Create a preference for a given user", "tags": []}
    },
    "/tama/accounts/users/{user_id}/preferences/{id}": {
      "patch": {"callbacks": {}, "description": "Update an existing preference with a new value", "operationId": "update-user-preference (2)", "parameters": [{"description": "The ID of the user", "in": "path", "name": "user_id", "required": true, "schema": {"type": "string"}}, {"description": "The id or type of the preference", "in": "path", "name": "id", "required": true, "schema": {"type": "string"}}], "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/PreferenceRequest"}}}, "description": "The attribute of the preference", "required": true}, "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/PreferenceResponse"}}}, "description": "Preference"}}, "security": [{"oauth": ["all"]}], "summary": "Update an existing preference", "tags": []},
      "put": {"callbacks": {}, "description": "Update an existing preference with a new value", "operationId": "update-user-preference", "parameters": [{"description": "The ID of the user", "in": "path", "name": "user_id", "required": true, "schema": {"type": "string"}}, {"description": "The id or type of the preference", "in": "path", "name": "id", "required": true, "schema": {"type": "string"}}], "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/PreferenceRequest"}}}, "description": "The attribute of the preference", "required": true}, "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/PreferenceResponse"}}}, "description": "Preference"}}, "security": [{"oauth": ["all"]}], "summary": "Update an existing preference", "tags": []}
    },
    "/tama/conversation/messages/{message_id}/artifacts": {
      "post": {"callbacks": {}, "description": "Create an artifact to be displayed to the user for a given message_id", "operationId": "create-message-artifact", "parameters": [{"description": "The ID of the message to create the artifact for", "in": "path", "name": "message_id", "required": true, "schema": {"type": "string"}}], "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/ArtifactRequest"}}}, "description": "The attributes of the artifact", "required": false}, "responses": {"201": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/ArtifactResponse"}}}, "description": "Artifact"}}, "security": [{"oauth": ["all"]}], "summary": "Create an artifact to display to the user", "tags": []}
    }
  },
  "security": [{"oauth": []}],
  "servers": [{"url": "http://localhost:4001", "variables": {}}],
  "tags": []
}
JSON
  )
}

resource "tama_specification" "test_spec" {
  space_id = tama_space.test_space.id
  version  = "1.0.0"
  endpoint = "http://localhost:4001"
  schema   = jsonencode(local.openapi_spec)

  wait_for {
    field {
      name = "current_state"
      in   = ["completed"]
    }
  }
}`, timestamp) + fmt.Sprintf(`

resource "tama_source_identity" "test" {
  specification_id = tama_specification.test_spec.id
  identifier       = %[1]q
  client_id        = %[2]q
  client_secret    = %[3]q

  validation {
    path   = %[4]q
    method = %[5]q
    codes  = %[6]s
  }
}
`, identifier, clientID, clientSecret, validationPath, validationMethod, validationCodes)
}

func testAccSourceIdentityResourceConfigWaitFor(identifier, apiKey, validationPath, validationMethod, validationCodes string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-identity-wait-%d"
  type = "root"
}

resource "tama_specification" "test_spec" {
  space_id = tama_space.test_space.id
  version  = "1.0.0"
  endpoint = "https://elasticsearch.arrakis.upmaru.network"
  schema   = jsonencode(jsondecode(file("${path.module}/testdata/elasticsearch_schema.json")))

  wait_for {
    field {
      name = "current_state"
      in   = ["completed"]
    }
  }
}`, timestamp) + fmt.Sprintf(`

resource "tama_source_identity" "test" {
  specification_id = tama_specification.test_spec.id
  identifier       = %[1]q
  api_key          = %[2]q

  validation {
    path   = %[3]q
    method = %[4]q
    codes  = %[5]s
  }

  wait_for {
    field {
      name = "provision_state"
      in   = ["active"]
    }
  }
}
`, identifier, apiKey, validationPath, validationMethod, validationCodes)
}

func testAccSourceIdentityResourceConfigWaitForMultiple(identifier, apiKey, validationPath, validationMethod, validationCodes string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-identity-wait-multiple-%d"
  type = "root"
}

resource "tama_specification" "test_spec" {
  space_id = tama_space.test_space.id
  version  = "1.0.0"
  endpoint = "https://elasticsearch.arrakis.upmaru.network"
  schema   = jsonencode(jsondecode(file("${path.module}/testdata/elasticsearch_schema.json")))

  wait_for {
    field {
      name = "current_state"
      in   = ["completed"]
    }
  }
}`, timestamp) + fmt.Sprintf(`

resource "tama_source_identity" "test" {
  specification_id = tama_specification.test_spec.id
  identifier       = %[1]q
  api_key          = %[2]q

  validation {
    path   = %[3]q
    method = %[4]q
    codes  = %[5]s
  }

  wait_for {
    field {
      name = "provision_state"
      in   = ["active"]
    }
    field {
      name = "current_state"
      in   = ["active", "failed"]
    }
  }
}
`, identifier, apiKey, validationPath, validationMethod, validationCodes)
}
