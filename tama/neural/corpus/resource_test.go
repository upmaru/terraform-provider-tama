// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package corpus_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccCorpusResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCorpusResourceConfig(fmt.Sprintf("test-corpus-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class_corpus.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class_corpus.test", "class_id"),
					resource.TestCheckResourceAttr("tama_class_corpus.test", "name", "Test Corpus"),
					resource.TestCheckResourceAttr("tama_class_corpus.test", "main", "true"),
					resource.TestCheckResourceAttr("tama_class_corpus.test", "template", "{{ data.something }}"),
					resource.TestCheckResourceAttrSet("tama_class_corpus.test", "slug"),
					resource.TestCheckResourceAttrSet("tama_class_corpus.test", "provision_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_class_corpus.test",
				ImportState:       true,
				ImportStateVerify: false, // Skip verification due to class_id not being available in import
			},
			// Update and Read testing
			{
				Config: testAccCorpusResourceConfigUpdate(fmt.Sprintf("test-corpus-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class_corpus.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class_corpus.test", "class_id"),
					resource.TestCheckResourceAttr("tama_class_corpus.test", "name", "Updated Corpus"),
					resource.TestCheckResourceAttr("tama_class_corpus.test", "main", "false"),
					resource.TestCheckResourceAttr("tama_class_corpus.test", "template", "{{ data.updated }}"),
					resource.TestCheckResourceAttrSet("tama_class_corpus.test", "slug"),
					resource.TestCheckResourceAttrSet("tama_class_corpus.test", "provision_state"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccCorpusResource_DefaultMain(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCorpusResourceConfigDefaultMain(fmt.Sprintf("test-corpus-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class_corpus.test", "id"),
					resource.TestCheckResourceAttr("tama_class_corpus.test", "main", "false"),
					resource.TestCheckResourceAttrSet("tama_class_corpus.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccCorpusResource_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCorpusResourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// First corpus
					resource.TestCheckResourceAttrSet("tama_class_corpus.test1", "id"),
					resource.TestCheckResourceAttr("tama_class_corpus.test1", "name", "Main Corpus"),
					resource.TestCheckResourceAttr("tama_class_corpus.test1", "main", "true"),
					resource.TestCheckResourceAttrSet("tama_class_corpus.test1", "provision_state"),
					// Second corpus
					resource.TestCheckResourceAttrSet("tama_class_corpus.test2", "id"),
					resource.TestCheckResourceAttr("tama_class_corpus.test2", "name", "Secondary Corpus"),
					resource.TestCheckResourceAttr("tama_class_corpus.test2", "main", "false"),
					resource.TestCheckResourceAttrSet("tama_class_corpus.test2", "provision_state"),
				),
			},
		},
	})
}

func TestAccCorpusResource_ClassIdChange(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCorpusResourceConfigClassChange("initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class_corpus.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class_corpus.test", "class_id"),
				),
			},
			{
				Config: testAccCorpusResourceConfigClassChange("updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class_corpus.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class_corpus.test", "class_id"),
				),
			},
		},
	})
}

func testAccCorpusResourceConfig(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
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

resource "tama_class_corpus" "test" {
  class_id = tama_class.test.id
  name     = "Test Corpus"
  main     = true
  template = "{{ data.something }}"
}
`, spaceName)
}

func testAccCorpusResourceConfigUpdate(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
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

resource "tama_class_corpus" "test" {
  class_id = tama_class.test.id
  name     = "Updated Corpus"
  main     = false
  template = "{{ data.updated }}"
}
`, spaceName)
}

func testAccCorpusResourceConfigDefaultMain(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = %[1]q
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
    }
    required = ["tool_id"]
  })
}

resource "tama_class_corpus" "test" {
  class_id = tama_class.test.id
  name     = "Default Main Corpus"
  template = "{{ data.default }}"
}
`, spaceName)
}

func testAccCorpusResourceConfigMultiple() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "test-space-%d"
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

resource "tama_class_corpus" "test1" {
  class_id = tama_class.test.id
  name     = "Main Corpus"
  main     = true
  template = "{{ data.main }}"
}

resource "tama_class_corpus" "test2" {
  class_id = tama_class.test.id
  name     = "Secondary Corpus"
  main     = false
  template = "{{ data.secondary }}"
}
`, timestamp)
}

func testAccCorpusResourceConfigClassChange(suffix string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_%s" {
  name = "test-space-%s-%d"
  type = "root"
}

resource "tama_class" "test_%s" {
  space_id = tama_space.test_%s.id
  schema_json = jsonencode({
    title = "action-call-%s"
    description = "An action call is a request to execute an action."
    type = "object"
    properties = {
      tool_id = {
        description = "The ID of the tool to execute"
        type        = "string"
      }
    }
    required = ["tool_id"]
  })
}

resource "tama_class_corpus" "test" {
  class_id = tama_class.test_%s.id
  name     = "Test Corpus"
  main     = true
  template = "{{ data.test }}"
}
`, suffix, suffix, timestamp, suffix, suffix, suffix, suffix)
}
