// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package source_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccSourceDataSource_ById(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create source first, then read it using data source
			{
				Config: testAccSourceDataSourceConfig_ById("test-source", "model", "https://api.example.com", "test-api-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check the resource
					resource.TestCheckResourceAttr("tama_source.test", "name", "test-source"),
					resource.TestCheckResourceAttr("tama_source.test", "type", "model"),
					resource.TestCheckResourceAttr("tama_source.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttrSet("tama_source.test", "id"),
					resource.TestCheckResourceAttrSet("tama_source.test", "space_id"),
					resource.TestCheckResourceAttrSet("tama_source.test", "provision_state"),
					// Check the data source
					resource.TestCheckResourceAttr("data.tama_source.test", "name", "test-source"),
					resource.TestCheckResourceAttr("data.tama_source.test", "type", "model"),
					resource.TestCheckResourceAttr("data.tama_source.test", "endpoint", "https://api.example.com"),
					resource.TestCheckResourceAttrSet("data.tama_source.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_source.test", "slug"),
					resource.TestCheckResourceAttrSet("data.tama_source.test", "space_id"),
					resource.TestCheckResourceAttrSet("data.tama_source.test", "provision_state"),
					// Verify that the IDs match
					resource.TestCheckResourceAttrPair("tama_source.test", "id", "data.tama_source.test", "id"),
				),
			},
		},
	})
}

func TestAccSourceDataSource_BySpecificationAndSlug(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create specification (which creates the source automatically), then read source using data source by specification_id and slug
			{
				Config: testAccSourceDataSourceConfig_BySpecificationAndSlug(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check the specification
					resource.TestCheckResourceAttr("tama_specification.test", "version", "1.0.0"),
					resource.TestCheckResourceAttr("tama_specification.test", "endpoint", "https://elasticsearch.arrakis.upmaru.network"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "id"),
					// Check the data source by specification_id and slug (the source is created by the specification)
					resource.TestCheckResourceAttr("data.tama_source.test", "slug", "elasticsearch-index-creation-and-alias-api"),
					resource.TestCheckResourceAttrSet("data.tama_source.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_source.test", "name"),
					resource.TestCheckResourceAttrSet("data.tama_source.test", "type"),
					resource.TestCheckResourceAttrSet("data.tama_source.test", "endpoint"),
					resource.TestCheckResourceAttrSet("data.tama_source.test", "space_id"),
					resource.TestCheckResourceAttrSet("data.tama_source.test", "provision_state"),
					// Verify specification_id is passed through correctly
					resource.TestCheckResourceAttrPair("tama_specification.test", "id", "data.tama_source.test", "specification_id"),
				),
			},
		},
	})
}

func TestAccSourceDataSource_InvalidConfiguration_NoParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSourceDataSourceConfig_InvalidNoParameters(),
				ExpectError: regexp.MustCompile("Either 'id' or both 'specification_id' and 'slug' must be provided"),
			},
		},
	})
}

func TestAccSourceDataSource_InvalidConfiguration_BothMethods(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSourceDataSourceConfig_InvalidBothMethods("test-source", "model", "https://api.example.com", "test-api-key"),
				ExpectError: regexp.MustCompile("Cannot provide both 'id' and 'specification_id'/'slug' simultaneously"),
			},
		},
	})
}

func TestAccSourceDataSource_InvalidConfiguration_OnlySpecificationId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSourceDataSourceConfig_InvalidOnlySpecificationId("test-source", "model", "https://api.example.com", "test-api-key"),
				ExpectError: regexp.MustCompile("Either 'id' or both 'specification_id' and 'slug' must be provided"),
			},
		},
	})
}

func TestAccSourceDataSource_InvalidConfiguration_OnlySlug(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSourceDataSourceConfig_InvalidOnlySlug("test-source", "model", "https://api.example.com", "test-api-key"),
				ExpectError: regexp.MustCompile("Either 'id' or both 'specification_id' and 'slug' must be provided"),
			},
		},
	})
}

func TestAccSourceDataSource_NonExistentId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSourceDataSourceConfig_NonExistentId(),
				ExpectError: regexp.MustCompile("Unable to read source"),
			},
		},
	})
}

func TestAccSourceDataSource_NonExistentSlug(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSourceDataSourceConfig_NonExistentSlug("https://api.example.com/v1"),
				ExpectError: regexp.MustCompile("Unable to read source"),
			},
		},
	})
}

// Configuration functions

func testAccSourceDataSourceConfig_ById(name, sourceType, endpoint, apiKey string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-source-ds-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_source" "test" {
  space_id = tama_space.test_space.id
  name     = %[1]q
  type     = %[2]q
  endpoint = %[3]q
  api_key  = %[4]q
}

data "tama_source" "test" {
  id = tama_source.test.id
}
`, name, sourceType, endpoint, apiKey)
}

func testAccSourceDataSourceConfig_BySpecificationAndSlug() string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-source-ds-%d"
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

data "tama_source" "test" {
  specification_id = tama_specification.test.id
  slug = "elasticsearch-index-creation-and-alias-api"
}
`, timestamp)
}

