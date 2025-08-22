// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package processor

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RoleMappingModel describes the role mapping data model.
type RoleMappingModel struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

// TemplateModel describes the template data model.
type TemplateModel struct {
	Type    types.String `tfsdk:"type"`
	Content types.String `tfsdk:"content"`
}

// CompletionConfigModel describes the completion configuration data model.
type CompletionConfigModel struct {
	Temperature  types.Float64      `tfsdk:"temperature"`
	ToolChoice   types.String       `tfsdk:"tool_choice"`
	RoleMappings []RoleMappingModel `tfsdk:"role_mappings"`
	Parameters   types.String       `tfsdk:"parameters"`
}

// EmbeddingConfigModel describes the embedding configuration data model.
type EmbeddingConfigModel struct {
	MaxTokens types.Int64     `tfsdk:"max_tokens"`
	Templates []TemplateModel `tfsdk:"templates"`
}

// RerankingConfigModel describes the reranking configuration data model.
type RerankingConfigModel struct {
	TopN types.Int64 `tfsdk:"top_n"`
}

// ProcessorModel describes the common processor data model.
type ProcessorModel struct {
	Id      types.String `tfsdk:"id"`
	ModelId types.String `tfsdk:"model_id"`
	Type    types.String `tfsdk:"type"`
}

// NeuralProcessorModel for neural processors.
type NeuralProcessorModel struct {
	ProcessorModel
	SpaceId    types.String           `tfsdk:"space_id"`
	Completion *CompletionConfigModel `tfsdk:"completion"`
	Embedding  *EmbeddingConfigModel  `tfsdk:"embedding"`
	Reranking  *RerankingConfigModel  `tfsdk:"reranking"`
}

// PerceptionProcessorModel for perception processors.
type PerceptionProcessorModel struct {
	ProcessorModel
	ThoughtId  types.String           `tfsdk:"thought_id"`
	Completion *CompletionConfigModel `tfsdk:"completion"`
	Embedding  *EmbeddingConfigModel  `tfsdk:"embedding"`
	Reranking  *RerankingConfigModel  `tfsdk:"reranking"`
}

// Legacy models for backward compatibility.
type BaseResourceModel struct {
	Id         types.String           `tfsdk:"id"`
	ModelId    types.String           `tfsdk:"model_id"`
	Type       types.String           `tfsdk:"type"`
	Completion *CompletionConfigModel `tfsdk:"completion"`
	Embedding  *EmbeddingConfigModel  `tfsdk:"embedding"`
	Reranking  *RerankingConfigModel  `tfsdk:"reranking"`
}

type ThoughtProcessorResourceModel struct {
	BaseResourceModel
	ThoughtId types.String `tfsdk:"thought_id"`
}

type SpaceProcessorResourceModel struct {
	BaseResourceModel
	SpaceId types.String `tfsdk:"space_id"`
}
