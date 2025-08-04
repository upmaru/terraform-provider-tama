// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package input

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
	"github.com/upmaru/tama-go/perception/module"
)

var _ resource.Resource = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *tama.Client
}

type ResourceModel struct {
	Id              types.String `tfsdk:"id"`
	ThoughtId       types.String `tfsdk:"thought_id"`
	ThoughtModuleId types.String `tfsdk:"thought_module_id"`
	Type            types.String `tfsdk:"type"`
	ClassCorpusId   types.String `tfsdk:"class_corpus_id"`
	ProvisionState  types.String `tfsdk:"provision_state"`
}

func validateType(t string) error {
	switch t {
	case "concept", "entity":
		return nil
	default:
		return fmt.Errorf("unsupported type: %s", t)
	}
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_thought_module_input"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Thought Module Input resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Input identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"thought_id": schema.StringAttribute{
				MarkdownDescription: "ID of the thought this input belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"thought_module_id": schema.StringAttribute{
				MarkdownDescription: "ID of the thought module (computed from API)",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of input, concept or entity",
				Required:            true,
			},
			"class_corpus_id": schema.StringAttribute{
				MarkdownDescription: "Class corpus ID related to thought",
				Required:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the input",
				Computed:            true,
			},
		},
	}
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	if err := validateType(data.Type.ValueString()); err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
		return
	}

	createReq := module.CreateInputRequest{
		Input: module.CreateInputData{
			Type:          data.Type.ValueString(),
			ClassCorpusID: data.ClassCorpusId.ValueString(),
		},
	}

	tflog.Debug(ctx, "Creating input", map[string]any{
		"thought_id": data.ThoughtId.ValueString(),
		"type":       createReq.Input.Type,
	})

	inputResponse, err := r.client.Perception.Module.CreateInput(data.ThoughtId.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create input, got error: %s", err))
		return
	}

	// Map creation result to the data model
	data.Id = types.StringValue(inputResponse.ID)
	data.ThoughtModuleId = types.StringValue(inputResponse.ThoughtModuleID)
	data.ProvisionState = types.StringValue(inputResponse.ProvisionState)

	// Write logs
	tflog.Trace(ctx, "created a module input resource")

	// Save successful data
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save into Terraform state
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...) // Read state

	inputResponse, err := r.client.Perception.Module.GetInput(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read input, got error: %s", err))
		return
	}

	// Refresh the data
	data.Id = types.StringValue(inputResponse.ID)
	data.ThoughtId = types.StringValue(inputResponse.ThoughtID)
	data.ThoughtModuleId = types.StringValue(inputResponse.ThoughtModuleID)
	data.Type = types.StringValue(inputResponse.Type)
	data.ClassCorpusId = types.StringValue(inputResponse.ClassCorpusID)
	data.ProvisionState = types.StringValue(inputResponse.ProvisionState)

	// Save updated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save into Terraform state
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...) // Read plan

	if err := validateType(data.Type.ValueString()); err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
		return
	}

	updateRequest := module.UpdateInputRequest{
		Input: module.UpdateInputData{
			Type:          data.Type.ValueString(),
			ClassCorpusID: data.ClassCorpusId.ValueString(),
		},
	}

	tflog.Debug(ctx, "Updating input", map[string]any{
		"id":   data.Id.ValueString(),
		"type": updateRequest.Input.Type,
	})

	inputResponse, err := r.client.Perception.Module.UpdateInput(data.Id.ValueString(), updateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update input, got error: %s", err))
		return
	}

	// Update state
	data.Id = types.StringValue(inputResponse.ID)
	data.ThoughtId = types.StringValue(inputResponse.ThoughtID)
	data.ThoughtModuleId = types.StringValue(inputResponse.ThoughtModuleID)
	data.Type = types.StringValue(inputResponse.Type)
	data.ClassCorpusId = types.StringValue(inputResponse.ClassCorpusID)
	data.ProvisionState = types.StringValue(inputResponse.ProvisionState)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...) // Read state

	err := r.client.Perception.Module.DeleteInput(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete input, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "Deleted input", map[string]any{
		"id": data.Id.ValueString(),
	})
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	inputResponse, err := r.client.Perception.Module.GetInput(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read input for import, got error: %s", err))
		return
	}

	var data ResourceModel
	data.Id = types.StringValue(inputResponse.ID)
	data.ThoughtId = types.StringValue(inputResponse.ThoughtID)
	data.ThoughtModuleId = types.StringValue(inputResponse.ThoughtModuleID)
	data.Type = types.StringValue(inputResponse.Type)
	data.ClassCorpusId = types.StringValue(inputResponse.ClassCorpusID)
	data.ProvisionState = types.StringValue(inputResponse.ProvisionState)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save
}
