// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package operation

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	tama "github.com/upmaru/tama-go"
	"github.com/upmaru/tama-go/neural/class"
	"github.com/upmaru/terraform-provider-tama/internal/wait"
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
	Id           types.String   `tfsdk:"id"`
	ClassId      types.String   `tfsdk:"class_id"`
	ChainIds     types.List     `tfsdk:"chain_ids"`
	NodeType     types.String   `tfsdk:"node_type"`
	CurrentState types.String   `tfsdk:"current_state"`
	NodeIds      types.List     `tfsdk:"node_ids"`
	WaitFor      []wait.WaitFor `tfsdk:"wait_for"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_class_operation"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Neural Class Operation resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Operation identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"class_id": schema.StringAttribute{
				MarkdownDescription: "ID of the class this operation belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"chain_ids": schema.ListAttribute{
				MarkdownDescription: "List of chain IDs for this operation",
				ElementType:         types.StringType,
				Required:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"node_type": schema.StringAttribute{
				MarkdownDescription: "Type of node (explicit or reactive). Defaults to 'reactive'",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("reactive"),
				Validators: []validator.String{
					stringvalidator.OneOf("explicit", "reactive"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"current_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the operation",
				Computed:            true,
			},
			"node_ids": schema.ListAttribute{
				MarkdownDescription: "List of node IDs created by this operation",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
		Blocks: wait.WaitForBlockSchema(),
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

	tflog.Trace(ctx, "creating class operation resource")

	// Convert chain_ids from types.List to []string
	var chainIds []string
	resp.Diagnostics.Append(data.ChainIds.ElementsAs(ctx, &chainIds, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating operation with chain IDs", map[string]any{
		"class_id":  data.ClassId.ValueString(),
		"chain_ids": chainIds,
	})

	// Prepare the create request
	createReq := class.CreateOperationRequest{
		Operation: class.CreateOperationData{
			ChainIDs: chainIds,
		},
	}

	// Set node type if provided
	if !data.NodeType.IsNull() && !data.NodeType.IsUnknown() {
		nodeType := data.NodeType.ValueString()
		createReq.Operation.NodeType = &nodeType
		tflog.Debug(ctx, "Setting node type", map[string]any{
			"node_type": nodeType,
		})
	}

	// Create the operation
	classOperationService := class.NewService(r.client.GetHTTPClient())
	tflog.Debug(ctx, "Calling CreateOperation API", map[string]any{
		"class_id": data.ClassId.ValueString(),
		"request":  createReq,
	})
	operation, err := classOperationService.CreateOperation(data.ClassId.ValueString(), createReq)
	if err != nil {
		tflog.Error(ctx, "Failed to create operation", map[string]any{
			"error":    err.Error(),
			"class_id": data.ClassId.ValueString(),
			"request":  createReq,
		})
		resp.Diagnostics.AddError(
			"Error creating class operation",
			fmt.Sprintf("Could not create class operation: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Operation created successfully", map[string]any{
		"operation_id":  operation.ID,
		"current_state": operation.CurrentState,
		"class_id":      operation.ClassID,
		"node_ids":      operation.NodeIDs,
	})

	// Update the model with the response data
	data.Id = types.StringValue(operation.ID)
	data.CurrentState = types.StringValue(operation.CurrentState)

	// Convert node IDs to types.List
	nodeIds := make([]types.String, len(operation.NodeIDs))
	for i, nodeId := range operation.NodeIDs {
		nodeIds[i] = types.StringValue(nodeId)
	}
	nodeIdsList, diags := types.ListValueFrom(ctx, types.StringType, nodeIds)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.NodeIds = nodeIdsList

	// Set default node_type if it wasn't provided
	if data.NodeType.IsNull() || data.NodeType.IsUnknown() {
		data.NodeType = types.StringValue("reactive")
	}

	// Handle wait_for conditions if specified
	if len(data.WaitFor) > 0 {
		getOperationFunc := func(id string) (interface{}, error) {
			return classOperationService.GetOperation(data.ClassId.ValueString(), id)
		}
		for _, waitFor := range data.WaitFor {
			err := wait.ForConditions(ctx, getOperationFunc, data.Id.ValueString(), waitFor.Field, 10*time.Minute)
			if err != nil {
				resp.Diagnostics.AddError("Wait Condition Failed", fmt.Sprintf("Unable to satisfy wait conditions: %s", err))
				return
			}
		}
	}

	tflog.Trace(ctx, "created class operation resource")

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

	tflog.Trace(ctx, "reading class operation resource")

	// Get the operation from the API
	classOperationService := class.NewService(r.client.GetHTTPClient())
	operation, err := classOperationService.GetOperation(data.ClassId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading class operation",
			fmt.Sprintf("Could not read class operation %s: %s", data.Id.ValueString(), err),
		)
		return
	}

	// Update the model with the response data
	data.CurrentState = types.StringValue(operation.CurrentState)

	// Convert node IDs to types.List
	nodeIds := make([]types.String, len(operation.NodeIDs))
	for i, nodeId := range operation.NodeIDs {
		nodeIds[i] = types.StringValue(nodeId)
	}
	nodeIdsList, diags := types.ListValueFrom(ctx, types.StringType, nodeIds)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.NodeIds = nodeIdsList

	tflog.Trace(ctx, "read class operation resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Operations are immutable, so Update should not be called
	resp.Diagnostics.AddError(
		"Update not supported",
		"Class operations are immutable and cannot be updated. Please destroy and recreate the resource.",
	)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Operations are immutable and cannot be deleted through the API
	// We just remove it from the state
	tflog.Trace(ctx, "deleting class operation resource (removing from state only)")
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: class_id:operation_id
	classId, operationId, err := parseImportId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing import ID",
			fmt.Sprintf("Could not parse import ID %s: %s. Expected format: class_id:operation_id", req.ID, err),
		)
		return
	}

	// Set the basic IDs and let the subsequent Read operation fill in the computed fields
	data := ResourceModel{
		Id:           types.StringValue(operationId),
		ClassId:      types.StringValue(classId),
		ChainIds:     types.ListNull(types.StringType), // Will be ignored during verification
		NodeType:     types.StringValue("reactive"),    // Default, will be ignored during verification
		NodeIds:      types.ListNull(types.StringType), // Initialize as null, Read will populate
		CurrentState: types.StringNull(),               // Initialize as null, Read will populate
	}

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func parseImportId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected format class_id:operation_id, got %d parts", len(parts))
	}

	return parts[0], parts[1], nil
}