func testAccSourceDataSourceConfig_InvalidNoParameters() string {
	return acceptance.ProviderConfig + `
data "tama_source" "test" {
  # No parameters provided - should fail
}
`
}

func testAccSourceDataSourceConfig_InvalidBothMethods(name, sourceType, endpoint, apiKey string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-source-ds-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_specification" "test" {
  space_id = tama_space.test_space.id
  schema = jsonencode({
    openapi = "3.0.3"
    info = {
      title = "Test API"
      version = "1.0.0"
      description = "Test specification"
    }
    servers = [
      {
        url = "https://api.example.com/v1"
      }
    ]
    paths = {
      "/messages" = {
        post = {
          operationId = "createMessage"
          summary = "Create message"
          requestBody = {
            required = true
            content = {
              "application/json" = {
                schema = {
                  type = "object"
                  properties = {
                    message = {
                      type = "string"
                      description = "A message property"
                    }
                  }
                  required = ["message"]
                }
              }
            }
          }
          responses = {
            "200" = {
              description = "Success"
            }
          }
        }
      }
    }
  })
  version  = "1.0.0"
  endpoint = "https://api.example.com/v1"
}

resource "tama_source" "test" {
  space_id = tama_space.test_space.id
  name     = %[1]q
  type     = %[2]q
  endpoint = %[3]q
  api_key  = %[4]q
}

data "tama_source" "test" {
  id               = tama_source.test.id
  specification_id = tama_specification.test.id
  slug            = "tmdb-api"
}
`, name, sourceType, endpoint, apiKey)
}

func testAccSourceDataSourceConfig_InvalidOnlySpecificationId(name, sourceType, endpoint, apiKey string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-source-ds-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_specification" "test" {
  space_id = tama_space.test_space.id
  schema = jsonencode({
    openapi = "3.0.3"
    info = {
      title = "Test API"
      version = "1.0.0"
      description = "Test specification"
    }
    servers = [
      {
        url = "https://api.example.com/v1"
      }
    ]
    paths = {
      "/messages" = {
        post = {
          operationId = "createMessage"
          summary = "Create message"
          requestBody = {
            required = true
            content = {
              "application/json" = {
                schema = {
                  type = "object"
                  properties = {
                    message = {
                      type = "string"
                      description = "A message property"
                    }
                  }
                  required = ["message"]
                }
              }
            }
          }
          responses = {
            "200" = {
              description = "Success"
            }
          }
        }
      }
    }
  })
  version  = "1.0.0"
  endpoint = "https://api.example.com/v1"
}

resource "tama_source" "test" {
  space_id = tama_space.test_space.id
  name     = %[1]q
  type     = %[2]q
  endpoint = %[3]q
  api_key  = %[4]q
}

data "tama_source" "test" {
  specification_id = tama_specification.test.id
  # Missing slug - should fail
}
`, name, sourceType, endpoint, apiKey)
}

func testAccSourceDataSourceConfig_InvalidOnlySlug(name, sourceType, endpoint, apiKey string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-source-ds-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_source" "test" {
  space_id = tama_space.test_space.id
  name     = %[1]q
  type     = %[2]q
  endpoint = %[3]q
  api_key  = %[4]q
}

data "tama_source" "test" {
  # Missing specification_id - should fail
  slug = "tmdb-api"
}
`, name, sourceType, endpoint, apiKey)
}

func testAccSourceDataSourceConfig_NonExistentId() string {
	return acceptance.ProviderConfig + `
data "tama_source" "test" {
  id = "non-existent-id"
}
`
}

func testAccSourceDataSourceConfig_NonExistentSlug(endpoint string) string {
	timestamp := time.Now().UnixNano()
	return acceptance.ProviderConfig + fmt.Sprintf(`
resource "tama_space" "test_space" {
  name = "test-space-for-source-ds-%d"
  type = "root"
}`, timestamp) + fmt.Sprintf(`

resource "tama_specification" "test" {
  space_id = tama_space.test_space.id
  schema = jsonencode({
    openapi = "3.0.3"
    info = {
      title = "Test API"
      version = "1.0.0"
      description = "Test specification"
    }
    servers = [
      {
        url = %[1]q
      }
    ]
    paths = {
      "/messages" = {
        post = {
          operationId = "createMessage"
          summary = "Create message"
          requestBody = {
            required = true
            content = {
              "application/json" = {
                schema = {
                  type = "object"
                  properties = {
                    message = {
                      type = "string"
                      description = "A message property"
                    }
                  }
                  required = ["message"]
                }
              }
            }
          }
          responses = {
            "200" = {
              description = "Success"
            }
          }
        }
      }
    }
  })
  version  = "1.0.0"
  endpoint = %[1]q
}

data "tama_source" "test" {
  specification_id = tama_specification.test.id
  slug            = "non-existent-slug"
}
`, endpoint)
}
