# Tama Chain Data Source Example

This example demonstrates how to use the `tama_chain` data source to fetch information about existing perception chains in the Tama provider.

## Overview

The `tama_chain` data source allows you to retrieve information about an existing chain by its ID. This is useful when you need to reference chains created outside of your current Terraform configuration or when building modular configurations.

## Usage

```hcl
# Fetch information about an existing chain by ID
data "tama_chain" "example" {
  id = "chain-12345"
}

# Use the chain data source to create another chain in the same space
resource "tama_chain" "related_chain" {
  space_id = data.tama_chain.example.space_id
  name     = "Related Chain for ${data.tama_chain.example.name}"
}
```

## Data Source Attributes

### Required

- `id` - (String) Chain identifier to look up.

### Computed

- `space_id` - (String) ID of the space this chain belongs to.
- `name` - (String) Name of the chain.
- `slug` - (String) URL-friendly slug generated from the chain name.
- `current_state` - (String) Current state of the chain.

## Example Configurations

### Basic Data Source Lookup

```hcl
data "tama_chain" "existing" {
  id = "chain-12345"
}

output "chain_name" {
  value = data.tama_chain.existing.name
}
```

### Using Chain Data to Create Related Resources

```hcl
# Fetch existing chain
data "tama_chain" "main_chain" {
  id = var.chain_id
}

# Create a new chain in the same space
resource "tama_chain" "backup_chain" {
  space_id = data.tama_chain.main_chain.space_id
  name     = "Backup for ${data.tama_chain.main_chain.name}"
}

# Create a class in the same space
resource "tama_class" "chain_schema" {
  space_id = data.tama_chain.main_chain.space_id
  schema_json = jsonencode({
    title       = "Chain Processing Schema"
    description = "Schema for processing data in ${data.tama_chain.main_chain.name}"
    type        = "object"
    properties = {
      input = {
        type        = "string"
        description = "Input data to process"
      }
      chain_id = {
        type        = "string"
        description = "Reference to the processing chain"
      }
    }
    required = ["input", "chain_id"]
  })
}
```

### Using Variables for Chain ID

```hcl
variable "chain_id" {
  description = "ID of the chain to fetch"
  type        = string
}

data "tama_chain" "variable_example" {
  id = var.chain_id
}
```

## Outputs

The example includes several outputs to demonstrate accessing chain data source attributes:

- `chain_name` - Name of the chain
- `chain_slug` - Slug of the chain
- `chain_space_id` - Space ID that contains the chain
- `chain_current_state` - Current state of the chain
- `chain_id` - ID of the chain

## Running the Example

1. Set your Tama API credentials:
   ```bash
   export TAMA_API_KEY="your-tama-api-key"
   export TAMA_BASE_URL="https://api.tama.io"  # Optional, defaults to this URL
   ```

2. Set the chain ID variable:
   ```bash
   export TF_VAR_chain_id="your-chain-id"
   ```

3. Initialize Terraform:
   ```bash
   terraform init
   ```

4. Plan the deployment:
   ```bash
   terraform plan
   ```

5. Apply the configuration:
   ```bash
   terraform apply
   ```

## Use Cases

- **Configuration Reference**: Reference chains created in other Terraform configurations
- **Modular Architecture**: Build modular configurations that reference existing chains
- **Cross-Environment Setup**: Reference chains from different environments
- **Resource Dependencies**: Create resources that depend on existing chains
- **Information Gathering**: Extract chain information for use in other systems

## Integration Examples

### With Other Data Sources

```hcl
# Fetch chain and its space
data "tama_chain" "processing_chain" {
  id = var.processing_chain_id
}

data "tama_space" "chain_space" {
  id = data.tama_chain.processing_chain.space_id
}

# Create a source in the same space
resource "tama_source" "chain_source" {
  space_id = data.tama_space.chain_space.id
  name     = "Source for ${data.tama_chain.processing_chain.name}"
  type     = "model"
  endpoint = "https://api.example.com/v1"
  api_key  = var.api_key
}
```

### With Local Values

```hcl
data "tama_chain" "main" {
  id = var.main_chain_id
}

locals {
  chain_info = {
    id    = data.tama_chain.main.id
    name  = data.tama_chain.main.name
    slug  = data.tama_chain.main.slug
    space = data.tama_chain.main.space_id
  }
}
```

## Error Handling

If the chain ID doesn't exist or is inaccessible, Terraform will return an error:

```
Error: Client Error: Unable to read chain, got error: chain not found
```

Make sure the chain ID is correct and your API key has permission to access it.

## Related Resources

- `tama_chain` resource - For creating new chains
- `tama_space` data source - For fetching space information
- `tama_class` resource - For creating schemas in the same space
- `tama_source` resource - For creating AI providers in the same space

## Notes

- The chain must exist before the data source can fetch it
- The API key must have read access to the chain
- All computed attributes are read-only
- Use this data source when you need to reference existing chains