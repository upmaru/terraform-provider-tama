terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

provider "tama" {
  # Configuration options
}

# Create a space
resource "tama_space" "example" {
  name = "example-space"
  type = "root"
}

# Create a prompt within the space
resource "tama_prompt" "example" {
  space_id = tama_space.example.id
  name     = "example-prompt"
  content  = "You are a helpful assistant. Please analyze the provided context."
  role     = "system"
}

# Create a perception chain
resource "tama_chain" "example" {
  space_id = tama_space.example.id
  name     = "example-chain"
}

# Create a thought within the chain
resource "tama_thought" "example" {
  chain_id = tama_chain.example.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

# Create a thought context
resource "tama_thought_context" "example" {
  thought_id = tama_thought.example.id
  prompt_id  = tama_prompt.example.id
  layer      = 0
}

# Create a second prompt for different context
resource "tama_prompt" "secondary" {
  space_id = tama_space.example.id
  name     = "secondary-prompt"
  content  = "Please provide additional analysis on the topic."
  role     = "user"
}

# You can also create additional contexts with different layers
resource "tama_thought_context" "secondary" {
  thought_id = tama_thought.example.id
  prompt_id  = tama_prompt.secondary.id
  layer      = 1
}

# Output the context information
output "thought_context_id" {
  description = "The ID of the created thought context"
  value       = tama_thought_context.example.id
}

output "thought_context_provision_state" {
  description = "The provision state of the thought context"
  value       = tama_thought_context.example.provision_state
}
