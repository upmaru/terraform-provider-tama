# Example configuration for tama_chain data source

terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# Fetch information about an existing chain by ID
data "tama_chain" "example" {
  id = "chain-12345"
}

# Use the chain data source to create another chain in the same space
resource "tama_chain" "related_chain" {
  space_id = data.tama_chain.example.space_id
  name     = "Related Chain for ${data.tama_chain.example.name}"
}

# Create a class in the same space as the chain
resource "tama_class" "chain_related" {
  space_id = data.tama_chain.example.space_id
  schema_json = jsonencode({
    title       = "Chain Processing Schema"
    description = "Schema for processing data in ${data.tama_chain.example.name}"
    type        = "object"
    properties = {
      input = {
        type        = "string"
        description = "Input data to process"
      }
      chain_id = {
        type        = "string"
        description = "Reference to the processing chain"
      }
    }
    required = ["input", "chain_id"]
  })
}

# Variable for chain ID (optional)
variable "chain_id" {
  description = "ID of the chain to fetch"
  type        = string
  default     = "chain-12345"
}

# Alternative example using variable
data "tama_chain" "variable_example" {
  id = var.chain_id
}

# Output the chain information
output "chain_name" {
  description = "Name of the chain"
  value       = data.tama_chain.example.name
}

output "chain_slug" {
  description = "Slug of the chain"
  value       = data.tama_chain.example.slug
}

output "chain_space_id" {
  description = "Space ID that contains the chain"
  value       = data.tama_chain.example.space_id
}

output "chain_current_state" {
  description = "Current state of the chain"
  value       = data.tama_chain.example.current_state
}

output "chain_id" {
  description = "ID of the chain"
  value       = data.tama_chain.example.id
}
