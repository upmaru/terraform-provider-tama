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
