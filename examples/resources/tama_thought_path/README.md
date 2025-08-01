# Tama Thought Path Resource Example

This example demonstrates how to use the `tama_modular_thought_path` resource to create perception paths in the Tama provider.

## Overview

Perception paths define relationships between thoughts and target classes, enabling AI modules to understand how different thoughts should be connected to specific data structures. Each path specifies a target class and can optionally include parameters to control how the relationship is established.

## Usage

```hcl
# Create a path linking a thought to a target class
resource "tama_modular_thought_path" "content_categorization" {
  thought_id      = tama_modular_thought.content_analysis.id
  target_class_id = tama_class.content_categories.id
  
  parameters = jsonencode({
    relation = "similarity"
    similarity = {
      threshold = 0.8
    }
    max_results = 10
  })
}
```

## Resource Attributes

### Required

- `thought_id` - (String) ID of the thought this path belongs to. Changing this forces replacement of the resource.
- `target_class_id` - (String) ID of the target class for the path relationship.

### Optional

- `parameters` - (String) Path parameters as JSON string. Defines how the relationship should be established.

### Computed

- `id` - (String) Path identifier.

## Example Configurations

### Basic Path without Parameters

```hcl
resource "tama_modular_thought_path" "simple_connection" {
  thought_id      = tama_modular_thought.content_analysis.id
  target_class_id = tama_class.output_schema.id
}
```

### Path with Similarity Parameters

```hcl
resource "tama_modular_thought_path" "similarity_based" {
  thought_id      = tama_modular_thought.content_matching.id
  target_class_id = tama_class.content_categories.id
  
  parameters = jsonencode({
    relation = "similarity"
    similarity = {
      threshold = 0.9
      algorithm = "cosine"
    }
    max_results = 5
  })
}
```

### Path with Complex Filtering

```hcl
resource "tama_modular_thought_path" "filtered_results" {
  thought_id      = tama_modular_thought.content_search.id
  target_class_id = tama_class.search_results.id
  
  parameters = jsonencode({
    relation = "similarity"
    similarity = {
      threshold = 0.7
    }
    filters = {
      category = ["technology", "ai"]
      published_after = "2023-01-01"
    }
    max_results = 20
    sort_by = "relevance"
  })
}
```

### Multiple Paths for Different Relations

```hcl
# Path for similarity-based connections
resource "tama_modular_thought_path" "similarity_path" {
  thought_id      = tama_modular_thought.content_processor.id
  target_class_id = tama_class.similar_content.id
  
  parameters = jsonencode({
    relation = "similarity"
    similarity = {
      threshold = 0.8
    }
  })
}

# Path for classification connections
resource "tama_modular_thought_path" "classification_path" {
  thought_id      = tama_modular_thought.content_processor.id
  target_class_id = tama_class.content_categories.id
  
  parameters = jsonencode({
    relation = "classification"
    confidence_threshold = 0.9
  })
}

# Path for extraction connections
resource "tama_modular_thought_path" "extraction_path" {
  thought_id      = tama_modular_thought.content_processor.id
  target_class_id = tama_class.extracted_entities.id
  
  parameters = jsonencode({
    relation = "extraction"
    extract_types = ["person", "organization", "location"]
  })
}
```

## Parameter Types

### Similarity Parameters

```json
{
  "relation": "similarity",
  "similarity": {
    "threshold": 0.8,
    "algorithm": "cosine"
  },
  "max_results": 10
}
```

### Classification Parameters

```json
{
  "relation": "classification",
  "confidence_threshold": 0.9,
  "categories": ["category1", "category2"],
  "multi_label": true
}
```

### Extraction Parameters

```json
{
  "relation": "extraction",
  "extract_types": ["entity", "keyword", "sentiment"],
  "confidence_threshold": 0.7
}
```

### Filtering Parameters

```json
{
  "relation": "similarity",
  "similarity": {
    "threshold": 0.75
  },
  "filters": {
    "field_name": "field_value",
    "date_range": {
      "start": "2023-01-01",
      "end": "2023-12-31"
    }
  },
  "max_results": 50,
  "sort_by": "score"
}
```

## Complete Example

This example shows a full setup with spaces, chains, classes, thoughts, and paths:

