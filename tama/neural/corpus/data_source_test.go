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

func TestAccCorpusDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCorpusDataSourceConfig("test-corpus"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test", "id"),
					resource.TestCheckResourceAttr("data.tama_class_corpus.test", "name", "Test Corpus"),
					resource.TestCheckResourceAttr("data.tama_class_corpus.test", "main", "true"),
					resource.TestCheckResourceAttr("data.tama_class_corpus.test", "template", "{{ data.something }}"),
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test", "slug"),
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccCorpusDataSource_MainFalse(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCorpusDataSourceConfigMainFalse("test-corpus-false"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test", "id"),
					resource.TestCheckResourceAttr("data.tama_class_corpus.test", "name", "Secondary Corpus"),
					resource.TestCheckResourceAttr("data.tama_class_corpus.test", "main", "false"),
					resource.TestCheckResourceAttr("data.tama_class_corpus.test", "template", "{{ data.secondary }}"),
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test", "slug"),
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccCorpusDataSource_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCorpusDataSourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check first corpus
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test_main", "id"),
					resource.TestCheckResourceAttr("data.tama_class_corpus.test_main", "name", "Main Corpus"),
					resource.TestCheckResourceAttr("data.tama_class_corpus.test_main", "main", "true"),
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test_main", "provision_state"),

					// Check second corpus
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test_secondary", "id"),
					resource.TestCheckResourceAttr("data.tama_class_corpus.test_secondary", "name", "Secondary Corpus"),
					resource.TestCheckResourceAttr("data.tama_class_corpus.test_secondary", "main", "false"),
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test_secondary", "provision_state"),
				),
			},
		},
	})
}

func TestAccCorpusDataSource_VerifyAllAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCorpusDataSourceConfig("verify-attrs"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify all required attributes are present
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test", "name"),
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test", "slug"),
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test", "main"),
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test", "template"),
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test", "provision_state"),

					// Verify that provision_state is not empty
					resource.TestCheckNoResourceAttr("data.tama_class_corpus.test", "provision_state.#"),
				),
			},
		},
	})
}

func TestAccCorpusDataSource_TemplateContent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCorpusDataSourceConfigComplexTemplate("template-test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test", "id"),
					resource.TestCheckResourceAttr("data.tama_class_corpus.test", "template", "{{ data.user.name }} - {{ data.context.timestamp }}"),
					// Verify the template contains expected content
					resource.TestCheckResourceAttrWith("data.tama_class_corpus.test", "template", func(value string) error {
						if value == "" {
							return fmt.Errorf("template should not be empty")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccCorpusDataSource_StateVerification(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCorpusDataSourceConfig("state-test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test", "name"),
					resource.TestCheckResourceAttrSet("data.tama_class_corpus.test", "provision_state"),
				),
			},
		},
	})
}

func testAccCorpusDataSourceConfig(name string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s-%d"
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

data "tama_class_corpus" "test" {
  id = tama_class_corpus.test.id
}
`, name, timestamp)
}

func testAccCorpusDataSourceConfigMainFalse(name string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s-%d"
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
  name     = "Secondary Corpus"
  main     = false
  template = "{{ data.secondary }}"
}

data "tama_class_corpus" "test" {
  id = tama_class_corpus.test.id
}
`, name, timestamp)
}

func testAccCorpusDataSourceConfigMultiple() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "multi-test-%d"
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

resource "tama_class_corpus" "test_main" {
  class_id = tama_class.test.id
  name     = "Main Corpus"
  main     = true
  template = "{{ data.main }}"
}

resource "tama_class_corpus" "test_secondary" {
  class_id = tama_class.test.id
  name     = "Secondary Corpus"
  main     = false
  template = "{{ data.secondary }}"
}

data "tama_class_corpus" "test_main" {
  id = tama_class_corpus.test_main.id
}

data "tama_class_corpus" "test_secondary" {
  id = tama_class_corpus.test_secondary.id
}
`, timestamp)
}

func testAccCorpusDataSourceConfigComplexTemplate(name string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test" {
  name = "%s-%d"
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
  name     = "Complex Template Corpus"
  main     = true
  template = "{{ data.user.name }} - {{ data.context.timestamp }}"
}

data "tama_class_corpus" "test" {
  id = tama_class_corpus.test.id
}
`, name, timestamp)
}
