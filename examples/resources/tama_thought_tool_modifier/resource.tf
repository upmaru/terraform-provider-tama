terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

resource "tama_space" "example" {
  name = "trusted-tool-modifier-example"
  type = "root"
}

resource "tama_specification" "search" {
  space_id = tama_space.example.id
  version  = "1.0.0"
  endpoint = "https://search.example.com"
  schema = jsonencode({
    openapi = "3.1.0"
    info = {
      title   = "Scoped Search API"
      version = "1.0.0"
    }
    servers = [
      {
        url = "https://search.example.com"
      }
    ]
    paths = {
      "/search" = {
        post = {
          operationId = "scoped-search"
          requestBody = {
            required = true
            content = {
              "application/json" = {
                schema = {
                  type = "object"
                  properties = {
                    search = {
                      type = "object"
                      properties = {
                        query = {
                          type = "string"
                        }
                        scope = {
                          type = "object"
                          properties = {
                            user_id = {
                              type = "string"
                            }
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
          responses = {
            "200" = {
              description = "Search results"
            }
          }
        }
      }
    }
  })

  wait_for {
    field {
      name = "current_state"
      in   = ["completed"]
    }
  }
}

resource "tama_chain" "example" {
  space_id = tama_space.example.id
  name     = "Scoped Search Chain"
}

resource "tama_modular_thought" "example" {
  chain_id = tama_chain.example.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

data "tama_action" "search" {
  specification_id = tama_specification.search.id
  identifier       = "scoped-search"
}

resource "tama_thought_tool" "search" {
  thought_id = tama_modular_thought.example.id
  action_id  = data.tama_action.search.id
}

# The source selects trusted runtime metadata; actor identifiers are not stored
# as Terraform values. Missing search.scope parents are skipped, while a call
# with no actor metadata fails closed.
resource "tama_thought_tool_modifier" "actor_scope" {
  thought_tool_id   = tama_thought_tool.search.id
  index             = 0
  target            = "/body/search/scope/user_id"
  on_missing_parent = "skip"
  on_missing_source = "error"

  source {
    type = "metadata"
    path = "actor_identifier"
  }
}
