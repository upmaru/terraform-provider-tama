terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# First, create a space to contain the specification
resource "tama_space" "example" {
  name = "example-space"
  type = "root"
}

# Create a specification that will contain the source identities
resource "tama_specification" "api_service" {
  space_id = tama_space.example.id
  version  = "1.0.0"
  endpoint = "https://api.example.com"
  schema = jsonencode({
    openapi = "3.0.3"
    info = {
      title       = "Example API"
      version     = "1.0.0"
      description = "Example API for source identity demonstration"
    }
    paths = {
      "/health" = {
        get = {
          summary     = "Health check endpoint"
          description = "Check if the API is healthy"
          responses = {
            "200" = {
              description = "API is healthy"
              content = {
                "application/json" = {
                  schema = {
                    type = "object"
                    properties = {
                      status = { type = "string" }
                    }
                  }
                }
              }
            }
          }
        }
      }
      "/status" = {
        post = {
          summary     = "Status endpoint"
          description = "Get detailed status information"
          responses = {
            "200" = {
              description = "Status information"
            }
            "201" = {
              description = "Status created"
            }
          }
        }
      }
    }
  })
}

# Basic API Key identity with simple health check validation
resource "tama_source_identity" "api_key" {
  specification_id = tama_specification.api_service.id
  identifier       = "ApiKey"
  api_key          = "your-api-key-here"

  validation {
    path   = "/health"
    method = "GET"
    codes  = [200]
  }
}

# Bearer Token identity with multiple acceptable status codes
resource "tama_source_identity" "bearer_token" {
  specification_id = tama_specification.api_service.id
  identifier       = "BearerToken"
  api_key          = "bearer-token-value"

  validation {
    path   = "/status"
    method = "POST"
    codes  = [200, 201]
  }
}

# Custom header identity with complex validation path
resource "tama_source_identity" "custom_header" {
  specification_id = tama_specification.api_service.id
  identifier       = "CustomHeader"
  api_key          = "custom-header-value"

  validation {
    path   = "/api/v1/health?check=all"
    method = "GET"
    codes  = [200, 202, 204]
  }
}

# Example using variables for configuration
variable "api_key" {
  description = "API key for the service"
  type        = string
  sensitive   = true
  default     = "default-api-key"
}

variable "specification_id" {
  description = "ID of the specification"
  type        = string
  default     = ""
}

# Identity using variables (useful for different environments)
resource "tama_source_identity" "variable_example" {
  specification_id = var.specification_id != "" ? var.specification_id : tama_specification.api_service.id
  identifier       = "VariableApiKey"
  api_key          = var.api_key

  validation {
    path   = "/health"
    method = "GET"
    codes  = [200]
  }
}

# Example with different HTTP methods
resource "tama_source_identity" "put_method" {
  specification_id = tama_specification.api_service.id
  identifier       = "PutMethod"
  api_key          = "put-method-key"

  validation {
    path   = "/status"
    method = "PUT"
    codes  = [200, 204]
  }
}

resource "tama_source_identity" "patch_method" {
  specification_id = tama_specification.api_service.id
  identifier       = "PatchMethod"
  api_key          = "patch-method-key"

  validation {
    path   = "/status"
    method = "PATCH"
    codes  = [200, 202]
  }
}

# Example for webhook validation
resource "tama_source_identity" "webhook" {
  specification_id = tama_specification.api_service.id
  identifier       = "WebhookSecret"
  api_key          = "webhook-secret-key"

  validation {
    path   = "/webhook/validate"
    method = "POST"
    codes  = [200, 201, 202]
  }
}

# Outputs to show the created identities
output "api_key_identity_id" {
  description = "ID of the API key identity"
  value       = tama_source_identity.api_key.id
}

output "api_key_identity_state" {
  description = "States of the API key identity"
  value = {
    provision_state = tama_source_identity.api_key.provision_state
    current_state   = tama_source_identity.api_key.current_state
  }
}

output "bearer_token_identity_id" {
  description = "ID of the bearer token identity"
  value       = tama_source_identity.bearer_token.id
}

output "custom_header_identity_id" {
  description = "ID of the custom header identity"
  value       = tama_source_identity.custom_header.id
}

output "all_identity_ids" {
  description = "List of all created identity IDs"
  value = [
    tama_source_identity.api_key.id,
    tama_source_identity.bearer_token.id,
    tama_source_identity.custom_header.id,
    tama_source_identity.variable_example.id,
    tama_source_identity.put_method.id,
    tama_source_identity.patch_method.id,
    tama_source_identity.webhook.id,
  ]
}

# Example showing dependency between specification and identities
output "specification_with_identities" {
  description = "Specification ID with its associated identities"
  value = {
    specification_id = tama_specification.api_service.id
    identities = {
      api_key       = tama_source_identity.api_key.id
      bearer_token  = tama_source_identity.bearer_token.id
      custom_header = tama_source_identity.custom_header.id
      webhook       = tama_source_identity.webhook.id
    }
  }
}
