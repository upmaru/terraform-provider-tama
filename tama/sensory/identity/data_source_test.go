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

func TestAccSourceIdentityDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSourceIdentityDataSourceConfig("ApiKey", "test-api-key", "/health", "GET", "[200]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the data source attributes match the resource
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "identifier", "ApiKey"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.path", "/health"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.method", "GET"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.codes.#", "1"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.codes.0", "200"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "specification_id"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "current_state"),
					// Verify the data source ID matches the resource ID
					resource.TestCheckResourceAttrPair("data.tama_source_identity.test", "id", "tama_source_identity.test", "id"),
					resource.TestCheckResourceAttrPair("data.tama_source_identity.test", "specification_id", "tama_source_identity.test", "specification_id"),
					resource.TestCheckResourceAttrPair("data.tama_source_identity.test", "identifier", "tama_source_identity.test", "identifier"),
					resource.TestCheckResourceAttrPair("data.tama_source_identity.test", "provision_state", "tama_source_identity.test", "provision_state"),
					resource.TestCheckResourceAttrPair("data.tama_source_identity.test", "current_state", "tama_source_identity.test", "current_state"),
					resource.TestCheckResourceAttrPair("data.tama_source_identity.test", "validation.path", "tama_source_identity.test", "validation.path"),
					resource.TestCheckResourceAttrPair("data.tama_source_identity.test", "validation.method", "tama_source_identity.test", "validation.method"),
					resource.TestCheckResourceAttrPair("data.tama_source_identity.test", "validation.codes.#", "tama_source_identity.test", "validation.codes.#"),
				),
			},
		},
	})
}

func TestAccSourceIdentityDataSource_MultipleStatusCodes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceIdentityDataSourceConfig("BearerToken", "test-bearer-token", "/status", "POST", "[200, 201, 202]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "identifier", "BearerToken"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.path", "/status"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.method", "POST"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.codes.#", "3"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.codes.0", "200"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.codes.1", "201"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.codes.2", "202"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "specification_id"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "current_state"),
				),
			},
		},
	})
}

func TestAccSourceIdentityDataSource_ComplexValidationPath(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceIdentityDataSourceConfig("CustomHeader", "test-custom-header", "/api/v1/health?check=all", "GET", "[200, 204]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "identifier", "CustomHeader"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.path", "/api/v1/health?check=all"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.method", "GET"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.codes.#", "2"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.codes.0", "200"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.codes.1", "204"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "specification_id"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "current_state"),
				),
			},
		},
	})
}

func TestAccSourceIdentityDataSource_DifferentHttpMethods(t *testing.T) {
	testCases := []struct {
		name   string
		method string
	}{
		{"GET method", "GET"},
		{"POST method", "POST"},
		{"PUT method", "PUT"},
		{"PATCH method", "PATCH"},
		{"DELETE method", "DELETE"},
		{"HEAD method", "HEAD"},
		{"OPTIONS method", "OPTIONS"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccSourceIdentityDataSourceConfig("ApiKey", "test-api-key", "/health", tc.method, "[200]"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.method", tc.method),
							resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.path", "/health"),
							resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.codes.#", "1"),
							resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.codes.0", "200"),
							resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "provision_state"),
							resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "current_state"),
						),
					},
				},
			})
		})
	}
}

func TestAccSourceIdentityDataSource_DifferentIdentifiers(t *testing.T) {
	testCases := []struct {
		name       string
		identifier string
	}{
		{"API Key", "ApiKey"},
		{"Bearer token", "BearerToken"},
		{"Basic auth", "BasicAuth"},
		{"Custom header", "CustomHeader"},
		{"JWT token", "JWTToken"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccSourceIdentityDataSourceConfig(tc.identifier, "test-api-key", "/health", "GET", "[200]"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("data.tama_source_identity.test", "identifier", tc.identifier),
							resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.path", "/health"),
							resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.method", "GET"),
							resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "provision_state"),
							resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "current_state"),
						),
					},
				},
			})
		})
	}
}

func TestAccSourceIdentityDataSource_NonExistentIdentity(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSourceIdentityDataSourceConfigNonExistent(),
				ExpectError: regexp.MustCompile("Unable to read source identity"),
			},
		},
	})
}

