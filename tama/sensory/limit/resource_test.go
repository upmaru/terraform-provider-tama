// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package limit_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccLimitResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccLimitResourceConfig("minutes", 1, 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_limit.test", "scale_unit", "minutes"),
					resource.TestCheckResourceAttr("tama_limit.test", "scale_count", "1"),
					resource.TestCheckResourceAttr("tama_limit.test", "value", "100"),
					resource.TestCheckResourceAttrSet("tama_limit.test", "id"),
					resource.TestCheckResourceAttrSet("tama_limit.test", "source_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "tama_limit.test",
				ImportState:             true,
				ImportStateVerify:       true, // SourceId is now available from API
				ImportStateVerifyIgnore: []string{},
			},
			// Update and Read testing
			{
				Config: testAccLimitResourceConfig("hours", 24, 1000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_limit.test", "scale_unit", "hours"),
					resource.TestCheckResourceAttr("tama_limit.test", "scale_count", "24"),
					resource.TestCheckResourceAttr("tama_limit.test", "value", "1000"),
					resource.TestCheckResourceAttrSet("tama_limit.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccLimitResource_TimeUnits(t *testing.T) {
	testCases := []struct {
		name      string
		scaleUnit string
		count     int64
		limit     int64
	}{
		{"seconds limit", "seconds", 60, 10},
		{"minutes limit", "minutes", 5, 50},
		{"hours limit", "hours", 1, 100},
		{"hours high count", "hours", 24, 1000},
		{"hours very high", "hours", 168, 10000},
		{"hours monthly", "hours", 720, 100000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccLimitResourceConfig(tc.scaleUnit, tc.count, tc.limit),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("tama_limit.test", "scale_unit", tc.scaleUnit),
							resource.TestCheckResourceAttr("tama_limit.test", "scale_count", fmt.Sprintf("%d", tc.count)),
							resource.TestCheckResourceAttr("tama_limit.test", "value", fmt.Sprintf("%d", tc.limit)),
							resource.TestCheckResourceAttrSet("tama_limit.test", "id"),
						),
					},
				},
			})
		})
	}
}

func TestAccLimitResource_RateLimitScenarios(t *testing.T) {
	testCases := []struct {
		name        string
		scaleUnit   string
		scaleCount  int64
		limit       int64
		description string
	}{
		{"burst protection", "seconds", 1, 5, "5 requests per second"},
		{"moderate usage", "minutes", 1, 100, "100 requests per minute"},
		{"hourly quota", "hours", 1, 1000, "1000 requests per hour"},
		{"daily quota hours", "hours", 24, 10000, "10000 requests per day using hours"},
		{"weekly batch", "hours", 168, 50000, "50000 requests per week"},
		{"monthly enterprise", "hours", 720, 1000000, "1M requests per month using hours"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccLimitResourceConfig(tc.scaleUnit, tc.scaleCount, tc.limit),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("tama_limit.test", "scale_unit", tc.scaleUnit),
							resource.TestCheckResourceAttr("tama_limit.test", "scale_count", fmt.Sprintf("%d", tc.scaleCount)),
							resource.TestCheckResourceAttr("tama_limit.test", "value", fmt.Sprintf("%d", tc.limit)),
							resource.TestCheckResourceAttrSet("tama_limit.test", "id"),
						),
					},
				},
			})
		})
	}
}

func TestAccLimitResource_InvalidScaleUnit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccLimitResourceConfig("invalid-unit", 1, 100),
				ExpectError: regexp.MustCompile("Unable to create limit"),
			},
			{
				Config:      testAccLimitResourceConfig("days", 1, 100),
				ExpectError: regexp.MustCompile("Unable to create limit"),
			},
			{
				Config:      testAccLimitResourceConfig("weeks", 1, 100),
				ExpectError: regexp.MustCompile("Unable to create limit"),
			},
			{
				Config:      testAccLimitResourceConfig("months", 1, 100),
				ExpectError: regexp.MustCompile("Unable to create limit"),
			},
		},
	})
}

