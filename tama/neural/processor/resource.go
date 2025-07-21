// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package space_processor

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	tama "github.com/upmaru/tama-go"
	"github.com/upmaru/tama-go/neural"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

func NewResource() resource.Resource {
	return &Resource{}
}

// Resource defines the resource implementation.
type Resource struct {
	client *tama.Client
}

// CompletionConfigModel describes the completion configuration data model.
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

type CompletionConfigModel struct {
	Temperature  types.Float64      `tfsdk:"temperature"`
	ToolChoice   types.String       `tfsdk:"tool_choice"`
	RoleMappings []RoleMappingModel `tfsdk:"role_mappings"`
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

// ResourceModel describes the resource data model.
type ResourceModel struct {
	Id               types.String            `tfsdk:"id"`
	SpaceId          types.String            `tfsdk:"space_id"`
	ModelId          types.String            `tfsdk:"model_id"`
	Type             types.String            `tfsdk:"type"`
	CompletionConfig []CompletionConfigModel `tfsdk:"completion_config"`
	EmbeddingConfig  []EmbeddingConfigModel  `tfsdk:"embedding_config"`
	RerankingConfig  []RerankingConfigModel  `tfsdk:"reranking_config"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space_processor"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Neural Space Processor resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Processor identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "ID of the space this processor belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"model_id": schema.StringAttribute{
				MarkdownDescription: "ID of the model this processor uses",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of processor: automatically determined from the configuration block provided (completion_config, embedding_config, or reranking_config)",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"completion_config": schema.ListNestedBlock{
				MarkdownDescription: "Configuration for completion type processors",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"temperature": schema.Float64Attribute{
							MarkdownDescription: "Sampling temperature (default: 0.8)",
							Optional:            true,
						},
						"tool_choice": schema.StringAttribute{
							MarkdownDescription: "Tool choice strategy: required, auto, or any (default: required)",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("required", "auto", "any"),
							},
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
					},
				},
			},
			"embedding_config": schema.ListNestedBlock{
				MarkdownDescription: "Configuration for embedding type processors",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"max_tokens": schema.Int64Attribute{
							MarkdownDescription: "Maximum number of tokens (default: 512)",
							Optional:            true,
						},
						"templates": schema.ListNestedAttribute{
							MarkdownDescription: "Templates for embedding processing",
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										MarkdownDescription: "Template type (e.g., 'query', 'document')",
										Required:            true,
									},
									"content": schema.StringAttribute{
										MarkdownDescription: "Template content with placeholders",
										Required:            true,
									},
								},
							},
						},
					},
				},
			},
			"reranking_config": schema.ListNestedBlock{
				MarkdownDescription: "Configuration for reranking type processors",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"top_n": schema.Int64Attribute{
							MarkdownDescription: "Number of top results to return (default: 3)",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tama.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *tama.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that exactly one config is provided and determine the type
	if err := r.ValidateConfiguration(data); err != nil {
		resp.Diagnostics.AddError("Configuration Error", err.Error())
		return
	}

	// Determine type from configuration
	processorType, err := r.determineTypeFromConfig(data)
	if err != nil {
		resp.Diagnostics.AddError("Configuration Error", err.Error())
		return
	}

	// Set the type in the data model
	data.Type = types.StringValue(processorType)

	// Build configuration based on type
	var config map[string]any
	switch processorType {
	case "completion":
		if len(data.CompletionConfig) > 0 {
			completionConfig := data.CompletionConfig[0]
			config = map[string]any{}
			if !completionConfig.Temperature.IsNull() {
				config["temperature"] = completionConfig.Temperature.ValueFloat64()
			}
			if !completionConfig.ToolChoice.IsNull() {
				config["tool_choice"] = completionConfig.ToolChoice.ValueString()
			}
			if len(completionConfig.RoleMappings) > 0 {
				var roleMappings []map[string]any
				for _, mapping := range completionConfig.RoleMappings {
					roleMappings = append(roleMappings, map[string]any{
						"from": mapping.From.ValueString(),
						"to":   mapping.To.ValueString(),
					})
				}
				config["role_mappings"] = roleMappings
			}
		}
	case "embedding":
		if len(data.EmbeddingConfig) > 0 {
			embeddingConfig := data.EmbeddingConfig[0]
			config = map[string]any{}
			if !embeddingConfig.MaxTokens.IsNull() {
				config["max_tokens"] = embeddingConfig.MaxTokens.ValueInt64()
			}
			if len(embeddingConfig.Templates) > 0 {
				var templates []map[string]any
				for _, template := range embeddingConfig.Templates {
					templates = append(templates, map[string]any{
						"type":    template.Type.ValueString(),
						"content": template.Content.ValueString(),
					})
				}
				config["templates"] = templates
			}
		}
	case "reranking":
		if len(data.RerankingConfig) > 0 {
			rerankingConfig := data.RerankingConfig[0]
			config = map[string]any{}
			if !rerankingConfig.TopN.IsNull() {
				config["top_n"] = rerankingConfig.TopN.ValueInt64()
			}
		}
	}

	// Create processor using the Tama client
	createRequest := neural.CreateProcessorRequest{
		Processor: neural.ProcessorRequestData{
			Type:          processorType,
			Configuration: config,
		},
	}

	tflog.Debug(ctx, "Creating processor", map[string]any{
		"space_id": data.SpaceId.ValueString(),
		"model_id": data.ModelId.ValueString(),
		"type":     processorType,
		"config":   config,
	})

	processorResponse, err := r.client.Neural.CreateProcessor(data.SpaceId.ValueString(), data.ModelId.ValueString(), createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create processor, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.Id = types.StringValue(processorResponse.ID)
	data.ModelId = types.StringValue(processorResponse.ModelID)
	data.Type = types.StringValue(processorResponse.Type)

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a processor resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get processor from API
	processorResponse, err := r.client.Neural.GetProcessor(data.SpaceId.ValueString(), data.ModelId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read processor, got error: %s", err))
		return
	}

	// Update the model with the latest data
	data.ModelId = types.StringValue(processorResponse.ModelID)
	data.Type = types.StringValue(processorResponse.Type)

	// Update configuration blocks based on the type and API response
	r.updateConfigurationFromResponse(processorResponse, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that exactly one config is provided and determine the type
	if err := r.ValidateConfiguration(data); err != nil {
		resp.Diagnostics.AddError("Configuration Error", err.Error())
		return
	}

	// Determine type from configuration
	processorType, err := r.determineTypeFromConfig(data)
	if err != nil {
		resp.Diagnostics.AddError("Configuration Error", err.Error())
		return
	}

	// Set the type in the data model
	data.Type = types.StringValue(processorType)

	// Build configuration based on type
	var config map[string]any
	switch processorType {
	case "completion":
		if len(data.CompletionConfig) > 0 {
			completionConfig := data.CompletionConfig[0]
			config = map[string]any{}
			if !completionConfig.Temperature.IsNull() {
				config["temperature"] = completionConfig.Temperature.ValueFloat64()
			}
			if !completionConfig.ToolChoice.IsNull() {
				config["tool_choice"] = completionConfig.ToolChoice.ValueString()
			}
			if len(completionConfig.RoleMappings) > 0 {
				var roleMappings []map[string]any
				for _, mapping := range completionConfig.RoleMappings {
					roleMappings = append(roleMappings, map[string]any{
						"from": mapping.From.ValueString(),
						"to":   mapping.To.ValueString(),
					})
				}
				config["role_mappings"] = roleMappings
			}
		}
	case "embedding":
		if len(data.EmbeddingConfig) > 0 {
			embeddingConfig := data.EmbeddingConfig[0]
			config = map[string]any{}
			if !embeddingConfig.MaxTokens.IsNull() {
				config["max_tokens"] = embeddingConfig.MaxTokens.ValueInt64()
			}
			if len(embeddingConfig.Templates) > 0 {
				var templates []map[string]any
				for _, template := range embeddingConfig.Templates {
					templates = append(templates, map[string]any{
						"type":    template.Type.ValueString(),
						"content": template.Content.ValueString(),
					})
				}
				config["templates"] = templates
			}
		}
	case "reranking":
		if len(data.RerankingConfig) > 0 {
			rerankingConfig := data.RerankingConfig[0]
			config = map[string]any{}
			if !rerankingConfig.TopN.IsNull() {
				config["top_n"] = rerankingConfig.TopN.ValueInt64()
			}
		}
	}

	// Update processor using the Tama client
	updateRequest := neural.UpdateProcessorRequest{
		Processor: neural.UpdateProcessorData{
			Type:          processorType,
			Configuration: config,
		},
	}

	tflog.Debug(ctx, "Updating processor", map[string]any{
		"id":       data.Id.ValueString(),
		"model_id": data.ModelId.ValueString(),
		"type":     processorType,
		"config":   config,
	})

	processorResponse, err := r.client.Neural.UpdateProcessor(data.SpaceId.ValueString(), data.ModelId.ValueString(), updateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update processor, got error: %s", err))
		return
	}

	// Update the model with the response data
	data.ModelId = types.StringValue(processorResponse.ModelID)
	data.Type = types.StringValue(processorResponse.Type)

	// Update configuration blocks based on the type and API response
	r.updateConfigurationFromResponse(processorResponse, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete processor using the Tama client
	tflog.Debug(ctx, "Deleting processor", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Neural.DeleteProcessor(data.SpaceId.ValueString(), data.ModelId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete processor, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse the compound ID to extract space_id and model_id
	// The import ID should be in the format "space_id/model_id"
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in the format 'space_id/model_id'",
		)
		return
	}

	spaceID := parts[0]
	modelID := parts[1]

	// Get processor from API to populate state
	processorResponse, err := r.client.Neural.GetProcessor(spaceID, modelID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import processor, got error: %s", err))
		return
	}

	// Create model from API response
	data := ResourceModel{
		Id:      types.StringValue(processorResponse.ID),
		SpaceId: types.StringValue(spaceID),
		ModelId: types.StringValue(modelID),
		Type:    types.StringValue(processorResponse.Type),
	}

	// Update configuration blocks based on the type and API response
	r.updateConfigurationFromResponse(processorResponse, &data)

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// determineTypeFromConfig determines the processor type based on which configuration block is provided.
func (r *Resource) determineTypeFromConfig(data ResourceModel) (string, error) {
	configCount := 0
	var processorType string

	if len(data.CompletionConfig) > 0 {
		configCount++
		processorType = "completion"
	}
	if len(data.EmbeddingConfig) > 0 {
		configCount++
		processorType = "embedding"
	}
	if len(data.RerankingConfig) > 0 {
		configCount++
		processorType = "reranking"
	}

	if configCount == 0 {
		return "", fmt.Errorf("exactly one configuration block must be provided (completion_config, embedding_config, or reranking_config)")
	}

	if configCount > 1 {
		return "", fmt.Errorf("only one configuration block can be provided")
	}

	return processorType, nil
}

// ValidateConfiguration ensures exactly one configuration block is provided.
func (r *Resource) ValidateConfiguration(data ResourceModel) error {
	_, err := r.determineTypeFromConfig(data)
	return err
}

// updateConfigurationFromResponse updates the configuration blocks in the model based on the API response.
func (r *Resource) updateConfigurationFromResponse(processor *neural.Processor, data *ResourceModel) {
	switch processor.Type {
	case "completion":
		if processor.Configuration != nil {
			// Preserve existing config values if they exist
			var config CompletionConfigModel
			if len(data.CompletionConfig) > 0 {
				config = data.CompletionConfig[0]
			}

			if temperature, ok := processor.Configuration["temperature"]; ok {
				if strVal, ok := temperature.(string); ok {
					if floatVal, err := strconv.ParseFloat(strVal, 64); err == nil {
						config.Temperature = types.Float64Value(floatVal)
					}
				}
			} else if config.Temperature.IsNull() {
				config.Temperature = types.Float64Null()
			}

			if toolChoice, ok := processor.Configuration["tool_choice"]; ok {
				if val, ok := toolChoice.(string); ok {
					config.ToolChoice = types.StringValue(val)
				}
			} else if config.ToolChoice.IsNull() {
				config.ToolChoice = types.StringNull()
			}

			if roleMappings, ok := processor.Configuration["role_mappings"]; ok {
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
					config.RoleMappings = roleMappingModels
				}
			}
			data.CompletionConfig = []CompletionConfigModel{config}
		}
		data.EmbeddingConfig = []EmbeddingConfigModel{}
		data.RerankingConfig = []RerankingConfigModel{}

	case "embedding":
		if processor.Configuration != nil {
			// Preserve existing config values if they exist
			var config EmbeddingConfigModel
			if len(data.EmbeddingConfig) > 0 {
				config = data.EmbeddingConfig[0]
			}

			if maxTokens, ok := processor.Configuration["max_tokens"]; ok {
				if val, ok := maxTokens.(float64); ok {
					config.MaxTokens = types.Int64Value(int64(val))
				}
			} else if config.MaxTokens.IsNull() {
				config.MaxTokens = types.Int64Null()
			}

			if templates, ok := processor.Configuration["templates"]; ok {
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
					config.Templates = templateModels
				}
			}
			data.EmbeddingConfig = []EmbeddingConfigModel{config}
		}
		data.CompletionConfig = []CompletionConfigModel{}
		data.RerankingConfig = []RerankingConfigModel{}

	case "reranking":
		if processor.Configuration != nil {
			// Preserve existing config values if they exist
			var config RerankingConfigModel
			if len(data.RerankingConfig) > 0 {
				config = data.RerankingConfig[0]
			}

			if topN, ok := processor.Configuration["top_n"]; ok {
				if val, ok := topN.(float64); ok {
					config.TopN = types.Int64Value(int64(val))
				}
			} else if config.TopN.IsNull() {
				config.TopN = types.Int64Null()
			}
			data.RerankingConfig = []RerankingConfigModel{config}
		}
		data.CompletionConfig = []CompletionConfigModel{}
		data.EmbeddingConfig = []EmbeddingConfigModel{}
	}
}