```hcl
# Create a space for the AI processing
resource "tama_space" "ai_workspace" {
  name = "Content Processing Workspace"
  type = "root"
}

# Create a chain for content processing
resource "tama_chain" "content_pipeline" {
  space_id = tama_space.ai_workspace.id
  name     = "Content Analysis Pipeline"
}

# Create target classes for different types of connections
resource "tama_class" "content_categories" {
  space_id = tama_space.ai_workspace.id
  schema_json = jsonencode({
    title = "Content Categories"
    type  = "object"
    properties = {
      category = {
        type        = "string"
        description = "Content category"
      }
      confidence = {
        type        = "number"
        description = "Classification confidence"
      }
    }
  })
}

resource "tama_class" "similar_content" {
  space_id = tama_space.ai_workspace.id
  schema_json = jsonencode({
    title = "Similar Content"
    type  = "object"
    properties = {
      title = {
        type        = "string"
        description = "Content title"
      }
      similarity_score = {
        type        = "number"
        description = "Similarity score"
      }
      url = {
        type        = "string"
        description = "Content URL"
      }
    }
  })
}

# Create thoughts for processing content
resource "tama_modular_thought" "content_analyzer" {
  chain_id = tama_chain.content_pipeline.id
  relation = "analysis"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "analysis"
    })
  }
}

resource "tama_modular_thought" "content_classifier" {
  chain_id = tama_chain.content_pipeline.id
  relation = "classification"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "classification"
    })
  }
}

# Create paths linking thoughts to target classes
resource "tama_modular_thought_path" "analysis_to_categories" {
  thought_id      = tama_modular_thought.content_analyzer.id
  target_class_id = tama_class.content_categories.id
  
  parameters = jsonencode({
    relation = "classification"
    confidence_threshold = 0.8
  })
}

resource "tama_modular_thought_path" "analysis_to_similar" {
  thought_id      = tama_modular_thought.content_analyzer.id
  target_class_id = tama_class.similar_content.id
  
  parameters = jsonencode({
    relation = "similarity"
    similarity = {
      threshold = 0.7
    }
    max_results = 15
  })
}

resource "tama_modular_thought_path" "classifier_to_categories" {
  thought_id      = tama_modular_thought.content_classifier.id
  target_class_id = tama_class.content_categories.id
  
  parameters = jsonencode({
    relation = "classification"
    confidence_threshold = 0.9
  })
}
```

## Outputs

The example includes several outputs to demonstrate accessing path attributes:

- `analysis_categories_path_id` - ID of the analysis to categories path
- `analysis_similar_path_id` - ID of the analysis to similar content path
- `classifier_categories_path_id` - ID of the classifier to categories path
- `content_categories_class_id` - ID of the content categories class
- `similar_content_class_id` - ID of the similar content class
- `analyzer_thought_id` - ID of the content analyzer thought
- `classifier_thought_id` - ID of the content classifier thought

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

- **Content Recommendation**: Link content analysis thoughts to similarity classes for recommendations
- **Document Classification**: Connect classification thoughts to category classes for automated tagging
- **Entity Extraction**: Link extraction thoughts to entity classes for structured data extraction
- **Search Enhancement**: Connect search thoughts to result classes for improved search functionality
- **Content Discovery**: Link analysis thoughts to discovery classes for content exploration
- **Data Enrichment**: Connect enrichment thoughts to schema classes for structured data enhancement
- **Semantic Search**: Link semantic thoughts to similarity classes for advanced search capabilities

## Related Resources

- `tama_space` - Required parent resource for chains and classes
- `tama_chain` - Required parent resource for thoughts
- `tama_modular_thought` - Required source for path connections
- `tama_class` - Required target for path connections
- `tama_source` - AI providers used by the modules in thoughts

## Notes

- Paths require existing thoughts and classes to be created first
- The `thought_id` cannot be changed after creation (forces replacement)
- Parameters must be valid JSON when provided
- Different relation types support different parameter structures
- Multiple paths can connect the same thought to different classes
- Paths enable AI modules to understand data relationships and processing flows

## Best Practices

1. **Clear Relations**: Use descriptive relation types that indicate the path's purpose
2. **Parameter Validation**: Ensure path parameters are properly formatted JSON
3. **Threshold Selection**: Choose appropriate similarity thresholds based on your use case
4. **Result Limits**: Set reasonable max_results limits to control performance
5. **Class Design**: Design target classes with clear, well-defined schemas
6. **Path Organization**: Group related paths logically for easier management
7. **Testing**: Test path configurations with sample data to ensure expected behavior
