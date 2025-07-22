# Example configuration for tama_class resource

terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# Create a space first (required for class)
resource "tama_space" "example" {
  name = "AI Development Space"
  type = "root"
}

# Create a class with schema block - action call example
resource "tama_class" "action_call" {
  space_id = tama_space.example.id

  schema {
    title       = "action-call"
    description = "An action call is a request to execute an action."
    type        = "object"
    required    = ["tool_id", "parameters", "code", "content_type", "content"]
    strict      = true
    properties = jsonencode({
      tool_id = {
        type        = "string"
        description = "The ID of the tool to execute"
      }
      parameters = {
        type        = "object"
        description = "The parameters to pass to the action"
      }
      code = {
        type        = "integer"
        description = "The status of the action call"
      }
      content_type = {
        type        = "string"
        description = "The content type of the response"
      }
      content = {
        type        = "object"
        description = "The response from the action"
      }
    })
  }
}

# Create a class with schema_json - collection example
resource "tama_class" "collection" {
  space_id = tama_space.example.id
  schema_json = jsonencode({
    title       = "collection"
    description = "A collection is a group of entities that can be queried."
    type        = "object"
    properties = {
      space = {
        type        = "string"
        description = "Slug of the space"
      }
      name = {
        description = "The name of the collection"
        type        = "string"
      }
      created_at = {
        description = "The unix timestamp when the collection was created"
        type        = "integer"
      }
      items = {
        description = "An array of objects"
        items = {
          type = "object"
        }
        type = "array"
      }
    }
    required = ["items", "space", "name", "created_at"]
  })
}

# Create a class with complex nested schema using schema block
resource "tama_class" "entity_network" {
  space_id = tama_space.example.id

  schema {
    title       = "entity-network"
    description = <<-EOT
      A entity network is records the connections between entities.

      ## Fields:
      - edges: An array of entity ids that are connected to the entity.
    EOT
    type        = "object"
    required    = ["edges"]
    strict      = false
    properties = jsonencode({
      edges = {
        type        = "object"
        description = "An array of entity ids that are connected to the entity."
      }
    })
  }
}

# Example with a simpler schema block (no jsonencode needed for simple cases)
resource "tama_class" "simple_message" {
  space_id = tama_space.example.id

  schema {
    title       = "simple-message"
    description = "A simple message schema with basic string properties"
    type        = "object"
    required    = ["message", "sender"]
    strict      = true
    # For simple schemas, you can still use string literals
    properties = "{\"message\":{\"type\":\"string\",\"description\":\"The message content\"},\"sender\":{\"type\":\"string\",\"description\":\"Who sent the message\"}}"
  }
}

# Using variables for schema definition with schema_json
variable "action_call_schema" {
  description = "Schema definition for action call class"
  type = object({
    title       = string
    description = string
    type        = string
    properties = object({
      tool_id = object({
        description = string
        type        = string
      })
      parameters = object({
        description = string
        type        = string
      })
      code = object({
        description = string
        type        = string
      })
    })
    required = list(string)
  })
  default = {
    title       = "action-call"
    description = "An action call is a request to execute an action."
    type        = "object"
    properties = {
      tool_id = {
        description = "The ID of the tool to execute"
        type        = "string"
      }
      parameters = {
        description = "The parameters to pass to the action"
        type        = "object"
      }
      code = {
        description = "The status of the action call"
        type        = "integer"
      }
    }
    required = ["tool_id", "parameters", "code"]
  }
}

resource "tama_class" "action_from_variable" {
  space_id    = tama_space.example.id
  schema_json = jsonencode(var.action_call_schema)
}

# Example using schema from a local JSON file
resource "tama_class" "user_from_file" {
  space_id    = tama_space.example.id
  schema_json = jsonencode(jsondecode(file("${path.module}/schemas/user-schema.json")))
}

# Output the class IDs
output "action_call_class_id" {
  description = "ID of the action call class"
  value       = tama_class.action_call.id
}

output "collection_class_id" {
  description = "ID of the collection class"
  value       = tama_class.collection.id
}

output "entity_network_class_id" {
  description = "ID of the entity network class"
  value       = tama_class.entity_network.id
}

output "action_from_variable_class_id" {
  description = "ID of the action class created from variable"
  value       = tama_class.action_from_variable.id
}

output "user_from_file_class_id" {
  description = "ID of the user class created from JSON file"
  value       = tama_class.user_from_file.id
}

# Output class details
output "class_details" {
  description = "Details of all created classes"
  value = {
    action_call = {
      id            = tama_class.action_call.id
      name          = tama_class.action_call.name
      description   = tama_class.action_call.description
      current_state = tama_class.action_call.current_state
      schema_title  = tama_class.action_call.schema[0].title
    }
    collection = {
      id            = tama_class.collection.id
      name          = tama_class.collection.name
      description   = tama_class.collection.description
      current_state = tama_class.collection.current_state
    }
    entity_network = {
      id            = tama_class.entity_network.id
      name          = tama_class.entity_network.name
      description   = tama_class.entity_network.description
      current_state = tama_class.entity_network.current_state
      schema_title  = tama_class.entity_network.schema[0].title
    }
  }
}