func TestAccSourceIdentityDataSource_EmptyId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSourceIdentityDataSourceConfigEmptyId(),
				ExpectError: regexp.MustCompile("Invalid Attribute Value|Attribute id string length must be at least 1"),
			},
		},
	})
}

func TestAccSourceIdentityDataSource_StateValues(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceIdentityDataSourceConfig("ApiKey", "test-api-key", "/health", "GET", "[200]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "identifier", "ApiKey"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "specification_id"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "provision_state"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "current_state"),
					// Verify that states are not empty
					resource.TestMatchResourceAttr("data.tama_source_identity.test", "provision_state", regexp.MustCompile(".+")),
					resource.TestMatchResourceAttr("data.tama_source_identity.test", "current_state", regexp.MustCompile(".+")),
				),
			},
		},
	})
}

func TestAccSourceIdentityDataSource_ValidationCodes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceIdentityDataSourceConfig("WebhookSecret", "webhook-secret", "/webhook/validate", "POST", "[200, 201, 202, 204]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "identifier", "WebhookSecret"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.codes.#", "4"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.codes.0", "200"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.codes.1", "201"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.codes.2", "202"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test", "validation.codes.3", "204"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test", "specification_id"),
				),
			},
		},
	})
}

func TestAccSourceIdentityDataSource_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceIdentityDataSourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// First identity data source
					resource.TestCheckResourceAttr("data.tama_source_identity.test1", "identifier", "ApiKey"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test1", "validation.path", "/health"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test1", "validation.method", "GET"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test1", "validation.codes.#", "1"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test1", "validation.codes.0", "200"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test1", "id"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test1", "specification_id"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test1", "provision_state"),
					// Second identity data source
					resource.TestCheckResourceAttr("data.tama_source_identity.test2", "identifier", "BearerToken"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test2", "validation.path", "/status"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test2", "validation.method", "POST"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test2", "validation.codes.#", "2"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test2", "validation.codes.0", "200"),
					resource.TestCheckResourceAttr("data.tama_source_identity.test2", "validation.codes.1", "201"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test2", "id"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test2", "specification_id"),
					resource.TestCheckResourceAttrSet("data.tama_source_identity.test2", "provision_state"),
					// Verify data sources match their corresponding resources
					resource.TestCheckResourceAttrPair("data.tama_source_identity.test1", "id", "tama_source_identity.test1", "id"),
					resource.TestCheckResourceAttrPair("data.tama_source_identity.test2", "id", "tama_source_identity.test2", "id"),
				),
			},
		},
	})
}

func testAccSourceIdentityDataSourceConfig(identifier, apiKey, validationPath, validationMethod, validationCodes string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-identity-ds-%d"
  type = "root"
}

resource "tama_specification" "test_spec" {
  space_id = tama_space.test_space.id
  version  = "1.0.0"
  endpoint = "https://api.example.com"
  schema = {
    "type" = "object"
    "properties" = {
      "message" = {
        "type" = "string"
      }
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

data "tama_source_identity" "test" {
  id = tama_source_identity.test.id
}
`, identifier, apiKey, validationPath, validationMethod, validationCodes)
}

func testAccSourceIdentityDataSourceConfigNonExistent() string {
	return acceptance.ProviderConfig + `
data "tama_source_identity" "non_existent" {
  id = "non-existent-identity-id-12345"
}
`
}

func testAccSourceIdentityDataSourceConfigEmptyId() string {
	return acceptance.ProviderConfig + `
data "tama_source_identity" "empty_id" {
  id = ""
}
`
}

func testAccSourceIdentityDataSourceConfigMultiple() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-multiple-identity-ds-%d"
  type = "root"
}

resource "tama_specification" "test_spec" {
  space_id = tama_space.test_space.id
  version  = "1.0.0"
  endpoint = "https://api.example.com"
  schema = {
    "type" = "object"
    "properties" = {
      "message" = {
        "type" = "string"
      }
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
  identifier       = "BearerToken"
  api_key          = "test-api-key-2"

  validation {
    path   = "/status"
    method = "POST"
    codes  = [200, 201]
  }
}

data "tama_source_identity" "test1" {
  id = tama_source_identity.test1.id
}

data "tama_source_identity" "test2" {
  id = tama_source_identity.test2.id
}
`
}
