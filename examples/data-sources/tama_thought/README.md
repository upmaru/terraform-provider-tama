# Tama Thought Data Source Example

This example demonstrates how to use the `tama_thought` data source to fetch information about existing perception thoughts in the Tama provider.

## Overview

The thought data source allows you to retrieve information about existing thoughts by their ID. This is useful for referencing thoughts in other resources, analyzing thought configurations, or building dependent infrastructure.

## Usage

```hcl
# Fetch information about an existing thought by ID
data "tama_thought" "example" {
  id = "thought-12345"
}

# Use the thought data to create related resources
resource "tama_thought" "related_thought" {
  chain_id = data.tama_thought.example.chain_id
  relation = "analysis"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "analysis"
    })
  }
}
```

## Data Source Attributes

### Required

- `id` - (String) Thought identifier to fetch.

### Exported

- `id` - (String) Thought identifier.
- `chain_id` - (String) ID of the chain this thought belongs to.
- `output_class_id` - (String) ID of the output class for this thought (if any).
- `relation` - (String) Relation type for the thought.
- `current_state` - (String) Current state of the thought.
- `index` - (Number) Index position of the thought in the chain.
- `module` - (List) Module configuration block containing:
  - `reference` - (String) Module reference.
  - `parameters` - (String) Module parameters as JSON string.

## Example Configurations

### Basic Data Source Usage

```hcl
data "tama_thought" "existing_thought" {
  id = "thought-abc123"
}

output "thought_relation" {
  value = data.tama_thought.existing_thought.relation
}
```

### Creating Related Thoughts

```hcl
# Fetch an existing thought
data "tama_thought" "base_thought" {
  id = "thought-12345"
}

# Create another thought in the same chain
resource "tama_thought" "follow_up" {
  chain_id = data.tama_thought.base_thought.chain_id
  relation = "summary"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "summary"
    })
  }
}

# Create a validation thought using the same output class
resource "tama_thought" "validator" {
  chain_id        = data.tama_thought.base_thought.chain_id
  output_class_id = data.tama_thought.base_thought.output_class_id
  relation        = "validation"

  module {
    reference = "tama/identities/validate"
  }
}
```

### Analyzing Thought Configurations

```hcl
# Fetch multiple thoughts for comparison
data "tama_thought" "thought_1" {
  id = "thought-11111"
}

data "tama_thought" "thought_2" {
  id = "thought-22222"
}

data "tama_thought" "thought_3" {
  id = "thought-33333"
}

# Local values for analysis
locals {
  # Parse parameters from thoughts
  thought_1_params = jsondecode(data.tama_thought.thought_1.module[0].parameters)
  thought_2_params = jsondecode(data.tama_thought.thought_2.module[0].parameters)
  
  # Check if thoughts are in the same chain
  same_chain = data.tama_thought.thought_1.chain_id == data.tama_thought.thought_2.chain_id
  
  # Get unique chain IDs
  chain_ids = toset([
    data.tama_thought.thought_1.chain_id,
    data.tama_thought.thought_2.chain_id,
    data.tama_thought.thought_3.chain_id
  ])
  
  # Group thoughts by module type
  generate_thoughts = [
    for thought in [data.tama_thought.thought_1, data.tama_thought.thought_2, data.tama_thought.thought_3] :
    thought if startswith(thought.module[0].reference, "tama/agentic/generate")
  ]
  
  validate_thoughts = [
    for thought in [data.tama_thought.thought_1, data.tama_thought.thought_2, data.tama_thought.thought_3] :
    thought if startswith(thought.module[0].reference, "tama/identities/validate")
  ]
}
```

### Using Variables and Outputs

```hcl
variable "thought_ids" {
  description = "List of thought IDs to fetch"
  type        = list(string)
  default     = ["thought-1", "thought-2", "thought-3"]
}

# Fetch thoughts using for_each
data "tama_thought" "thoughts" {
  for_each = toset(var.thought_ids)
  id       = each.value
}

# Output comprehensive thought information
output "thoughts_summary" {
  description = "Summary of all fetched thoughts"
  value = {
    for k, thought in data.tama_thought.thoughts : k => {
      id               = thought.id
      relation         = thought.relation
      chain_id         = thought.chain_id
      index           = thought.index
      module_reference = thought.module[0].reference
      has_output_class = thought.output_class_id != null && thought.output_class_id != ""
      current_state   = thought.current_state
    }
  }
}
```

## Common Use Cases

### Chain Analysis

