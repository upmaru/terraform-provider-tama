// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package processor

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ProcessorConfig represents a generic processor configuration interface.
type ProcessorConfig interface {
	GetCompletion() *CompletionConfigModel
	GetEmbedding() *EmbeddingConfigModel
	GetReranking() *RerankingConfigModel
}

// Implement ProcessorConfig for new model types.
func (m *NeuralProcessorModel) GetCompletion() *CompletionConfigModel {
	return m.Completion
}

func (m *NeuralProcessorModel) GetEmbedding() *EmbeddingConfigModel {
	return m.Embedding
}

func (m *NeuralProcessorModel) GetReranking() *RerankingConfigModel {
	return m.Reranking
}

func (m *PerceptionProcessorModel) GetCompletion() *CompletionConfigModel {
	return m.Completion
}

func (m *PerceptionProcessorModel) GetEmbedding() *EmbeddingConfigModel {
	return m.Embedding
}

func (m *PerceptionProcessorModel) GetReranking() *RerankingConfigModel {
	return m.Reranking
}

// Legacy model implementations.
func (m *ThoughtProcessorResourceModel) GetCompletion() *CompletionConfigModel {
	return m.Completion
}

func (m *ThoughtProcessorResourceModel) GetEmbedding() *EmbeddingConfigModel {
	return m.Embedding
}

func (m *ThoughtProcessorResourceModel) GetReranking() *RerankingConfigModel {
	return m.Reranking
}

func (m *SpaceProcessorResourceModel) GetCompletion() *CompletionConfigModel {
	return m.Completion
}

func (m *SpaceProcessorResourceModel) GetEmbedding() *EmbeddingConfigModel {
	return m.Embedding
}

func (m *SpaceProcessorResourceModel) GetReranking() *RerankingConfigModel {
	return m.Reranking
}

// DetermineTypeFromConfig determines the processor type based on which configuration block is provided.
func DetermineTypeFromConfig(config ProcessorConfig) (string, error) {
	configCount := 0
	var processorType string

	if config.GetCompletion() != nil {
		configCount++
		processorType = "completion"
	}
	if config.GetEmbedding() != nil {
		configCount++
		processorType = "embedding"
	}
	if config.GetReranking() != nil {
		configCount++
		processorType = "reranking"
	}

	if configCount == 0 {
		return "", fmt.Errorf("exactly one configuration block must be provided (completion, embedding, or reranking)")
	}

	if configCount > 1 {
		return "", fmt.Errorf("only one configuration block can be provided")
	}

	return processorType, nil
}

// BuildConfig builds the configuration map for API requests.
func BuildConfig(config ProcessorConfig, processorType string, isCreate bool) (map[string]any, error) {
	switch processorType {
	case "completion":
		return buildCompletionConfig(config.GetCompletion())
	case "embedding":
		return buildEmbeddingConfig(config.GetEmbedding())
	case "reranking":
		return buildRerankingConfig(config.GetReranking())
	default:
		return nil, fmt.Errorf("unknown processor type: %s", processorType)
	}
}

func buildCompletionConfig(completion *CompletionConfigModel) (map[string]any, error) {
	if completion == nil {
		return map[string]any{}, nil
	}

	config := map[string]any{}

	if !completion.Temperature.IsNull() && !completion.Temperature.IsUnknown() {
		config["temperature"] = completion.Temperature.ValueFloat64()
	}

	if !completion.ToolChoice.IsNull() && !completion.ToolChoice.IsUnknown() {
		config["tool_choice"] = completion.ToolChoice.ValueString()
	}

	if len(completion.RoleMappings) > 0 {
		var roleMappings []map[string]any
		for _, mapping := range completion.RoleMappings {
			roleMappings = append(roleMappings, map[string]any{
				"from": mapping.From.ValueString(),
				"to":   mapping.To.ValueString(),
			})
		}
		config["role_mappings"] = roleMappings
	}

	// Parse parameters if provided
	if !completion.Parameters.IsNull() && !completion.Parameters.IsUnknown() && completion.Parameters.ValueString() != "" {
		var parametersMap map[string]any
		if err := json.Unmarshal([]byte(completion.Parameters.ValueString()), &parametersMap); err != nil {
			return nil, fmt.Errorf("unable to parse parameters as JSON: %s", err)
		}
		config["parameters"] = parametersMap
	}

	return config, nil
}

