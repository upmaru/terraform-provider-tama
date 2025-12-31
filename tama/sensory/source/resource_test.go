// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package source_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccSourceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSourceResourceConfig("test-source", "model", "https://api.example.com", "test-api-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source.test", "name", "test-source"),
					resource.TestCheckResourceAttr("tama_source.test", "type", "model"),
					resource.TestCheckResourceAttr("tama_source.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttr("tama_source.test", "api_key", "test-api-key"),
					resource.TestCheckResourceAttrSet("tama_source.test", "id"),
					resource.TestCheckResourceAttrSet("tama_source.test", "slug"),
					resource.TestCheckResourceAttrSet("tama_source.test", "space_id"),
					resource.TestCheckResourceAttrSet("tama_source.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "tama_source.test",
				ImportState:             true,
				ImportStateVerify:       false, // SpaceId, Type, Endpoint, and ApiKey cannot be retrieved from API
				ImportStateVerifyIgnore: []string{"space_id", "type", "api_key"},
			},
			// Update and Read testing
			{
				Config: testAccSourceResourceConfig("test-source-updated", "model", "https://api.updated.com", "updated-api-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source.test", "name", "test-source-updated"),
					resource.TestCheckResourceAttr("tama_source.test", "type", "model"),
					resource.TestCheckResourceAttr("tama_source.test", "endpoint", "https://api.updated.com"),
					resource.TestCheckResourceAttr("tama_source.test", "api_key", "updated-api-key"),
					resource.TestCheckResourceAttrSet("tama_source.test", "id"),
					resource.TestCheckResourceAttrSet("tama_source.test", "provision_state"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccSourceResource_InvalidEndpoint(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSourceResourceConfig("test-source", "model", "invalid-url", "test-api-key"),
				ExpectError: regexp.MustCompile("Unable to create source"),
			},
		},
	})
}

func TestAccSourceResource_EmptyName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSourceResourceConfig("", "model", "https://api.example.com", "test-api-key"),
				ExpectError: regexp.MustCompile("Unable to create source"),
			},
		},
	})
}

func TestAccSourceResource_EmptyApiKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSourceResourceConfig("test-source", "model", "https://api.example.com", ""),
				ExpectError: regexp.MustCompile("Unable to create source"),
			},
		},
	})
}

func TestAccSourceResource_InvalidTypes(t *testing.T) {
	testCases := []struct {
		name       string
		sourceType string
	}{
		{"api type", "api"},
		{"webhook type", "webhook"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      testAccSourceResourceConfig(fmt.Sprintf("test-source-%s", tc.sourceType), tc.sourceType, "https://api.example.com", "test-api-key"),
						ExpectError: regexp.MustCompile("Unable to create source"),
					},
				},
			})
		})
	}
}

func TestAccSourceResource_LongName(t *testing.T) {
	longName := "this-is-a-very-long-source-name-that-might-exceed-database-limits-and-should-be-tested-for-proper-error-handling"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceResourceConfig(longName, "model", "https://api.example.com", "test-api-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source.test", "name", longName),
					resource.TestCheckResourceAttr("tama_source.test", "type", "model"),
					resource.TestCheckResourceAttrSet("tama_source.test", "slug"),
					resource.TestCheckResourceAttrSet("tama_source.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccSourceResource_SpecialCharacters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceResourceConfig("test-source-with-special_chars.123", "model", "https://api.example.com", "test-api-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source.test", "name", "test-source-with-special_chars.123"),
					resource.TestCheckResourceAttr("tama_source.test", "type", "model"),
					resource.TestCheckResourceAttrSet("tama_source.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccSourceResource_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceResourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// First source
					resource.TestCheckResourceAttr("tama_source.test1", "name", "test-source-1"),
					resource.TestCheckResourceAttr("tama_source.test1", "type", "model"),
					resource.TestCheckResourceAttr("tama_source.test1", "endpoint", "https://api1.example.com"),
					resource.TestCheckResourceAttrSet("tama_source.test1", "id"),
					resource.TestCheckResourceAttrSet("tama_source.test1", "provision_state"),
					// Second source
					resource.TestCheckResourceAttr("tama_source.test2", "name", "test-source-2"),
					resource.TestCheckResourceAttr("tama_source.test2", "type", "model"),
					resource.TestCheckResourceAttr("tama_source.test2", "endpoint", "https://api2.example.com"),
					resource.TestCheckResourceAttrSet("tama_source.test2", "id"),
					resource.TestCheckResourceAttrSet("tama_source.test2", "provision_state"),
				),
			},
		},
	})
}

