terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# Example of fetching an existing specification by ID
data "tama_specification" "elasticsearch" {
  id = "spec-12345678-1234-1234-1234-123456789abc"
}

# Output the fetched specification details
output "elasticsearch_specification" {
  description = "Details of the Elasticsearch specification"
  value = {
    id              = data.tama_specification.elasticsearch.id
    space_id        = data.tama_specification.elasticsearch.space_id
    version         = data.tama_specification.elasticsearch.version
    endpoint        = data.tama_specification.elasticsearch.endpoint
    current_state   = data.tama_specification.elasticsearch.current_state
    provision_state = data.tama_specification.elasticsearch.provision_state
  }
}

# Example of using the fetched schema in a local value
locals {
  # Parse the OpenAPI schema from the specification
  elasticsearch_schema = jsondecode(data.tama_specification.elasticsearch.schema)

  # Extract API info
  api_info = local.elasticsearch_schema.info

  # Get available paths
  api_paths = keys(local.elasticsearch_schema.paths)
}

# Output parsed schema information
output "elasticsearch_api_info" {
  description = "API information from the specification"
  value = {
    title       = local.api_info.title
    version     = local.api_info.version
    description = local.api_info.description
  }
}

output "elasticsearch_api_paths" {
  description = "Available API paths"
  value       = local.api_paths
}

# Example of using specification data to configure other resources
# This could be used to dynamically configure API clients or monitoring

# Create a monitoring configuration based on the specification
resource "local_file" "api_monitoring_config" {
  filename = "monitoring-config.json"
  content = jsonencode({
    service_name = local.api_info.title
    version      = local.api_info.version
    endpoint     = data.tama_specification.elasticsearch.endpoint
    health_check = {
      enabled = true
      path    = contains(local.api_paths, "/health") ? "/health" : "/"
    }
    metrics = {
      enabled = true
      paths   = local.api_paths
    }
  })
}

# Example with variable input for specification ID
variable "specification_id" {
  description = "ID of the specification to fetch"
  type        = string
  default     = ""
}

# Conditionally fetch specification if ID is provided
data "tama_specification" "dynamic" {
  count = var.specification_id != "" ? 1 : 0
  id    = var.specification_id
}

# Output for dynamic specification
output "dynamic_specification" {
  description = "Dynamically fetched specification details"
  value = var.specification_id != "" ? {
    id              = data.tama_specification.dynamic[0].id
    version         = data.tama_specification.dynamic[0].version
    endpoint        = data.tama_specification.dynamic[0].endpoint
    current_state   = data.tama_specification.dynamic[0].current_state
    provision_state = data.tama_specification.dynamic[0].provision_state
  } : null
}

# Example of using specification data for validation
locals {
  # Validate that the specification has required OpenAPI structure
  is_valid_openapi = can(jsondecode(data.tama_specification.elasticsearch.schema).openapi)

  # Check if specification is ready for use
  is_ready = (
    data.tama_specification.elasticsearch.provision_state == "active" &&
    data.tama_specification.elasticsearch.current_state == "ready"
  )
}

# Output validation results
output "specification_validation" {
  description = "Validation results for the specification"
  value = {
    is_valid_openapi = local.is_valid_openapi
    is_ready         = local.is_ready
    raw_schema       = data.tama_specification.elasticsearch.schema
  }
}

# Example of using multiple specifications for comparison
data "tama_specification" "api_v1" {
  id = "spec-v1-12345678-1234-1234-1234-123456789abc"
}

data "tama_specification" "api_v2" {
  id = "spec-v2-12345678-1234-1234-1234-123456789abc"
}

# Compare versions
locals {
  v1_info = jsondecode(data.tama_specification.api_v1.schema).info
  v2_info = jsondecode(data.tama_specification.api_v2.schema).info
}

output "api_version_comparison" {
  description = "Comparison of API versions"
  value = {
    v1 = {
      version  = local.v1_info.version
      endpoint = data.tama_specification.api_v1.endpoint
      state    = data.tama_specification.api_v1.provision_state
    }
    v2 = {
      version  = local.v2_info.version
      endpoint = data.tama_specification.api_v2.endpoint
      state    = data.tama_specification.api_v2.provision_state
    }
  }
}
