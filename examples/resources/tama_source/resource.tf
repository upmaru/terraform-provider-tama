# Example configuration for tama_source resource

terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# Create a space first (sources belong to spaces)
resource "tama_space" "example" {
  name = "AI Services Space"
  type = "root"
}

# Create a source for Mistral AI
resource "tama_source" "mistral" {
  space_id = tama_space.example.id
  name     = "Mistral AI Source"
  type     = "model"
  endpoint = "https://api.mistral.ai/v1"
  api_key  = var.mistral_api_key
}

# Create a source for OpenAI
resource "tama_source" "openai" {
  space_id = tama_space.example.id
  name     = "OpenAI Source"
  type     = "model"
  endpoint = "https://api.openai.com/v1"
  api_key  = var.openai_api_key
}

# Variables for API keys
variable "mistral_api_key" {
  description = "API key for Mistral AI"
  type        = string
  sensitive   = true
}

variable "openai_api_key" {
  description = "API key for OpenAI"
  type        = string
  sensitive   = true
}

# Output the source IDs
output "mistral_source_id" {
  description = "ID of the Mistral source"
  value       = tama_source.mistral.id
}

output "openai_source_id" {
  description = "ID of the OpenAI source"
  value       = tama_source.openai.id
}
