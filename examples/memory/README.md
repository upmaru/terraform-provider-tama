# Memory Prompt Examples

This directory contains examples for using the Tama Memory Prompt resources and data sources.

## Overview

The Memory Prompt resources allow you to manage AI prompts within Tama spaces. Prompts can have different roles (system, user, assistant) and are used to define behavior and templates for AI interactions.

## Resources

### `tama_prompt`

Manages a memory prompt resource with the following attributes:

- `space_id` (Required) - ID of the space this prompt belongs to
- `name` (Required) - Name of the prompt
- `content` (Required) - Content of the prompt
- `role` (Required) - Role associated with the prompt (system, user)
- `id` (Computed) - Prompt identifier
- `slug` (Computed) - Slug for the prompt
- `current_state` (Computed) - Current state of the prompt

### `tama_prompt` Data Source

Fetches information about an existing prompt using its ID.

## Usage

### Basic Example

```hcl
resource "tama_space" "workspace" {
  name = "my-workspace"
  type = "root"
}

resource "tama_prompt" "system_prompt" {
  space_id = tama_space.workspace.id
  name     = "helpful-assistant"
  content  = "You are a helpful AI assistant."
  role     = "system"
}

data "tama_prompt" "system_prompt_data" {
  id = tama_prompt.system_prompt.id
}
```

### Advanced Example

See `main.tf` for a comprehensive example that includes:

- Multiple prompts with different roles
- System prompts for different use cases
- Template prompts for users and assistants
- Data sources to fetch prompt information
- Outputs showing prompt details

## Prompt Roles

### System Role
Used to define the AI's behavior, personality, and guidelines. These prompts set the context for how the AI should respond.

```hcl
resource "tama_prompt" "system_assistant" {
  space_id = tama_space.workspace.id
  name     = "general-assistant"
  content  = "You are a helpful AI assistant..."
  role     = "system"
}
```

### User Role
Represents example user inputs or templates for user messages.

```hcl
resource "tama_prompt" "user_template" {
  space_id = tama_space.workspace.id
  name     = "question-template"
  content  = "I need help with: [DESCRIBE_YOUR_PROBLEM]"
  role     = "user"
}
```

## Running the Examples

1. Set your environment variables:
   ```bash
   export TAMA_BASE_URL="https://your-tama-instance.com"
   export TAMA_API_KEY="your-api-key"
   ```

2. Initialize Terraform:
   ```bash
   terraform init
   ```

3. Plan the deployment:
   ```bash
   terraform plan
   ```

4. Apply the configuration:
   ```bash
   terraform apply
   ```

## Best Practices

1. **Organize by Space**: Group related prompts within the same space for better organization.

2. **Descriptive Names**: Use clear, descriptive names for your prompts to make them easy to identify.

3. **Role Consistency**: Use appropriate roles for different types of prompts:
   - `system` for AI behavior definition
   - `user` for example user inputs or templates

4. **Content Structure**: Structure your prompt content clearly with proper formatting and instructions.

5. **Template Usage**: Create reusable templates for common prompt patterns.

## Import Existing Prompts

You can import existing prompts into Terraform management:

```bash
terraform import tama_prompt.example <prompt-id>
```

Note: After importing, you may need to update your configuration to match the existing prompt's attributes.