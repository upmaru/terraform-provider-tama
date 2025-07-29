// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wait

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/thedevsaddam/gojsonq/v2"
)

// WaitForField represents a field condition for waiting.
type WaitForField struct {
	Name types.String `tfsdk:"name"`
	In   types.List   `tfsdk:"in"`
}

// WaitFor represents the wait_for configuration.
type WaitFor struct {
	Field []WaitForField `tfsdk:"field"`
}

// WaitForBlockSchema returns the common schema block for wait_for functionality.
func WaitForBlockSchema() map[string]schema.Block {
	return map[string]schema.Block{
		"wait_for": schema.ListNestedBlock{
			MarkdownDescription: "If set, will wait until either all of conditions are satisfied, or until timeout is reached",
			NestedObject: schema.NestedBlockObject{
				Blocks: map[string]schema.Block{
					"field": schema.ListNestedBlock{
						MarkdownDescription: "Condition criteria for a field",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "Name of the field to check (JSON path)",
									Required:            true,
								},
								"in": schema.ListAttribute{
									MarkdownDescription: "List of acceptable values for the field",
									Required:            true,
									ElementType:         types.StringType,
								},
							},
						},
					},
				},
			},
		},
	}
}

// ForConditions waits for specified field conditions to be met on a resource.
// This is a generic function that can be used by any resource that needs wait functionality.
func ForConditions(ctx context.Context, getResourceFunc func(string) (any, error), resourceId string, conditions []WaitForField, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for conditions")
		case <-ticker.C:
			// Get current resource state
			resource, err := getResourceFunc(resourceId)
			if err != nil {
				return fmt.Errorf("failed to get resource: %s", err)
			}

			// Convert to JSON for querying
			jsonBytes, err := json.Marshal(resource)
			if err != nil {
				return fmt.Errorf("failed to marshal resource to JSON: %s", err)
			}

			// Check all conditions
			allConditionsMet := true
			gq := gojsonq.New().FromString(string(jsonBytes))

			for _, condition := range conditions {
				// Find the value at the specified field name
				value := gq.Reset().Find(condition.Name.ValueString())
				if value == nil {
					allConditionsMet = false
					break
				}

				// Convert to string for comparison
				stringVal := fmt.Sprintf("%v", value)

				// Get the list of acceptable values
				var acceptableValues []string
				diags := condition.In.ElementsAs(ctx, &acceptableValues, false)
				if diags.HasError() {
					return fmt.Errorf("failed to parse acceptable values for field '%s'", condition.Name.ValueString())
				}

				// Check if the current value is in the list of acceptable values
				if !slices.Contains(acceptableValues, stringVal) {
					allConditionsMet = false
					break
				}
			}

			if allConditionsMet {
				return nil
			}
		}
	}
}
