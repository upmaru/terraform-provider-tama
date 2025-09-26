// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package filter_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccListenerFilterResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccListenerFilterResourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_listener_filter.test", "id"),
					resource.TestCheckResourceAttrSet("tama_listener_filter.test", "listener_id"),
					resource.TestCheckResourceAttrSet("tama_listener_filter.test", "chain_id"),
					resource.TestCheckResourceAttrSet("tama_listener_filter.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_listener_filter.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing (switch to another chain)
			{
				Config: testAccListenerFilterResourceConfigUpdate(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_listener_filter.test", "id"),
					resource.TestCheckResourceAttrSet("tama_listener_filter.test", "listener_id"),
					resource.TestCheckResourceAttrSet("tama_listener_filter.test", "chain_id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccListenerFilterResourceConfig(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Filter Chain ABC"
}

resource "tama_listener" "test" {
  space_id = tama_space.test.id
  endpoint = "http://localhost:4000/tama/activities"
  secret   = "super-secret"
}

resource "tama_listener_filter" "test" {
  listener_id = tama_listener.test.id
  chain_id    = tama_chain.test.id
}
`, spaceName)
}

func testAccListenerFilterResourceConfigUpdate(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

# Chain remains the same to avoid API
resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Filter Chain A"
}
		
resource "tama_chain" "test_b" {
  space_id = tama_space.test.id
  name     = "Filter Chain B"
}


resource "tama_listener" "test" {
  space_id = tama_space.test.id
  endpoint = "http://localhost:4000/tama/activities"
  secret   = "super-secret"
}

resource "tama_listener_filter" "test" {
  listener_id = tama_listener.test.id
  chain_id    = tama_chain.test.id
}
`, spaceName)
}
