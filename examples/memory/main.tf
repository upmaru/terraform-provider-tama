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

# Create a space for organizing prompts
resource "tama_space" "ai_workspace" {
  name = "ai-assistant-workspace"
  type = "root"
}

# System prompt for general AI assistant behavior
resource "tama_prompt" "system_assistant" {
  space_id = tama_space.ai_workspace.id
  name     = "general-assistant"
  content  = <<-EOT
    You are a helpful AI assistant. Your primary goals are to:

    1. Provide accurate and helpful information
    2. Be respectful and professional in all interactions
    3. Ask clarifying questions when needed
    4. Admit when you don't know something
    5. Offer to help find solutions or alternatives

    Always maintain a friendly and supportive tone while being concise and informative.
  EOT
  role     = "system"
}

# System prompt for code assistance
resource "tama_prompt" "code_assistant" {
  space_id = tama_space.ai_workspace.id
  name     = "code-assistant"
  content  = <<-EOT
    You are an expert software engineer with deep knowledge across multiple programming languages and frameworks.

    When helping with code:
    - Provide clear, well-commented examples
    - Explain the reasoning behind your solutions
    - Suggest best practices and potential improvements
    - Point out common pitfalls or considerations
    - Offer alternative approaches when applicable

    Focus on writing clean, maintainable, and efficient code.
  EOT
  role     = "system"
}

# User prompt template for asking questions
resource "tama_prompt" "user_question_template" {
  space_id = tama_space.ai_workspace.id
  name     = "question-template"
  content  = "I need help with: [DESCRIBE_YOUR_PROBLEM_HERE]. Please provide a detailed explanation and any relevant examples."
  role     = "user"
}

# User prompt for general help requests
resource "tama_prompt" "user_help_request" {
  space_id = tama_space.ai_workspace.id
  name     = "help-request"
  content  = "Can you help me understand this concept and provide some examples?"
  role     = "user"
}

# Data source to fetch system assistant prompt
data "tama_prompt" "system_assistant_data" {
  id = tama_prompt.system_assistant.id
}

# Data source to fetch code assistant prompt
data "tama_prompt" "code_assistant_data" {
  id = tama_prompt.code_assistant.id
}

# Data source to fetch user question template
data "tama_prompt" "user_template_data" {
  id = tama_prompt.user_question_template.id
}

# Outputs to show prompt information
output "system_assistant_info" {
  description = "Information about the system assistant prompt"
  value = {
    id            = data.tama_prompt.system_assistant_data.id
    name          = data.tama_prompt.system_assistant_data.name
    slug          = data.tama_prompt.system_assistant_data.slug
    role          = data.tama_prompt.system_assistant_data.role
    current_state = data.tama_prompt.system_assistant_data.current_state
  }
}

output "code_assistant_info" {
  description = "Information about the code assistant prompt"
  value = {
    id            = data.tama_prompt.code_assistant_data.id
    name          = data.tama_prompt.code_assistant_data.name
    slug          = data.tama_prompt.code_assistant_data.slug
    role          = data.tama_prompt.code_assistant_data.role
    current_state = data.tama_prompt.code_assistant_data.current_state
  }
}

output "workspace_prompts" {
  description = "Summary of all prompts in the workspace"
  value = {
    space_id   = tama_space.ai_workspace.id
    space_name = tama_space.ai_workspace.name
    prompts = {
      system_assistant = {
        id   = tama_prompt.system_assistant.id
        name = tama_prompt.system_assistant.name
        slug = tama_prompt.system_assistant.slug
        role = tama_prompt.system_assistant.role
      }
      code_assistant = {
        id   = tama_prompt.code_assistant.id
        name = tama_prompt.code_assistant.name
        slug = tama_prompt.code_assistant.slug
        role = tama_prompt.code_assistant.role
      }
      user_template = {
        id   = tama_prompt.user_question_template.id
        name = tama_prompt.user_question_template.name
        slug = tama_prompt.user_question_template.slug
        role = tama_prompt.user_question_template.role
      }
      user_help_request = {
        id   = tama_prompt.user_help_request.id
        name = tama_prompt.user_help_request.name
        slug = tama_prompt.user_help_request.slug
        role = tama_prompt.user_help_request.role
      }
    }
  }
}
