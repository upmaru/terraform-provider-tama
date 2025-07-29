// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package specification_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
	"github.com/upmaru/terraform-provider-tama/tama/sensory/specification/testhelpers"
)

func TestAccSpecificationDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create a specification resource first
			{
				Config: testAccSpecificationDataSourceConfig("1.0.0", "https://api.example.com", testhelpers.MustMarshalJSON(testhelpers.TestSchema())),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Data source checks
					resource.TestCheckResourceAttr("data.tama_specification.test", "version", "1.0.0"),
					resource.TestCheckResourceAttr("data.tama_specification.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttrSet("data.tama_specification.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_specification.test", "space_id"),
					resource.TestCheckResourceAttrSet("data.tama_specification.test", "schema"),
					resource.TestCheckResourceAttrSet("data.tama_specification.test", "current_state"),
					resource.TestCheckResourceAttrSet("data.tama_specification.test", "provision_state"),
					// Verify that states are not empty
					resource.TestMatchResourceAttr("data.tama_specification.test", "current_state", regexp.MustCompile(".+")),
					resource.TestMatchResourceAttr("data.tama_specification.test", "provision_state", regexp.MustCompile(".+")),
				),
			},
		},
	})
}

func TestAccSpecificationDataSource_InvalidId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSpecificationDataSourceConfigInvalidId(),
				ExpectError: regexp.MustCompile("Unable to read specification"),
			},
		},
	})
}

func TestAccSpecificationDataSource_NonExistentId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSpecificationDataSourceConfigNonExistent(),
				ExpectError: regexp.MustCompile("Unable to read specification"),
			},
		},
	})
}

func TestAccSpecificationDataSource_ComplexSchema(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpecificationDataSourceConfig("1.0.0", "https://api.example.com", testhelpers.MustMarshalJSON(testhelpers.TestComplexSchema())),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Data source checks
					resource.TestCheckResourceAttr("data.tama_specification.test", "version", "1.0.0"),
					resource.TestCheckResourceAttr("data.tama_specification.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttrSet("data.tama_specification.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_specification.test", "schema"),
					resource.TestCheckResourceAttrSet("data.tama_specification.test", "provision_state"),
					// Verify schema contains expected OpenAPI content
					resource.TestMatchResourceAttr("data.tama_specification.test", "schema", regexp.MustCompile("openapi")),
					resource.TestMatchResourceAttr("data.tama_specification.test", "schema", regexp.MustCompile("metadata")),
					resource.TestMatchResourceAttr("data.tama_specification.test", "schema", regexp.MustCompile("configuration")),
				),
			},
		},
	})
}

func TestAccSpecificationDataSource_States(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpecificationDataSourceConfig("1.0.0", "https://api.example.com", testhelpers.MustMarshalJSON(testhelpers.TestSchema())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_specification.test", "version", "1.0.0"),
					resource.TestCheckResourceAttr("data.tama_specification.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttrSet("data.tama_specification.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_specification.test", "current_state"),
					resource.TestCheckResourceAttrSet("data.tama_specification.test", "provision_state"),
					// Verify that states are not empty
					resource.TestMatchResourceAttr("data.tama_specification.test", "current_state", regexp.MustCompile(".+")),
					resource.TestMatchResourceAttr("data.tama_specification.test", "provision_state", regexp.MustCompile(".+")),
				),
			},
		},
	})
}

func TestAccSpecificationDataSource_SpecialVersionFormats(t *testing.T) {
	testCases := []struct {
		name    string
		version string
	}{
		{"semantic version", "1.2.3"},
		{"version with prerelease", "2.0.0-alpha.1"},
		{"version with build metadata", "1.0.0+20230101"},
		{"complex version", "3.1.0-beta.2+build.456"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccSpecificationDataSourceConfig(tc.version, "https://api.example.com", testhelpers.MustMarshalJSON(testhelpers.TestSchema())),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("data.tama_specification.test", "version", tc.version),
							resource.TestCheckResourceAttr("data.tama_specification.test", "endpoint", "https://api.example.com"),
							resource.TestCheckResourceAttrSet("data.tama_specification.test", "id"),
							resource.TestCheckResourceAttrSet("data.tama_specification.test", "provision_state"),
						),
					},
				},
			})
		})
	}
}

func TestAccSpecificationDataSource_LongEndpoint(t *testing.T) {
	longEndpoint := "https://very-long-domain-name-that-might-exceed-some-limits.example.com/api/v1/specifications/endpoint/with/many/path/segments"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpecificationDataSourceConfig("1.0.0", longEndpoint, testhelpers.MustMarshalJSON(testhelpers.TestSchema())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_specification.test", "version", "1.0.0"),
					resource.TestCheckResourceAttr("data.tama_specification.test", "endpoint", longEndpoint),
					resource.TestCheckResourceAttrSet("data.tama_specification.test", "provision_state"),
				),
			},
		},
	})
}

func testAccSpecificationDataSourceConfig(version, endpoint, schema string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-spec-ds-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_specification" "test" {
  space_id = tama_space.test_space.id
  version  = %[1]q
  endpoint = %[2]q
  schema   = %[3]q

  wait_for {
    field {
      name = "current_state"
      in   = ["completed"]
    }
  }
}

data "tama_specification" "test" {
  id = tama_specification.test.id
}
`, version, endpoint, schema)
}

func testAccSpecificationDataSourceConfigInvalidId() string {
	return acceptance.ProviderConfig + `
data "tama_specification" "test" {
  id = "invalid-id"
}
`
}

func testAccSpecificationDataSourceConfigNonExistent() string {
	return acceptance.ProviderConfig + `
data "tama_specification" "test" {
  id = "spec-00000000-0000-0000-0000-000000000000"
}
`
}