func buildEmbeddingConfig(embedding *EmbeddingConfigModel) (map[string]any, error) {
	if embedding == nil {
		return map[string]any{}, nil
	}

	config := map[string]any{}

	if !embedding.MaxTokens.IsNull() && !embedding.MaxTokens.IsUnknown() {
		config["max_tokens"] = embedding.MaxTokens.ValueInt64()
	}

	if len(embedding.Templates) > 0 {
		var templates []map[string]any
		for _, template := range embedding.Templates {
			templates = append(templates, map[string]any{
				"type":    template.Type.ValueString(),
				"content": template.Content.ValueString(),
			})
		}
		config["templates"] = templates
	}

	return config, nil
}

func buildRerankingConfig(reranking *RerankingConfigModel) (map[string]any, error) {
	if reranking == nil {
		return map[string]any{}, nil
	}

	config := map[string]any{}

	// Parse parameters if provided
	if !reranking.Parameters.IsNull() && !reranking.Parameters.IsUnknown() && reranking.Parameters.ValueString() != "" {
		var parametersMap map[string]any
		if err := json.Unmarshal([]byte(reranking.Parameters.ValueString()), &parametersMap); err != nil {
			return nil, fmt.Errorf("unable to parse parameters as JSON: %s", err)
		}
		config["parameters"] = parametersMap
	}

	return config, nil
}

func updateCompletionFromResponse(processorConfig map[string]any, config ProcessorConfig) {
	// Get existing config or create new one
	var completionConfig CompletionConfigModel
	if existingCompletion := config.GetCompletion(); existingCompletion != nil {
		completionConfig = *existingCompletion
	}

	if temperature, ok := processorConfig["temperature"]; ok {
		if floatVal, ok := temperature.(float64); ok {
			completionConfig.Temperature = types.Float64Value(floatVal)
		} else if strVal, ok := temperature.(string); ok {
			if floatVal, err := strconv.ParseFloat(strVal, 64); err == nil {
				completionConfig.Temperature = types.Float64Value(floatVal)
			}
		}
	}

	if toolChoice, ok := processorConfig["tool_choice"]; ok {
		if val, ok := toolChoice.(string); ok {
			completionConfig.ToolChoice = types.StringValue(val)
		}
	}

	if roleMappings, ok := processorConfig["role_mappings"]; ok {
		if mappings, ok := roleMappings.([]any); ok && len(mappings) > 0 {
			var roleMappingModels []RoleMappingModel
			for _, mapping := range mappings {
				if mappingMap, ok := mapping.(map[string]any); ok {
					if from, ok := mappingMap["from"].(string); ok {
						if to, ok := mappingMap["to"].(string); ok {
							roleMappingModels = append(roleMappingModels, RoleMappingModel{
								From: types.StringValue(from),
								To:   types.StringValue(to),
							})
						}
					}
				}
			}
			completionConfig.RoleMappings = roleMappingModels
		}
	}

	// Handle parameters from response
	if parameters, ok := processorConfig["parameters"]; ok {
		if paramMap, ok := parameters.(map[string]any); ok && len(paramMap) > 0 {
			parametersJSON, err := json.Marshal(paramMap)
			if err == nil {
				// Only update if user didn't provide parameters or if the values are different
				if completionConfig.Parameters.IsNull() || completionConfig.Parameters.IsUnknown() {
					completionConfig.Parameters = types.StringValue(string(parametersJSON))
				}
			}
		} else if completionConfig.Parameters.IsNull() || completionConfig.Parameters.IsUnknown() {
			completionConfig.Parameters = types.StringValue("")
		}
	} else if completionConfig.Parameters.IsNull() || completionConfig.Parameters.IsUnknown() {
		completionConfig.Parameters = types.StringValue("")
	}

	// Update the config - this needs to be handled by the specific model type
	updateCompletionInConfig(config, &completionConfig)
}

