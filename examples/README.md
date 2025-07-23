# Tama Provider Examples

This directory contains examples that demonstrate how to use the Tama Terraform provider. The examples are organized by resource type and functionality.

## Structure

The examples are organized according to Terraform documentation conventions:

### Provider Configuration

- **[provider/provider.tf](provider/provider.tf)** - Basic provider configuration example

### Resources

- **[resources/tama_space/](resources/tama_space/)** - Neural Space resource examples
- **[resources/tama_chain/](resources/tama_chain/)** - Perception Chain resource examples
- **[resources/tama_source/](resources/tama_source/)** - Sensory Source resource examples  
- **[resources/tama_model/](resources/tama_model/)** - Sensory Model resource examples
- **[resources/tama_limit/](resources/tama_limit/)** - Sensory Limit resource examples

### Data Sources

- **[data-sources/tama_space/](data-sources/tama_space/)** - Space data source examples
- **[data-sources/tama_chain/](data-sources/tama_chain/)** - Chain data source examples
- **[data-sources/tama_source/](data-sources/tama_source/)** - Source data source examples
- **[data-sources/tama_model/](data-sources/tama_model/)** - Model data source examples
- **[data-sources/tama_limit/](data-sources/tama_limit/)** - Limit data source examples

### Complete Example

- **[complete-example.tf](complete-example.tf)** - Comprehensive example showing all resources working together

## Quick Start

1. **Configure the provider** with your API credentials:

```hcl
provider "tama" {
  base_url = "https://api.tama.io"
  api_key  = var.tama_api_key
}
```

2. **Create a space** (neural resource):

```hcl
resource "tama_space" "example" {
  name = "My AI Space"
  type = "root"
}
```

3. **Create a chain** (perception resource):

```hcl
resource "tama_chain" "example" {
  space_id = tama_space.example.id
  name     = "Identity Validation"
}
```

4. **Create a source** (sensory resource):

```hcl
resource "tama_source" "mistral" {
  space_id = tama_space.example.id
  name     = "Mistral AI"
  type     = "model"
  endpoint = "https://api.mistral.ai/v1"
  api_key  = var.mistral_api_key
}
```

5. **Add models and limits**:

```hcl
resource "tama_model" "mistral_small" {
  source_id  = tama_source.mistral.id
  identifier = "mistral-small-latest"
  path       = "/chat/completions"
}

resource "tama_limit" "rate_limit" {
  source_id   = tama_source.mistral.id
  scale_unit  = "minutes"
  scale_count = 1
  limit       = 100
}
```

## Resource Relationships

The Tama provider resources have the following hierarchy:

```
Space (Neural)
├── Chain (Perception)
└── Source (Sensory)
    ├── Model (Sensory)
    └── Limit (Sensory)
```

- **Spaces** are top-level containers for organizing AI services
- **Chains** represent AI processing pipelines and belong to a space
- **Sources** represent external AI providers/APIs and belong to a space
- **Models** define specific AI models available from a source
- **Limits** define rate limiting rules for a source

## Environment Variables

You can configure the provider using environment variables:

- `TAMA_BASE_URL` - Base URL for the Tama API
- `TAMA_API_KEY` - API key for authentication

## Running Examples

1. Set your API keys:
```bash
export TAMA_API_KEY="your-tama-api-key"
export TF_VAR_mistral_api_key="your-mistral-api-key"
export TF_VAR_openai_api_key="your-openai-api-key"
```

2. Initialize Terraform:
```bash
terraform init
```

3. Plan and apply:
```bash
terraform plan
terraform apply
```

## Example Use Cases

### Basic Setup
Start with the individual resource examples to understand each resource type.

### Multi-Provider Setup
Use the complete example to set up multiple AI providers with proper rate limiting.

### Data Source Integration
Use data source examples to reference existing resources in your configurations.

### Production Deployment
The complete example demonstrates a production-ready setup with:
- Separate spaces for production and testing
- Multiple AI providers (Mistral, OpenAI, Anthropic)
- Comprehensive rate limiting
- Configuration file generation for external systems

## Documentation

For detailed documentation on each resource and data source, refer to the provider documentation or use `terraform plan` to see the available attributes and their descriptions.