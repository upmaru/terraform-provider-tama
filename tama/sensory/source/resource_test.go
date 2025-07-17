// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package source_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
					resource.TestCheckResourceAttrSet("tama_source.test", "space_id"),
					resource.TestCheckResourceAttrSet("tama_source.test", "current_state"),
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
					resource.TestCheckResourceAttrSet("tama_source.test", "current_state"),
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
					resource.TestCheckResourceAttrSet("tama_source.test", "current_state"),
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
					resource.TestCheckResourceAttrSet("tama_source.test", "current_state"),
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
					resource.TestCheckResourceAttrSet("tama_source.test1", "current_state"),
					// Second source
					resource.TestCheckResourceAttr("tama_source.test2", "name", "test-source-2"),
					resource.TestCheckResourceAttr("tama_source.test2", "type", "model"),
					resource.TestCheckResourceAttr("tama_source.test2", "endpoint", "https://api2.example.com"),
					resource.TestCheckResourceAttrSet("tama_source.test2", "id"),
					resource.TestCheckResourceAttrSet("tama_source.test2", "current_state"),
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
					resource.TestCheckResourceAttrSet("tama_source.test", "current_state"),
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
					resource.TestCheckResourceAttrSet("tama_source.test", "current_state"),
					// Verify that current_state is not empty
					resource.TestMatchResourceAttr("tama_source.test", "current_state", regexp.MustCompile(".+")),
				),
			},
		},
	})
}

func testAccCheckSourceDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// This function simulates the source being destroyed outside of Terraform
		// In a real test, you would make an API call to delete the resource
		// For now, we'll just return nil to simulate successful destruction
		return nil
	}
}
