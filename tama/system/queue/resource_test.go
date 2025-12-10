// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package queue_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccQueueResource(t *testing.T) {
	queueName := fmt.Sprintf("conversation-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueueResourceConfig(queueName, 24),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_queue.test", "id"),
					resource.TestCheckResourceAttr("tama_queue.test", "role", "oracle"),
					resource.TestCheckResourceAttr("tama_queue.test", "name", queueName),
					resource.TestCheckResourceAttr("tama_queue.test", "concurrency", "24"),
				),
			},
			{
				ResourceName:      "tama_queue.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueResourceConfig(queueName, 48),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_queue.test", "id"),
					resource.TestCheckResourceAttr("tama_queue.test", "role", "oracle"),
					resource.TestCheckResourceAttr("tama_queue.test", "name", queueName),
					resource.TestCheckResourceAttr("tama_queue.test", "concurrency", "48"),
				),
			},
		},
	})
}

func testAccQueueResourceConfig(name string, concurrency int) string {
	return fmt.Sprintf(`
resource "tama_queue" "test" {
  role        = "oracle"
  name        = "%s"
  concurrency = %d
}
`, name, concurrency)
}
