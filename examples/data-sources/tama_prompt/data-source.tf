terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

provider "tama" {
  # Configuration will be provided via environment variables:
  # TAMA_BASE_URL and TAMA_API_KEY
}

# Data source to fetch an existing prompt by ID
data "tama_prompt" "existing_system_prompt" {
  id = "prompt-123"
}

# Data source to fetch another existing prompt
data "tama_prompt" "existing_user_prompt" {
  id = "prompt-456"
}

# Data source to fetch a code assistant prompt
data "tama_prompt" "code_assistant" {
  id = "prompt-789"
}

# Outputs to display the fetched prompt information
output "system_prompt_info" {
  description = "Information about the system prompt"
  value = {
    id            = data.tama_prompt.existing_system_prompt.id
    name          = data.tama_prompt.existing_system_prompt.name
    slug          = data.tama_prompt.existing_system_prompt.slug
    role          = data.tama_prompt.existing_system_prompt.role
    current_state = data.tama_prompt.existing_system_prompt.current_state
    space_id      = data.tama_prompt.existing_system_prompt.space_id
  }
}

output "user_prompt_info" {
  description = "Information about the user prompt"
  value = {
    id            = data.tama_prompt.existing_user_prompt.id
    name          = data.tama_prompt.existing_user_prompt.name
    slug          = data.tama_prompt.existing_user_prompt.slug
    role          = data.tama_prompt.existing_user_prompt.role
    current_state = data.tama_prompt.existing_user_prompt.current_state
    space_id      = data.tama_prompt.existing_user_prompt.space_id
  }
}

output "code_assistant_content" {
  description = "Content of the code assistant prompt"
  value       = data.tama_prompt.code_assistant.content
  sensitive   = true
}

output "all_prompts_summary" {
  description = "Summary of all fetched prompts"
  value = {
    system_prompt = {
      id   = data.tama_prompt.existing_system_prompt.id
      name = data.tama_prompt.existing_system_prompt.name
      role = data.tama_prompt.existing_system_prompt.role
    }
    user_prompt = {
      id   = data.tama_prompt.existing_user_prompt.id
      name = data.tama_prompt.existing_user_prompt.name
      role = data.tama_prompt.existing_user_prompt.role
    }
    code_assistant = {
      id   = data.tama_prompt.code_assistant.id
      name = data.tama_prompt.code_assistant.name
      role = data.tama_prompt.code_assistant.role
    }
  }
}
