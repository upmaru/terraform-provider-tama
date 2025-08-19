# Example usage of tama_thought_tool_initializer resource

# Import initializer with custom parameters
resource "tama_thought_tool_initializer" "import_initializer" {
  thought_tool_id = tama_thought_tool.example.id
  reference       = "tama/initializers/import"
  parameters = jsonencode({
    resources = [
      {
        type     = "concept"
        relation = "import-relation"
        scope    = "space"
      }
    ]
  })
}

# Preload initializer with standard structure
resource "tama_thought_tool_initializer" "preload_initializer" {
  thought_tool_id = tama_thought_tool.example.id
  reference       = "tama/initializers/preload"
  index           = 1
  parameters = jsonencode({
    record = {
      rejections = []
    }
    parents = []
    concept = {
      relations  = ["description", "overview"]
      embeddings = "include"
      content = {
        action = "merge"
        merge = {
          name     = "tool-merge"
          location = "root"
        }
      }
    }
    children = []
  })
}

# Simple preload initializer without custom parameters
resource "tama_thought_tool_initializer" "simple_preload" {
  thought_tool_id = tama_thought_tool.example.id
  reference       = "tama/initializers/preload"
}

# Example thought tool (prerequisite)
resource "tama_thought_tool" "example" {
  thought_id = tama_modular_thought.example.id
  action_id  = data.tama_action.example.id
}

# Example modular thought (prerequisite)
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

# Example chain (prerequisite)
resource "tama_chain" "example" {
  space_id = tama_space.example.id
  name     = "Example Chain"
}

# Example space (prerequisite)
resource "tama_space" "example" {
  name = "example-space"
  type = "root"
}

# Example action data source (prerequisite)
data "tama_action" "example" {
  specification_id = tama_specification.example.id
  identifier       = "create-index"
}

# Example specification (prerequisite)
resource "tama_specification" "example" {
  space_id = tama_space.example.id
  version  = "1.0.0"
  endpoint = "https://api.example.com"
  schema = jsonencode({
    openapi = "3.0.0"
    info = {
      title   = "Example API"
      version = "1.0.0"
    }
    paths = {
      "/create-index" = {
        post = {
          operationId = "create-index"
          summary     = "Create an index"
          responses = {
            "200" = {
              description = "Success"
            }
          }
        }
      }
    }
  })
}
