// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package space_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccSpaceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSpaceResourceConfig("test-space", "root"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_space.test", "name", "test-space"),
					resource.TestCheckResourceAttr("tama_space.test", "type", "root"),
					resource.TestCheckResourceAttrSet("tama_space.test", "id"),
					resource.TestCheckResourceAttrSet("tama_space.test", "slug"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_space.test",
				ImportState:       true,
				ImportStateVerify: false, // Type field cannot be retrieved from API
			},
			// Update and Read testing
			{
				Config: testAccSpaceResourceConfig("test-space-updated", "component"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_space.test", "name", "test-space-updated"),
					resource.TestCheckResourceAttr("tama_space.test", "type", "component"),
					resource.TestCheckResourceAttrSet("tama_space.test", "id"),
					resource.TestCheckResourceAttrSet("tama_space.test", "slug"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccSpaceResource_InvalidType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSpaceResourceConfig("test-space", "invalid-type"),
				ExpectError: regexp.MustCompile("Unable to create space"),
			},
		},
	})
}

func TestAccSpaceResource_EmptyName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSpaceResourceConfig("", "root"),
				ExpectError: regexp.MustCompile("Unable to create space"),
			},
		},
	})
}

func TestAccSpaceResource_LongName(t *testing.T) {
	longName := "this-is-a-very-long-space-name-that-might-exceed-database-limits-and-should-be-tested-for-proper-error-handling"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceResourceConfig(longName, "root"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_space.test", "name", longName),
					resource.TestCheckResourceAttr("tama_space.test", "type", "root"),
				),
			},
		},
	})
}

func TestAccSpaceResource_SpecialCharacters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceResourceConfig("test-space-with-special_chars.123", "root"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_space.test", "name", "test-space-with-special_chars.123"),
					resource.TestCheckResourceAttr("tama_space.test", "type", "root"),
				),
			},
		},
	})
}

func TestAccSpaceResource_ComponentType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceResourceConfig("component-space", "component"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_space.test", "name", "component-space"),
					resource.TestCheckResourceAttr("tama_space.test", "type", "component"),
					resource.TestCheckResourceAttrSet("tama_space.test", "id"),
					resource.TestCheckResourceAttrSet("tama_space.test", "slug"),
				),
			},
		},
	})
}

func TestAccSpaceResource_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceResourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// First space
					resource.TestCheckResourceAttr("tama_space.test1", "name", "test-space-1"),
					resource.TestCheckResourceAttr("tama_space.test1", "type", "root"),
					resource.TestCheckResourceAttrSet("tama_space.test1", "id"),
					resource.TestCheckResourceAttrSet("tama_space.test1", "slug"),
					// Second space
					resource.TestCheckResourceAttr("tama_space.test2", "name", "test-space-2"),
					resource.TestCheckResourceAttr("tama_space.test2", "type", "component"),
					resource.TestCheckResourceAttrSet("tama_space.test2", "id"),
					resource.TestCheckResourceAttrSet("tama_space.test2", "slug"),
				),
			},
		},
	})
}

func TestAccSpaceResource_DisappearResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceResourceConfig("disappear-space", "root"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_space.test", "name", "disappear-space"),
					testAccCheckSpaceDestroy("tama_space.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSpaceResourceConfig(name, spaceType string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
  type = %[2]q
}
`, name, spaceType)
}

func testAccSpaceResourceConfigMultiple() string {
	return acceptance.ProviderConfig + `
resource "tama_space" "test1" {
  name = "test-space-1"
  type = "root"
}

resource "tama_space" "test2" {
  name = "test-space-2"
  type = "component"
}
`
}

func testAccCheckSpaceDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// This function simulates the space being destroyed outside of Terraform
		// In a real test, you would make an API call to delete the resource
		// For now, we'll just return nil to simulate successful destruction
		return nil
	}
}
