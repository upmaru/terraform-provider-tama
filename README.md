# Terraform Provider for Tama

A Terraform provider for managing Tama API resources, enabling Infrastructure as Code for AI service provisioning and management.

## Features

This provider supports managing the following Tama resources:

- **Neural Resources:**
  - `tama_space` - Neural spaces for organizing AI services
  - `tama_space_processor` - Processors within neural spaces for AI model operations

- **Sensory Resources:**
  - `tama_source` - AI service sources (e.g., OpenAI, Mistral, Anthropic)
  - `tama_model` - AI models available from sources
  - `tama_limit` - Rate limiting configurations for sources
  - `tama_specification` - OpenAPI 3.0 schema definitions with endpoints and versioning
  - `tama_source_identity` - Source identities with API credentials and validation endpoints

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
  client_id  = var.tama_client_id
  client_secret = var.tama_client_secret
  # Required: your Tama API key
  timeout  = 30                     # Optional: request timeout in seconds
}
```

You can also use environment variables:
- `TAMA_BASE_URL` - Base URL for the Tama API
- `TAMA_CLIENT_ID` - Your Tama Client ID
- `TAMA_CLIENT_SECRET` - Your Tama Client Secret

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

# Create an OpenAPI specification
resource "tama_specification" "elasticsearch" {
  space_id = tama_space.production.id
  version  = "3.1.0"
  endpoint = "https://elasticsearch.arrakis.upmaru.network"
  schema = jsonencode({
    openapi = "3.0.3"
    info = {
      title       = "Elasticsearch API"
      version     = "3.1.0"
      description = "Search API for Elasticsearch"
    }
    paths = {
      "/search" = {
        post = {
          summary = "Execute search query"
          requestBody = {
            required = true
            content = {
              "application/json" = {
                schema = {
                  type = "object"
                  properties = {
                    query = {
                      type        = "string"
                      description = "The search query"
                    }
                    index = {
                      type        = "string"
                      description = "The Elasticsearch index to search"
                    }
                  }
                  required = ["query", "index"]
                }
              }
            }
          }
          responses = {
            "200" = {
              description = "Successful search response"
            }
          }
        }
      }
    }
  })
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

### tama_space_processor

Manages processors within neural spaces for AI model operations. Processors are configured with different types and corresponding configurations based on their intended use.

```hcl
# Completion Processor
resource "tama_space_processor" "completion" {
  space_id = tama_space.example.id
  model_id = tama_model.gpt4.id
  type     = "completion"

  completion_config {
    temperature = 0.7
    tool_choice = "auto"
    role_mappings = jsonencode([
      {
        from = "user"
        to   = "human"
      },
      {
        from = "assistant"
        to   = "ai"
      }
    ])
  }
}

# Embedding Processor
resource "tama_space_processor" "embedding" {
  space_id = tama_space.example.id
  model_id = tama_model.embedding.id
  type     = "embedding"

  embedding_config {
    max_tokens = 512
    templates = jsonencode([
      {
        type    = "query"
        content = "Query: {text}"
      },
      {
        type    = "document"
        content = "Document: {text}"
      }
    ])
  }
}

