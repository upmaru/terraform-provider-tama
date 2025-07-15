# Complete example showing all Tama resources working together
# This example demonstrates the relationship between spaces, sources, models, and limits

terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# Configure the Tama provider
provider "tama" {
  base_url = var.tama_base_url
  api_key  = var.tama_api_key
  timeout  = 30
}

# Variables
variable "tama_base_url" {
  description = "Base URL for Tama API"
  type        = string
  default     = "https://api.tama.io"
}

variable "tama_api_key" {
  description = "API key for Tama"
  type        = string
  sensitive   = true
}

variable "mistral_api_key" {
  description = "API key for Mistral AI"
  type        = string
  sensitive   = true
}

variable "openai_api_key" {
  description = "API key for OpenAI"
  type        = string
  sensitive   = true
}

variable "anthropic_api_key" {
  description = "API key for Anthropic"
  type        = string
  sensitive   = true
}

# Create the main AI services space
resource "tama_space" "ai_services" {
  name = "AI Services Production"
  type = "root"
}

# Create a component space for testing
resource "tama_space" "ai_testing" {
  name = "AI Testing Environment"
  type = "component"
}

# Create sources for different AI providers in production
resource "tama_source" "mistral_prod" {
  space_id = tama_space.ai_services.id
  name     = "Mistral AI Production"
  type     = "model"
  endpoint = "https://api.mistral.ai/v1"
  api_key  = var.mistral_api_key
}

resource "tama_source" "openai_prod" {
  space_id = tama_space.ai_services.id
  name     = "OpenAI Production"
  type     = "model"
  endpoint = "https://api.openai.com/v1"
  api_key  = var.openai_api_key
}

resource "tama_source" "anthropic_prod" {
  space_id = tama_space.ai_services.id
  name     = "Anthropic Production"
  type     = "model"
  endpoint = "https://api.anthropic.com"
  api_key  = var.anthropic_api_key
}

# Create test sources in the testing space
resource "tama_source" "mistral_test" {
  space_id = tama_space.ai_testing.id
  name     = "Mistral AI Testing"
  type     = "model"
  endpoint = "https://api.mistral.ai/v1"
  api_key  = var.mistral_api_key
}

# Create models for Mistral
resource "tama_model" "mistral_small" {
  source_id  = tama_source.mistral_prod.id
  identifier = "mistral-small-latest"
  path       = "/chat/completions"
}

resource "tama_model" "mistral_medium" {
  source_id  = tama_source.mistral_prod.id
  identifier = "mistral-medium-latest"
  path       = "/chat/completions"
}

resource "tama_model" "mistral_large" {
  source_id  = tama_source.mistral_prod.id
  identifier = "mistral-large-latest"
  path       = "/chat/completions"
}

# Create models for OpenAI
resource "tama_model" "gpt_3_5_turbo" {
  source_id  = tama_source.openai_prod.id
  identifier = "gpt-3.5-turbo"
  path       = "/chat/completions"
}

resource "tama_model" "gpt_4" {
  source_id  = tama_source.openai_prod.id
  identifier = "gpt-4"
  path       = "/chat/completions"
}

resource "tama_model" "gpt_4_turbo" {
  source_id  = tama_source.openai_prod.id
  identifier = "gpt-4-turbo"
  path       = "/chat/completions"
}

# Create models for Anthropic
resource "tama_model" "claude_3_haiku" {
  source_id  = tama_source.anthropic_prod.id
  identifier = "claude-3-haiku-20240307"
  path       = "/v1/messages"
}

resource "tama_model" "claude_3_sonnet" {
  source_id  = tama_source.anthropic_prod.id
  identifier = "claude-3-sonnet-20240229"
  path       = "/v1/messages"
}

resource "tama_model" "claude_3_opus" {
  source_id  = tama_source.anthropic_prod.id
  identifier = "claude-3-opus-20240229"
  path       = "/v1/messages"
}

# Create test model
resource "tama_model" "mistral_small_test" {
  source_id  = tama_source.mistral_test.id
  identifier = "mistral-small-latest"
  path       = "/chat/completions"
}

# Create production limits for Mistral (generous limits)
resource "tama_limit" "mistral_per_second" {
  source_id   = tama_source.mistral_prod.id
  scale_unit  = "seconds"
  scale_count = 1
  limit       = 20
}

resource "tama_limit" "mistral_per_minute" {
  source_id   = tama_source.mistral_prod.id
  scale_unit  = "minutes"
  scale_count = 1
  limit       = 500
}

resource "tama_limit" "mistral_per_hour" {
  source_id   = tama_source.mistral_prod.id
  scale_unit  = "hours"
  scale_count = 1
  limit       = 10000
}

# Create production limits for OpenAI (moderate limits)
resource "tama_limit" "openai_per_minute" {
  source_id   = tama_source.openai_prod.id
  scale_unit  = "minutes"
  scale_count = 1
  limit       = 100
}

resource "tama_limit" "openai_per_hour" {
  source_id   = tama_source.openai_prod.id
  scale_unit  = "hours"
  scale_count = 1
  limit       = 2000
}

