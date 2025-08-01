# Tama Thought Path Data Source Example

This example demonstrates how to use the `tama_modular_thought_path` data source to fetch information about existing perception paths in the Tama provider.

## Overview

The `tama_modular_thought_path` data source allows you to retrieve information about existing paths that link thoughts to target classes. This is useful for:

- Referencing existing paths in other resources
- Extracting path parameters for analysis or duplication
- Building conditional logic based on path configurations
- Integrating with existing AI processing workflows

## Usage

```hcl
# Fetch information about an existing thought path by ID
data "tama_modular_thought_path" "existing_path" {
  id = "path-12345"
}

# Use the path information to create related resources
resource "tama_modular_thought_path" "similar_path" {
  thought_id      = data.tama_modular_thought_path.existing_path.thought_id
  target_class_id = tama_class.new_target.id
  parameters      = data.tama_modular_thought_path.existing_path.parameters
}
```

## Data Source Attributes

### Required

- `id` - (String) ID of the thought path to fetch.

### Computed (Read-Only)

- `thought_id` - (String) ID of the thought this path belongs to.
- `target_class_id` - (String) ID of the target class for the path relationship.
- `parameters` - (String) Path parameters as JSON string.

## Example Configurations

### Basic Path Lookup

```hcl
data "tama_modular_thought_path" "content_classification_path" {
  id = "path-abc123"
}

output "path_thought_id" {
  value = data.tama_modular_thought_path.content_classification_path.thought_id
}

output "path_target_class" {
  value = data.tama_modular_thought_path.content_classification_path.target_class_id
}
```

### Using Path Data for Conditional Logic

```hcl
# Fetch an existing path
data "tama_modular_thought_path" "existing_similarity_path" {
  id = var.similarity_path_id
}

# Parse the parameters to make decisions
locals {
  path_params = jsondecode(data.tama_modular_thought_path.existing_similarity_path.parameters)
  is_high_threshold = try(local.path_params.similarity.threshold, 0) > 0.8
  has_filters = can(local.path_params.filters)
}

# Create a new path based on the existing one's configuration
resource "tama_modular_thought_path" "optimized_path" {
  thought_id      = data.tama_modular_thought_path.existing_similarity_path.thought_id
  target_class_id = data.tama_modular_thought_path.existing_similarity_path.target_class_id
  
  parameters = jsonencode({
    relation = "similarity"
    similarity = {
      # Increase threshold if the existing one is high
      threshold = local.is_high_threshold ? 0.9 : 0.7
      algorithm = "semantic"
    }
    max_results = local.is_high_threshold ? 5 : 15
    # Preserve existing filters if they exist
    filters = local.has_filters ? local.path_params.filters : {}
  })
}
```

### Fetching Multiple Paths for Analysis

```hcl
# Variables for path IDs
variable "classification_path_id" {
  description = "ID of the classification path"
  type        = string
}

variable "similarity_path_id" {
  description = "ID of the similarity path"
  type        = string
}

variable "extraction_path_id" {
  description = "ID of the extraction path"
  type        = string
}

# Fetch multiple paths
data "tama_modular_thought_path" "classification_path" {
  id = var.classification_path_id
}

data "tama_modular_thought_path" "similarity_path" {
  id = var.similarity_path_id
}

data "tama_modular_thought_path" "extraction_path" {
  id = var.extraction_path_id
}

# Analyze the paths
locals {
  classification_params = jsondecode(data.tama_modular_thought_path.classification_path.parameters)
  similarity_params     = jsondecode(data.tama_modular_thought_path.similarity_path.parameters)
  extraction_params     = jsondecode(data.tama_modular_thought_path.extraction_path.parameters)
  
  # Check if paths share the same thought
  same_thought = (
    data.tama_modular_thought_path.classification_path.thought_id == 
    data.tama_modular_thought_path.similarity_path.thought_id
  )
  
  # Get unique target classes
  target_classes = toset([
    data.tama_modular_thought_path.classification_path.target_class_id,
    data.tama_modular_thought_path.similarity_path.target_class_id,
    data.tama_modular_thought_path.extraction_path.target_class_id
  ])
  
  # Extract thresholds for comparison
  classification_confidence = try(local.classification_params.confidence_threshold, 0.5)
  similarity_threshold      = try(local.similarity_params.similarity.threshold, 0.5)
  extraction_confidence     = try(local.extraction_params.confidence_threshold, 0.5)
}
```

