# Example configuration for tama_thought data source

terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# Fetch information about an existing thought by ID
data "tama_thought" "example" {
  id = "thought-12345"
}

# Use the thought data source to create another thought in the same chain
resource "tama_thought" "related_thought" {
  chain_id = data.tama_thought.example.chain_id
  relation = "analysis"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "analysis"
    })
  }
}

# Create a thought that references the same output class
resource "tama_thought" "validation_thought" {
  chain_id        = data.tama_thought.example.chain_id
  output_class_id = data.tama_thought.example.output_class_id
  relation        = "validation"

  module {
    reference = "tama/identities/validate"
  }
}

# Fetch multiple thoughts for comparison
data "tama_thought" "first_thought" {
  id = "thought-11111"
}

data "tama_thought" "second_thought" {
  id = "thought-22222"
}

# Variable for thought ID (optional)
variable "thought_id" {
  description = "ID of the thought to fetch"
  type        = string
  default     = "thought-12345"
}

# Alternative example using variable
data "tama_thought" "variable_example" {
  id = var.thought_id
}

# Local values for processing thought data
locals {
  # Extract module parameters for analysis
  thought_parameters = jsondecode(data.tama_thought.example.module[0].parameters)

  # Check if thoughts are in the same chain
  same_chain = data.tama_thought.first_thought.chain_id == data.tama_thought.second_thought.chain_id

  # Get unique chain IDs from multiple thoughts
  chain_ids = toset([
    data.tama_thought.first_thought.chain_id,
    data.tama_thought.second_thought.chain_id,
    data.tama_thought.example.chain_id
  ])
}

# Output the thought information
output "thought_id" {
  description = "ID of the thought"
  value       = data.tama_thought.example.id
}

output "thought_relation" {
  description = "Relation type of the thought"
  value       = data.tama_thought.example.relation
}

output "thought_chain_id" {
  description = "Chain ID that contains the thought"
  value       = data.tama_thought.example.chain_id
}

output "thought_index" {
  description = "Index position of the thought in the chain"
  value       = data.tama_thought.example.index
}

output "thought_current_state" {
  description = "Current state of the thought"
  value       = data.tama_thought.example.current_state
}

output "thought_output_class_id" {
  description = "Output class ID (if any)"
  value       = data.tama_thought.example.output_class_id
}

output "module_reference" {
  description = "Module reference used by the thought"
  value       = data.tama_thought.example.module[0].reference
}

output "module_parameters" {
  description = "Module parameters (raw JSON)"
  value       = data.tama_thought.example.module[0].parameters
}

output "parsed_parameters" {
  description = "Parsed module parameters"
  value       = local.thought_parameters
}

output "thoughts_same_chain" {
  description = "Whether first and second thoughts are in the same chain"
  value       = local.same_chain
}

output "unique_chain_ids" {
  description = "Unique chain IDs from all fetched thoughts"
  value       = local.chain_ids
}

# Conditional outputs based on thought attributes
output "has_output_class" {
  description = "Whether the thought has an output class defined"
  value       = data.tama_thought.example.output_class_id != null && data.tama_thought.example.output_class_id != ""
}

output "module_type" {
  description = "Type of module (extracted from reference)"
  value       = split("/", data.tama_thought.example.module[0].reference)[1]
}

output "module_category" {
  description = "Category of module (extracted from reference)"
  value       = split("/", data.tama_thought.example.module[0].reference)[0]
}
