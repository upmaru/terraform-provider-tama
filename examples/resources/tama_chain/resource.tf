# Example configuration for tama_chain resource

terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# Create a space first (required for chain)
resource "tama_space" "example" {
  name = "Perception Space"
  type = "root"
}

# Create a chain for identity validation
resource "tama_chain" "identity_validation" {
  space_id = tama_space.example.id
  name     = "Identity Validation"
}

# Create a chain for content analysis
resource "tama_chain" "content_analysis" {
  space_id = tama_space.example.id
  name     = "Content Analysis Pipeline"
}

# Create a chain for data processing
resource "tama_chain" "data_processing" {
  space_id = tama_space.example.id
  name     = "Data Processing Chain"
}

# Output the chain information
output "identity_validation_chain_id" {
  description = "ID of the identity validation chain"
  value       = tama_chain.identity_validation.id
}

output "identity_validation_chain_slug" {
  description = "Slug of the identity validation chain"
  value       = tama_chain.identity_validation.slug
}

output "content_analysis_chain_id" {
  description = "ID of the content analysis chain"
  value       = tama_chain.content_analysis.id
}

output "data_processing_chain_state" {
  description = "Current state of the data processing chain"
  value       = tama_chain.data_processing.provision_state
}

output "space_id" {
  description = "ID of the space containing the chains"
  value       = tama_space.example.id
}