### Creating Derived Paths

```hcl
# Fetch a base path to use as a template
data "tama_modular_thought_path" "base_similarity_path" {
  id = "path-template-123"
}

# Create paths with variations of the base configuration
resource "tama_modular_thought_path" "strict_similarity" {
  thought_id      = data.tama_modular_thought_path.base_similarity_path.thought_id
  target_class_id = data.tama_modular_thought_path.base_similarity_path.target_class_id
  
  parameters = jsonencode(merge(
    jsondecode(data.tama_modular_thought_path.base_similarity_path.parameters),
    {
      similarity = {
        threshold = 0.95
        algorithm = "semantic"
      }
      max_results = 3
    }
  ))
}

resource "tama_modular_thought_path" "relaxed_similarity" {
  thought_id      = data.tama_modular_thought_path.base_similarity_path.thought_id
  target_class_id = data.tama_modular_thought_path.base_similarity_path.target_class_id
  
  parameters = jsonencode(merge(
    jsondecode(data.tama_modular_thought_path.base_similarity_path.parameters),
    {
      similarity = {
        threshold = 0.6
        algorithm = "cosine"
      }
      max_results = 25
    }
  ))
}
```

### Path Discovery and Filtering

```hcl
# Use with for_each to process multiple paths
variable "path_ids" {
  description = "List of path IDs to analyze"
  type        = list(string)
  default     = ["path-1", "path-2", "path-3"]
}

data "tama_modular_thought_path" "paths" {
  for_each = toset(var.path_ids)
  id       = each.value
}

# Filter paths by criteria
locals {
  # Find paths with similarity relations
  similarity_paths = {
    for k, path in data.tama_modular_thought_path.paths :
    k => path
    if can(jsondecode(path.parameters).relation) && 
       jsondecode(path.parameters).relation == "similarity"
  }
  
  # Find paths with high thresholds
  high_threshold_paths = {
    for k, path in data.tama_modular_thought_path.paths :
    k => path
    if can(jsondecode(path.parameters).similarity.threshold) && 
       jsondecode(path.parameters).similarity.threshold > 0.8
  }
  
  # Group paths by thought
  paths_by_thought = {
    for k, path in data.tama_modular_thought_path.paths :
    path.thought_id => k...
  }
}
```

## Complete Example with Integration

```hcl
terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# Variables for existing resources
variable "existing_path_id" {
  description = "ID of an existing path to reference"
  type        = string
  default     = "path-example-123"
}

# Fetch existing path information
data "tama_modular_thought_path" "reference_path" {
  id = var.existing_path_id
}

# Get the thought and target class information
data "tama_modular_thought" "source_thought" {
  id = data.tama_modular_thought_path.reference_path.thought_id
}

data "tama_class" "target_class" {
  id = data.tama_modular_thought_path.reference_path.target_class_id
}

# Parse existing parameters
locals {
  existing_params = jsondecode(data.tama_modular_thought_path.reference_path.parameters)
  
  # Extract useful information
  relation_type = try(local.existing_params.relation, "unknown")
  has_similarity = can(local.existing_params.similarity)
  similarity_threshold = try(local.existing_params.similarity.threshold, 0.5)
  max_results = try(local.existing_params.max_results, 10)
  
  # Create enhanced parameters
  enhanced_params = {
    relation = local.relation_type
    similarity = local.has_similarity ? {
      threshold = max(local.similarity_threshold, 0.8)
      algorithm = "semantic"
    } : null
    max_results = min(local.max_results * 2, 50)
    boost_recent = true
    quality_filter = true
  }
}

# Create a new class for enhanced results
resource "tama_class" "enhanced_results" {
  space_id = data.tama_modular_thought.source_thought.chain_id # Assuming chain_id relates to space
  schema_json = jsonencode({
    title = "Enhanced ${data.tama_class.target_class.schema_json.title}"
    type  = "object"
    properties = merge(
      jsondecode(data.tama_class.target_class.schema_json).properties,
      {
        enhancement_score = {
          type        = "number"
          description = "Enhancement quality score"
        }
        processed_at = {
          type        = "string"
          description = "Processing timestamp"
          format      = "date-time"
        }
      }
    )
  })
}

# Create an enhanced path
resource "tama_modular_thought_path" "enhanced_path" {
  thought_id      = data.tama_modular_thought_path.reference_path.thought_id
  target_class_id = tama_class.enhanced_results.id
  
  parameters = jsonencode(local.enhanced_params)
}
```

