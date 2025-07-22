// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package planmodifier

import (
	"bytes"
	"context"
	"encoding/json"
	"sort"

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

// NormalizeJSON normalizes JSON by sorting keys recursively
func NormalizeJSON(jsonStr string) (string, error) {
	if jsonStr == "" {
		return "", nil
	}

	var obj any
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return "", err
	}

	normalized := normalizeValue(obj)

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false) // Optional: prevents escaping of <, >, &
	encoder.SetIndent("", "")    // No indentation for compact output

	if err := encoder.Encode(normalized); err != nil {
		return "", err
	}

	// Remove the trailing newline that json.Encoder adds
	result := buf.String()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	return result, nil
}

// normalizeValue recursively processes values to ensure consistent ordering
func normalizeValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		// Create a new map and process keys in sorted order
		normalized := make(map[string]any)

		// Get all keys and sort them
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Process each key-value pair recursively
		for _, k := range keys {
			normalized[k] = normalizeValue(val[k])
		}
		return normalized

	case []any:
		// Process array elements recursively
		normalized := make([]any, len(val))
		for i, elem := range val {
			normalized[i] = normalizeValue(elem)
		}
		return normalized

	default:
		// For primitive types (string, number, bool, null), return as-is
		return val
	}
}
