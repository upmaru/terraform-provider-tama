package modifier

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	tama "github.com/upmaru/tama-go"
	"github.com/upmaru/tama-go/motor"
	internalplanmodifier "github.com/upmaru/terraform-provider-tama/internal/planmodifier"
)

var _ resource.Resource = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

func NewResource() resource.Resource { return &Resource{} }

type Resource struct{ client *tama.Client }

type ResourceModel struct {
	Id             types.String `tfsdk:"id"`
	ActionId       types.String `tfsdk:"action_id"`
	Name           types.String `tfsdk:"name"`
	Schema         types.String `tfsdk:"schema"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_action_modifier"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Motor Action Modifier resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Modifier identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"action_id": schema.StringAttribute{
				MarkdownDescription: "ID of the action this modifier belongs to",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the modifier",
				Required:            true,
			},
			"schema": schema.StringAttribute{
				MarkdownDescription: "Modifier schema as JSON string",
				Required:            true,
				PlanModifiers:       []planmodifier.String{internalplanmodifier.JSONNormalize()},
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the modifier",
				Computed:            true,
			},
		},
	}
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil { return }
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
	if resp.Diagnostics.HasError() { return }

	// Parse schema JSON
	var schemaMap map[string]any
	if err := json.Unmarshal([]byte(data.Schema.ValueString()), &schemaMap); err != nil {
		resp.Diagnostics.AddError("Invalid Schema", fmt.Sprintf("Unable to parse schema JSON: %s", err))
		return
	}

	createReq := motor.CreateModifierRequest{
		Modifier: motor.ModifierRequestData{
			Name:   data.Name.ValueString(),
			Schema: schemaMap,
		},
	}

	tflog.Debug(ctx, "Creating action modifier", map[string]any{
		"action_id": data.ActionId.ValueString(),
		"name":      createReq.Modifier.Name,
	})

	created, err := r.client.Motor.CreateModifier(data.ActionId.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create modifier, got error: %s", err))
		return
	}

	data.Id = types.StringValue(created.ID)
	data.ActionId = types.StringValue(created.ActionID)
	data.Name = types.StringValue(created.Name)
	data.ProvisionState = types.StringValue(created.ProvisionState)
	// Normalize and set schema from response
	if b, err := json.Marshal(created.Schema); err == nil {
		if normalized, nerr := internalplanmodifier.NormalizeJSON(string(b)); nerr == nil {
			data.Schema = types.StringValue(normalized)
		}
	}

	tflog.Trace(ctx, "created an action modifier resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save state
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...) // Read state
	if resp.Diagnostics.HasError() { return }

	mod, err := r.client.Motor.GetModifier(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read modifier, got error: %s", err))
		return
	}

	data.ActionId = types.StringValue(mod.ActionID)
	data.Name = types.StringValue(mod.Name)
	data.ProvisionState = types.StringValue(mod.ProvisionState)
	// Normalize and set schema from response
	if b, err := json.Marshal(mod.Schema); err == nil {
		if normalized, nerr := internalplanmodifier.NormalizeJSON(string(b)); nerr == nil {
			data.Schema = types.StringValue(normalized)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...) // Read plan
	if resp.Diagnostics.HasError() { return }

	var update motor.UpdateModifierRequest
	update.Modifier = motor.UpdateModifierData{}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		update.Modifier.Name = data.Name.ValueString()
	}
	if !data.Schema.IsNull() && !data.Schema.IsUnknown() && data.Schema.ValueString() != "" {
		var schemaMap map[string]any
		if err := json.Unmarshal([]byte(data.Schema.ValueString()), &schemaMap); err != nil {
			resp.Diagnostics.AddError("Invalid Schema", fmt.Sprintf("Unable to parse schema JSON: %s", err))
			return
		}
		update.Modifier.Schema = schemaMap
	}

	tflog.Debug(ctx, "Updating action modifier", map[string]any{
		"id":   data.Id.ValueString(),
		"name": update.Modifier.Name,
	})

	updated, err := r.client.Motor.UpdateModifier(data.Id.ValueString(), update)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update modifier, got error: %s", err))
		return
	}

	data.Name = types.StringValue(updated.Name)
	data.ActionId = types.StringValue(updated.ActionID)
	data.ProvisionState = types.StringValue(updated.ProvisionState)
	// Normalize and set schema from response
	if b, err := json.Marshal(updated.Schema); err == nil {
		if normalized, nerr := internalplanmodifier.NormalizeJSON(string(b)); nerr == nil {
			data.Schema = types.StringValue(normalized)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...) // Read state
	if resp.Diagnostics.HasError() { return }

	if err := r.client.Motor.DeleteModifier(data.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete modifier, got error: %s", err))
		return
	}
	// No state to set; Terraform will remove resource from state after successful delete
		tflog.Debug(ctx, "Deleted action modifier", map[string]any{"id": data.Id.ValueString()})
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	mod, err := r.client.Motor.GetModifier(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read modifier for import, got error: %s", err))
		return
	}

	var data ResourceModel
	data.Id = types.StringValue(mod.ID)
	data.ActionId = types.StringValue(mod.ActionID)
	data.Name = types.StringValue(mod.Name)
	data.ProvisionState = types.StringValue(mod.ProvisionState)
	if b, err := json.Marshal(mod.Schema); err == nil {
		if normalized, nerr := internalplanmodifier.NormalizeJSON(string(b)); nerr == nil {
			data.Schema = types.StringValue(normalized)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save
}