## Outputs and Analysis

```hcl
# Output original path information
output "original_path_id" {
  description = "ID of the original path"
  value       = data.tama_modular_thought_path.reference_path.id
}

output "original_thought_id" {
  description = "ID of the source thought"
  value       = data.tama_modular_thought_path.reference_path.thought_id
}

output "original_target_class" {
  description = "ID of the original target class"
  value       = data.tama_modular_thought_path.reference_path.target_class_id
}

output "original_parameters" {
  description = "Original path parameters"
  value       = data.tama_modular_thought_path.reference_path.parameters
}

# Output parsed information
output "relation_type" {
  description = "Type of relation used in the path"
  value       = local.relation_type
}

output "similarity_threshold" {
  description = "Similarity threshold (if applicable)"
  value       = local.has_similarity ? local.similarity_threshold : null
}

output "max_results_limit" {
  description = "Maximum results limit"
  value       = local.max_results
}

# Output enhanced path information
output "enhanced_path_id" {
  description = "ID of the enhanced path"
  value       = tama_modular_thought_path.enhanced_path.id
}

output "enhanced_target_class" {
  description = "ID of the enhanced target class"
  value       = tama_class.enhanced_results.id
}

# Conditional outputs
output "is_similarity_based" {
  description = "Whether the path uses similarity relations"
  value       = local.has_similarity
}

output "is_high_threshold" {
  description = "Whether the similarity threshold is considered high"
  value       = local.has_similarity && local.similarity_threshold > 0.8
}

# Comparison outputs
output "threshold_improvement" {
  description = "How much the threshold was improved"
  value       = local.has_similarity ? max(local.similarity_threshold, 0.8) - local.similarity_threshold : 0
}

output "results_multiplier" {
  description = "Multiplier applied to max results"
  value       = min(local.max_results * 2, 50) / local.max_results
}
```

## Running the Example

1. Set your Tama API credentials:
   ```bash
   export TAMA_API_KEY="your-tama-api-key"
   export TAMA_BASE_URL="https://api.tama.io"  # Optional
   ```

2. Set the path ID variable:
   ```bash
   export TF_VAR_existing_path_id="your-path-id"
   ```

3. Initialize and run Terraform:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## Use Cases

- **Path Analysis**: Understand existing path configurations and parameters
- **Configuration Migration**: Copy path configurations between environments
- **Parameter Optimization**: Create improved versions of existing paths
- **Workflow Integration**: Reference existing paths in new AI processing workflows
- **Monitoring Setup**: Extract path information for monitoring and alerting
- **Template Creation**: Use existing paths as templates for new configurations
- **Relationship Mapping**: Understand how thoughts connect to target classes

## Related Resources

- `tama_modular_thought_path` (resource) - Create new thought paths
- `tama_modular_thought` (data source) - Get information about the source thought
- `tama_class` (data source) - Get information about the target class
- `tama_chain` (data source) - Get information about the parent chain
- `tama_space` (data source) - Get information about the parent space

## Notes

- The data source provides read-only access to existing paths
- Parameters are returned as JSON strings and need to be parsed for analysis
- All computed attributes reflect the current state of the path
- Use `jsondecode()` and `jsonencode()` functions for parameter manipulation
- Consider using `try()` function when accessing optional parameter fields

## Best Practices

1. **Error Handling**: Use `try()` and `can()` functions when accessing optional parameters
2. **Parameter Validation**: Validate parameter structure before using in new resources
3. **State Management**: Consider the implications of referencing external resources
4. **Documentation**: Document the purpose and expected structure of referenced paths
5. **Version Control**: Keep track of path IDs and their purposes in your configuration
6. **Testing**: Test parameter parsing logic with various path configurations
7. **Monitoring**: Monitor referenced paths for changes that might affect dependent resources