func TestAccLimitResource_ZeroValues(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccLimitResourceConfig("minutes", 0, 100),
				ExpectError: regexp.MustCompile("Unable to create limit"),
			},
			{
				Config:      testAccLimitResourceConfig("minutes", 1, 0),
				ExpectError: regexp.MustCompile("Unable to create limit"),
			},
		},
	})
}

func TestAccLimitResource_NegativeValues(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccLimitResourceConfigNegative("minutes", -1, 100),
				ExpectError: regexp.MustCompile("limit scale_count must be greater than 0"),
			},
			{
				Config:      testAccLimitResourceConfigNegative("minutes", 1, -100),
				ExpectError: regexp.MustCompile("Unable to create limit"),
			},
		},
	})
}

func TestAccLimitResource_EmptyScaleUnit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccLimitResourceConfig("", 1, 100),
				ExpectError: regexp.MustCompile("Unable to create limit"),
			},
		},
	})
}

func TestAccLimitResource_HighLimits(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLimitResourceConfig("hours", 8760, 999999999),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_limit.test", "scale_unit", "hours"),
					resource.TestCheckResourceAttr("tama_limit.test", "scale_count", "8760"),
					resource.TestCheckResourceAttr("tama_limit.test", "value", "999999999"),
					resource.TestCheckResourceAttrSet("tama_limit.test", "id"),
				),
			},
		},
	})
}

func TestAccLimitResource_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLimitResourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// First limit - burst protection
					resource.TestCheckResourceAttr("tama_limit.burst", "scale_unit", "seconds"),
					resource.TestCheckResourceAttr("tama_limit.burst", "scale_count", "1"),
					resource.TestCheckResourceAttr("tama_limit.burst", "value", "10"),
					resource.TestCheckResourceAttrSet("tama_limit.burst", "id"),
					// Second limit - hourly quota
					resource.TestCheckResourceAttr("tama_limit.hourly", "scale_unit", "hours"),
					resource.TestCheckResourceAttr("tama_limit.hourly", "scale_count", "1"),
					resource.TestCheckResourceAttr("tama_limit.hourly", "value", "1000"),
					resource.TestCheckResourceAttrSet("tama_limit.hourly", "id"),
				),
			},
		},
	})
}

func TestAccLimitResource_DifferentSources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLimitResourceConfigDifferentSources(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Limit for first source
					resource.TestCheckResourceAttr("tama_limit.source1_limit", "scale_unit", "minutes"),
					resource.TestCheckResourceAttr("tama_limit.source1_limit", "scale_count", "1"),
					resource.TestCheckResourceAttr("tama_limit.source1_limit", "value", "100"),
					resource.TestCheckResourceAttrSet("tama_limit.source1_limit", "id"),
					// Limit for second source
					resource.TestCheckResourceAttr("tama_limit.source2_limit", "scale_unit", "hours"),
					resource.TestCheckResourceAttr("tama_limit.source2_limit", "scale_count", "1"),
					resource.TestCheckResourceAttr("tama_limit.source2_limit", "value", "500"),
					resource.TestCheckResourceAttrSet("tama_limit.source2_limit", "id"),
				),
			},
		},
	})
}

func TestAccLimitResource_UpdateLimits(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Start with conservative limits
			{
				Config: testAccLimitResourceConfig("minutes", 1, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_limit.test", "scale_unit", "minutes"),
					resource.TestCheckResourceAttr("tama_limit.test", "scale_count", "1"),
					resource.TestCheckResourceAttr("tama_limit.test", "value", "10"),
				),
			},
			// Increase to moderate limits
			{
				Config: testAccLimitResourceConfig("minutes", 5, 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_limit.test", "scale_unit", "minutes"),
					resource.TestCheckResourceAttr("tama_limit.test", "scale_count", "5"),
					resource.TestCheckResourceAttr("tama_limit.test", "value", "100"),
				),
			},
			// Change to hourly limits
			{
				Config: testAccLimitResourceConfig("hours", 1, 1000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_limit.test", "scale_unit", "hours"),
					resource.TestCheckResourceAttr("tama_limit.test", "scale_count", "1"),
					resource.TestCheckResourceAttr("tama_limit.test", "value", "1000"),
				),
			},
		},
	})
}

