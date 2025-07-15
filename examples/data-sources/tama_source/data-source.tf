# Example configuration for tama_source data source

terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# Fetch information about an existing source by ID
data "tama_source" "example" {
  id = "source-12345"
}

# Use the data source output in other resources
resource "tama_model" "example" {
  source_id  = data.tama_source.example.id
  identifier = "example-model-v1"
  path       = "/chat/completions"
}

resource "tama_limit" "example" {
  source_id   = data.tama_source.example.id
  scale_unit  = "minutes"
  scale_count = 1
  limit       = 100
}

# Output the source information
output "source_name" {
  description = "Name of the source"
  value       = data.tama_source.example.name
}

output "source_type" {
  description = "Type of the source"
  value       = data.tama_source.example.type
}

output "source_endpoint" {
  description = "Endpoint of the source"
  value       = data.tama_source.example.endpoint
}

output "source_id" {
  description = "ID of the source"
  value       = data.tama_source.example.id
}
