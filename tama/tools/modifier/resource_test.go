// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package modifier_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	tama "github.com/upmaru/tama-go"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

const modifierResourceName = "tama_thought_tool_modifier.test"

func TestAccThoughtToolModifierResource(t *testing.T) {
	spaceName := fmt.Sprintf("test-space-%d", time.Now().UnixNano())
	var initialID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtToolModifierConfig(
					spaceName,
					0,
					"/body/search/scope/user_id",
					"skip",
					"error",
					"actor_identifier",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(modifierResourceName, "id"),
					resource.TestCheckResourceAttrSet(modifierResourceName, "thought_tool_id"),
					resource.TestCheckResourceAttr(modifierResourceName, "index", "0"),
					resource.TestCheckResourceAttr(modifierResourceName, "target", "/body/search/scope/user_id"),
					resource.TestCheckResourceAttr(modifierResourceName, "on_missing_parent", "skip"),
					resource.TestCheckResourceAttr(modifierResourceName, "on_missing_source", "error"),
					resource.TestCheckResourceAttr(modifierResourceName, "source.type", "metadata"),
					resource.TestCheckResourceAttr(modifierResourceName, "source.path", "actor_identifier"),
					resource.TestCheckResourceAttr(modifierResourceName, "provision_state", "active"),
					captureResourceID(&initialID),
				),
			},
			{
				ResourceName:      modifierResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccThoughtToolModifierConfig(
					spaceName,
					0,
					"/body/search/query",
					"error",
					"skip",
					"current_timestamp",
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(modifierResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr(modifierResourceName, "id", &initialID),
					resource.TestCheckResourceAttr(modifierResourceName, "target", "/body/search/query"),
					resource.TestCheckResourceAttr(modifierResourceName, "on_missing_parent", "error"),
					resource.TestCheckResourceAttr(modifierResourceName, "on_missing_source", "skip"),
					resource.TestCheckResourceAttr(modifierResourceName, "source.path", "current_timestamp"),
				),
			},
			{
				Config: testAccThoughtToolModifierConfig(
					spaceName,
					0,
					"/body/search/scope/user_id",
					"skip",
					"error",
					"actor_identifier",
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(modifierResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.TestCheckResourceAttrPtr(modifierResourceName, "id", &initialID),
			},
			{
				Config: testAccThoughtToolModifierConfig(
					spaceName,
					1,
					"/body/search/scope/user_id",
					"skip",
					"error",
					"actor_identifier",
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(modifierResourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(modifierResourceName, "index", "1"),
					checkResourceIDChanged(&initialID),
				),
			},
			{
				Config: testAccThoughtToolModifierBaseConfig(spaceName),
			},
			{
				Config: testAccThoughtToolModifierConfig(
					spaceName,
					0,
					"/body/search/scope/user_id",
					"skip",
					"error",
					"actor_identifier",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr(modifierResourceName, "id", &initialID),
					resource.TestCheckResourceAttr(modifierResourceName, "provision_state", "active"),
				),
			},
		},
	})
}

func TestAccThoughtToolModifierResource_InvalidTarget(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtToolModifierConfig(
					fmt.Sprintf("test-space-%d", time.Now().UnixNano()),
					0,
					"/body/search/missing",
					"skip",
					"error",
					"actor_identifier",
				),
				ExpectError: regexp.MustCompile(`(?s)Unable to create thought tool modifier.*target`),
			},
		},
	})
}

func TestAccThoughtToolModifierResource_ExternalDeactivation(t *testing.T) {
	spaceName := fmt.Sprintf("test-space-%d", time.Now().UnixNano())
	var modifierID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtToolModifierConfig(
					spaceName,
					0,
					"/body/search/scope/user_id",
					"skip",
					"error",
					"actor_identifier",
				),
				Check:              captureAndDeactivateModifier(&modifierID),
				ExpectNonEmptyPlan: true,
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				Check:              checkModifierAbsent,
			},
		},
	})
}

func TestAccThoughtToolModifierResource_DeleteAlreadyInactive(t *testing.T) {
	spaceName := fmt.Sprintf("test-space-%d", time.Now().UnixNano())
	var modifierID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThoughtToolModifierConfig(
					spaceName,
					0,
					"/body/search/scope/user_id",
					"skip",
					"error",
					"actor_identifier",
				),
				Check: captureResourceID(&modifierID),
			},
			{
				Config: testAccThoughtToolModifierBaseConfig(spaceName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(modifierResourceName, plancheck.ResourceActionDestroy),
						deactivateModifierPlanCheck{id: &modifierID},
					},
				},
			},
		},
	})
}

