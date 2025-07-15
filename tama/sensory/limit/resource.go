// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package limit

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
	Id         types.String `tfsdk:"id"`
	SourceId   types.String `tfsdk:"source_id"`
	ScaleUnit  types.String `tfsdk:"scale_unit"`
	ScaleCount types.Int64  `tfsdk:"scale_count"`
	Limit      types.Int64  `tfsdk:"limit"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_limit"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Sensory Limit resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Limit identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_id": schema.StringAttribute{
				MarkdownDescription: "ID of the source this limit belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scale_unit": schema.StringAttribute{
				MarkdownDescription: "Unit for the scaling period (e.g., 'seconds', 'minutes', 'hours')",
				Required:            true,
			},
			"scale_count": schema.Int64Attribute{
				MarkdownDescription: "Number of scale units for the limit period",
				Required:            true,
			},
			"limit": schema.Int64Attribute{
				MarkdownDescription: "The limit value for the specified period",
				Required:            true,
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

	// Create limit using the Tama client
	createRequest := sensory.CreateLimitRequest{
		Limit: sensory.LimitRequestData{
			ScaleUnit:  data.ScaleUnit.ValueString(),
			ScaleCount: int(data.ScaleCount.ValueInt64()),
			Limit:      int(data.Limit.ValueInt64()),
		},
	}

	tflog.Debug(ctx, "Creating limit", map[string]interface{}{
		"source_id":   data.SourceId.ValueString(),
		"scale_unit":  data.ScaleUnit.ValueString(),
		"scale_count": data.ScaleCount.ValueInt64(),
		"limit":       data.Limit.ValueInt64(),
	})

	limitResponse, err := r.client.Sensory.CreateLimit(data.SourceId.ValueString(), createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create limit, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.Id = types.StringValue(limitResponse.ID)
	data.ScaleUnit = types.StringValue(limitResponse.ScaleUnit)
	data.ScaleCount = types.Int64Value(int64(limitResponse.ScaleCount))
	data.Limit = types.Int64Value(int64(limitResponse.Limit))

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a limit resource")

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

	// Get limit from API
	limitResponse, err := r.client.Sensory.GetLimit(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read limit, got error: %s", err))
		return
	}

	// Update the model with the latest data
	data.ScaleUnit = types.StringValue(limitResponse.ScaleUnit)
	data.ScaleCount = types.Int64Value(int64(limitResponse.ScaleCount))
	data.Limit = types.Int64Value(int64(limitResponse.Limit))

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

	// Update limit using the Tama client
	updateRequest := sensory.UpdateLimitRequest{
		Limit: sensory.UpdateLimitData{
			ScaleUnit:  data.ScaleUnit.ValueString(),
			ScaleCount: int(data.ScaleCount.ValueInt64()),
			Limit:      int(data.Limit.ValueInt64()),
		},
	}

	tflog.Debug(ctx, "Updating limit", map[string]interface{}{
		"id":          data.Id.ValueString(),
		"scale_unit":  data.ScaleUnit.ValueString(),
		"scale_count": data.ScaleCount.ValueInt64(),
		"limit":       data.Limit.ValueInt64(),
	})

	limitResponse, err := r.client.Sensory.UpdateLimit(data.Id.ValueString(), updateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update limit, got error: %s", err))
		return
	}

	// Update the model with the response data
	data.ScaleUnit = types.StringValue(limitResponse.ScaleUnit)
	data.ScaleCount = types.Int64Value(int64(limitResponse.ScaleCount))
	data.Limit = types.Int64Value(int64(limitResponse.Limit))

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

	// Delete limit using the Tama client
	tflog.Debug(ctx, "Deleting limit", map[string]interface{}{
		"id": data.Id.ValueString(),
	})

	err := r.client.Sensory.DeleteLimit(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete limit, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get limit from API to populate state
	limitResponse, err := r.client.Sensory.GetLimit(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import limit, got error: %s", err))
		return
	}

	// Create model from API response
	data := ResourceModel{
		Id:         types.StringValue(limitResponse.ID),
		ScaleUnit:  types.StringValue(limitResponse.ScaleUnit),
		ScaleCount: types.Int64Value(int64(limitResponse.ScaleCount)),
		Limit:      types.Int64Value(int64(limitResponse.Limit)),
		// SourceId cannot be retrieved from API response
		// This will need to be manually set after import
		SourceId: types.StringValue(""),
	}

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
