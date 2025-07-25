terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

provider "tama" {
  # Configuration will be provided via environment variables
  # TAMA_BASE_URL and TAMA_API_KEY
}

# Data source to fetch an existing space processor
data "tama_space_processor" "example" {
  space_id = "space-123"
  type     = "completion"
}

# Data source to fetch another processor
data "tama_space_processor" "completion_processor" {
  space_id = "space-abc"
  type     = "embedding"
}

# Outputs to display the fetched processor information
output "processor_id" {
  description = "The ID of the processor"
  value       = data.tama_space_processor.example.id
}

output "processor_model_id" {
  description = "The model ID of the processor"
  value       = data.tama_space_processor.example.model_id
}

output "processor_type" {
  description = "The type of the processor"
  value       = data.tama_space_processor.example.type
}

output "processor_current_state" {
  description = "The current state of the processor"
  value       = data.tama_space_processor.example.provision_state
}

output "completion_config" {
  description = "Completion configuration if available"
  value       = length(data.tama_space_processor.completion_processor.completion_config) > 0 ? data.tama_space_processor.completion_processor.completion_config[0] : null
}

output "embedding_config" {
  description = "Embedding configuration if available"
  value       = length(data.tama_space_processor.example.embedding_config) > 0 ? data.tama_space_processor.example.embedding_config[0] : null
}

output "reranking_config" {
  description = "Reranking configuration if available"
  value       = length(data.tama_space_processor.example.reranking_config) > 0 ? data.tama_space_processor.example.reranking_config[0] : null
}
