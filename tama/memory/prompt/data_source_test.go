// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package prompt_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccPromptDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptDataSourceConfig("test-prompt", "You are a helpful assistant", "system"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_prompt.test", "name", "test-prompt"),
					resource.TestCheckResourceAttr("data.tama_prompt.test", "content", "You are a helpful assistant"),
					resource.TestCheckResourceAttr("data.tama_prompt.test", "role", "system"),
					resource.TestCheckResourceAttrSet("data.tama_prompt.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_prompt.test", "space_id"),
					resource.TestCheckResourceAttrSet("data.tama_prompt.test", "slug"),
					resource.TestCheckResourceAttrSet("data.tama_prompt.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccPromptDataSource_SystemRole(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptDataSourceConfig("system-prompt", "You are a coding assistant", "system"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_prompt.test", "name", "system-prompt"),
					resource.TestCheckResourceAttr("data.tama_prompt.test", "content", "You are a coding assistant"),
					resource.TestCheckResourceAttr("data.tama_prompt.test", "role", "system"),
					resource.TestCheckResourceAttrSet("data.tama_prompt.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_prompt.test", "slug"),
				),
			},
		},
	})
}

func TestAccPromptDataSource_UserRole(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptDataSourceConfig("user-prompt", "Please help me solve this", "user"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_prompt.test", "name", "user-prompt"),
					resource.TestCheckResourceAttr("data.tama_prompt.test", "content", "Please help me solve this"),
					resource.TestCheckResourceAttr("data.tama_prompt.test", "role", "user"),
					resource.TestCheckResourceAttrSet("data.tama_prompt.test", "id"),
				),
			},
		},
	})
}

func TestAccPromptDataSource_LongContent(t *testing.T) {
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
				Config: testAccPromptDataSourceConfig("detailed-prompt", longContent, "system"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_prompt.test", "name", "detailed-prompt"),
					resource.TestCheckResourceAttr("data.tama_prompt.test", "content", longContent),
					resource.TestCheckResourceAttr("data.tama_prompt.test", "role", "system"),
					resource.TestCheckResourceAttrSet("data.tama_prompt.test", "id"),
				),
			},
		},
	})
}

func TestAccPromptDataSource_SpecialCharacters(t *testing.T) {
	specialContent := "Handle these characters: !@#$%^&*()_+-=[]{}|;':\",./<>?"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptDataSourceConfig("special-chars", specialContent, "system"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_prompt.test", "name", "special-chars"),
					resource.TestCheckResourceAttr("data.tama_prompt.test", "content", specialContent),
					resource.TestCheckResourceAttr("data.tama_prompt.test", "role", "system"),
					resource.TestCheckResourceAttrSet("data.tama_prompt.test", "id"),
				),
			},
		},
	})
}

func TestAccPromptDataSource_ContentWithNewlines(t *testing.T) {
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
				Config: testAccPromptDataSourceConfig("multiline-prompt", contentWithNewlines, "system"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_prompt.test", "name", "multiline-prompt"),
					resource.TestCheckResourceAttr("data.tama_prompt.test", "content", contentWithNewlines),
					resource.TestCheckResourceAttr("data.tama_prompt.test", "role", "system"),
				),
			},
		},
	})
}

func TestAccPromptDataSource_ComplexScenario(t *testing.T) {
	complexContent := `You are an expert software engineer with deep knowledge in:
- Go programming language
- Terraform providers
- API design and development
- Testing methodologies

When responding to queries:
1. Provide clear, concise explanations
2. Include relevant code examples when appropriate
3. Suggest best practices
4. Mention potential pitfalls or considerations

Always structure your responses with proper formatting and clear sections.`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptDataSourceConfig("complex-expert", complexContent, "system"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_prompt.test", "name", "complex-expert"),
					resource.TestCheckResourceAttr("data.tama_prompt.test", "content", complexContent),
					resource.TestCheckResourceAttr("data.tama_prompt.test", "role", "system"),
					resource.TestCheckResourceAttrSet("data.tama_prompt.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_prompt.test", "space_id"),
					resource.TestCheckResourceAttrSet("data.tama_prompt.test", "slug"),
					resource.TestCheckResourceAttrSet("data.tama_prompt.test", "provision_state"),
				),
			},
		},
	})
}

func TestAccPromptDataSource_MinimalContent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptDataSourceConfig("minimal", "Hi", "user"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tama_prompt.test", "name", "minimal"),
					resource.TestCheckResourceAttr("data.tama_prompt.test", "content", "Hi"),
					resource.TestCheckResourceAttr("data.tama_prompt.test", "role", "user"),
					resource.TestCheckResourceAttrSet("data.tama_prompt.test", "id"),
				),
			},
		},
	})
}

func testAccPromptDataSourceConfig(name, content, role string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-prompt-ds-%d"
  type = "root"
}

resource "tama_prompt" "test" {
  space_id = tama_space.test_space.id
  name     = %[2]q
  content  = %[3]q
  role     = %[4]q
}

data "tama_prompt" "test" {
  id = tama_prompt.test.id
}
`, timestamp, name, content, role)
}
