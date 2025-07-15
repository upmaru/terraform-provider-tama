# Example configuration for tama_space data source

terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# Fetch information about an existing space by ID
data "tama_space" "example" {
  id = "space-12345"
}

# Use the data source output in other resources
resource "tama_source" "example" {
  space_id = data.tama_space.example.id
  name     = "Example Source in ${data.tama_space.example.name}"
  type     = "model"
  endpoint = "https://api.example.com/v1"
  api_key  = var.example_api_key
}

# Variable for API key
variable "example_api_key" {
  description = "API key for the example service"
  type        = string
  sensitive   = true
}

# Output the space information
output "space_name" {
  description = "Name of the space"
  value       = data.tama_space.example.name
}

output "space_type" {
  description = "Type of the space"
  value       = data.tama_space.example.type
}

output "space_id" {
  description = "ID of the space"
  value       = data.tama_space.example.id
}
