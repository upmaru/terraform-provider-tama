# Example configuration for tama_limit resource

terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# Create a space first
resource "tama_space" "example" {
  name = "AI Services Space"
  type = "root"
}

# Create a source for Mistral AI
resource "tama_source" "mistral" {
  space_id = tama_space.example.id
  name     = "Mistral AI Source"
  type     = "model"
  endpoint = "https://api.mistral.ai/v1"
  api_key  = var.mistral_api_key
}

# Create limits for the source
resource "tama_limit" "requests_per_second" {
  source_id   = tama_source.mistral.id
  scale_unit  = "seconds"
  scale_count = 1
  limit       = 10
}

resource "tama_limit" "requests_per_minute" {
  source_id   = tama_source.mistral.id
  scale_unit  = "minutes"
  scale_count = 1
  limit       = 100
}

resource "tama_limit" "requests_per_hour" {
  source_id   = tama_source.mistral.id
  scale_unit  = "hours"
  scale_count = 1
  limit       = 1000
}

# Create a source for OpenAI with different limits
resource "tama_source" "openai" {
  space_id = tama_space.example.id
  name     = "OpenAI Source"
  type     = "model"
  endpoint = "https://api.openai.com/v1"
  api_key  = var.openai_api_key
}

# Create more restrictive limits for OpenAI
resource "tama_limit" "openai_requests_per_minute" {
  source_id   = tama_source.openai.id
  scale_unit  = "minutes"
  scale_count = 1
  limit       = 50
}

resource "tama_limit" "openai_requests_per_day" {
  source_id   = tama_source.openai.id
  scale_unit  = "hours"
  scale_count = 24
  limit       = 5000
}

# Variables for API keys
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

# Output the limit IDs
output "mistral_per_second_limit_id" {
  description = "ID of the Mistral per-second limit"
  value       = tama_limit.requests_per_second.id
}

output "mistral_per_minute_limit_id" {
  description = "ID of the Mistral per-minute limit"
  value       = tama_limit.requests_per_minute.id
}

output "mistral_per_hour_limit_id" {
  description = "ID of the Mistral per-hour limit"
  value       = tama_limit.requests_per_hour.id
}

output "openai_per_minute_limit_id" {
  description = "ID of the OpenAI per-minute limit"
  value       = tama_limit.openai_requests_per_minute.id
}

output "openai_per_day_limit_id" {
  description = "ID of the OpenAI per-day limit"
  value       = tama_limit.openai_requests_per_day.id
}