func updateEmbeddingFromResponse(processorConfig map[string]any, config ProcessorConfig) {
	// Get existing config or create new one
	var embeddingConfig EmbeddingConfigModel
	if existingEmbedding := config.GetEmbedding(); existingEmbedding != nil {
		embeddingConfig = *existingEmbedding
	}

	if maxTokens, ok := processorConfig["max_tokens"]; ok {
		if val, ok := maxTokens.(float64); ok {
			embeddingConfig.MaxTokens = types.Int64Value(int64(val))
		} else if intVal, ok := maxTokens.(int64); ok {
			embeddingConfig.MaxTokens = types.Int64Value(intVal)
		} else if intVal, ok := maxTokens.(int); ok {
			embeddingConfig.MaxTokens = types.Int64Value(int64(intVal))
		}
	}

	if templates, ok := processorConfig["templates"]; ok {
		if tmplList, ok := templates.([]any); ok && len(tmplList) > 0 {
			var templateModels []TemplateModel
			for _, template := range tmplList {
				if templateMap, ok := template.(map[string]any); ok {
					if tmplType, ok := templateMap["type"].(string); ok {
						if content, ok := templateMap["content"].(string); ok {
							templateModels = append(templateModels, TemplateModel{
								Type:    types.StringValue(tmplType),
								Content: types.StringValue(content),
							})
						}
					}
				}
			}
			embeddingConfig.Templates = templateModels
		}
	}

	updateEmbeddingInConfig(config, &embeddingConfig)
}

func updateRerankingFromResponse(processorConfig map[string]any, config ProcessorConfig) {
	// Get existing config or create new one
	var rerankingConfig RerankingConfigModel
	if existingReranking := config.GetReranking(); existingReranking != nil {
		rerankingConfig = *existingReranking
	}

	// Handle parameters from response
	if parameters, ok := processorConfig["parameters"]; ok {
		if paramMap, ok := parameters.(map[string]any); ok && len(paramMap) > 0 {
			parametersJSON, err := json.Marshal(paramMap)
			if err == nil {
				if rerankingConfig.Parameters.IsNull() || rerankingConfig.Parameters.IsUnknown() {
					rerankingConfig.Parameters = types.StringValue(string(parametersJSON))
				}
			}
		} else if rerankingConfig.Parameters.IsNull() || rerankingConfig.Parameters.IsUnknown() {
			rerankingConfig.Parameters = types.StringValue("")
		}
	} else if rerankingConfig.Parameters.IsNull() || rerankingConfig.Parameters.IsUnknown() {
		rerankingConfig.Parameters = types.StringValue("")
	}

	updateRerankingInConfig(config, &rerankingConfig)
}

// These functions need to be implemented differently for each model type.
func updateCompletionInConfig(config ProcessorConfig, completion *CompletionConfigModel) {
	switch c := config.(type) {
	case *NeuralProcessorModel:
		c.Completion = completion
		c.Embedding = nil
		c.Reranking = nil
	case *PerceptionProcessorModel:
		c.Completion = completion
		c.Embedding = nil
		c.Reranking = nil
	case *ThoughtProcessorResourceModel:
		c.Completion = completion
		c.Embedding = nil
		c.Reranking = nil
	case *SpaceProcessorResourceModel:
		c.Completion = completion
		c.Embedding = nil
		c.Reranking = nil
	}
}

