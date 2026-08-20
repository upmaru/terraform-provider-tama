// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package modifier

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	tama "github.com/upmaru/tama-go"
	"github.com/upmaru/tama-go/tools"
)

var _ resource.Resource = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *tama.Client
}

type SourceModel struct {
	Type types.String `tfsdk:"type"`
	Path types.String `tfsdk:"path"`
}

type ResourceModel struct {
	Id              types.String `tfsdk:"id"`
	ThoughtToolId   types.String `tfsdk:"thought_tool_id"`
	Index           types.Int64  `tfsdk:"index"`
	Target          types.String `tfsdk:"target"`
	OnMissingParent types.String `tfsdk:"on_missing_parent"`
	OnMissingSource types.String `tfsdk:"on_missing_source"`
	Source          *SourceModel `tfsdk:"source"`
	ProvisionState  types.String `tfsdk:"provision_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_thought_tool_modifier"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a trusted Tama Thought Tool Modifier resource. The `source` selects trusted runtime metadata; it is not a secret value stored in Terraform. `on_missing_parent = \"skip\"` leaves calls unchanged when the target parent is absent, while `on_missing_source = \"error\"` fails closed when required metadata is unavailable. Changing `thought_tool_id` or `index` replaces the resource. Destroy deactivates the modifier, and exact recreation can reuse its ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Modifier identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"thought_tool_id": schema.StringAttribute{
				MarkdownDescription: "ID of the thought tool this modifier belongs to.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"index": schema.Int64Attribute{
				MarkdownDescription: "Modifier execution index. Valid values are 0 through 2147483647.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(0, math.MaxInt32),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"target": schema.StringAttribute{
				MarkdownDescription: "JSON Pointer target in the thought tool's effective arguments.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"on_missing_parent": schema.StringAttribute{
				MarkdownDescription: "Behavior when the target's parent is missing. Must be `error` or `skip`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(tools.ModifierMissingPolicyError, tools.ModifierMissingPolicySkip),
				},
			},
			"on_missing_source": schema.StringAttribute{
				MarkdownDescription: "Behavior when trusted source metadata is missing. Must be `error` or `skip`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(tools.ModifierMissingPolicyError, tools.ModifierMissingPolicySkip),
				},
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current provision state of the modifier.",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"source": schema.SingleNestedBlock{
				MarkdownDescription: "Trusted metadata source to inject at the target.",
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Trusted source type. Version 1 supports only `metadata`.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf(tools.ModifierSourceTypeMetadata),
						},
					},
					"path": schema.StringAttribute{
						MarkdownDescription: "Allowlisted metadata path to inject.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								tools.ModifierSourcePathActorIdentifier,
								tools.ModifierSourcePathOriginEntityIdentifier,
								tools.ModifierSourcePathCurrentTimestamp,
							),
						},
					},
				},
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
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Source == nil {
		resp.Diagnostics.AddError("Missing Modifier Source", "The source block must be configured.")
		return
	}

	index := data.Index.ValueInt64()
	if index < 0 || index > math.MaxInt32 {
		resp.Diagnostics.AddError("Invalid Modifier Index", "The index must be between 0 and 2147483647.")
		return
	}

	createReq := tools.CreateModifierRequest{
		Modifier: tools.ModifierRequestData{
			Index:           int(index),
			Target:          data.Target.ValueString(),
			OnMissingParent: data.OnMissingParent.ValueString(),
			OnMissingSource: data.OnMissingSource.ValueString(),
			Source: tools.ModifierSource{
				Type: data.Source.Type.ValueString(),
				Path: data.Source.Path.ValueString(),
			},
		},
	}

	tflog.Debug(ctx, "Creating thought tool modifier", map[string]any{
		"thought_tool_id": data.ThoughtToolId.ValueString(),
		"index":           createReq.Modifier.Index,
		"target":          createReq.Modifier.Target,
		"source_type":     createReq.Modifier.Source.Type,
		"source_path":     createReq.Modifier.Source.Path,
	})

	modifier, err := r.client.Tools.CreateModifier(data.ThoughtToolId.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create thought tool modifier, got error: %s", err))
		return
	}

	data = modelFromModifier(modifier)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading thought tool modifier", map[string]any{"id": data.Id.ValueString()})

	modifier, err := r.client.Tools.GetModifier(data.Id.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read thought tool modifier, got error: %s", err))
		return
	}

	data = modelFromModifier(modifier)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Source == nil {
		resp.Diagnostics.AddError("Missing Modifier Source", "The source block must be configured.")
		return
	}

	source := tools.ModifierSource{
		Type: data.Source.Type.ValueString(),
		Path: data.Source.Path.ValueString(),
	}
	updateReq := tools.UpdateModifierRequest{
		Modifier: tools.UpdateModifierData{
			Target:          data.Target.ValueString(),
			OnMissingParent: data.OnMissingParent.ValueString(),
			OnMissingSource: data.OnMissingSource.ValueString(),
			Source:          &source,
		},
	}

	tflog.Debug(ctx, "Updating thought tool modifier", map[string]any{
		"id":          data.Id.ValueString(),
		"target":      updateReq.Modifier.Target,
		"source_type": source.Type,
		"source_path": source.Path,
	})

	modifier, err := r.client.Tools.UpdateModifier(data.Id.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update thought tool modifier, got error: %s", err))
		return
	}

	data = modelFromModifier(modifier)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting thought tool modifier", map[string]any{"id": data.Id.ValueString()})

	if err := r.client.Tools.DeleteModifier(data.Id.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete thought tool modifier, got error: %s", err))
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing thought tool modifier", map[string]any{"id": req.ID})

	modifier, err := r.client.Tools.GetModifier(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read thought tool modifier for import, got error: %s", err))
		return
	}

	data := modelFromModifier(modifier)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func modelFromModifier(modifier *tools.Modifier) ResourceModel {
	return ResourceModel{
		Id:              types.StringValue(modifier.ID),
		ThoughtToolId:   types.StringValue(modifier.ThoughtToolID),
		Index:           types.Int64Value(int64(modifier.Index)),
		Target:          types.StringValue(modifier.Target),
		OnMissingParent: types.StringValue(modifier.OnMissingParent),
		OnMissingSource: types.StringValue(modifier.OnMissingSource),
		Source: &SourceModel{
			Type: types.StringValue(modifier.Source.Type),
			Path: types.StringValue(modifier.Source.Path),
		},
		ProvisionState: types.StringValue(modifier.ProvisionState),
	}
}

func isNotFound(err error) bool {
	var apiErr *tools.Error
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}
