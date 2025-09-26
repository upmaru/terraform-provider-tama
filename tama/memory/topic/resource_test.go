// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package topic_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccListenerTopicResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccListenerTopicResourceConfig(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_listener_topic.test", "id"),
					resource.TestCheckResourceAttrSet("tama_listener_topic.test", "listener_id"),
					resource.TestCheckResourceAttrSet("tama_listener_topic.test", "class_id"),
					resource.TestCheckResourceAttrSet("tama_listener_topic.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_listener_topic.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing (switch to another class)
			{
				Config: testAccListenerTopicResourceConfigUpdate(fmt.Sprintf("test-space-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_listener_topic.test", "id"),
					resource.TestCheckResourceAttrSet("tama_listener_topic.test", "listener_id"),
					resource.TestCheckResourceAttrSet("tama_listener_topic.test", "class_id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccListenerTopicResourceConfig(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_class" "test" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "action-call"
    description = "An action call is a request to execute an action."
    type = "object"
    properties = {
      tool_id = {
        description = "The ID of the tool to execute"
        type        = "string"
      }
      parameters = {
        description = "The parameters to pass to the action"
        type        = "object"
      }
    }
    required = ["tool_id", "parameters"]
  })
}

resource "tama_listener" "test" {
  space_id = tama_space.test.id
  endpoint = "http://localhost:4000/tama/activities"
  secret   = "super-secret"
}

resource "tama_listener_topic" "test" {
  listener_id = tama_listener.test.id
  class_id    = tama_class.test.id
}
`, spaceName)
}

func testAccListenerTopicResourceConfigUpdate(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_class" "test_a" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "schema-a"
    description = "Schema A"
    type = "object"
    properties = {
      name = {
        description = "Name"
        type        = "string"
      }
    }
    required = ["name"]
  })
}

resource "tama_class" "test_b" {
  space_id = tama_space.test.id
  schema_json = jsonencode({
    title = "schema-b"
    description = "Schema B"
    type = "object"
    properties = {
      id = {
        description = "ID"
        type        = "string"
      }
    }
    required = ["id"]
  })
}

resource "tama_listener" "test" {
  space_id = tama_space.test.id
  endpoint = "http://localhost:4000/tama/activities"
  secret   = "super-secret"
}

resource "tama_listener_topic" "test" {
  listener_id = tama_listener.test.id
  class_id    = tama_class.test_b.id
}
`, spaceName)
}
