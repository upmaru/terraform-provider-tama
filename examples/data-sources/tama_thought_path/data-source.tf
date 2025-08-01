# Example configuration for tama_modular_thought_path data source

terraform {
  required_providers {
    tama = {
      source = "upmaru/tama"
    }
  }
}

# Fetch information about an existing thought path by ID
data "tama_modular_thought_path" "example" {
  id = "path-12345"
}

# Use the path data source to create a similar path with different parameters
resource "tama_modular_thought_path" "derived_path" {
  thought_id      = data.tama_modular_thought_path.example.thought_id
  target_class_id = data.tama_modular_thought_path.example.target_class_id

  # Modify the existing parameters
  parameters = jsonencode(merge(
    jsondecode(data.tama_modular_thought_path.example.parameters),
    {
      max_results  = 20
      boost_recent = true
    }
  ))
}

# Fetch multiple paths for comparison
data "tama_modular_thought_path" "classification_path" {
  id = "path-classification-123"
}

data "tama_modular_thought_path" "similarity_path" {
  id = "path-similarity-456"
}

data "tama_modular_thought_path" "extraction_path" {
  id = "path-extraction-789"
}

# Variable for dynamic path ID lookup
variable "path_id" {
  description = "ID of the path to fetch"
  type        = string
  default     = "path-12345"
}

# Alternative example using variable
data "tama_modular_thought_path" "variable_example" {
  id = var.path_id
}

# Get related resources based on path information
data "tama_modular_thought" "source_thought" {
  id = data.tama_modular_thought_path.example.thought_id
}

data "tama_class" "target_class" {
  id = data.tama_modular_thought_path.example.target_class_id
}

# Local values for processing path data
locals {
  # Parse parameters from different paths
  example_params        = jsondecode(data.tama_modular_thought_path.example.parameters)
  classification_params = jsondecode(data.tama_modular_thought_path.classification_path.parameters)
  similarity_params     = jsondecode(data.tama_modular_thought_path.similarity_path.parameters)
  extraction_params     = jsondecode(data.tama_modular_thought_path.extraction_path.parameters)

  # Extract relation types
  example_relation        = try(local.example_params.relation, "unknown")
  classification_relation = try(local.classification_params.relation, "unknown")
  similarity_relation     = try(local.similarity_params.relation, "unknown")
  extraction_relation     = try(local.extraction_params.relation, "unknown")

  # Check if paths share the same thought
  same_thought_classification = (
    data.tama_modular_thought_path.example.thought_id ==
    data.tama_modular_thought_path.classification_path.thought_id
  )

  same_thought_similarity = (
    data.tama_modular_thought_path.example.thought_id ==
    data.tama_modular_thought_path.similarity_path.thought_id
  )

  # Get unique target classes from all fetched paths
  target_classes = toset([
    data.tama_modular_thought_path.example.target_class_id,
    data.tama_modular_thought_path.classification_path.target_class_id,
    data.tama_modular_thought_path.similarity_path.target_class_id,
    data.tama_modular_thought_path.extraction_path.target_class_id
  ])

  # Extract thresholds and confidence scores
  similarity_threshold      = try(local.similarity_params.similarity.threshold, 0.5)
  classification_confidence = try(local.classification_params.confidence_threshold, 0.5)
  extraction_confidence     = try(local.extraction_params.confidence_threshold, 0.5)

  # Check for advanced features
  has_similarity_config = can(local.similarity_params.similarity)
  has_filters           = can(local.similarity_params.filters)
  has_boost_recent      = try(local.similarity_params.boost_recent, false)

  # Analyze max results across paths
  max_results_values = [
    try(local.example_params.max_results, 10),
    try(local.classification_params.max_results, 10),
    try(local.similarity_params.max_results, 10),
    try(local.extraction_params.max_results, 10)
  ]

  avg_max_results = sum(local.max_results_values) / length(local.max_results_values)

  # Create optimized parameters based on analysis
  optimized_similarity_params = {
    relation = "similarity"
    similarity = {
      threshold = max(local.similarity_threshold, 0.8)
      algorithm = "semantic"
    }
    max_results    = ceil(local.avg_max_results * 1.5)
    boost_recent   = true
    quality_filter = true
  }
}

# Create new paths based on analysis of existing ones
resource "tama_modular_thought_path" "optimized_similarity" {
  thought_id      = data.tama_modular_thought_path.similarity_path.thought_id
  target_class_id = data.tama_modular_thought_path.similarity_path.target_class_id

  parameters = jsonencode(local.optimized_similarity_params)
}

# Create a consolidated path that combines insights from multiple paths
resource "tama_modular_thought_path" "consolidated" {
  thought_id      = data.tama_modular_thought_path.example.thought_id
  target_class_id = data.tama_modular_thought_path.example.target_class_id

  parameters = jsonencode({
    relation = local.example_relation
    similarity = local.has_similarity_config ? {
      threshold = max(local.similarity_threshold, 0.7)
      algorithm = "hybrid"
      } : {
      threshold = 0.8
      algorithm = "semantic"
    }
    classification = {
      confidence_threshold = max(local.classification_confidence, 0.8)
      multi_label          = true
    }
    extraction = {
      confidence_threshold = max(local.extraction_confidence, 0.7)
      extract_types        = ["all"]
    }
    max_results  = ceil(local.avg_max_results)
    filters      = local.has_filters ? local.similarity_params.filters : {}
    boost_recent = true
  })
}

