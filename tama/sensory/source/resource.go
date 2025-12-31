// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package source

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	tama "github.com/upmaru/tama-go"
	"github.com/upmaru/tama-go/sensory"
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

// ResourceModel describes the resource data model.
type ResourceModel struct {
	Id             types.String  `tfsdk:"id"`
	SpaceId        types.String  `tfsdk:"space_id"`
	Name           types.String  `tfsdk:"name"`
	Slug           types.String  `tfsdk:"slug"`
	Type           types.String  `tfsdk:"type"`
	Endpoint       types.String  `tfsdk:"endpoint"`
	ApiKey         types.String  `tfsdk:"api_key"`
	ProvisionState types.String  `tfsdk:"provision_state"`
	Request        *RequestModel `tfsdk:"request"`
}

// RequestModel describes the request configuration
type RequestModel struct {
	Headers         []HeaderModel         `tfsdk:"headers"`
	SessionAffinity *SessionAffinityModel `tfsdk:"session_affinity"`
}

// HeaderModel describes a request header
type HeaderModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// SessionAffinityModel describes session affinity configuration
type SessionAffinityModel struct {
	Location types.String `tfsdk:"location"`
	Key      types.String `tfsdk:"key"`
	Value    types.String `tfsdk:"value"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Sensory Source resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Source identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "ID of the space this source belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the source",
				Required:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Source slug (generated from name)",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the source (e.g., 'model')",
				Required:            true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "API endpoint URL for the source",
				Required:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key for authenticating with the source",
				Required:            true,
				Sensitive:           true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the source ('active' or 'inactive')",
				Computed:            true,
			},
			"request": schema.SingleNestedAttribute{
				MarkdownDescription: "Request configuration for the source",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"headers": schema.ListNestedAttribute{
						MarkdownDescription: "Custom headers to include in requests",
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "Header name",
									Required:            true,
								},
								"value": schema.StringAttribute{
									MarkdownDescription: "Header value",
									Required:            true,
								},
							},
						},
					},
					"session_affinity": schema.SingleNestedAttribute{
						MarkdownDescription: "Session affinity configuration",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"location": schema.StringAttribute{
								MarkdownDescription: "Location of the session affinity value (header or body)",
								Required:            true,
							},
							"key": schema.StringAttribute{
								MarkdownDescription: "Key for the session affinity",
								Required:            true,
							},
							"value": schema.StringAttribute{
								MarkdownDescription: "Value for the session affinity (e.g., 'actor_id')",
								Required:            true,
							},
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

	// Create source using the Tama client
	createRequest := sensory.CreateSourceRequest{
		Source: sensory.SourceRequestData{
			Name:     data.Name.ValueString(),
			Type:     data.Type.ValueString(),
			Endpoint: data.Endpoint.ValueString(),
			Credential: sensory.SourceCredential{
				APIKey: data.ApiKey.ValueString(),
			},
		},
	}

	// Add request configuration if provided
	if data.Request != nil {
		requestData := &sensory.Request{}

		// Add headers if provided
		if len(data.Request.Headers) > 0 {
			requestData.Headers = make([]sensory.Header, len(data.Request.Headers))
			for i, h := range data.Request.Headers {
				requestData.Headers[i] = sensory.Header{
					Name:  h.Name.ValueString(),
					Value: h.Value.ValueString(),
				}
			}
		}

		// Add session affinity if provided
		if data.Request.SessionAffinity != nil {
			requestData.SessionAffinity = &sensory.SessionAffinity{
				Location: data.Request.SessionAffinity.Location.ValueString(),
				Key:      data.Request.SessionAffinity.Key.ValueString(),
				Value:    data.Request.SessionAffinity.Value.ValueString(),
			}
		}

		createRequest.Source.Request = requestData
	}

	tflog.Debug(ctx, "Creating source", map[string]any{
		"space_id": data.SpaceId.ValueString(),
		"name":     data.Name.ValueString(),
		"type":     data.Type.ValueString(),
		"endpoint": data.Endpoint.ValueString(),
	})

	sourceResponse, err := r.client.Sensory.CreateSource(data.SpaceId.ValueString(), createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create source, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.Id = types.StringValue(sourceResponse.ID)
	data.Name = types.StringValue(sourceResponse.Name)
	data.Slug = types.StringValue(sourceResponse.Slug)
	data.Type = types.StringValue(sourceResponse.Type)
	data.SpaceId = types.StringValue(sourceResponse.SpaceID)
	data.ProvisionState = types.StringValue(sourceResponse.ProvisionState)
	data.Endpoint = types.StringValue(sourceResponse.Endpoint)
	// Note: API key is not returned in response, keep the original value

	// Populate request data from response if available
	if sourceResponse.Request != nil {
		data.Request = &RequestModel{}

		// Populate headers
		if len(sourceResponse.Request.Headers) > 0 {
			data.Request.Headers = make([]HeaderModel, len(sourceResponse.Request.Headers))
			for i, h := range sourceResponse.Request.Headers {
				data.Request.Headers[i] = HeaderModel{
					Name:  types.StringValue(h.Name),
					Value: types.StringValue(h.Value),
				}
			}
		}

		// Populate session affinity
		if sourceResponse.Request.SessionAffinity != nil {
			data.Request.SessionAffinity = &SessionAffinityModel{
				Location: types.StringValue(sourceResponse.Request.SessionAffinity.Location),
				Key:      types.StringValue(sourceResponse.Request.SessionAffinity.Key),
				Value:    types.StringValue(sourceResponse.Request.SessionAffinity.Value),
			}
		}
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a source resource")

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

	// Get source from API
	sourceResponse, err := r.client.Sensory.GetSource(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read source, got error: %s", err))
		return
	}

	// Update the model with the latest data
	data.Name = types.StringValue(sourceResponse.Name)
	data.Slug = types.StringValue(sourceResponse.Slug)
	data.Type = types.StringValue(sourceResponse.Type)
	data.SpaceId = types.StringValue(sourceResponse.SpaceID)
	data.ProvisionState = types.StringValue(sourceResponse.ProvisionState)
	data.Endpoint = types.StringValue(sourceResponse.Endpoint)
	// Note: API key is not returned in response, keep the original value

	// Populate request data from response if available
	if sourceResponse.Request != nil {
		data.Request = &RequestModel{}

		// Populate headers
		if len(sourceResponse.Request.Headers) > 0 {
			data.Request.Headers = make([]HeaderModel, len(sourceResponse.Request.Headers))
			for i, h := range sourceResponse.Request.Headers {
				data.Request.Headers[i] = HeaderModel{
					Name:  types.StringValue(h.Name),
					Value: types.StringValue(h.Value),
				}
			}
		}

		// Populate session affinity
		if sourceResponse.Request.SessionAffinity != nil {
			data.Request.SessionAffinity = &SessionAffinityModel{
				Location: types.StringValue(sourceResponse.Request.SessionAffinity.Location),
				Key:      types.StringValue(sourceResponse.Request.SessionAffinity.Key),
				Value:    types.StringValue(sourceResponse.Request.SessionAffinity.Value),
			}
		}
	} else {
		// Clear request data if not present in response
		data.Request = nil
	}

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

	// Update source using the Tama client
	updateRequest := sensory.UpdateSourceRequest{
		Source: sensory.UpdateSourceData{
			Name:     data.Name.ValueString(),
			Type:     data.Type.ValueString(),
			Endpoint: data.Endpoint.ValueString(),
			Credential: &sensory.SourceCredential{
				APIKey: data.ApiKey.ValueString(),
			},
		},
	}

	// Add request configuration if provided
	if data.Request != nil {
		requestData := &sensory.Request{}

		// Add headers if provided
		if len(data.Request.Headers) > 0 {
			requestData.Headers = make([]sensory.Header, len(data.Request.Headers))
			for i, h := range data.Request.Headers {
				requestData.Headers[i] = sensory.Header{
					Name:  h.Name.ValueString(),
					Value: h.Value.ValueString(),
				}
			}
		}

		// Add session affinity if provided
		if data.Request.SessionAffinity != nil {
			requestData.SessionAffinity = &sensory.SessionAffinity{
				Location: data.Request.SessionAffinity.Location.ValueString(),
				Key:      data.Request.SessionAffinity.Key.ValueString(),
				Value:    data.Request.SessionAffinity.Value.ValueString(),
			}
		}

		updateRequest.Source.Request = requestData
	}

	tflog.Debug(ctx, "Updating source", map[string]any{
		"id":       data.Id.ValueString(),
		"name":     data.Name.ValueString(),
		"type":     data.Type.ValueString(),
		"endpoint": data.Endpoint.ValueString(),
	})

	sourceResponse, err := r.client.Sensory.UpdateSource(data.Id.ValueString(), updateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update source, got error: %s", err))
		return
	}

	// Update the model with the response data
	data.Name = types.StringValue(sourceResponse.Name)
	data.Slug = types.StringValue(sourceResponse.Slug)
	data.Type = types.StringValue(sourceResponse.Type)
	data.SpaceId = types.StringValue(sourceResponse.SpaceID)
	data.ProvisionState = types.StringValue(sourceResponse.ProvisionState)
	data.Endpoint = types.StringValue(sourceResponse.Endpoint)
	// Note: API key is not returned in response, keep the original value

	// Populate request data from response if available
	if sourceResponse.Request != nil {
		data.Request = &RequestModel{}

		// Populate headers
		if len(sourceResponse.Request.Headers) > 0 {
			data.Request.Headers = make([]HeaderModel, len(sourceResponse.Request.Headers))
			for i, h := range sourceResponse.Request.Headers {
				data.Request.Headers[i] = HeaderModel{
					Name:  types.StringValue(h.Name),
					Value: types.StringValue(h.Value),
				}
			}
		}

		// Populate session affinity
		if sourceResponse.Request.SessionAffinity != nil {
			data.Request.SessionAffinity = &SessionAffinityModel{
				Location: types.StringValue(sourceResponse.Request.SessionAffinity.Location),
				Key:      types.StringValue(sourceResponse.Request.SessionAffinity.Key),
				Value:    types.StringValue(sourceResponse.Request.SessionAffinity.Value),
			}
		}
	} else {
		// Clear request data if not present in response
		data.Request = nil
	}

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

	// Delete source using the Tama client
	tflog.Debug(ctx, "Deleting source", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Sensory.DeleteSource(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete source, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get source from API to populate state
	sourceResponse, err := r.client.Sensory.GetSource(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import source, got error: %s", err))
		return
	}

	// Create model from API response
	data := ResourceModel{
		Id:             types.StringValue(sourceResponse.ID),
		Name:           types.StringValue(sourceResponse.Name),
		Slug:           types.StringValue(sourceResponse.Slug),
		Type:           types.StringValue(sourceResponse.Type),
		SpaceId:        types.StringValue(sourceResponse.SpaceID),
		ProvisionState: types.StringValue(sourceResponse.ProvisionState),
		Endpoint:       types.StringValue(sourceResponse.Endpoint),
		// ApiKey cannot be retrieved from API response
		// This will need to be manually set after import
		ApiKey: types.StringValue(""),
	}

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
