# Example configuration for tama_space resource

terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# Create a root space
resource "tama_space" "example" {
  name = "Production Space"
  type = "root"
}

# Create a component space
resource "tama_space" "component_example" {
  name = "AI Models Component"
  type = "component"
}

# Output the space IDs
output "space_id" {
  description = "ID of the production space"
  value       = tama_space.example.id
}

output "component_space_id" {
  description = "ID of the component space"
  value       = tama_space.component_example.id
}
