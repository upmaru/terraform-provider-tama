// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package modifier_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

// testCheckSchemaJSONEqual compares the schema JSON string semantically
func testCheckSchemaJSONEqual(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["tama_action_modifier.test"]
		if !ok {
			return fmt.Errorf("Not found: %s", "tama_action_modifier.test")
		}

		actual := rs.Primary.Attributes["schema"]

		var expectedObj, actualObj any
		if err := json.Unmarshal([]byte(expected), &expectedObj); err != nil {
			return fmt.Errorf("Expected value is not valid JSON: %v", err)
		}
		if err := json.Unmarshal([]byte(actual), &actualObj); err != nil {
			return fmt.Errorf("Actual value is not valid JSON: %v", err)
		}

		expectedNorm, _ := json.Marshal(expectedObj)
		actualNorm, _ := json.Marshal(actualObj)
		if string(expectedNorm) != string(actualNorm) {
			return fmt.Errorf("JSON values are not equal.\nExpected: %s\nActual: %s", expected, actual)
		}
		return nil
	}
}

func TestAccActionModifierResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccActionModifierResourceConfig("region", `{"type":"string","description":"the region the user is in"}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_action_modifier.test", "name", "region"),
					resource.TestCheckResourceAttrSet("tama_action_modifier.test", "id"),
					resource.TestCheckResourceAttrSet("tama_action_modifier.test", "provision_state"),
					// action_id should match the data source action id
					resource.TestCheckResourceAttrPair("tama_action_modifier.test", "action_id", "data.tama_action.test", "id"),
					testCheckSchemaJSONEqual(`{"type":"string","description":"the region the user is in"}`),
				),
			},
			// Import
			{
				ResourceName:      "tama_action_modifier.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update
			{
				Config: testAccActionModifierResourceConfig("user-region", `{"type":"string","description":"user selected region"}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_action_modifier.test", "name", "user-region"),
					testCheckSchemaJSONEqual(`{"type":"string","description":"user selected region"}`),
				),
			},
		},
	})
}

func testAccActionModifierResourceConfig(name, schema string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-action-modifier-%d"
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
  identifier       = "create-index"
}

resource "tama_action_modifier" "test" {
  action_id = data.tama_action.test.id
  name      = %q
  schema    = jsonencode(jsondecode(%q))
}
`, timestamp, name, schema)
}
