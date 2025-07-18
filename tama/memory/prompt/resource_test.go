// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package prompt_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccPromptResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPromptResourceConfig("test-prompt", "You are a helpful assistant", "system"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_prompt.test", "name", "test-prompt"),
					resource.TestCheckResourceAttr("tama_prompt.test", "content", "You are a helpful assistant"),
					resource.TestCheckResourceAttr("tama_prompt.test", "role", "system"),
					resource.TestCheckResourceAttrSet("tama_prompt.test", "id"),
					resource.TestCheckResourceAttrSet("tama_prompt.test", "space_id"),
					resource.TestCheckResourceAttrSet("tama_prompt.test", "slug"),
					resource.TestCheckResourceAttrSet("tama_prompt.test", "current_state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "tama_prompt.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccPromptResourceConfig("updated-prompt", "You are an expert assistant", "user"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_prompt.test", "name", "updated-prompt"),
					resource.TestCheckResourceAttr("tama_prompt.test", "slug", "updated-prompt"),
					resource.TestCheckResourceAttr("tama_prompt.test", "content", "You are an expert assistant"),
					resource.TestCheckResourceAttr("tama_prompt.test", "role", "user"),
					resource.TestCheckResourceAttrSet("tama_prompt.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccPromptResource_SystemRole(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptResourceConfig("system-prompt", "You are a coding assistant specialized in Go", "system"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_prompt.test", "name", "system-prompt"),
					resource.TestCheckResourceAttr("tama_prompt.test", "content", "You are a coding assistant specialized in Go"),
					resource.TestCheckResourceAttr("tama_prompt.test", "role", "system"),
					resource.TestCheckResourceAttrSet("tama_prompt.test", "id"),
					resource.TestCheckResourceAttrSet("tama_prompt.test", "slug"),
				),
			},
		},
	})
}

func TestAccPromptResource_UserRole(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptResourceConfig("user-prompt", "Please help me with this problem", "user"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_prompt.test", "name", "user-prompt"),
					resource.TestCheckResourceAttr("tama_prompt.test", "content", "Please help me with this problem"),
					resource.TestCheckResourceAttr("tama_prompt.test", "role", "user"),
					resource.TestCheckResourceAttrSet("tama_prompt.test", "id"),
					resource.TestCheckResourceAttrSet("tama_prompt.test", "slug"),
				),
			},
		},
	})
}

func TestAccPromptResource_LongContent(t *testing.T) {
	longContent := `You are an AI assistant that helps users with various tasks.
Your primary responsibilities include:
1. Answering questions accurately and helpfully
2. Providing clear explanations
3. Being respectful and professional
4. Admitting when you don't know something
5. Offering to help find solutions

Please always maintain a helpful and friendly tone while being informative and concise.`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptResourceConfig("detailed-prompt", longContent, "system"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_prompt.test", "name", "detailed-prompt"),
					resource.TestCheckResourceAttr("tama_prompt.test", "content", longContent),
					resource.TestCheckResourceAttr("tama_prompt.test", "role", "system"),
					resource.TestCheckResourceAttrSet("tama_prompt.test", "id"),
				),
			},
		},
	})
}

func TestAccPromptResource_SpecialCharacters(t *testing.T) {
	specialContent := "Handle these characters: !@#$%^&*()_+-=[]{}|;':\",./<>?"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptResourceConfig("special-chars", specialContent, "system"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_prompt.test", "name", "special-chars"),
					resource.TestCheckResourceAttr("tama_prompt.test", "content", specialContent),
					resource.TestCheckResourceAttr("tama_prompt.test", "role", "system"),
					resource.TestCheckResourceAttrSet("tama_prompt.test", "id"),
				),
			},
		},
	})
}

func TestAccPromptResource_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptResourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_prompt.system", "name", "system-prompt"),
					resource.TestCheckResourceAttr("tama_prompt.system", "role", "system"),
					resource.TestCheckResourceAttr("tama_prompt.user", "name", "user-prompt"),
					resource.TestCheckResourceAttr("tama_prompt.user", "role", "user"),
				),
			},
		},
	})
}

func TestAccPromptResource_DifferentSpaces(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptResourceConfigDifferentSpaces(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_prompt.space1", "name", "prompt-space1"),
					resource.TestCheckResourceAttr("tama_prompt.space2", "name", "prompt-space2"),
					resource.TestCheckResourceAttrPair("tama_prompt.space1", "space_id", "tama_space.test_space1", "id"),
					resource.TestCheckResourceAttrPair("tama_prompt.space2", "space_id", "tama_space.test_space2", "id"),
				),
			},
		},
	})
}

func TestAccPromptResource_DisappearResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		CheckDestroy:             nil, // We expect this to be recreated
		Steps: []resource.TestStep{
			{
				Config: testAccPromptResourceConfig("disappear-test", "This prompt will disappear", "system"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_prompt.test", "name", "disappear-test"),
				),
			},
		},
	})
}

func TestAccPromptResource_InvalidRole(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPromptResourceConfig("invalid-role", "Test content", "assistant"),
				ExpectError: regexp.MustCompile("role is invalid"),
			},
		},
	})
}

func TestAccPromptResource_EmptyContent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPromptResourceConfig("empty-content", "", "system"),
				ExpectError: regexp.MustCompile(".*content.*required.*"),
			},
		},
	})
}

func TestAccPromptResource_EmptyName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPromptResourceConfig("", "Test content", "system"),
				ExpectError: regexp.MustCompile(".*name.*required.*"),
			},
		},
	})
}

func TestAccPromptResource_ContentWithNewlines(t *testing.T) {
	contentWithNewlines := `You are a helpful assistant.

Please follow these guidelines:
- Be respectful
- Provide accurate information
- Ask for clarification when needed

Thank you!`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptResourceConfig("multiline-prompt", contentWithNewlines, "system"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_prompt.test", "name", "multiline-prompt"),
					resource.TestCheckResourceAttr("tama_prompt.test", "content", contentWithNewlines),
					resource.TestCheckResourceAttr("tama_prompt.test", "role", "system"),
				),
			},
		},
	})
}

func testAccPromptResourceConfig(name, content, role string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-prompt-%d"
  type = "root"
}

resource "tama_prompt" "test" {
  space_id = tama_space.test_space.id
  name     = %[2]q
  content  = %[3]q
  role     = %[4]q
}
`, timestamp, name, content, role)
}

func testAccPromptResourceConfigMultiple() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-prompts-%d"
  type = "root"
}

resource "tama_prompt" "system" {
  space_id = tama_space.test_space.id
  name     = "system-prompt"
  content  = "You are a helpful system assistant"
  role     = "system"
}

resource "tama_prompt" "user" {
  space_id = tama_space.test_space.id
  name     = "user-prompt"
  content  = "Please help me with this task"
  role     = "user"
}
`, timestamp)
}

func testAccPromptResourceConfigDifferentSpaces() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space1" {
  name = "test-space1-for-prompt-%d"
  type = "root"
}

resource "tama_space" "test_space2" {
  name = "test-space2-for-prompt-%d"
  type = "root"
}

resource "tama_prompt" "space1" {
  space_id = tama_space.test_space1.id
  name     = "prompt-space1"
  content  = "Content for space 1"
  role     = "system"
}

resource "tama_prompt" "space2" {
  space_id = tama_space.test_space2.id
  name     = "prompt-space2"
  content  = "Content for space 2"
  role     = "system"
}
`, timestamp, timestamp)
}
