# Example configuration for tama_limit data source

terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# Fetch information about an existing limit by ID
data "tama_limit" "example" {
  id = "limit-12345"
}

# Use the data source output to reference limit information
locals {
  limit_info = {
    id          = data.tama_limit.example.id
    scale_unit  = data.tama_limit.example.scale_unit
    scale_count = data.tama_limit.example.scale_count
    limit       = data.tama_limit.example.limit
  }

  # Calculate total seconds for the limit period
  period_seconds = (
    data.tama_limit.example.scale_unit == "seconds" ? data.tama_limit.example.scale_count :
    data.tama_limit.example.scale_unit == "minutes" ? data.tama_limit.example.scale_count * 60 :
    data.tama_limit.example.scale_unit == "hours" ? data.tama_limit.example.scale_count * 3600 :
    data.tama_limit.example.scale_count
  )
}

# Example of using limit data in monitoring configuration
resource "local_file" "limit_config" {
  content = jsonencode({
    limit = {
      id              = data.tama_limit.example.id
      scale_unit      = data.tama_limit.example.scale_unit
      scale_count     = data.tama_limit.example.scale_count
      limit_value     = data.tama_limit.example.limit
      period_seconds  = local.period_seconds
      rate_per_second = data.tama_limit.example.limit / local.period_seconds
    }
  })
  filename = "limit-config.json"
}

# Output the limit information
output "limit_id" {
  description = "ID of the limit"
  value       = data.tama_limit.example.id
}

output "limit_scale_unit" {
  description = "Scale unit of the limit"
  value       = data.tama_limit.example.scale_unit
}

output "limit_scale_count" {
  description = "Scale count of the limit"
  value       = data.tama_limit.example.scale_count
}

output "limit_value" {
  description = "Limit value"
  value       = data.tama_limit.example.limit
}

output "limit_info" {
  description = "Complete limit information"
  value       = local.limit_info
}

output "rate_per_second" {
  description = "Calculated rate per second"
  value       = data.tama_limit.example.limit / local.period_seconds
}
