// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package planmodifier

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// JSONNormalize returns a plan modifier that normalizes JSON strings to prevent
// formatting differences from causing unnecessary updates.
//
// This plan modifier compares JSON strings semantically rather than as raw strings.
// If the planned value and state value are semantically equivalent JSON, it will
// suppress the diff and keep the existing state value.
func JSONNormalize() planmodifier.String {
	return jsonNormalizePlanModifier{}
}

// jsonNormalizePlanModifier implements a plan modifier that normalizes JSON strings
// to prevent formatting differences from causing unnecessary updates.
type jsonNormalizePlanModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m jsonNormalizePlanModifier) Description(_ context.Context) string {
	return "Normalizes JSON strings to prevent formatting differences"
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m jsonNormalizePlanModifier) MarkdownDescription(_ context.Context) string {
	return "Normalizes JSON strings to prevent formatting differences"
}

// PlanModifyString implements the plan modification logic for JSON strings.
func (m jsonNormalizePlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If either value is null/unknown, no modification needed
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() ||
		req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	planString := req.PlanValue.ValueString()
	stateString := req.StateValue.ValueString()

	// If strings are identical, no need to normalize
	if planString == stateString {
		return
	}

	// Normalize both JSON strings for comparison
	planJSON, planErr := NormalizeJSON(planString)
	stateJSON, stateErr := NormalizeJSON(stateString)

	// If normalization succeeds and they're semantically equal, keep the state value
	if planErr == nil && stateErr == nil && planJSON == stateJSON {
		resp.PlanValue = req.StateValue
		return
	}

	// If there's an error normalizing the plan value, but the state value is valid,
	// keep the plan value (let other validation catch the error)
	// Otherwise, proceed with the planned value
}

// NormalizeJSON normalizes a JSON string by parsing and re-marshaling it.
// This ensures consistent formatting regardless of input formatting.
func NormalizeJSON(jsonStr string) (string, error) {
	if jsonStr == "" {
		return "", nil
	}

	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return "", err
	}

	normalized, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}

	return string(normalized), nil
}