resource "tama_limit" "openai_per_day" {
  source_id   = tama_source.openai_prod.id
  scale_unit  = "hours"
  scale_count = 24
  limit       = 20000
}

# Create production limits for Anthropic (conservative limits)
resource "tama_limit" "anthropic_per_minute" {
  source_id   = tama_source.anthropic_prod.id
  scale_unit  = "minutes"
  scale_count = 1
  limit       = 50
}

resource "tama_limit" "anthropic_per_hour" {
  source_id   = tama_source.anthropic_prod.id
  scale_unit  = "hours"
  scale_count = 1
  limit       = 1000
}

# Create test limits (very restrictive)
resource "tama_limit" "mistral_test_per_minute" {
  source_id   = tama_source.mistral_test.id
  scale_unit  = "minutes"
  scale_count = 1
  limit       = 10
}

resource "tama_limit" "mistral_test_per_hour" {
  source_id   = tama_source.mistral_test.id
  scale_unit  = "hours"
  scale_count = 1
  limit       = 100
}

# Data sources to fetch information about created resources
data "tama_space" "ai_services_data" {
  id = tama_space.ai_services.id
}

data "tama_source" "mistral_data" {
  id = tama_source.mistral_prod.id
}

data "tama_model" "mistral_large_data" {
  id = tama_model.mistral_large.id
}

data "tama_limit" "mistral_per_second_data" {
  id = tama_limit.mistral_per_second.id
}

# Local values for organizing outputs
locals {
  production_sources = {
    mistral   = tama_source.mistral_prod.id
    openai    = tama_source.openai_prod.id
    anthropic = tama_source.anthropic_prod.id
  }

  mistral_models = {
    small  = tama_model.mistral_small.id
    medium = tama_model.mistral_medium.id
    large  = tama_model.mistral_large.id
  }

  openai_models = {
    gpt_3_5_turbo = tama_model.gpt_3_5_turbo.id
    gpt_4         = tama_model.gpt_4.id
    gpt_4_turbo   = tama_model.gpt_4_turbo.id
  }

  anthropic_models = {
    claude_3_haiku  = tama_model.claude_3_haiku.id
    claude_3_sonnet = tama_model.claude_3_sonnet.id
    claude_3_opus   = tama_model.claude_3_opus.id
  }
}

# Outputs
output "spaces" {
  description = "Created spaces"
  value = {
    production = {
      id   = tama_space.ai_services.id
      name = tama_space.ai_services.name
      type = tama_space.ai_services.type
    }
    testing = {
      id   = tama_space.ai_testing.id
      name = tama_space.ai_testing.name
      type = tama_space.ai_testing.type
    }
  }
}

output "production_sources" {
  description = "Production source IDs"
  value       = local.production_sources
}

output "models" {
  description = "Created models organized by provider"
  value = {
    mistral   = local.mistral_models
    openai    = local.openai_models
    anthropic = local.anthropic_models
  }
}

output "limits_summary" {
  description = "Summary of rate limits by provider"
  value = {
    mistral = {
      per_second = tama_limit.mistral_per_second.limit
      per_minute = tama_limit.mistral_per_minute.limit
      per_hour   = tama_limit.mistral_per_hour.limit
    }
    openai = {
      per_minute = tama_limit.openai_per_minute.limit
      per_hour   = tama_limit.openai_per_hour.limit
      per_day    = tama_limit.openai_per_day.limit
    }
    anthropic = {
      per_minute = tama_limit.anthropic_per_minute.limit
      per_hour   = tama_limit.anthropic_per_hour.limit
    }
  }
}

output "data_source_examples" {
  description = "Examples of data source outputs"
  value = {
    space_name     = data.tama_space.ai_services_data.name
    source_name    = data.tama_source.mistral_data.name
    model_id       = data.tama_model.mistral_large_data.identifier
    limit_per_sec  = data.tama_limit.mistral_per_second_data.limit
  }
}

# Configuration file for external systems
resource "local_file" "ai_config" {
  content = jsonencode({
    spaces = {
      production = tama_space.ai_services.id
      testing    = tama_space.ai_testing.id
    }
    sources = local.production_sources
    models = {
      mistral   = local.mistral_models
      openai    = local.openai_models
      anthropic = local.anthropic_models
    }
    limits = {
      mistral = {
        source_id   = tama_source.mistral_prod.id
        per_second  = tama_limit.mistral_per_second.limit
        per_minute  = tama_limit.mistral_per_minute.limit
        per_hour    = tama_limit.mistral_per_hour.limit
      }
      openai = {
        source_id  = tama_source.openai_prod.id
        per_minute = tama_limit.openai_per_minute.limit
        per_hour   = tama_limit.openai_per_hour.limit
        per_day    = tama_limit.openai_per_day.limit
      }
      anthropic = {
        source_id  = tama_source.anthropic_prod.id
        per_minute = tama_limit.anthropic_per_minute.limit
        per_hour   = tama_limit.anthropic_per_hour.limit
      }
    }
  })
  filename = "${path.module}/tama-config.json"
}
