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

# Example space
resource "tama_space" "example" {
  name = "example-space"
  type = "root"
}

# Example model (assuming you have a source)
resource "tama_model" "example" {
  source_id  = "your-source-id"
  identifier = "mistral-small-latest"
  path       = "/chat/completions"
  parameters = jsonencode({
    temperature = 0.8
    max_tokens  = 1500
  })
}

# Example completion processor
resource "tama_space_processor" "completion" {
  space_id = tama_space.example.id
  model_id = tama_model.example.id

  completion_config {
    temperature = 0.1
    tool_choice = "auto"
    role_mappings = [
      {
        from = "user"
        to   = "human"
      },
      {
        from = "assistant"
        to   = "ai"
      },
      {
        from = "system"
        to   = "system"
      }
    ]
  }
}

# Example embedding processor
resource "tama_space_processor" "embedding" {
  space_id = tama_space.example.id
  model_id = tama_model.example.id

  embedding_config {
    max_tokens = 512
    templates = [
      {
        type    = "query"
        content = "Search Query: {text}"
      },
      {
        type    = "document"
        content = "Document Content: {text}"
      },
      {
        type    = "passage"
        content = "Passage: {text}"
      }
    ]
  }
}

# Example reranking processor
resource "tama_space_processor" "reranking" {
  space_id = tama_space.example.id
  model_id = tama_model.example.id

  reranking_config {
    top_n = 5
  }
}

# Outputs
output "completion_processor_id" {
  value = tama_space_processor.completion.id
}

output "completion_role_mappings" {
  description = "Role mappings configured for completion processor"
  value       = tama_space_processor.completion.completion_config[0].role_mappings
}

output "embedding_processor_id" {
  value = tama_space_processor.embedding.id
}

output "embedding_templates" {
  description = "Templates configured for embedding processor"
  value       = tama_space_processor.embedding.embedding_config[0].templates
}

output "reranking_processor_id" {
  value = tama_space_processor.reranking.id
}