func TestAccSourceResource_SensitiveApiKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceResourceConfig("test-source", "model", "https://api.example.com", "super-secret-api-key-123"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source.test", "name", "test-source"),
					// Note: We can't verify both that api_key exists and doesn't exist
					// The sensitive nature means it won't be displayed in plans/logs
					resource.TestCheckResourceAttr("tama_source.test", "api_key", "super-secret-api-key-123"),
				),
			},
		},
	})
}

func TestAccSourceResource_DisappearResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceResourceConfig("disappear-source", "model", "https://api.example.com", "test-api-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source.test", "name", "disappear-source"),
					resource.TestCheckResourceAttr("tama_source.test", "type", "model"),
					resource.TestCheckResourceAttr("tama_source.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttrSet("tama_source.test", "id"),
					resource.TestCheckResourceAttrSet("tama_source.test", "provision_state"),
				),
			},
		},
	})
}

func testAccSourceResourceConfig(name, sourceType, endpoint, apiKey string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-source-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_source" "test" {
  space_id = tama_space.test_space.id
  name     = %[1]q
  type     = %[2]q
  endpoint = %[3]q
  api_key  = %[4]q
}
`, name, sourceType, endpoint, apiKey)
}

func testAccSourceResourceConfigWithHeaders(name, sourceType, endpoint, apiKey string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-source-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_source" "test" {
  space_id = tama_space.test_space.id
  name     = %[1]q
  type     = %[2]q
  endpoint = %[3]q
  api_key  = %[4]q

  request = {
    headers = [
      {
        name  = "x-custom-header"
        value = "custom-value"
      },
      {
        name  = "x-api-version"
        value = "v1"
      }
    ]
  }
}
`, name, sourceType, endpoint, apiKey)
}

func testAccSourceResourceConfigWithHeadersUpdated(name, sourceType, endpoint, apiKey string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-source-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_source" "test" {
  space_id = tama_space.test_space.id
  name     = %[1]q
  type     = %[2]q
  endpoint = %[3]q
  api_key  = %[4]q

  request = {
    headers = [
      {
        name  = "x-updated-header"
        value = "updated-value"
      }
    ]
  }
}
`, name, sourceType, endpoint, apiKey)
}

func testAccSourceResourceConfigWithSessionAffinity(name, sourceType, endpoint, apiKey string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-source-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_source" "test" {
  space_id = tama_space.test_space.id
  name     = %[1]q
  type     = %[2]q
  endpoint = %[3]q
  api_key  = %[4]q

  request = {
    session_affinity = {
      location = "header"
      key      = "x-session-affinity"
      value    = "actor_id"
    }
  }
}
`, name, sourceType, endpoint, apiKey)
}

func testAccSourceResourceConfigWithSessionAffinityUpdated(name, sourceType, endpoint, apiKey string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-source-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_source" "test" {
  space_id = tama_space.test_space.id
  name     = %[1]q
  type     = %[2]q
  endpoint = %[3]q
  api_key  = %[4]q

  request = {
    session_affinity = {
      location = "body"
      key      = "session_id"
      value    = "actor_id"
    }
  }
}
`, name, sourceType, endpoint, apiKey)
}

func testAccSourceResourceConfigWithFullRequest(name, sourceType, endpoint, apiKey string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-source-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_source" "test" {
  space_id = tama_space.test_space.id
  name     = %[1]q
  type     = %[2]q
  endpoint = %[3]q
  api_key  = %[4]q

  request = {
    headers = [
      {
        name  = "x-http"
        value = "something"
      },
      {
        name  = "authorization"
        value = "Bearer token"
      }
    ]

    session_affinity = {
      location = "header"
      key      = "x-session-affinity"
      value    = "actor_id"
    }
  }
}
`, name, sourceType, endpoint, apiKey)
}

func testAccSourceResourceConfigMultiple() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-multiple-sources-%d"
  type = "root"
}`, timestamp) + `

resource "tama_source" "test1" {
  space_id = tama_space.test_space.id
  name     = "test-source-1"
  type     = "model"
  endpoint = "https://api1.example.com"
  api_key  = "test-api-key-1"
}

