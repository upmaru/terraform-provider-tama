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

func TestAccSpecificationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSpecificationResourceConfig("3.1.0", "https://elasticsearch.arrakis.upmaru.network", testhelpers.MustMarshalJSON(testhelpers.TestSchema())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_specification.test", "version", "3.1.0"),
					resource.TestCheckResourceAttr("tama_specification.test", "endpoint", "https://elasticsearch.arrakis.upmaru.network"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "id"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "space_id"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "schema"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "current_state"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_specification.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore current_state as it may differ based on timing
				ImportStateVerifyIgnore: []string{"current_state"},
			},
			// Update and Read testing
			{
				Config: testAccSpecificationResourceConfig("3.2.0", "https://elasticsearch-updated.arrakis.upmaru.network", testhelpers.MustMarshalJSON(testhelpers.TestSchemaUpdated())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_specification.test", "version", "3.2.0"),
					resource.TestCheckResourceAttr("tama_specification.test", "endpoint", "https://elasticsearch-updated.arrakis.upmaru.network"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "id"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "space_id"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "schema"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "current_state"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "provision_state"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccSpecificationResource_EmptyVersion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSpecificationResourceConfig("", "https://elasticsearch.arrakis.upmaru.network", testhelpers.MustMarshalJSON(testhelpers.TestSchema())),
				ExpectError: regexp.MustCompile("Unable to create specification"),
			},
		},
	})
}

func TestAccSpecificationResource_InvalidSchema(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSpecificationResourceConfigInvalidSchema("3.1.0", "https://elasticsearch.arrakis.upmaru.network"),
				ExpectError: regexp.MustCompile("Invalid Schema"),
			},
		},
	})
}

func TestAccSpecificationResource_ComplexSchema(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpecificationResourceConfig("1.0.0", "https://api.example.com", testhelpers.MustMarshalJSON(testhelpers.TestComplexSchema())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_specification.test", "version", "1.0.0"),
					resource.TestCheckResourceAttr("tama_specification.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "id"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "schema"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccSpecificationResource_JSONNormalization(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpecificationResourceConfig("1.0.0", "https://api.example.com", testhelpers.MustMarshalJSON(testhelpers.TestSchemaWithWhitespace())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_specification.test", "version", "1.0.0"),
					resource.TestCheckResourceAttr("tama_specification.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "schema"),
				),
			},
			// Update with semantically identical but differently formatted JSON
			{
				Config: testAccSpecificationResourceConfig("1.0.0", "https://api.example.com", testhelpers.MustMarshalJSON(testhelpers.TestSchemaCompact())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_specification.test", "version", "1.0.0"),
					resource.TestCheckResourceAttr("tama_specification.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "schema"),
				),
			},
		},
	})
}

func TestAccSpecificationResource_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpecificationResourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// First specification
					resource.TestCheckResourceAttr("tama_specification.test1", "version", "1.0.0"),
					resource.TestCheckResourceAttr("tama_specification.test1", "endpoint", "https://api1.example.com"),
					resource.TestCheckResourceAttrSet("tama_specification.test1", "id"),
					resource.TestCheckResourceAttrSet("tama_specification.test1", "provision_state"),
					// Second specification
					resource.TestCheckResourceAttr("tama_specification.test2", "version", "2.0.0"),
					resource.TestCheckResourceAttr("tama_specification.test2", "endpoint", "https://api2.example.com"),
					resource.TestCheckResourceAttrSet("tama_specification.test2", "id"),
					resource.TestCheckResourceAttrSet("tama_specification.test2", "provision_state"),
				),
			},
		},
	})
}

func TestAccSpecificationResource_LongEndpoint(t *testing.T) {
	longEndpoint := "https://very-long-domain-name-that-might-exceed-some-limits.example.com/api/v1/specifications/endpoint"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpecificationResourceConfig("1.0.0", longEndpoint, testhelpers.MustMarshalJSON(testhelpers.TestSchema())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_specification.test", "version", "1.0.0"),
					resource.TestCheckResourceAttr("tama_specification.test", "endpoint", longEndpoint),
					resource.TestCheckResourceAttrSet("tama_specification.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccSpecificationResource_SpecialCharactersInVersion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpecificationResourceConfig("1.0.0-beta.1+build.123", "https://api.example.com", testhelpers.MustMarshalJSON(testhelpers.TestSchema())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_specification.test", "version", "1.0.0-beta.1+build.123"),
					resource.TestCheckResourceAttr("tama_specification.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccSpecificationResource_CurrentState(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpecificationResourceConfig("1.0.0", "https://api.example.com", testhelpers.MustMarshalJSON(testhelpers.TestSchema())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_specification.test", "version", "1.0.0"),
					resource.TestCheckResourceAttr("tama_specification.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "id"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "current_state"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "provision_state"),
					// Verify that states are not empty
					resource.TestMatchResourceAttr("tama_specification.test", "current_state", regexp.MustCompile(".+")),
					resource.TestMatchResourceAttr("tama_specification.test", "provision_state", regexp.MustCompile(".+")),
				),
			},
		},
	})
}

func testAccSpecificationResourceConfig(version, endpoint, schema string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-spec-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_specification" "test" {
  space_id = tama_space.test_space.id
  version  = %[1]q
  endpoint = %[2]q
  schema   = %[3]q
}
`, version, endpoint, schema)
}

func testAccSpecificationResourceConfigInvalidSchema(version, endpoint string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-spec-invalid-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_specification" "test" {
  space_id = tama_space.test_space.id
  version  = %[1]q
  endpoint = %[2]q
  schema   = "invalid json {"
}
`, version, endpoint)
}

