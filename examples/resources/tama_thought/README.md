# Tama Thought Resource Example

This example demonstrates how to use the `tama_thought` resource to create perception thoughts in the Tama provider.

## Overview

Perception thoughts are individual processing steps within a chain that define how AI modules should process data. Each thought has a specific relation type, references an AI module, and can optionally include parameters and output schemas.

## Usage

```hcl
# Create a chain first (required for thoughts)
resource "tama_chain" "processing_pipeline" {
  space_id = tama_space.example.id
  name     = "Content Processing Pipeline"
}

# Basic thought with generate module
resource "tama_thought" "content_description" {
  chain_id = tama_chain.processing_pipeline.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}
```

## Resource Attributes

### Required

- `chain_id` - (String) ID of the chain this thought belongs to. Changing this forces replacement of the resource.
- `relation` - (String) Relation type for the thought (e.g., 'description', 'analysis', 'validation', 'summary').
- `module` - (Block) Module configuration for the thought (exactly one required).
  - `reference` - (String) Module reference (e.g., 'tama/agentic/generate', 'tama/identities/validate').
  - `parameters` - (String, Optional) Module parameters as JSON string.

### Optional

- `output_class_id` - (String) ID of the output class for structured output validation.

### Computed

- `id` - (String) Thought identifier.
- `provision_state` - (String) Current state of the thought (managed by the API).
- `index` - (Number) Index position of the thought in the chain.

## Example Configurations

### Basic Thought with Parameters

```hcl
resource "tama_thought" "content_analysis" {
  chain_id = tama_chain.processing_pipeline.id
  relation = "analysis"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "analysis"
    })
  }
}
```

### Thought with Output Class (Structured Output)

```hcl
resource "tama_class" "validation_schema" {
  space_id = tama_space.example.id
  schema_json = jsonencode({
    title = "Validation Output Schema"
    type  = "object"
    properties = {
      valid = {
        type = "boolean"
        description = "Whether the input is valid"
      }
      confidence = {
        type = "number"
        description = "Confidence score (0-1)"
      }
    }
    required = ["valid"]
  })
}

resource "tama_thought" "content_validation" {
  chain_id        = tama_chain.processing_pipeline.id
  output_class_id = tama_class.validation_schema.id
  relation        = "validation"

  module {
    reference = "tama/identities/validate"
  }
}
```

### Thought without Parameters

```hcl
resource "tama_thought" "identity_validation" {
  chain_id = tama_chain.processing_pipeline.id
  relation = "validation"

  module {
    reference = "tama/identities/validate"
  }
}
```

### Multiple Thoughts in Processing Order

```hcl
resource "tama_thought" "step_1_description" {
  chain_id = tama_chain.processing_pipeline.id
  relation = "description"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
}

resource "tama_thought" "step_2_analysis" {
  chain_id = tama_chain.processing_pipeline.id
  relation = "analysis"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "analysis"
    })
  }
}

resource "tama_thought" "step_3_validation" {
  chain_id        = tama_chain.processing_pipeline.id
  output_class_id = tama_class.validation_schema.id
  relation        = "validation"

  module {
    reference = "tama/identities/validate"
  }
}

resource "tama_thought" "step_4_summary" {
  chain_id = tama_chain.processing_pipeline.id
  relation = "summary"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "summary"
    })
  }
}
```

## Available Modules

### tama/agentic/generate
- **Purpose**: General AI content generation
- **Required Parameters**: `relation` (string)
- **Common Relations**: description, analysis, summary, classification
- **Example**:
  ```hcl
  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "description"
    })
  }
  ```

### tama/identities/validate
- **Purpose**: Identity and content validation
- **Parameters**: None required (module handles validation logic internally)
- **Common Relations**: validation, verification
- **Often used with**: Output classes for structured validation results
- **Example**:
  ```hcl
  module {
    reference = "tama/identities/validate"
  }
  ```

## Common Relation Types

- **description** - Generate descriptive content about input
- **analysis** - Perform analytical processing on input
- **validation** - Validate input against rules or schemas
- **summary** - Create summaries of input content
- **classification** - Classify or categorize input
- **extraction** - Extract specific information from input
- **transformation** - Transform input from one format to another

## Outputs

The example includes several outputs to demonstrate accessing thought attributes:

- `description_thought_id` - ID of the description thought
- `description_thought_index` - Index position of the description thought in the chain
- `analysis_thought_id` - ID of the analysis thought
- `validation_thought_id` - ID of the validation thought
- `validation_thought_state` - Current state of the validation thought
- `summary_thought_id` - ID of the summary thought
- `chain_id` - ID of the processing chain containing the thoughts
- `validation_class_id` - ID of the validation output class

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

- **Content Processing Pipelines**: Create multi-step workflows for processing content
- **Identity Verification**: Build validation workflows for user identity verification
- **Document Analysis**: Set up thoughts for analyzing and extracting information from documents
- **Sentiment Analysis**: Create thoughts for analyzing text sentiment and emotions
- **Data Transformation**: Define thoughts for transforming data between formats
- **Quality Assurance**: Build validation thoughts for content quality checking
- **Classification Workflows**: Create thoughts for categorizing and tagging content

## Related Resources

- `tama_space` - Required parent resource for chains and classes
- `tama_chain` - Required parent resource for thoughts
- `tama_class` - Optional output schema definition for structured results
- `tama_source` - AI providers used by the modules

## Notes

- Thoughts require an existing chain to be created first
- The `chain_id` cannot be changed after creation (forces replacement)
- Thoughts are executed in the order of their index within the chain
- The `index` is automatically assigned by the API based on creation order
- Module parameters must be valid JSON when provided
- Output classes provide structured validation for thought results
- Some modules (like `tama/identities/validate`) work without explicit parameters
- The `provision_state` reflects the thought's processing status and is managed by the API

## Best Practices

1. **Descriptive Relations**: Use clear, descriptive relation names that indicate the thought's purpose
2. **Parameter Validation**: Ensure module parameters are properly formatted JSON
3. **Output Classes**: Use output classes for thoughts that need structured, validated outputs
4. **Chain Organization**: Organize thoughts in logical processing order within chains
5. **Module Selection**: Choose appropriate modules based on the processing requirements
6. **Error Handling**: Monitor thought states to ensure proper processing flow