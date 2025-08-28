// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package processor

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	jsonplanmodifier "github.com/upmaru/terraform-provider-tama/internal/planmodifier"
)

// GetProcessorBlocks returns the common processor blocks (completion, embedding, reranking).
func GetProcessorBlocks(includeValidation bool) map[string]schema.Block {
	blocks := map[string]schema.Block{
		"completion": schema.SingleNestedBlock{
			MarkdownDescription: "Configuration for completion type processors",
			Attributes:          getCompletionAttributes(includeValidation),
		},
		"embedding": schema.SingleNestedBlock{
			MarkdownDescription: "Configuration for embedding type processors",
			Attributes:          getEmbeddingAttributes(),
		},
		"reranking": schema.SingleNestedBlock{
			MarkdownDescription: "Configuration for reranking type processors",
			Attributes:          getRerankingAttributes(),
		},
	}
	return blocks
}

// GetNeuralProcessorSchema returns the complete schema for neural processors.
func GetNeuralProcessorSchema() (map[string]schema.Attribute, map[string]schema.Block) {
	attributes := GetBaseAttributes()
	attributes["space_id"] = GetSpaceIdAttribute()
	blocks := GetProcessorBlocks(true) // Include validation for neural
	return attributes, blocks
}

// GetPerceptionProcessorSchema returns the complete schema for perception processors.
func GetPerceptionProcessorSchema() (map[string]schema.Attribute, map[string]schema.Block) {
	attributes := GetBaseAttributes()
	attributes["thought_id"] = GetThoughtIdAttribute()
	blocks := GetProcessorBlocks(false) // No validation for perception
	return attributes, blocks
}

func getCompletionAttributes(includeValidation bool) map[string]schema.Attribute {
	attributes := map[string]schema.Attribute{
		"temperature": schema.Float64Attribute{
			MarkdownDescription: "Sampling temperature",
			Optional:            true,
			Computed:            true,
		},
		"tool_choice": schema.StringAttribute{
			MarkdownDescription: "Tool choice strategy",
			Optional:            true,
			Computed:            true,
		},
		"role_mappings": schema.ListNestedAttribute{
			MarkdownDescription: "Role mappings for conversation roles",
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"from": schema.StringAttribute{
						MarkdownDescription: "Source role name",
						Required:            true,
					},
					"to": schema.StringAttribute{
						MarkdownDescription: "Target role name",
						Required:            true,
					},
				},
			},
		},
		"parameters": schema.StringAttribute{
			MarkdownDescription: "Additional parameters as JSON string",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				jsonplanmodifier.JSONNormalize(),
			},
		},
	}

	// Add validation for neural processor
	if includeValidation {
		if toolChoiceAttr, ok := attributes["tool_choice"]; ok {
			if stringAttr, ok := toolChoiceAttr.(schema.StringAttribute); ok {
				stringAttr.MarkdownDescription = "Tool choice strategy: required, auto, or any (default: required)"
				stringAttr.Validators = []validator.String{
					stringvalidator.OneOf("required", "auto", "any"),
				}
				attributes["tool_choice"] = stringAttr
			}
		}

		if tempAttr, ok := attributes["temperature"]; ok {
			if float64Attr, ok := tempAttr.(schema.Float64Attribute); ok {
				float64Attr.MarkdownDescription = "Sampling temperature (default: 0.8)"
				attributes["temperature"] = float64Attr
			}
		}

		if paramsAttr, ok := attributes["parameters"]; ok {
			if stringAttr, ok := paramsAttr.(schema.StringAttribute); ok {
				stringAttr.MarkdownDescription = "Additional parameters as JSON string (e.g., '{\"max_tokens\": 1000, \"stop\": [\"\\n\"]}')"
				attributes["parameters"] = stringAttr
			}
		}
	}

	return attributes
}

func getEmbeddingAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"max_tokens": schema.Int64Attribute{
			MarkdownDescription: "Maximum number of tokens",
			Optional:            true,
			Computed:            true,
		},
		"templates": schema.ListNestedAttribute{
			MarkdownDescription: "Templates for embedding processing",
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Template type",
						Required:            true,
					},
					"content": schema.StringAttribute{
						MarkdownDescription: "Template content",
						Required:            true,
					},
				},
			},
		},
	}
}

func getRerankingAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"parameters": schema.StringAttribute{
			MarkdownDescription: "Additional parameters as JSON string",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				jsonplanmodifier.JSONNormalize(),
			},
		},
	}
}

// GetBaseAttributes returns the common base attributes (id, model_id, type).
func GetBaseAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Processor identifier",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"model_id": schema.StringAttribute{
			MarkdownDescription: "ID of the model this processor uses",
			Required:            true,
		},
		"type": schema.StringAttribute{
			MarkdownDescription: "Type of processor (e.g., 'completion', 'embedding', 'reranking')",
			Computed:            true,
		},
	}
}

// GetThoughtIdAttribute returns the thought_id attribute for perception processors.
func GetThoughtIdAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "ID of the thought this processor belongs to",
		Required:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

// GetSpaceIdAttribute returns the space_id attribute for neural processors.
func GetSpaceIdAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "ID of the space this processor belongs to",
		Required:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}
