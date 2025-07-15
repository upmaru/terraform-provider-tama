# Terraform Provider for Tama

A Terraform provider for managing Tama API resources, enabling Infrastructure as Code for AI service provisioning and management.

## Features

This provider supports managing the following Tama resources:

- **Neural Resources:**
  - `tama_space` - Neural spaces for organizing AI services

- **Sensory Resources:**
  - `tama_source` - AI service sources (e.g., OpenAI, Mistral, Anthropic)
  - `tama_model` - AI models available from sources
  - `tama_limit` - Rate limiting configurations for sources

- **Data Sources:**
  - All resources above have corresponding data sources for referencing existing resources

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23 (for development)

## Using the Provider

### Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}
```

### Authentication

Configure the provider with your Tama API credentials:

```hcl
provider "tama" {
  base_url = "https://api.tama.io"  # Optional: defaults to https://api.tama.io
  api_key  = var.tama_api_key       # Required: your Tama API key
  timeout  = 30                     # Optional: request timeout in seconds
}
```

You can also use environment variables:
- `TAMA_BASE_URL` - Base URL for the Tama API
- `TAMA_API_KEY` - Your Tama API key

### Quick Start Example

```hcl
# Create a space
resource "tama_space" "production" {
  name = "Production AI Services"
  type = "root"
}

# Create a source for OpenAI
resource "tama_source" "openai" {
  space_id = tama_space.production.id
  name     = "OpenAI Production"
  type     = "model"
  endpoint = "https://api.openai.com/v1"
  api_key  = var.openai_api_key
}

# Add a GPT-4 model
resource "tama_model" "gpt4" {
  source_id  = tama_source.openai.id
  identifier = "gpt-4"
  path       = "/chat/completions"
}

# Set rate limits
resource "tama_limit" "openai_rate_limit" {
  source_id   = tama_source.openai.id
  scale_unit  = "minutes"
  scale_count = 1
  limit       = 100
}
```

## Resource Documentation

### tama_space

Manages neural spaces for organizing AI services.

```hcl
resource "tama_space" "example" {
  name = "My AI Space"
  type = "root"  # or "component"
}
```

**Arguments:**
- `name` (Required) - Name of the space
- `type` (Required) - Type of space: "root" or "component"

**Attributes:**
- `id` - Unique identifier for the space
- `slug` - URL-friendly identifier

### tama_source

Manages AI service sources within a space.

```hcl
resource "tama_source" "example" {
  space_id = tama_space.example.id
  name     = "OpenAI Source"
  type     = "model"
  endpoint = "https://api.openai.com/v1"
  api_key  = var.api_key
}
```

**Arguments:**
- `space_id` (Required, Forces new resource) - ID of the space
- `name` (Required) - Name of the source
- `type` (Required) - Type of source (e.g., "model")
- `endpoint` (Required) - API endpoint URL
- `api_key` (Required, Sensitive) - API authentication key

**Attributes:**
- `id` - Unique identifier for the source

### tama_model

Manages AI models available from a source.

```hcl
resource "tama_model" "example" {
  source_id  = tama_source.example.id
  identifier = "gpt-4"
  path       = "/chat/completions"
}
```

**Arguments:**
- `source_id` (Required, Forces new resource) - ID of the source
- `identifier` (Required) - Model identifier (e.g., "gpt-4")
- `path` (Required) - API path for the model

**Attributes:**
- `id` - Unique identifier for the model

### tama_limit

Manages rate limits for sources.

```hcl
resource "tama_limit" "example" {
  source_id   = tama_source.example.id
  scale_unit  = "minutes"
  scale_count = 1
  limit       = 100
}
```

**Arguments:**
- `source_id` (Required, Forces new resource) - ID of the source
- `scale_unit` (Required) - Time unit: "seconds", "minutes", or "hours"
- `scale_count` (Required) - Number of time units
- `limit` (Required) - Maximum requests allowed in the time period

**Attributes:**
- `id` - Unique identifier for the limit

## Data Sources

All resources have corresponding data sources:

```hcl
data "tama_space" "example" {
  id = "space-123"
}

data "tama_source" "example" {
  id = "source-456"
}

data "tama_model" "example" {
  id = "model-789"
}

data "tama_limit" "example" {
  id = "limit-101"
}
```

## Examples

Comprehensive examples are available in the [examples/](examples/) directory:

- [Provider Configuration](examples/provider/provider.tf)
- [Resource Examples](examples/resources/)
- [Data Source Examples](examples/data-sources/)
- [Complete Multi-Provider Setup](examples/complete-example.tf)

## Resource Relationships

The Tama provider resources follow this hierarchy:

```
Space (Neural)
└── Source (Sensory)
    ├── Model (Sensory)
    └── Limit (Sensory)
```

- **Spaces** are top-level containers for organizing AI services
- **Sources** represent external AI providers/APIs and belong to a space
- **Models** define specific AI models available from a source
- **Limits** define rate limiting rules for a source

## Import

Resources can be imported using their IDs:

```bash
terraform import tama_space.example space-123
terraform import tama_source.example source-456
terraform import tama_model.example model-789
terraform import tama_limit.example limit-101
```

**Note:** Some attributes (like API keys and configuration parameters) may need to be manually set after import since they're not returned by the API.

## Development

### Building the Provider

```bash
go build
```

### Running Tests

```bash
go test ./...
```

### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Support

- [Issues](https://github.com/upmaru/terraform-provider-tama/issues)
- [Examples](examples/)
- [Tama API Documentation](https://api.tama.io/docs)

## License

This project is licensed under the MPL-2.0 License - see the [LICENSE](LICENSE) file for details.