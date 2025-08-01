// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package thought_processor

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	tama "github.com/upmaru/tama-go"
	"github.com/upmaru/tama-go/perception"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &DataSource{}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// DataSource defines the data source implementation.
type DataSource struct {
	client *tama.Client
}

// DataSourceModel describes the data source data model.
type DataSourceModel struct {
	Id               types.String            `tfsdk:"id"`
	ThoughtId        types.String            `tfsdk:"thought_id"`
	ModelId          types.String            `tfsdk:"model_id"`
	Type             types.String            `tfsdk:"type"`
	ProvisionState   types.String            `tfsdk:"provision_state"`
	CompletionConfig []CompletionConfigModel `tfsdk:"completion_config"`
	EmbeddingConfig  []EmbeddingConfigModel  `tfsdk:"embedding_config"`
	RerankingConfig  []RerankingConfigModel  `tfsdk:"reranking_config"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_thought_processor"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Tama Perception Thought Processor",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Processor identifier",
				Computed:            true,
			},
			"thought_id": schema.StringAttribute{
				MarkdownDescription: "ID of the thought this processor belongs to",
				Required:            true,
			},
			"model_id": schema.StringAttribute{
				MarkdownDescription: "ID of the model this processor uses",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Processor type",
				Required:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the processor",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"completion_config": schema.ListNestedBlock{
				MarkdownDescription: "Configuration for completion type processors",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"temperature": schema.Float64Attribute{
							MarkdownDescription: "Sampling temperature",
							Computed:            true,
						},
						"tool_choice": schema.StringAttribute{
							MarkdownDescription: "Tool choice strategy",
							Computed:            true,
						},
						"role_mappings": schema.ListNestedAttribute{
							MarkdownDescription: "Role mappings for conversation roles",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"from": schema.StringAttribute{
										MarkdownDescription: "Source role name",
										Computed:            true,
									},
									"to": schema.StringAttribute{
										MarkdownDescription: "Target role name",
										Computed:            true,
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
							MarkdownDescription: "Maximum number of tokens",
							Computed:            true,
						},
						"templates": schema.ListNestedAttribute{
							MarkdownDescription: "Templates for embedding processing",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										MarkdownDescription: "Template type",
										Computed:            true,
									},
									"content": schema.StringAttribute{
										MarkdownDescription: "Template content",
										Computed:            true,
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
							MarkdownDescription: "Number of top results to return",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *DataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tama.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *tama.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get processor from API
	tflog.Debug(ctx, "Reading processor", map[string]any{
		"thought_id": data.ThoughtId.ValueString(),
		"type":       data.Type.ValueString(),
	})

	processorResponse, err := d.client.Perception.GetProcessor(data.ThoughtId.ValueString(), data.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read processor, got error: %s", err))
		return
	}

	// Map response to data source schema
	data.Id = types.StringValue(processorResponse.ID)
	data.ThoughtId = types.StringValue(processorResponse.ThoughtID)
	data.ModelId = types.StringValue(processorResponse.ModelID)
	data.Type = types.StringValue(processorResponse.Type)
	data.ProvisionState = types.StringValue(processorResponse.ProvisionState)

	// Update configuration blocks based on the type and API response
	d.updateConfigurationFromResponse(processorResponse, &data)

	// Write logs using the tflog package
	tflog.Trace(ctx, "read a processor data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// updateConfigurationFromResponse updates the configuration blocks in the model based on the API response.
func (d *DataSource) updateConfigurationFromResponse(processor *perception.Processor, data *DataSourceModel) {
	switch processor.Type {
	case "completion":
		if processor.Configuration != nil {
			config := CompletionConfigModel{}
			if temperature, ok := processor.Configuration["temperature"]; ok {
				if str, ok := temperature.(string); ok {
					if val, err := strconv.ParseFloat(str, 64); err == nil {
						config.Temperature = types.Float64Value(val)
					} else {
						config.Temperature = types.Float64Null()
					}
				} else {
					config.Temperature = types.Float64Null()
				}
			} else {
				config.Temperature = types.Float64Null()
			}
			if toolChoice, ok := processor.Configuration["tool_choice"]; ok {
				if val, ok := toolChoice.(string); ok {
					config.ToolChoice = types.StringValue(val)
				} else {
					config.ToolChoice = types.StringNull()
				}
			} else {
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
			config := EmbeddingConfigModel{}
			if maxTokens, ok := processor.Configuration["max_tokens"]; ok {
				if val, ok := maxTokens.(float64); ok {
					config.MaxTokens = types.Int64Value(int64(val))
				} else {
					config.MaxTokens = types.Int64Null()
				}
			} else {
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
			config := RerankingConfigModel{}
			if topN, ok := processor.Configuration["top_n"]; ok {
				if val, ok := topN.(float64); ok {
					config.TopN = types.Int64Value(int64(val))
				} else {
					config.TopN = types.Int64Null()
				}
			} else {
				config.TopN = types.Int64Null()
			}
			data.RerankingConfig = []RerankingConfigModel{config}
		}
		data.CompletionConfig = []CompletionConfigModel{}
		data.EmbeddingConfig = []EmbeddingConfigModel{}
	}
}
