// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package space_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccSpaceDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceDataSourceConfig("test-space", "root"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_space.test", "name"),
					resource.TestCheckResourceAttr("data.tama_space.test", "type", "root"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "slug"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccSpaceDataSource_RootType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceDataSourceConfig("root-space", "root"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_space.test", "name"),
					resource.TestCheckResourceAttr("data.tama_space.test", "type", "root"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "slug"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccSpaceDataSource_ComponentType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceDataSourceConfig("component-space", "component"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_space.test", "name"),
					resource.TestCheckResourceAttr("data.tama_space.test", "type", "component"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "slug"),
				),
			},
		},
	})
}

func TestAccSpaceDataSource_LongName(t *testing.T) {
	longName := "this-is-a-very-long-space-name-that-tests-the-handling-of-longer-names"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceDataSourceConfig(longName, "root"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_space.test", "name"),
					resource.TestCheckResourceAttr("data.tama_space.test", "type", "root"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "slug"),
				),
			},
		},
	})
}

func TestAccSpaceDataSource_SpecialCharacters(t *testing.T) {
	specialName := "test-space-with-numbers-123"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceDataSourceConfig(specialName, "component"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_space.test", "name"),
					resource.TestCheckResourceAttr("data.tama_space.test", "type", "component"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "slug"),
				),
			},
		},
	})
}

func TestAccSpaceDataSource_MultipleSpaces(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceDataSourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check first space (root)
					resource.TestCheckResourceAttrSet("data.tama_space.test_root", "name"),
					resource.TestCheckResourceAttr("data.tama_space.test_root", "type", "root"),
					resource.TestCheckResourceAttrSet("data.tama_space.test_root", "id"),
					resource.TestCheckResourceAttrSet("data.tama_space.test_root", "slug"),

					// Check second space (component)
					resource.TestCheckResourceAttrSet("data.tama_space.test_component", "name"),
					resource.TestCheckResourceAttr("data.tama_space.test_component", "type", "component"),
					resource.TestCheckResourceAttrSet("data.tama_space.test_component", "id"),
					resource.TestCheckResourceAttrSet("data.tama_space.test_component", "slug"),
				),
			},
		},
	})
}

func TestAccSpaceDataSource_VerifyAllAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceDataSourceConfig("verify-attrs", "root"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify all required attributes are present
					resource.TestCheckResourceAttrSet("data.tama_space.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "name"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "type"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "slug"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "provision_state"),

					// Verify specific values
					resource.TestCheckResourceAttrSet("data.tama_space.test", "name"),
					resource.TestCheckResourceAttr("data.tama_space.test", "type", "root"),
				),
			},
		},
	})
}

func TestAccSpaceDataSource_MinimalName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceDataSourceConfig("x", "root"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_space.test", "name"),
					resource.TestCheckResourceAttr("data.tama_space.test", "type", "root"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "slug"),
				),
			},
		},
	})
}

func TestAccSpaceDataSource_ComplexScenario(t *testing.T) {
	complexName := "ai-development-workspace-v2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceDataSourceConfig(complexName, "component"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_space.test", "name"),
					resource.TestCheckResourceAttr("data.tama_space.test", "type", "component"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "slug"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccSpaceDataSource_StateVerification(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceDataSourceConfig("state-test", "root"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_space.test", "name"),
					resource.TestCheckResourceAttr("data.tama_space.test", "type", "root"),
					resource.TestCheckResourceAttrSet("data.tama_space.test", "provision_state"),
					// Verify that provision_state is not empty
					resource.TestCheckNoResourceAttr("data.tama_space.test", "provision_state.#"),
				),
			},
		},
	})
}

func testAccSpaceDataSourceConfig(name, spaceType string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s-%d"
  type = %q
}

data "tama_space" "test" {
  id = tama_space.test.id
}
`, name, timestamp, spaceType)
}

func testAccSpaceDataSourceConfigMultiple() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_root" {
  name = "multi-test-root-%d"
  type = "root"
}

resource "tama_space" "test_component" {
  name = "multi-test-component-%d"
  type = "component"
}

data "tama_space" "test_root" {
  id = tama_space.test_root.id
}

data "tama_space" "test_component" {
  id = tama_space.test_component.id
}
`, timestamp, timestamp)
}
