# Example configuration for tama_class data source

terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# Fetch information about an existing class by ID
data "tama_class" "example" {
  id = "class-12345"
}

# Alternative: Fetch class using space_id and name (new in v0.2.16)
data "tama_class" "class_by_space_and_name" {
  space_id = tama_space.some_space.id
  name     = "class-proxy"
}

# Alternative: Fetch class using specification_id and name
data "tama_class" "class_by_spec_and_name" {
  specification_id = "spec-67890"
  name             = "user-profile"
}

# Use the data source output in other resources
resource "tama_source" "example" {
  space_id = data.tama_class.example.space_id
  name     = "Source for ${data.tama_class.example.name}"
  type     = "model"
  endpoint = "https://api.example.com/v1"
  api_key  = var.example_api_key
}

# Create a new class in the same space as the referenced class
resource "tama_class" "related" {
  space_id = data.tama_class.example.space_id
  schema_json = jsonencode({
    title       = "Related Class Schema"
    description = "A schema related to ${data.tama_class.example.name}"
    type        = "object"
    properties = {
      reference_id = {
        type        = "string"
        description = "Reference to ${data.tama_class.example.name}"
      }
      data = {
        type = "object"
      }
    }
    required = ["reference_id"]
  })
}

# Variable for API key
variable "example_api_key" {
  description = "API key for the example service"
  type        = string
  sensitive   = true
}

# Output the class information
output "class_id" {
  description = "ID of the class"
  value       = data.tama_class.example.id
}

output "class_name" {
  description = "Name of the class"
  value       = data.tama_class.example.name
}

output "class_description" {
  description = "Description of the class"
  value       = data.tama_class.example.description
}

output "class_schema_json" {
  description = "Schema of the class as JSON string"
  value       = data.tama_class.example.schema_json
}

output "class_schema_title" {
  description = "Title from the schema block"
  value       = length(data.tama_class.example.schema) > 0 ? data.tama_class.example.schema[0].title : null
}

output "class_schema_type" {
  description = "Type from the schema block"
  value       = length(data.tama_class.example.schema) > 0 ? data.tama_class.example.schema[0].type : null
}

output "class_current_state" {
  description = "Current state of the class"
  value       = data.tama_class.example.provision_state
}

output "class_space_id" {
  description = "Space ID that contains the class"
  value       = data.tama_class.example.space_id
}

# Example of using multiple class data sources
data "tama_class" "user_class" {
  id = var.user_class_id
}

data "tama_class" "product_class" {
  id = var.product_class_id
}

# Variables for class IDs
variable "user_class_id" {
  description = "ID of the user class"
  type        = string
}

variable "product_class_id" {
  description = "ID of the product class"
  type        = string
}

# Create a relationship class using both referenced classes
resource "tama_class" "user_product_relationship" {
  space_id = data.tama_class.user_class.space_id
  schema = jsonencode({
    type = "object"
    properties = {
      user_id = {
        type        = "string"
        description = "Reference to user from ${data.tama_class.user_class.name}"
      }
      product_id = {
        type        = "string"
        description = "Reference to product from ${data.tama_class.product_class.name}"
      }
      relationship_type = {
        type = "string"
        enum = ["owner", "viewer", "editor"]
      }
      created_at = {
        type   = "string"
        format = "date-time"
      }
    }
    required = ["user_id", "product_id", "relationship_type"]
  })
}

# Output comprehensive class information
output "class_comparison" {
  description = "Comparison of multiple classes"
  value = {
    user_class = {
      id              = data.tama_class.user_class.id
      name            = data.tama_class.user_class.name
      space_id        = data.tama_class.user_class.space_id
      provision_state = data.tama_class.user_class.provision_state
    }
    product_class = {
      id              = data.tama_class.product_class.id
      name            = data.tama_class.product_class.name
      space_id        = data.tama_class.product_class.space_id
      provision_state = data.tama_class.product_class.provision_state
    }
    same_space = data.tama_class.user_class.space_id == data.tama_class.product_class.space_id
  }
}

# Example showing conditional logic based on class schema content
locals {
  user_schema_parsed = jsondecode(data.tama_class.user_class.schema)
  has_email_field    = contains(keys(local.user_schema_parsed.properties), "email")
}

# Conditional resource creation based on schema content
resource "tama_class" "email_validation" {
  count    = local.has_email_field ? 1 : 0
  space_id = data.tama_class.user_class.space_id
  schema = jsonencode({
    type = "object"
    properties = {
      email = {
        type   = "string"
        format = "email"
      }
      validated = {
        type = "boolean"
      }
      validation_date = {
        type   = "string"
        format = "date-time"
      }
    }
    required = ["email", "validated"]
  })
}

output "conditional_creation" {
  description = "Information about conditional resource creation"
  value = {
    user_class_has_email    = local.has_email_field
    email_validator_created = length(tama_class.email_validation) > 0
  }
}
