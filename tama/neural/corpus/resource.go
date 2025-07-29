// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package corpus

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

// ResourceModel describes the resource data model.
type ResourceModel struct {
	Id             types.String `tfsdk:"id"`
	ClassId        types.String `tfsdk:"class_id"`
	Name           types.String `tfsdk:"name"`
	Slug           types.String `tfsdk:"slug"`
	Main           types.Bool   `tfsdk:"main"`
	Template       types.String `tfsdk:"template"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_class_corpus"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Neural Class Corpus resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Corpus identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"class_id": schema.StringAttribute{
				MarkdownDescription: "ID of the class this corpus belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the corpus",
				Required:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Slug of the corpus",
				Computed:            true,
			},
			"main": schema.BoolAttribute{
				MarkdownDescription: "Whether this is the main corpus",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"template": schema.StringAttribute{
				MarkdownDescription: "Template for the corpus",
				Required:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the corpus",
				Computed:            true,
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

	// Create corpus using the Tama client
	createRequest := neural.CreateCorpusRequest{
		Corpus: neural.CorpusRequestData{
			Main:     data.Main.ValueBool(),
			Name:     data.Name.ValueString(),
			Template: data.Template.ValueString(),
		},
	}

	tflog.Debug(ctx, "Creating corpus", map[string]any{
		"class_id": data.ClassId.ValueString(),
		"name":     data.Name.ValueString(),
		"main":     data.Main.ValueBool(),
		"template": data.Template.ValueString(),
	})

	corpusResponse, err := r.client.Neural.CreateCorpus(data.ClassId.ValueString(), createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create corpus, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.Id = types.StringValue(corpusResponse.ID)
	data.Name = types.StringValue(corpusResponse.Name)
	data.Slug = types.StringValue(corpusResponse.Slug)
	data.Main = types.BoolValue(corpusResponse.Main)
	data.Template = types.StringValue(corpusResponse.Template)
	data.ProvisionState = types.StringValue(corpusResponse.ProvisionState)

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a corpus resource")

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

	// Get corpus from API
	corpusResponse, err := r.client.Neural.GetCorpus(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read corpus, got error: %s", err))
		return
	}

	// Update the model with the latest data
	data.Id = types.StringValue(corpusResponse.ID)
	data.Name = types.StringValue(corpusResponse.Name)
	data.Slug = types.StringValue(corpusResponse.Slug)
	data.Main = types.BoolValue(corpusResponse.Main)
	data.Template = types.StringValue(corpusResponse.Template)
	data.ProvisionState = types.StringValue(corpusResponse.ProvisionState)

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

	// Update corpus using the Tama client
	mainValue := data.Main.ValueBool()
	updateRequest := neural.UpdateCorpusRequest{
		Corpus: neural.UpdateCorpusData{
			Main:     &mainValue,
			Name:     data.Name.ValueString(),
			Template: data.Template.ValueString(),
		},
	}

	tflog.Debug(ctx, "Updating corpus", map[string]any{
		"id":       data.Id.ValueString(),
		"name":     data.Name.ValueString(),
		"main":     data.Main.ValueBool(),
		"template": data.Template.ValueString(),
	})

	corpusResponse, err := r.client.Neural.UpdateCorpus(data.Id.ValueString(), updateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update corpus, got error: %s", err))
		return
	}

	// Update the model with the response data
	data.Id = types.StringValue(corpusResponse.ID)
	data.Name = types.StringValue(corpusResponse.Name)
	data.Slug = types.StringValue(corpusResponse.Slug)
	data.Main = types.BoolValue(corpusResponse.Main)
	data.Template = types.StringValue(corpusResponse.Template)
	data.ProvisionState = types.StringValue(corpusResponse.ProvisionState)

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

	// Delete corpus using the Tama client
	tflog.Debug(ctx, "Deleting corpus", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Neural.DeleteCorpus(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete corpus, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get corpus from API to populate state
	corpusResponse, err := r.client.Neural.GetCorpus(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import corpus, got error: %s", err))
		return
	}

	// Create model from API response
	data := ResourceModel{
		Id:             types.StringValue(corpusResponse.ID),
		Name:           types.StringValue(corpusResponse.Name),
		Slug:           types.StringValue(corpusResponse.Slug),
		Main:           types.BoolValue(corpusResponse.Main),
		Template:       types.StringValue(corpusResponse.Template),
		ProvisionState: types.StringValue(corpusResponse.ProvisionState),
	}

	// Note: ClassId cannot be retrieved from corpus response, so we leave it null for import
	// Users will need to run terraform plan to see the required class_id

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
