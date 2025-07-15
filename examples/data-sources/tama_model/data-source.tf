# Example configuration for tama_model data source

terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# Fetch information about an existing model by ID
data "tama_model" "example" {
  id = "model-12345"
}

# Use the data source output to reference model information
locals {
  model_info = {
    id         = data.tama_model.example.id
    identifier = data.tama_model.example.identifier
    path       = data.tama_model.example.path
  }
}

# Example of using model data in other configurations
resource "local_file" "model_config" {
  content = jsonencode({
    model = {
      id         = data.tama_model.example.id
      identifier = data.tama_model.example.identifier
      api_path   = data.tama_model.example.path
    }
  })
  filename = "model-config.json"
}

# Output the model information
output "model_id" {
  description = "ID of the model"
  value       = data.tama_model.example.id
}

output "model_identifier" {
  description = "Identifier of the model"
  value       = data.tama_model.example.identifier
}

output "model_path" {
  description = "API path of the model"
  value       = data.tama_model.example.path
}

output "model_info" {
  description = "Complete model information"
  value       = local.model_info
}
