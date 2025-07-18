# Example configuration for tama_model resource with parameters

terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# Configure the Tama Provider
provider "tama" {
  api_key  = var.tama_api_key
  base_url = var.tama_base_url
}

# Create a neural space
resource "tama_space" "example" {
  name = "Example AI Space"
  type = "root"
}

# Create a source for AI models
resource "tama_source" "mistral" {
  space_id = tama_space.example.id
  name     = "Mistral AI"
  type     = "model"
  endpoint = "https://api.mistral.ai/v1"
  api_key  = var.mistral_api_key
}

# Create a model with simple parameters
resource "tama_model" "grok_mini" {
  source_id  = tama_source.mistral.id
  identifier = "grok-3-mini"
  path       = "/chat/completions"
  parameters = jsonencode({
    reasoning_effort = "low"
    temperature      = 0.8
  })
}

# Create a model with complex parameters
resource "tama_model" "gpt4_advanced" {
  source_id  = tama_source.mistral.id
  identifier = "gpt-4-turbo"
  path       = "/chat/completions"
  parameters = jsonencode({
    temperature       = 0.7
    max_tokens        = 2000
    top_p             = 0.9
    frequency_penalty = 0.1
    presence_penalty  = 0.1
    stream            = true
    stop              = ["\n", "###", "END"]
    reasoning_effort  = "medium"
    response_format = {
      type = "json_object"
    }
    tools = [
      {
        type = "function"
        function = {
          name        = "get_weather"
          description = "Get current weather information"
          parameters = {
            type = "object"
            properties = {
              location = {
                type        = "string"
                description = "City name"
              }
            }
            required = ["location"]
          }
        }
      }
    ]
  })
}

# Create a model for embeddings with specific parameters
resource "tama_model" "text_embedding" {
  source_id  = tama_source.mistral.id
  identifier = "text-embedding-3-large"
  path       = "/embeddings"
  parameters = jsonencode({
    dimensions      = 1536
    encoding_format = "float"
    batch_size      = 100
    timeout         = 30
  })
}

# Create a model without parameters (optional field)
resource "tama_model" "simple_model" {
  source_id  = tama_source.mistral.id
  identifier = "simple-chat"
  path       = "/chat/completions"
  # parameters is optional and can be omitted
}

# Variables
variable "tama_api_key" {
  description = "API key for Tama"
  type        = string
  sensitive   = true
}

variable "tama_base_url" {
  description = "Base URL for Tama API"
  type        = string
  default     = "https://api.tama.io"
}

variable "mistral_api_key" {
  description = "API key for Mistral AI"
  type        = string
  sensitive   = true
}

# Outputs
output "grok_mini_id" {
  description = "ID of the Grok Mini model"
  value       = tama_model.grok_mini.id
}

output "grok_mini_parameters" {
  description = "Parameters of the Grok Mini model"
  value       = tama_model.grok_mini.parameters
}

output "gpt4_advanced_id" {
  description = "ID of the GPT-4 Advanced model"
  value       = tama_model.gpt4_advanced.id
}

output "embedding_model_id" {
  description = "ID of the text embedding model"
  value       = tama_model.text_embedding.id
}
