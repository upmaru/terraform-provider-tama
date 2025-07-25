---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tama_prompt Data Source - tama"
subcategory: ""
description: |-
  Fetches information about a Tama Memory Prompt
---

# tama_prompt (Data Source)

Fetches information about a Tama Memory Prompt

## Example Usage

```terraform
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
    id              = data.tama_prompt.existing_system_prompt.id
    name            = data.tama_prompt.existing_system_prompt.name
    slug            = data.tama_prompt.existing_system_prompt.slug
    role            = data.tama_prompt.existing_system_prompt.role
    provision_state = data.tama_prompt.existing_system_prompt.provision_state
    space_id        = data.tama_prompt.existing_system_prompt.space_id
  }
}

output "user_prompt_info" {
  description = "Information about the user prompt"
  value = {
    id              = data.tama_prompt.existing_user_prompt.id
    name            = data.tama_prompt.existing_user_prompt.name
    slug            = data.tama_prompt.existing_user_prompt.slug
    role            = data.tama_prompt.existing_user_prompt.role
    provision_state = data.tama_prompt.existing_user_prompt.provision_state
    space_id        = data.tama_prompt.existing_user_prompt.space_id
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) Prompt identifier

### Read-Only

- `content` (String) Content of the prompt
- `name` (String) Name of the prompt
- `provision_state` (String) Current state of the prompt
- `role` (String) Role associated with the prompt (system or user)
- `slug` (String) Slug for the prompt
- `space_id` (String) ID of the space this prompt belongs to