```hcl
# Fetch a thought to analyze its chain
data "tama_thought" "chain_member" {
  id = "thought-12345"
}

# Get other chain information using the chain_id
data "tama_chain" "parent_chain" {
  id = data.tama_thought.chain_member.chain_id
}

output "chain_analysis" {
  value = {
    thought_id    = data.tama_thought.chain_member.id
    thought_index = data.tama_thought.chain_member.index
    chain_name    = data.tama_chain.parent_chain.name
    chain_state   = data.tama_chain.parent_chain.current_state
  }
}
```

### Module Configuration Replication

```hcl
# Fetch a thought with specific module configuration
data "tama_thought" "template_thought" {
  id = "thought-template-123"
}

# Replicate the module configuration in a new thought
resource "tama_thought" "replicated_thought" {
  chain_id = var.target_chain_id
  relation = "replicated_${data.tama_thought.template_thought.relation}"

  module {
    reference  = data.tama_thought.template_thought.module[0].reference
    parameters = data.tama_thought.template_thought.module[0].parameters
  }
}
```

### Conditional Resource Creation

```hcl
data "tama_thought" "conditional_thought" {
  id = "thought-12345"
}

# Create additional resources based on thought configuration
resource "tama_thought" "conditional_validator" {
  count = data.tama_thought.conditional_thought.output_class_id != null ? 1 : 0
  
  chain_id        = data.tama_thought.conditional_thought.chain_id
  output_class_id = data.tama_thought.conditional_thought.output_class_id
  relation        = "validation"

  module {
    reference = "tama/identities/validate"
  }
}
```

## Data Processing Examples

### Parameter Analysis

```hcl
data "tama_thought" "parameterized_thought" {
  id = "thought-with-params"
}

locals {
  # Parse and analyze parameters
  params = jsondecode(data.tama_thought.parameterized_thought.module[0].parameters)
  
  # Extract specific parameter values
  relation_param = lookup(local.params, "relation", "unknown")
  has_custom_params = length(keys(local.params)) > 1
  
  # Generate parameter summary
  param_summary = {
    total_params = length(keys(local.params))
    relation     = local.relation_param
    custom_keys  = [for k in keys(local.params) : k if k != "relation"]
  }
}

output "parameter_analysis" {
  value = local.param_summary
}
```

### Index-Based Operations

```hcl
data "tama_thought" "indexed_thought" {
  id = "thought-12345"
}

# Calculate relative positions
locals {
  is_first_thought = data.tama_thought.indexed_thought.index == 0
  is_early_thought = data.tama_thought.indexed_thought.index < 3
  
  # Create a descriptive label
  position_label = local.is_first_thought ? "initial" : local.is_early_thought ? "early" : "later"
}

output "thought_position" {
  value = {
    index     = data.tama_thought.indexed_thought.index
    position  = local.position_label
    is_first  = local.is_first_thought
  }
}
```

## Running the Examples

1. Set your Tama API credentials:
   ```bash
   export TAMA_API_KEY="your-tama-api-key"
   export TAMA_BASE_URL="https://api.tama.io"  # Optional
   ```

2. Replace example thought IDs with actual IDs from your Tama instance:
   ```hcl
   data "tama_thought" "example" {
     id = "your-actual-thought-id"
   }
   ```

3. Initialize Terraform:
   ```bash
   terraform init
   ```

4. Plan to see what data will be fetched:
   ```bash
   terraform plan
   ```

5. Apply to fetch the data:
   ```bash
   terraform apply
   ```

## Related Resources

- `tama_thought` (resource) - Create and manage thoughts
- `tama_chain` (data source) - Fetch chain information
- `tama_class` (data source) - Fetch output class schemas
- `tama_space` (data source) - Fetch space information

## Notes

- Thought IDs must exist in your Tama instance
- The data source is read-only and doesn't modify existing thoughts
- Module parameters are returned as JSON strings and may need parsing
- Index values reflect the thought's position in the processing chain
- Output class IDs will be null/empty if no output class is assigned
- Current state reflects the thought's processing status

## Best Practices

1. **Error Handling**: Always check if thought IDs exist before referencing them
2. **Parameter Parsing**: Use `jsondecode()` carefully with proper error handling
3. **Variable Usage**: Use variables for thought IDs to make configurations reusable
4. **Output Organization**: Structure outputs to provide meaningful information
5. **Conditional Logic**: Use conditional expressions when thought attributes may be null
6. **Documentation**: Comment your data source usage to explain the purpose