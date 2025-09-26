// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package listener_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccListenerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccListenerResourceConfig(
					fmt.Sprintf("test-listener-space-%d", time.Now().UnixNano()),
					"http://localhost:4000/tama/activities",
					"super-secret",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_listener.test", "id"),
					resource.TestCheckResourceAttrSet("tama_listener.test", "space_id"),
					resource.TestCheckResourceAttr("tama_listener.test", "endpoint", "http://localhost:4000/tama/activities"),
					resource.TestCheckResourceAttrSet("tama_listener.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "tama_listener.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
			// Update and Read testing
			{
				Config: testAccListenerResourceConfig(
					fmt.Sprintf("test-listener-space-%d", time.Now().UnixNano()),
					"http://localhost:5000/new-endpoint",
					"super-secret",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_listener.test", "id"),
					resource.TestCheckResourceAttr("tama_listener.test", "endpoint", "http://localhost:5000/new-endpoint"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccListenerResourceConfig(spaceName string, endpoint string, secret string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_listener" "test" {
  space_id = tama_space.test.id
  endpoint = "%s"
  secret   = "%s"
}
`, spaceName, endpoint, secret)
}
