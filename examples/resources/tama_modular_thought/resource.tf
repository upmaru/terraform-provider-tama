# Example configuration for tama_modular_thought resource

terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# Create a space first (required for chain)
resource "tama_space" "example" {
  name = "AI Processing Space"
  type = "root"
}

# Create a chain for the thoughts
resource "tama_chain" "processing_pipeline" {
  space_id = tama_space.example.id
  name     = "Content Processing Pipeline"
}

# Create an output class for structured validation
resource "tama_class" "validation_schema" {
  space_id = tama_space.example.id
  schema_json = jsonencode({
    title       = "Validation Output Schema"
    description = "Schema for validation results"
    type        = "object"
    properties = {
      valid = {
        type        = "boolean"
        description = "Whether the input is valid"
      }
      confidence = {
        type        = "number"
        description = "Confidence score (0-1)"
        minimum     = 0
        maximum     = 1
      }
      errors = {
        type        = "array"
        description = "List of validation errors"
        items = {
          type = "string"
        }
      }
    }
    required = ["valid"]
  })
}

# Basic thought with generate module and parameters
resource "tama_modular_thought" "content_description" {
  chain_id = tama_chain.processing_pipeline.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

# Thought for content analysis
resource "tama_modular_thought" "content_analysis" {
  chain_id = tama_chain.processing_pipeline.id
  relation = "analysis"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "analysis"
    })
  }
}

# Validation thought with output class and no parameters
resource "tama_modular_thought" "content_validation" {
  chain_id        = tama_chain.processing_pipeline.id
  output_class_id = tama_class.validation_schema.id
  relation        = "validation"

  module {
    reference = "tama/identities/validate"
  }
}

# Summary thought (minimal configuration)
resource "tama_modular_thought" "content_summary" {
  chain_id = tama_chain.processing_pipeline.id
  relation = "summary"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "summary"
    })
  }
}

# Output the thought information
output "description_thought_id" {
  description = "ID of the description thought"
  value       = tama_modular_thought.content_description.id
}

output "description_thought_index" {
  description = "Index position of the description thought in the chain"
  value       = tama_modular_thought.content_description.index
}

output "analysis_thought_id" {
  description = "ID of the analysis thought"
  value       = tama_modular_thought.content_analysis.id
}

output "validation_thought_id" {
  description = "ID of the validation thought"
  value       = tama_modular_thought.content_validation.id
}

output "validation_thought_state" {
  description = "Current state of the validation thought"
  value       = tama_modular_thought.content_validation.provision_state
}

output "summary_thought_id" {
  description = "ID of the summary thought"
  value       = tama_modular_thought.content_summary.id
}

output "chain_id" {
  description = "ID of the processing chain containing the thoughts"
  value       = tama_chain.processing_pipeline.id
}

output "validation_class_id" {
  description = "ID of the validation output class"
  value       = tama_class.validation_schema.id
}