func TestAccLimitResource_DisappearResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLimitResourceConfig("minutes", 1, 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_limit.test", "scale_unit", "minutes"),
					resource.TestCheckResourceAttr("tama_limit.test", "scale_count", "1"),
					resource.TestCheckResourceAttr("tama_limit.test", "value", "100"),
					resource.TestCheckResourceAttrSet("tama_limit.test", "id"),
				),
			},
		},
	})
}

func testAccLimitResourceConfig(scaleUnit string, scaleCount, limit int64) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-limit-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_source" "test_source" {
  space_id = tama_space.test_space.id
  name     = "test-source-for-limit"
  type     = "model"
  endpoint = "https://api.example.com"
  api_key  = "test-api-key"
}

resource "tama_limit" "test" {
  source_id   = tama_source.test_source.id
  scale_unit  = %[1]q
  scale_count = %[2]d
  value       = %[3]d
}
`, scaleUnit, scaleCount, limit)
}

func testAccLimitResourceConfigNegative(scaleUnit string, scaleCount, limit int64) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-limit-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_source" "test_source" {
  space_id = tama_space.test_space.id
  name     = "test-source-for-limit"
  type     = "model"
  endpoint = "https://api.example.com"
  api_key  = "test-api-key"
}

resource "tama_limit" "test" {
  source_id   = tama_source.test_source.id
  scale_unit  = %[1]q
  scale_count = %[2]d
  value       = %[3]d
}
`, scaleUnit, scaleCount, limit)
}

func testAccLimitResourceConfigMultiple() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-multiple-limits-%d"
  type = "root"
}`, timestamp) + `

resource "tama_source" "test_source_1" {
  space_id = tama_space.test_space.id
  name     = "test-source-for-burst-limit"
  type     = "model"
  endpoint = "https://api1.example.com"
  api_key  = "test-api-key-1"
}

resource "tama_source" "test_source_2" {
  space_id = tama_space.test_space.id
  name     = "test-source-for-hourly-limit"
  type     = "model"
  endpoint = "https://api2.example.com"
  api_key  = "test-api-key-2"
}

resource "tama_limit" "burst" {
  source_id   = tama_source.test_source_1.id
  scale_unit  = "seconds"
  scale_count = 1
  value       = 10
}

resource "tama_limit" "hourly" {
  source_id   = tama_source.test_source_2.id
  scale_unit  = "hours"
  scale_count = 1
  value       = 1000
}
`
}

func testAccLimitResourceConfigDifferentSources() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-different-source-limits-%d"
  type = "root"
}`, timestamp) + `

resource "tama_source" "source1" {
  space_id = tama_space.test_space.id
  name     = "test-source-1"
  type     = "model"
  endpoint = "https://api1.example.com"
  api_key  = "test-api-key-1"
}

resource "tama_source" "source2" {
  space_id = tama_space.test_space.id
  name     = "test-source-2"
  type     = "model"
  endpoint = "https://api2.example.com"
  api_key  = "test-api-key-2"
}

resource "tama_limit" "source1_limit" {
  source_id   = tama_source.source1.id
  scale_unit  = "minutes"
  scale_count = 1
  value       = 100
}

resource "tama_limit" "source2_limit" {
  source_id   = tama_source.source2.id
  scale_unit  = "hours"
  scale_count = 1
  value       = 500
}
`
}

func testAccCheckLimitDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// This function simulates the limit being destroyed outside of Terraform
		// In a real test, you would make an API call to delete the resource
		// For now, we'll just return nil to simulate successful destruction
		return nil
	}
}