func TestAccThoughtToolModifierResource_Validation(t *testing.T) {
	tests := map[string]struct {
		config string
		error  string
	}{
		"negative index": {
			config: testAccThoughtToolModifierValidationConfig(`index = -1`),
			error:  `between 0 and 2147483647`,
		},
		"index overflow": {
			config: testAccThoughtToolModifierValidationConfig(`index = 2147483648`),
			error:  `between 0 and 2147483647`,
		},
		"empty thought tool ID": {
			config: testAccThoughtToolModifierValidationConfig(`thought_tool_id = ""`),
			error:  `string length must be at least 1`,
		},
		"empty target": {
			config: testAccThoughtToolModifierValidationConfig(`target = ""`),
			error:  `string length must be at least 1`,
		},
		"invalid missing parent policy": {
			config: testAccThoughtToolModifierValidationConfig(`on_missing_parent = "ignore"`),
			error:  `value must be one of`,
		},
		"invalid missing source policy": {
			config: testAccThoughtToolModifierValidationConfig(`on_missing_source = "ignore"`),
			error:  `value must be one of`,
		},
		"absent source": {
			config: testAccThoughtToolModifierValidationConfig(`source = false`),
			error:  `Block source must have a configuration value`,
		},
		"invalid source type": {
			config: testAccThoughtToolModifierValidationConfig(`source_type = "request"`),
			error:  `value must be one of`,
		},
		"invalid source path": {
			config: testAccThoughtToolModifierValidationConfig(`source_path = "tenant_identifier"`),
			error:  `value must be one of`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      test.config,
						PlanOnly:    true,
						ExpectError: regexp.MustCompile(test.error),
					},
				},
			})
		})
	}
}

func captureResourceID(target *string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(modifierResourceName, "id", func(value string) error {
		*target = value
		return nil
	})
}

func checkResourceIDChanged(previous *string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(modifierResourceName, "id", func(value string) error {
		if value == *previous {
			return fmt.Errorf("expected modifier ID to change from %q after replacement", *previous)
		}
		return nil
	})
}

func captureAndDeactivateModifier(id *string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(modifierResourceName, "id", func(value string) error {
		*id = value
		return deleteModifier(value)
	})
}

func checkModifierAbsent(state *terraform.State) error {
	if _, ok := state.RootModule().Resources[modifierResourceName]; ok {
		return fmt.Errorf("expected %s to be removed from state", modifierResourceName)
	}
	return nil
}

type deactivateModifierPlanCheck struct {
	id *string
}

func (check deactivateModifierPlanCheck) CheckPlan(_ context.Context, _ plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	resp.Error = deleteModifier(*check.id)
}

func deleteModifier(id string) error {
	client, err := tama.NewClient(tama.Config{
		BaseURL:      os.Getenv("TAMA_BASE_URL"),
		ClientID:     os.Getenv("TAMA_CLIENT_ID"),
		ClientSecret: os.Getenv("TAMA_CLIENT_SECRET"),
		Scopes:       []string{"provision.all"},
	})
	if err != nil {
		return fmt.Errorf("create Tama client for external deactivation: %w", err)
	}

	if err := client.Tools.DeleteModifier(id); err != nil {
		return fmt.Errorf("externally deactivate modifier %q: %w", id, err)
	}
	return nil
}

func testAccThoughtToolModifierConfig(
	spaceName string,
	index int,
	target string,
	onMissingParent string,
	onMissingSource string,
	sourcePath string,
) string {
	return testAccThoughtToolModifierBaseConfig(spaceName) + fmt.Sprintf(`
resource "tama_thought_tool_modifier" "test" {
  thought_tool_id   = tama_thought_tool.test.id
  index             = %d
  target            = %q
  on_missing_parent = %q
  on_missing_source = %q

  source {
    type = "metadata"
    path = %q
  }
}
`, index, target, onMissingParent, onMissingSource, sourcePath)
}

func testAccThoughtToolModifierBaseConfig(spaceName string) string {
	return fmt.Sprintf(`
resource "tama_space" "test" {
  name = %q
  type = "root"
}

resource "tama_specification" "test" {
  space_id = tama_space.test.id
  version  = "1.0.0"
  endpoint = "https://search.example.com"
  schema   = jsonencode(jsondecode(file("${path.module}/testdata/search_schema.json")))

  wait_for {
    field {
      name = "current_state"
      in   = ["completed"]
    }
  }
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Modifier Test Chain"
}

resource "tama_modular_thought" "test" {
  chain_id = tama_chain.test.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

data "tama_action" "test" {
  specification_id = tama_specification.test.id
  identifier       = "scoped-search"
}

resource "tama_thought_tool" "test" {
  thought_id = tama_modular_thought.test.id
  action_id  = data.tama_action.test.id
}
`, spaceName)
}

func testAccThoughtToolModifierValidationConfig(override string) string {
	values := map[string]string{
		"thought_tool_id":   `"thought-tool-id"`,
		"index":             "0",
		"target":            `"/body/search/scope/user_id"`,
		"on_missing_parent": `"skip"`,
		"on_missing_source": `"error"`,
		"source":            "true",
		"source_type":       `"metadata"`,
		"source_path":       `"actor_identifier"`,
	}

	var key, value string
	if _, err := fmt.Sscanf(override, `%s = %s`, &key, &value); err == nil {
		values[key] = value
	}

	config := fmt.Sprintf(`
resource "tama_thought_tool_modifier" "test" {
  thought_tool_id   = %s
  index             = %s
  target            = %s
  on_missing_parent = %s
  on_missing_source = %s
`, values["thought_tool_id"], values["index"], values["target"], values["on_missing_parent"], values["on_missing_source"])

	if values["source"] == "true" {
		config += fmt.Sprintf(`
  source {
    type = %s
    path = %s
  }
`, values["source_type"], values["source_path"])
	}

	return config + "}\n"
}