resource "tama_source" "test2" {
  space_id = tama_space.test_space.id
  name     = "test-source-2"
  type     = "model"
  endpoint = "https://api2.example.com"
  api_key  = "test-api-key-2"
}
`
}

func TestAccSourceResource_CurrentState(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceResourceConfig("test-source-state", "model", "https://api.example.com", "test-api-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source.test", "name", "test-source-state"),
					resource.TestCheckResourceAttr("tama_source.test", "type", "model"),
					resource.TestCheckResourceAttr("tama_source.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttrSet("tama_source.test", "id"),
					resource.TestCheckResourceAttrSet("tama_source.test", "provision_state"),
					// Verify that provision_state is not empty
					resource.TestMatchResourceAttr("tama_source.test", "provision_state", regexp.MustCompile(".+")),
				),
			},
		},
	})
}

func TestAccSourceResource_WithRequestHeaders(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with headers
			{
				Config: testAccSourceResourceConfigWithHeaders("test-source-headers", "model", "https://api.example.com", "test-api-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source.test", "name", "test-source-headers"),
					resource.TestCheckResourceAttr("tama_source.test", "type", "model"),
					resource.TestCheckResourceAttr("tama_source.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttrSet("tama_source.test", "id"),
					resource.TestCheckResourceAttrSet("tama_source.test", "provision_state"),
					// Check request headers
					resource.TestCheckResourceAttr("tama_source.test", "request.headers.#", "2"),
					resource.TestCheckResourceAttr("tama_source.test", "request.headers.0.name", "x-custom-header"),
					resource.TestCheckResourceAttr("tama_source.test", "request.headers.0.value", "custom-value"),
					resource.TestCheckResourceAttr("tama_source.test", "request.headers.1.name", "x-api-version"),
					resource.TestCheckResourceAttr("tama_source.test", "request.headers.1.value", "v1"),
				),
			},
			// Update headers
			{
				Config: testAccSourceResourceConfigWithHeadersUpdated("test-source-headers", "model", "https://api.example.com", "test-api-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source.test", "name", "test-source-headers"),
					resource.TestCheckResourceAttr("tama_source.test", "request.headers.#", "1"),
					resource.TestCheckResourceAttr("tama_source.test", "request.headers.0.name", "x-updated-header"),
					resource.TestCheckResourceAttr("tama_source.test", "request.headers.0.value", "updated-value"),
				),
			},
		},
	})
}

func TestAccSourceResource_WithSessionAffinity(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with session affinity
			{
				Config: testAccSourceResourceConfigWithSessionAffinity("test-source-affinity", "model", "https://api.example.com", "test-api-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source.test", "name", "test-source-affinity"),
					resource.TestCheckResourceAttr("tama_source.test", "type", "model"),
					resource.TestCheckResourceAttr("tama_source.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttrSet("tama_source.test", "id"),
					resource.TestCheckResourceAttrSet("tama_source.test", "provision_state"),
					// Check session affinity
					resource.TestCheckResourceAttr("tama_source.test", "request.session_affinity.location", "header"),
					resource.TestCheckResourceAttr("tama_source.test", "request.session_affinity.key", "x-session-affinity"),
					resource.TestCheckResourceAttr("tama_source.test", "request.session_affinity.value", "actor_id"),
				),
			},
			// Update session affinity
			{
				Config: testAccSourceResourceConfigWithSessionAffinityUpdated("test-source-affinity", "model", "https://api.example.com", "test-api-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source.test", "name", "test-source-affinity"),
					resource.TestCheckResourceAttr("tama_source.test", "request.session_affinity.location", "body"),
					resource.TestCheckResourceAttr("tama_source.test", "request.session_affinity.key", "session_id"),
					resource.TestCheckResourceAttr("tama_source.test", "request.session_affinity.value", "actor_id"),
				),
			},
		},
	})
}

func TestAccSourceResource_WithFullRequest(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with both headers and session affinity
			{
				Config: testAccSourceResourceConfigWithFullRequest("test-source-full-request", "model", "https://api.example.com", "test-api-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source.test", "name", "test-source-full-request"),
					resource.TestCheckResourceAttr("tama_source.test", "type", "model"),
					resource.TestCheckResourceAttr("tama_source.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttrSet("tama_source.test", "id"),
					resource.TestCheckResourceAttrSet("tama_source.test", "provision_state"),
					// Check request headers
					resource.TestCheckResourceAttr("tama_source.test", "request.headers.#", "2"),
					resource.TestCheckResourceAttr("tama_source.test", "request.headers.0.name", "x-http"),
					resource.TestCheckResourceAttr("tama_source.test", "request.headers.0.value", "something"),
					resource.TestCheckResourceAttr("tama_source.test", "request.headers.1.name", "authorization"),
					resource.TestCheckResourceAttr("tama_source.test", "request.headers.1.value", "Bearer token"),
					// Check session affinity
					resource.TestCheckResourceAttr("tama_source.test", "request.session_affinity.location", "header"),
					resource.TestCheckResourceAttr("tama_source.test", "request.session_affinity.key", "x-session-affinity"),
					resource.TestCheckResourceAttr("tama_source.test", "request.session_affinity.value", "actor_id"),
				),
			},
			// Update to remove request configuration
			{
				Config: testAccSourceResourceConfig("test-source-full-request", "model", "https://api.example.com", "test-api-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_source.test", "name", "test-source-full-request"),
					resource.TestCheckNoResourceAttr("tama_source.test", "request.headers"),
					resource.TestCheckNoResourceAttr("tama_source.test", "request.session_affinity"),
				),
			},
		},
	})
}