# Reranking Processor
resource "tama_space_processor" "reranking" {
  space_id = tama_space.example.id
  model_id = tama_model.reranker.id
  type     = "reranking"

  reranking_config {
    top_n = 5
  }
}
```

**Arguments:**
- `space_id` (Required, Forces new resource) - ID of the space this processor belongs to
- `model_id` (Required) - ID of the model this processor uses
- `type` (Required) - Type of processor: "completion", "embedding", or "reranking"

**Configuration Blocks (exactly one required based on type):**

#### `completion_config`
Used when `type = "completion"`.
- `temperature` (Optional) - Sampling temperature (default: 0.8)
- `tool_choice` (Optional) - Tool choice strategy: "required", "auto", or "any" (default: "required")
- `role_mappings` (Optional) - Role mappings as JSON string

#### `embedding_config`
Used when `type = "embedding"`.
- `max_tokens` (Optional) - Maximum number of tokens (default: 512)
- `templates` (Optional) - Templates as JSON string

#### `reranking_config`
Used when `type = "reranking"`.
- `top_n` (Optional) - Number of top results to return (default: 3)

**Attributes:**
- `id` - Unique identifier for the processor

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

### tama_thought_module_input

Manages thought module inputs for connecting thought processes with class corpus data.

```hcl
resource "tama_thought_module_input" "example" {
  thought_id = tama_delegated_thought.example.id
  type              = "concept"
  class_corpus_id   = tama_class_corpus.example.id
}
```

**Arguments:**
- `thought_id` (Required, Forces new resource) - ID of the thought module this input belongs to
- `type` (Required) - Type of input: "concept" or "entity"
- `class_corpus_id` (Required) - Class corpus ID related to thought

**Attributes:**
- `id` - Unique identifier for the input
- `provision_state` - Current state of the input

### tama_specification

Manages OpenAPI 3.0 schema definitions with endpoints and versioning within a space.

```hcl
resource "tama_specification" "elasticsearch" {
  space_id = tama_space.example.id
  version  = "3.1.0"
  endpoint = "https://elasticsearch.arrakis.upmaru.network"
  schema = jsonencode({
    openapi = "3.0.3"
    info = {
      title       = "Elasticsearch API"
      version     = "3.1.0"
      description = "Search API for Elasticsearch"
    }
    paths = {
      "/search" = {
        post = {
          summary = "Execute search query"
          requestBody = {
            required = true
            content = {
              "application/json" = {
                schema = {
                  type = "object"
                  properties = {
                    query = {
                      type        = "string"
                      description = "The search query"
                    }
                    index = {
                      type        = "string"
                      description = "The Elasticsearch index to search"
                    }
                  }
                  required = ["query", "index"]
                }
              }
            }
          }
          responses = {
            "200" = {
              description = "Successful search response"
            }
          }
        }
      }
    }
  })
}
```

**Arguments:**
- `space_id` (Required, Forces new resource) - ID of the space this specification belongs to
- `schema` (Required) - OpenAPI 3.0 schema definition as JSON string
- `version` (Required) - Version of the specification
- `endpoint` (Required) - API endpoint URL for the specification

**Attributes:**
- `id` - Unique identifier for the specification
- `current_state` - Current state of the specification
- `provision_state` - Provision state of the specification

**Features:**
- **JSON Normalization**: The schema field uses JSON normalization to prevent unnecessary updates when the JSON structure is semantically identical but formatted differently
- **OpenAPI 3.0 Support**: Expects valid OpenAPI 3.0 specifications including `openapi` version, `info` object, and `paths` definitions
- **State Management**: Tracks both `current_state` and `provision_state` which are managed server-side

## Data Sources

All resources have corresponding data sources:

```hcl
data "tama_space" "example" {
  id = "space-123"
}

data "tama_space_processor" "example" {
  id = "processor-123"
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

data "tama_specification" "example" {
  id = "spec-123"
}

# Parse the OpenAPI schema from a specification
locals {
  api_schema = jsondecode(data.tama_specification.example.schema)
  api_info   = local.api_schema.info
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
├── Processor (Neural)
├── Specification (Sensory)
└── Source (Sensory)
    ├── Model (Sensory)
    └── Limit (Sensory)
```

- **Spaces** are top-level containers for organizing AI services
- **Processors** define AI processing capabilities within a space using specific models
- **Specifications** define OpenAPI 3.0 schemas with endpoints and versioning within a space
- **Sources** represent external AI providers/APIs and belong to a space
- **Models** define specific AI models available from a source
- **Limits** define rate limiting rules for a source

## Import

Resources can be imported using their IDs:

```bash
terraform import tama_space.example space-123
terraform import tama_space_processor.example space-123/model-456
terraform import tama_specification.example spec-123
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