func updateEmbeddingInConfig(config ProcessorConfig, embedding *EmbeddingConfigModel) {
	switch c := config.(type) {
	case *NeuralProcessorModel:
		c.Embedding = embedding
		c.Completion = nil
		c.Reranking = nil
	case *PerceptionProcessorModel:
		c.Embedding = embedding
		c.Completion = nil
		c.Reranking = nil
	case *ThoughtProcessorResourceModel:
		c.Embedding = embedding
		c.Completion = nil
		c.Reranking = nil
	case *SpaceProcessorResourceModel:
		c.Embedding = embedding
		c.Completion = nil
		c.Reranking = nil
	}
}

func updateRerankingInConfig(config ProcessorConfig, reranking *RerankingConfigModel) {
	switch c := config.(type) {
	case *NeuralProcessorModel:
		c.Reranking = reranking
		c.Completion = nil
		c.Embedding = nil
	case *PerceptionProcessorModel:
		c.Reranking = reranking
		c.Completion = nil
		c.Embedding = nil
	case *ThoughtProcessorResourceModel:
		c.Reranking = reranking
		c.Completion = nil
		c.Embedding = nil
	case *SpaceProcessorResourceModel:
		c.Reranking = reranking
		c.Completion = nil
		c.Embedding = nil
	}
}

// EnsureParametersInitialized ensures that parameters field is set to a known value before API response processing.
func EnsureParametersInitialized(config ProcessorConfig) {
	if completion := config.GetCompletion(); completion != nil && completion.Parameters.IsNull() {
		completion.Parameters = types.StringValue("")
	}
}

// DetermineProcessorType determines the processor type for the new model types.
func DetermineProcessorType(config ProcessorConfig) string {
	if config.GetCompletion() != nil {
		return "completion"
	}
	if config.GetEmbedding() != nil {
		return "embedding"
	}
	if config.GetReranking() != nil {
		return "reranking"
	}
	return ""
}

// BuildConfiguration builds the configuration for API requests using the new models.
func BuildConfiguration(config ProcessorConfig) map[string]any {
	processorType := DetermineProcessorType(config)
	if processorType == "" {
		return map[string]any{}
	}

	configMap, _ := BuildConfig(config, processorType, true)
	return configMap
}

// UpdateConfigurationFromResponse updates config from API response - simplified signature.
func UpdateConfigurationFromResponse(processorConfig map[string]any, config ProcessorConfig) {
	processorType := DetermineProcessorType(config)
	if processorType == "" {
		// During import, the config might not have blocks set yet,
		// so we need to determine the type from the configuration
		if _, hasTemp := processorConfig["temperature"]; hasTemp {
			processorType = "completion"
		} else if _, hasToolChoice := processorConfig["tool_choice"]; hasToolChoice {
			processorType = "completion"
		} else if _, hasRoleMappings := processorConfig["role_mappings"]; hasRoleMappings {
			processorType = "completion"
		} else if _, hasMaxTokens := processorConfig["max_tokens"]; hasMaxTokens {
			processorType = "embedding"
		} else if _, hasTemplates := processorConfig["templates"]; hasTemplates {
			processorType = "embedding"
		} else if _, hasParams := processorConfig["parameters"]; hasParams {
			processorType = "reranking"
		}
	}

	switch processorType {
	case "completion":
		updateCompletionFromResponse(processorConfig, config)
	case "embedding":
		updateEmbeddingFromResponse(processorConfig, config)
	case "reranking":
		updateRerankingFromResponse(processorConfig, config)
	}
}

// UpdateConfigurationFromResponseWithType updates config from API response with explicit processor type.
func UpdateConfigurationFromResponseWithType(processorConfig map[string]any, config ProcessorConfig, processorType string) {
	switch processorType {
	case "completion":
		updateCompletionFromResponse(processorConfig, config)
	case "embedding":
		updateEmbeddingFromResponse(processorConfig, config)
	case "reranking":
		updateRerankingFromResponse(processorConfig, config)
	}
}
