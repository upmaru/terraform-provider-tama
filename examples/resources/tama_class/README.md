# Tama Class Resource Examples

This directory contains examples for using the `tama_class` resource, which manages Class schemas in Tama Neural spaces.

## Overview

The `tama_class` resource supports two ways to define schemas:

1. **Schema Block** - Structured configuration with built-in validation
2. **Schema JSON** - Direct JSON string for file-based or variable-driven schemas

## Schema Block Approach

The schema block provides a structured way to define JSON schemas with validation and the optional `strict` attribute.

### Basic Example

```hcl
resource "tama_class" "action_call" {
  space_id = tama_space.example.id

  schema {
    title       = "action-call"
    description = "An action call is a request to execute an action."
    type        = "object"
    required    = ["tool_id", "parameters", "code"]
    strict      = true
    properties  = jsonencode({
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
    })
  }
}
```

### Schema Block Attributes

- `title` (Required): The title of the schema
- `description` (Required): Description of what the schema represents
- `type` (Required): JSON Schema type (usually "object" or "array")
- `properties` (Optional): JSON string defining the schema properties
- `required` (Optional): List of required property names
- `strict` (Optional): Boolean indicating whether strict validation should be applied

### Complex Schema Example

```hcl
resource "tama_class" "entity_network" {
  space_id = tama_space.example.id

  schema {
    title       = "entity-network"
    description = <<-EOT
      A entity network records the connections between entities.

      ## Fields:
      - edges: An array of entity ids that are connected to the entity.
    EOT
    type        = "object"
    required    = ["edges"]
    strict      = false
    properties  = jsonencode({
      edges = {
        type        = "object"
        description = "An array of entity ids that are connected to the entity."
        properties = {
          id = {
            type        = "integer"
            description = "The id of the entity that is connected to the entity."
          }
          level = {
            type        = "integer"
            description = "The level of the entity that is connected to the entity."
          }
          parent_id = {
            type        = ["integer", "null"]
            description = "The id of the parent entity that is connected to the entity."
          }
        }
        required = ["id", "level"]
      }
    })
  }
}
```

## Schema JSON Approach

The schema JSON approach allows you to define schemas using JSON strings, which is useful for loading from files or using with variables.

### From Variable Example

```hcl
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
```

### From File Example

```hcl
resource "tama_class" "user_from_file" {
  space_id    = tama_space.example.id
  schema_json = file("${path.module}/schemas/user-schema.json")
}
```

### Direct JSON Example

```hcl
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
```

## Required Schema Fields

Both approaches require that your JSON schema includes these mandatory fields:

- `title`: A descriptive title for the schema
- `description`: A description of what the schema represents

## Strict Attribute

The `strict` attribute is only available when using the schema block approach. When set to `true`, it enables stricter validation rules for the schema. This is useful for:

- Enforcing additional validation rules
- Preventing unexpected properties
- Ensuring data consistency

```hcl
schema {
  title       = "strict-schema"
  description = "A schema with strict validation enabled"
  type        = "object"
  strict      = true  # Enables strict validation
  # ... other attributes
}
```

## Best Practices

### When to Use Schema Block

- When you want built-in validation
- When you need the `strict` attribute
- For simple to moderately complex schemas
- When you prefer structured configuration

### When to Use Schema JSON

- When loading schemas from external files
- When using complex variable structures
- When you have existing JSON schema files
- For very complex nested schemas

### Example File Structure

```
your-terraform-project/
├── main.tf
├── variables.tf
├── schemas/
│   ├── user-schema.json
│   ├── action-call-schema.json
│   └── collection-schema.json
└── outputs.tf
```

## Error Handling

The resource provides validation for:

- Mutually exclusive usage (cannot specify both `schema` block and `schema_json`)
- Required fields in JSON schemas (`title` and `description`)
- Valid JSON syntax in properties and schema_json
- Proper schema structure

## Outputs

All class resources provide these computed outputs:

- `id`: The unique identifier of the class
- `name`: The computed name from the API
- `description`: The computed description from the API  
- `current_state`: The current state of the class
- `space_id`: The ID of the space containing the class

When using schema blocks, you can also access:
- `schema[0].title`: The schema title
- `schema[0].description`: The schema description
- `schema[0].type`: The schema type
- `schema[0].strict`: The strict validation setting

When using schema_json, you can access:
- `schema_json`: The complete schema as JSON string