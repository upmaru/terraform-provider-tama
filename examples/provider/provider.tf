terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

provider "tama" {
  # Configuration options
  base_url = "https://api.tama.io" # Optional: defaults to https://api.tama.io
  api_key  = var.tama_api_key      # Required: can also be set via TAMA_API_KEY env var
  timeout  = 30                    # Optional: timeout in seconds, defaults to 30
}

# Example variable for API key
variable "tama_api_key" {
  description = "API key for Tama"
  type        = string
  sensitive   = true
}
