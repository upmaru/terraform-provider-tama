// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package model_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccModelResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccModelResourceConfig("mistral-small-latest", "/chat/completions"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_model.test", "identifier", "mistral-small-latest"),
					resource.TestCheckResourceAttr("tama_model.test", "path", "/chat/completions"),
					resource.TestCheckResourceAttrSet("tama_model.test", "id"),
					resource.TestCheckResourceAttrSet("tama_model.test", "source_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "tama_model.test",
				ImportState:             true,
				ImportStateVerify:       false, // SourceId and Path cannot be retrieved from API
				ImportStateVerifyIgnore: []string{"source_id", "path"},
			},
			// Update and Read testing
			{
				Config: testAccModelResourceConfig("mistral-large-latest", "/v1/chat/completions"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_model.test", "identifier", "mistral-large-latest"),
					resource.TestCheckResourceAttr("tama_model.test", "path", "/v1/chat/completions"),
					resource.TestCheckResourceAttrSet("tama_model.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccModelResource_OpenAIModels(t *testing.T) {
	testCases := []struct {
		name       string
		identifier string
		path       string
	}{
		{"GPT-3.5", "gpt-3.5-turbo", "/v1/chat/completions"},
		{"GPT-4", "gpt-4", "/v1/chat/completions"},
		{"GPT-4 Turbo", "gpt-4-turbo", "/v1/chat/completions"},
		{"Text Davinci", "text-davinci-003", "/chat/completions"},
		{"Text Embedding", "text-embedding-ada-002", "/v1/embeddings"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccModelResourceConfig(tc.identifier, tc.path),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("tama_model.test", "identifier", tc.identifier),
							resource.TestCheckResourceAttr("tama_model.test", "path", tc.path),
							resource.TestCheckResourceAttrSet("tama_model.test", "id"),
						),
					},
				},
			})
		})
	}
}

func TestAccModelResource_AnthropicModels(t *testing.T) {
	testCases := []struct {
		name       string
		identifier string
		path       string
	}{
		{"Claude 3 Sonnet", "claude-3-sonnet-20240229", "/chat/completions"},
		{"Claude 3 Opus", "claude-3-opus-20240229", "/chat/completions"},
		{"Claude 3 Haiku", "claude-3-haiku-20240307", "/chat/completions"},
		{"Claude Instant", "claude-instant-1.2", "/chat/completions"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccModelResourceConfig(tc.identifier, tc.path),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("tama_model.test", "identifier", tc.identifier),
							resource.TestCheckResourceAttr("tama_model.test", "path", tc.path),
							resource.TestCheckResourceAttrSet("tama_model.test", "id"),
						),
					},
				},
			})
		})
	}
}

func TestAccModelResource_EmbeddingModels(t *testing.T) {
	testCases := []struct {
		name       string
		identifier string
		path       string
	}{
		{"OpenAI Text Embedding Ada", "text-embedding-ada-002", "/v1/embeddings"},
		{"OpenAI Text Embedding 3 Small", "text-embedding-3-small", "/v1/embeddings"},
		{"OpenAI Text Embedding 3 Large", "text-embedding-3-large", "/v1/embeddings"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccModelResourceConfig(tc.identifier, tc.path),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("tama_model.test", "identifier", tc.identifier),
							resource.TestCheckResourceAttr("tama_model.test", "path", tc.path),
							resource.TestCheckResourceAttrSet("tama_model.test", "id"),
						),
					},
				},
			})
		})
	}
}

func TestAccModelResource_InvalidIdentifier(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccModelResourceConfig("", "/chat/completions"),
				ExpectError: regexp.MustCompile("Unable to create model"),
			},
		},
	})
}

func TestAccModelResource_InvalidPath(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccModelResourceConfig("test-model", ""),
				ExpectError: regexp.MustCompile("Unable to create model"),
			},
		},
	})
}

func TestAccModelResource_PathVariations(t *testing.T) {
	testCases := []struct {
		name string
		path string
	}{
		{"Standard OpenAI", "/v1/chat/completions"},
		{"Chat completions", "/chat/completions"},
		{"Chat completions alt", "/chat/completions"},
		{"V1 chat completions", "/v1/chat/completions"},
		{"Basic chat", "/chat/completions"},
		{"Standard chat", "/chat/completions"},
		{"Embeddings API", "/v1/embeddings"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccModelResourceConfig("test-model", tc.path),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("tama_model.test", "identifier", "test-model"),
							resource.TestCheckResourceAttr("tama_model.test", "path", tc.path),
							resource.TestCheckResourceAttrSet("tama_model.test", "id"),
						),
					},
				},
			})
		})
	}
}