func TestAccSpecificationResource_WaitFor(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with wait_for
			{
				Config: testAccSpecificationResourceConfigWaitFor("1.0.0", "https://api.example.com", testhelpers.MustMarshalJSON(testhelpers.TestSchema())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_specification.test", "version", "1.0.0"),
					resource.TestCheckResourceAttr("tama_specification.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "id"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "space_id"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "schema"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "current_state"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "provision_state"),
					// Verify wait_for configuration is accepted and doesn't cause errors
					resource.TestCheckResourceAttr("tama_specification.test", "wait_for.#", "1"),
					resource.TestCheckResourceAttr("tama_specification.test", "wait_for.0.field.#", "1"),
					resource.TestCheckResourceAttr("tama_specification.test", "wait_for.0.field.0.key", "current_state"),
					resource.TestCheckResourceAttr("tama_specification.test", "wait_for.0.field.0.value", "completed"),
					resource.TestCheckResourceAttr("tama_specification.test", "wait_for.0.field.0.value_type", "eq"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_specification.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore wait_for during import as it's configuration-only
				// Also ignore current_state as it may differ after wait_for execution
				ImportStateVerifyIgnore: []string{"wait_for", "current_state"},
			},
		},
	})
}

func TestAccSpecificationResource_WaitForMultipleConditions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with multiple wait_for conditions
			{
				Config: testAccSpecificationResourceConfigWaitForMultiple("1.0.0", "https://api.example.com", testhelpers.MustMarshalJSON(testhelpers.TestSchema())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_specification.test", "version", "1.0.0"),
					resource.TestCheckResourceAttr("tama_specification.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "id"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "space_id"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "schema"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "current_state"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "provision_state"),
					// Verify multiple wait_for conditions
					resource.TestCheckResourceAttr("tama_specification.test", "wait_for.#", "1"),
					resource.TestCheckResourceAttr("tama_specification.test", "wait_for.0.field.#", "2"),
					resource.TestCheckResourceAttr("tama_specification.test", "wait_for.0.field.0.key", "current_state"),
					resource.TestCheckResourceAttr("tama_specification.test", "wait_for.0.field.0.value", "completed"),
					resource.TestCheckResourceAttr("tama_specification.test", "wait_for.0.field.0.value_type", "eq"),
					resource.TestCheckResourceAttr("tama_specification.test", "wait_for.0.field.1.key", "provision_state"),
					resource.TestCheckResourceAttr("tama_specification.test", "wait_for.0.field.1.value", "^(active|inactive)$"),
					resource.TestCheckResourceAttr("tama_specification.test", "wait_for.0.field.1.value_type", "regex"),
				),
			},
		},
	})
}

func TestAccSpecificationResource_WaitForRegex(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with regex wait_for condition
			{
				Config: testAccSpecificationResourceConfigWaitForRegex("1.0.0", "https://api.example.com", testhelpers.MustMarshalJSON(testhelpers.TestSchema())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_specification.test", "version", "1.0.0"),
					resource.TestCheckResourceAttr("tama_specification.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "id"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "space_id"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "schema"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "current_state"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "provision_state"),
					// Verify regex wait_for condition
					resource.TestCheckResourceAttr("tama_specification.test", "wait_for.#", "1"),
					resource.TestCheckResourceAttr("tama_specification.test", "wait_for.0.field.#", "1"),
					resource.TestCheckResourceAttr("tama_specification.test", "wait_for.0.field.0.key", "current_state"),
					resource.TestCheckResourceAttr("tama_specification.test", "wait_for.0.field.0.value", "^(completed|failed)$"),
					resource.TestCheckResourceAttr("tama_specification.test", "wait_for.0.field.0.value_type", "regex"),
				),
			},
		},
	})
}

func testAccSpecificationResourceConfigMultiple() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-multiple-specs-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_specification" "test1" {
  space_id = tama_space.test_space.id
  version  = "1.0.0"
  endpoint = "https://api1.example.com"
  schema   = %q
}

resource "tama_specification" "test2" {
  space_id = tama_space.test_space.id
  version  = "2.0.0"
  endpoint = "https://api2.example.com"
  schema   = %q
}
`, testhelpers.MustMarshalJSON(testhelpers.TestSchema()), testhelpers.MustMarshalJSON(testhelpers.TestSchemaUpdated()))
}

func testAccSpecificationResourceConfigWaitFor(version, endpoint, schema string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-spec-wait-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_specification" "test" {
  space_id = tama_space.test_space.id
  version  = %[1]q
  endpoint = %[2]q
  schema   = %[3]q

  wait_for {
    field {
      key   = "current_state"
      value = "completed"
    }
  }
}
`, version, endpoint, schema)
}

func testAccSpecificationResourceConfigWaitForMultiple(version, endpoint, schema string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-spec-wait-multiple-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_specification" "test" {
  space_id = tama_space.test_space.id
  version  = %[1]q
  endpoint = %[2]q
  schema   = %[3]q

  wait_for {
    field {
      key   = "current_state"
      value = "completed"
    }

    field {
      key        = "provision_state"
      value      = "^(active|inactive)$"
      value_type = "regex"
    }
  }
}
`, version, endpoint, schema)
}

func testAccSpecificationResourceConfigWaitForRegex(version, endpoint, schema string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-spec-wait-regex-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_specification" "test" {
  space_id = tama_space.test_space.id
  version  = %[1]q
  endpoint = %[2]q
  schema   = %[3]q

  wait_for {
    field {
      key        = "current_state"
      value      = "^(completed|failed)$"
      value_type = "regex"
    }
  }
}
`, version, endpoint, schema)
}
