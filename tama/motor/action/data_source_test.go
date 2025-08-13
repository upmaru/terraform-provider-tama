// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package action_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccActionDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create a specification resource first, then test the action data source
			{
				Config: testAccActionDataSourceConfig("create-index"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Data source checks
					resource.TestCheckResourceAttr("data.tama_action.test", "identifier", "create-index"),
					resource.TestCheckResourceAttrSet("data.tama_action.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_action.test", "path"),
					resource.TestCheckResourceAttrSet("data.tama_action.test", "method"),
					resource.TestCheckResourceAttrSet("data.tama_action.test", "specification_id"),
					// Verify that the specification_id matches the created specification
					resource.TestCheckResourceAttrPair("data.tama_action.test", "specification_id", "tama_specification.test", "id"),
				),
			},
		},
	})
}

func TestAccActionDataSource_InvalidSpecificationId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccActionDataSourceConfigInvalidSpecId(),
				ExpectError: regexp.MustCompile("Unable to read action"),
			},
		},
	})
}

func TestAccActionDataSource_InvalidIdentifier(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccActionDataSourceConfig("non-existent-action"),
				ExpectError: regexp.MustCompile("Unable to read action"),
			},
		},
	})
}

func TestAccActionDataSource_NonExistentSpecification(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccActionDataSourceConfigNonExistentSpec(),
				ExpectError: regexp.MustCompile("Unable to read action"),
			},
		},
	})
}

func TestAccActionDataSource_DifferentActionIdentifiers(t *testing.T) {
	testCases := []struct {
		name       string
		identifier string
	}{
		{"create index action", "create-index"},
		{"create or update document action", "create-or-update-document-with-id"},
		{"update aliases action", "update-aliases"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccActionDataSourceConfig(tc.identifier),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("data.tama_action.test", "identifier", tc.identifier),
							resource.TestCheckResourceAttrSet("data.tama_action.test", "id"),
							resource.TestCheckResourceAttrSet("data.tama_action.test", "path"),
							resource.TestCheckResourceAttrSet("data.tama_action.test", "method"),
							resource.TestCheckResourceAttrSet("data.tama_action.test", "specification_id"),
						),
					},
				},
			})
		})
	}
}

func TestAccActionDataSource_HTTPMethods(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionDataSourceConfig("create-index"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_action.test", "identifier", "create-index"),
					resource.TestCheckResourceAttrSet("data.tama_action.test", "method"),
					// Verify that method is a valid HTTP method
					resource.TestMatchResourceAttr("data.tama_action.test", "method", regexp.MustCompile("(?i)^(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)$")),
					resource.TestCheckResourceAttrSet("data.tama_action.test", "path"),
					// Verify that path starts with /
					resource.TestMatchResourceAttr("data.tama_action.test", "path", regexp.MustCompile("^/")),
				),
			},
		},
	})
}

func TestAccActionDataSource_SpecificationReference(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionDataSourceConfig("create-index"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_action.test", "identifier", "create-index"),
					resource.TestCheckResourceAttrSet("data.tama_action.test", "specification_id"),
					// Verify that specification_id matches the resource specification
					resource.TestCheckResourceAttrPair("data.tama_action.test", "specification_id", "tama_specification.test", "id"),
					// Verify that the action belongs to the correct specification
					resource.TestCheckResourceAttrSet("data.tama_action.test", "id"),
				),
			},
		},
	})
}

func testAccActionDataSourceConfig(identifier string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-action-ds-%d"
  type = "root"
}

resource "tama_specification" "test" {
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
}

data "tama_action" "test" {
  specification_id = tama_specification.test.id
  identifier       = %q
}
`, timestamp, identifier)
}

func testAccActionDataSourceConfigInvalidSpecId() string {
	return acceptance.ProviderConfig + `
data "tama_action" "test" {
  specification_id = "invalid-spec-id"
  identifier       = "create-index"
}
`
}

func testAccActionDataSourceConfigNonExistentSpec() string {
	return acceptance.ProviderConfig + `
data "tama_action" "test" {
  specification_id = "spec-00000000-0000-0000-0000-000000000000"
  identifier       = "create-index"
}
`
}

func TestAccActionDataSource_ByPathAndMethod(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test looking up action by path and method instead of identifier
			{
				Config: testAccActionDataSourceConfigByPathAndMethod("/_aliases", "POST"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Data source checks
					resource.TestCheckResourceAttr("data.tama_action.test_by_path_method", "path", "/_aliases"),
					resource.TestCheckResourceAttr("data.tama_action.test_by_path_method", "method", "post"), // Should be lowercased by the API
					resource.TestCheckResourceAttrSet("data.tama_action.test_by_path_method", "id"),
					resource.TestCheckResourceAttrSet("data.tama_action.test_by_path_method", "identifier"),
					resource.TestCheckResourceAttrSet("data.tama_action.test_by_path_method", "specification_id"),
					// Verify that the specification_id matches the created specification
					resource.TestCheckResourceAttrPair("data.tama_action.test_by_path_method", "specification_id", "tama_specification.test", "id"),
				),
			},
		},
	})
}

func testAccActionDataSourceConfigByPathAndMethod(path, method string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-action-ds-path-method-%d"
  type = "root"
}

resource "tama_specification" "test" {
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
}

data "tama_action" "test_by_path_method" {
  specification_id = tama_specification.test.id
  path             = %q
  method           = %q
}
`, timestamp, path, method)
}