func TestAccModelResource_LongIdentifier(t *testing.T) {
	longIdentifier := "this-is-a-very-long-model-identifier-that-might-exceed-database-limits-and-should-be-tested-for-proper-error-handling"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModelResourceConfig(longIdentifier, "/chat/completions"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_model.test", "identifier", longIdentifier),
					resource.TestCheckResourceAttr("tama_model.test", "path", "/chat/completions"),
				),
			},
		},
	})
}

func TestAccModelResource_SpecialCharacters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModelResourceConfig("model-with-special_chars.123", "/chat/completions"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_model.test", "identifier", "model-with-special_chars.123"),
					resource.TestCheckResourceAttr("tama_model.test", "path", "/chat/completions"),
				),
			},
		},
	})
}

func TestAccModelResource_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModelResourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// First model
					resource.TestCheckResourceAttr("tama_model.test1", "identifier", "gpt-3.5-turbo"),
					resource.TestCheckResourceAttr("tama_model.test1", "path", "/v1/chat/completions"),
					resource.TestCheckResourceAttrSet("tama_model.test1", "id"),
					// Second model
					resource.TestCheckResourceAttr("tama_model.test2", "identifier", "text-embedding-ada-002"),
					resource.TestCheckResourceAttr("tama_model.test2", "path", "/v1/embeddings"),
					resource.TestCheckResourceAttrSet("tama_model.test2", "id"),
				),
			},
		},
	})
}

func TestAccModelResource_DifferentSources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModelResourceConfigDifferentSources(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Model from first source
					resource.TestCheckResourceAttr("tama_model.openai", "identifier", "gpt-4"),
					resource.TestCheckResourceAttr("tama_model.openai", "path", "/v1/chat/completions"),
					resource.TestCheckResourceAttrSet("tama_model.openai", "id"),
					// Model from second source
					resource.TestCheckResourceAttr("tama_model.anthropic", "identifier", "claude-3-sonnet"),
					resource.TestCheckResourceAttr("tama_model.anthropic", "path", "/chat/completions"),
					resource.TestCheckResourceAttrSet("tama_model.anthropic", "id"),
				),
			},
		},
	})
}

func TestAccModelResource_DisappearResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModelResourceConfig("disappear-model", "/chat/completions"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tama_model.test", "identifier", "disappear-model"),
					resource.TestCheckResourceAttr("tama_model.test", "path", "/chat/completions"),
					resource.TestCheckResourceAttrSet("tama_model.test", "id"),
					resource.TestCheckResourceAttrSet("tama_model.test", "source_id"),
				),
			},
		},
	})
}

func testAccModelResourceConfig(identifier, path string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-model-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_source" "test_source" {
  space_id = tama_space.test_space.id
  name     = "test-source-for-model"
  type     = "model"
  endpoint = "https://api.example.com"
  api_key  = "test-api-key"
}

resource "tama_model" "test" {
  source_id  = tama_source.test_source.id
  identifier = %[1]q
  path       = %[2]q
}
`, identifier, path)
}

func testAccModelResourceConfigMultiple() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-multiple-models-%d"
  type = "root"
}`, timestamp) + `

resource "tama_source" "test_source" {
  space_id = tama_space.test_space.id
  name     = "test-source-for-multiple-models"
  type     = "model"
  endpoint = "https://api.example.com"
  api_key  = "test-api-key"
}

resource "tama_model" "test1" {
  source_id  = tama_source.test_source.id
  identifier = "gpt-3.5-turbo"
  path       = "/v1/chat/completions"
}

resource "tama_model" "test2" {
  source_id  = tama_source.test_source.id
  identifier = "text-embedding-ada-002"
  path       = "/v1/embeddings"
}
`
}

func testAccModelResourceConfigDifferentSources() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-different-sources-%d"
  type = "root"
}`, timestamp) + `

resource "tama_source" "openai" {
  space_id = tama_space.test_space.id
  name     = "openai-source"
  type     = "model"
  endpoint = "https://api.openai.com"
  api_key  = "openai-api-key"
}

resource "tama_source" "anthropic" {
  space_id = tama_space.test_space.id
  name     = "anthropic-source"
  type     = "model"
  endpoint = "https://api.anthropic.com"
  api_key  = "anthropic-api-key"
}

resource "tama_model" "openai" {
  source_id  = tama_source.openai.id
  identifier = "gpt-4"
  path       = "/v1/chat/completions"
}

resource "tama_model" "anthropic" {
  source_id  = tama_source.anthropic.id
  identifier = "claude-3-sonnet"
  path       = "/chat/completions"
}
`
}
