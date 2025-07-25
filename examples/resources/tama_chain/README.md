# Tama Chain Resource Example

This example demonstrates how to use the `tama_chain` resource to create perception chains in the Tama provider.

## Overview

Perception chains are AI processing pipelines that belong to a space and can be used to define workflows for data processing, content analysis, identity validation, and other AI-driven tasks.

## Usage

```hcl
# Create a space first (required for chain)
resource "tama_space" "example" {
  name = "Perception Space"
  type = "root"
}

# Create a chain for identity validation
resource "tama_chain" "identity_validation" {
  space_id = tama_space.example.id
  name     = "Identity Validation"
}
```

## Resource Attributes

### Required

- `space_id` - (String) ID of the space this chain belongs to. Changing this forces replacement of the resource.
- `name` - (String) Name of the chain.

### Computed

- `id` - (String) Chain identifier.
- `slug` - (String) URL-friendly slug generated from the chain name.
- `provision_state` - (String) Current state of the chain (managed by the API).

## Example Configurations

### Basic Chain

```hcl
resource "tama_chain" "basic" {
  space_id = tama_space.example.id
  name     = "Basic Processing Chain"
}
```

### Multiple Chains in Same Space

```hcl
resource "tama_chain" "identity_validation" {
  space_id = tama_space.example.id
  name     = "Identity Validation"
}

resource "tama_chain" "content_analysis" {
  space_id = tama_space.example.id
  name     = "Content Analysis Pipeline"
}

resource "tama_chain" "data_processing" {
  space_id = tama_space.example.id
  name     = "Data Processing Chain"
}
```

## Outputs

The example includes several outputs to demonstrate accessing chain attributes:

- `identity_validation_chain_id` - ID of the identity validation chain
- `identity_validation_chain_slug` - Slug of the identity validation chain
- `content_analysis_chain_id` - ID of the content analysis chain
- `data_processing_chain_state` - Current state of the data processing chain
- `space_id` - ID of the space containing the chains

## Running the Example

1. Set your Tama API credentials:
   ```bash
   export TAMA_API_KEY="your-tama-api-key"
   export TAMA_BASE_URL="https://api.tama.io"  # Optional, defaults to this URL
   ```

2. Initialize Terraform:
   ```bash
   terraform init
   ```

3. Plan the deployment:
   ```bash
   terraform plan
   ```

4. Apply the configuration:
   ```bash
   terraform apply
   ```

## Use Cases

- **Identity Validation**: Create chains for user identity verification workflows
- **Content Analysis**: Set up pipelines for analyzing and processing content
- **Data Processing**: Define workflows for data transformation and analysis
- **Sentiment Analysis**: Build chains for analyzing text sentiment
- **Document Processing**: Create pipelines for document analysis and extraction

## Related Resources

- `tama_space` - Required parent resource for chains
- `tama_class` - Can be created in the same space for data schema definition
- `tama_source` - AI providers that can be used within the processing chains

## Notes

- Chains require an existing space to be created first
- The `space_id` cannot be changed after creation (forces replacement)
- Chain names should be descriptive of their intended purpose
- The `slug` and `provision_state` are managed by the API and cannot be set directly