# Use for_each to analyze multiple paths dynamically
variable "analysis_path_ids" {
  description = "List of path IDs to analyze"
  type        = list(string)
  default     = ["path-1", "path-2", "path-3"]
}

data "tama_modular_thought_path" "analysis_paths" {
  for_each = toset(var.analysis_path_ids)
  id       = each.value
}

# Filter and group paths by relation type
locals {
  paths_by_relation = {
    for k, path in data.tama_modular_thought_path.analysis_paths :
    try(jsondecode(path.parameters).relation, "unknown") => k...
  }

  similarity_analysis_paths = {
    for k, path in data.tama_modular_thought_path.analysis_paths :
    k => path
    if can(jsondecode(path.parameters).relation) &&
    jsondecode(path.parameters).relation == "similarity"
  }

  high_threshold_paths = {
    for k, path in data.tama_modular_thought_path.analysis_paths :
    k => path
    if can(jsondecode(path.parameters).similarity.threshold) &&
    jsondecode(path.parameters).similarity.threshold > 0.8
  }
}

# Output the path information
output "example_path_id" {
  description = "ID of the example path"
  value       = data.tama_modular_thought_path.example.id
}

output "example_thought_id" {
  description = "ID of the thought connected to the example path"
  value       = data.tama_modular_thought_path.example.thought_id
}

output "example_target_class_id" {
  description = "ID of the target class for the example path"
  value       = data.tama_modular_thought_path.example.target_class_id
}

output "example_parameters" {
  description = "Parameters of the example path (raw JSON)"
  value       = data.tama_modular_thought_path.example.parameters
}

output "example_relation_type" {
  description = "Relation type extracted from example path"
  value       = local.example_relation
}

output "parsed_example_parameters" {
  description = "Parsed parameters from the example path"
  value       = local.example_params
}

# Comparison outputs
output "classification_path_id" {
  description = "ID of the classification path"
  value       = data.tama_modular_thought_path.classification_path.id
}

output "similarity_path_id" {
  description = "ID of the similarity path"
  value       = data.tama_modular_thought_path.similarity_path.id
}

output "extraction_path_id" {
  description = "ID of the extraction path"
  value       = data.tama_modular_thought_path.extraction_path.id
}

# Analysis outputs
output "paths_share_thought_classification" {
  description = "Whether example and classification paths share the same thought"
  value       = local.same_thought_classification
}

output "paths_share_thought_similarity" {
  description = "Whether example and similarity paths share the same thought"
  value       = local.same_thought_similarity
}

output "unique_target_classes" {
  description = "Unique target class IDs from all fetched paths"
  value       = local.target_classes
}

output "similarity_threshold" {
  description = "Similarity threshold from similarity path"
  value       = local.similarity_threshold
}

output "classification_confidence" {
  description = "Classification confidence threshold"
  value       = local.classification_confidence
}

output "extraction_confidence" {
  description = "Extraction confidence threshold"
  value       = local.extraction_confidence
}

output "average_max_results" {
  description = "Average max results across all paths"
  value       = local.avg_max_results
}

# Feature analysis outputs
output "has_similarity_configuration" {
  description = "Whether similarity path has similarity configuration"
  value       = local.has_similarity_config
}

output "has_filtering" {
  description = "Whether similarity path has filters configured"
  value       = local.has_filters
}

output "has_boost_recent" {
  description = "Whether similarity path has boost_recent enabled"
  value       = local.has_boost_recent
}

# Related resource outputs
output "source_thought_relation" {
  description = "Relation type of the source thought"
  value       = data.tama_modular_thought.source_thought.relation
}

output "source_thought_chain_id" {
  description = "Chain ID of the source thought"
  value       = data.tama_modular_thought.source_thought.chain_id
}

output "target_class_schema" {
  description = "Schema of the target class"
  value       = data.tama_class.target_class.schema_json
}

# Dynamic analysis outputs
output "paths_by_relation_type" {
  description = "Paths grouped by their relation type"
  value       = local.paths_by_relation
}

output "similarity_paths_count" {
  description = "Number of similarity-based paths"
  value       = length(local.similarity_analysis_paths)
}

output "high_threshold_paths_count" {
  description = "Number of paths with high similarity thresholds"
  value       = length(local.high_threshold_paths)
}

# Derived resource outputs
output "derived_path_id" {
  description = "ID of the derived path created from example"
  value       = tama_modular_thought_path.derived_path.id
}

output "optimized_path_id" {
  description = "ID of the optimized similarity path"
  value       = tama_modular_thought_path.optimized_similarity.id
}

output "consolidated_path_id" {
  description = "ID of the consolidated path"
  value       = tama_modular_thought_path.consolidated.id
}

output "optimized_parameters" {
  description = "Optimized parameters used for the new similarity path"
  value       = local.optimized_similarity_params
}
