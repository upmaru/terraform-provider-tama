// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package planmodifier

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestJSONNormalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		planValue         types.String
		stateValue        types.String
		expectSuppression bool
		description       string
	}{
		{
			name:              "identical strings",
			planValue:         types.StringValue(`{"key": "value"}`),
			stateValue:        types.StringValue(`{"key": "value"}`),
			expectSuppression: false,
			description:       "identical JSON strings should not be modified",
		},
		{
			name:              "semantically equal but different formatting",
			planValue:         types.StringValue(`{"key": "value", "number": 123}`),
			stateValue:        types.StringValue(`{"number":123,"key":"value"}`),
			expectSuppression: true,
			description:       "semantically equal JSON with different formatting should be suppressed",
		},
		{
			name: "pretty formatted vs minified",
			planValue: types.StringValue(`{
  "title": "dynamic",
  "description": "A dynamic schema",
  "properties": {
    "entity": {
      "description": "The record",
      "type": "object"
    }
  },
  "type": "object"
}`),
			stateValue:        types.StringValue(`{"description":"A dynamic schema","properties":{"entity":{"description":"The record","type":"object"}},"title":"dynamic","type":"object"}`),
			expectSuppression: true,
			description:       "pretty formatted vs minified JSON should be suppressed",
		},
		{
			name:              "different JSON content",
			planValue:         types.StringValue(`{"key": "value1"}`),
			stateValue:        types.StringValue(`{"key": "value2"}`),
			expectSuppression: false,
			description:       "different JSON content should not be suppressed",
		},
		{
			name:              "invalid JSON in plan",
			planValue:         types.StringValue(`{"key": invalid}`),
			stateValue:        types.StringValue(`{"key": "value"}`),
			expectSuppression: false,
			description:       "invalid JSON in plan should not be suppressed",
		},
		{
			name:              "invalid JSON in state",
			planValue:         types.StringValue(`{"key": "value"}`),
			stateValue:        types.StringValue(`{"key": invalid}`),
			expectSuppression: false,
			description:       "invalid JSON in state should not be suppressed",
		},
		{
			name:              "empty strings",
			planValue:         types.StringValue(""),
			stateValue:        types.StringValue(""),
			expectSuppression: false,
			description:       "empty strings should not be modified",
		},
		{
			name:              "empty vs empty object",
			planValue:         types.StringValue(""),
			stateValue:        types.StringValue("{}"),
			expectSuppression: false,
			description:       "empty string vs empty object should not be suppressed",
		},
		{
			name:              "null plan value",
			planValue:         types.StringNull(),
			stateValue:        types.StringValue(`{"key": "value"}`),
			expectSuppression: false,
			description:       "null plan value should not be modified",
		},
		{
			name:              "null state value",
			planValue:         types.StringValue(`{"key": "value"}`),
			stateValue:        types.StringNull(),
			expectSuppression: false,
			description:       "null state value should not be modified",
		},
		{
			name:              "unknown plan value",
			planValue:         types.StringUnknown(),
			stateValue:        types.StringValue(`{"key": "value"}`),
			expectSuppression: false,
			description:       "unknown plan value should not be modified",
		},
		{
			name:              "unknown state value",
			planValue:         types.StringValue(`{"key": "value"}`),
			stateValue:        types.StringUnknown(),
			expectSuppression: false,
			description:       "unknown state value should not be modified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			modifier := JSONNormalize()

			req := planmodifier.StringRequest{
				PlanValue:  tt.planValue,
				StateValue: tt.stateValue,
			}
			resp := &planmodifier.StringResponse{
				PlanValue: tt.planValue,
			}

			// Store original plan value to check if it changed
			originalPlanValue := resp.PlanValue

			modifier.PlanModifyString(context.Background(), req, resp)

			// Check if the plan value was modified (changed from original)
			planValueChanged := !resp.PlanValue.Equal(originalPlanValue)

			// For suppression to occur, the plan value should have changed AND
			// now equal the state value
			suppressed := planValueChanged && resp.PlanValue.Equal(tt.stateValue)

			if suppressed != tt.expectSuppression {
				t.Errorf("%s: expected suppression=%v, got suppression=%v", tt.description, tt.expectSuppression, suppressed)
				t.Errorf("  Original plan value: %v", originalPlanValue)
				t.Errorf("  Final plan value: %v", resp.PlanValue)
				t.Errorf("  State value: %v", tt.stateValue)
				t.Errorf("  Plan value changed: %v", planValueChanged)
			}
		})
	}
}

func TestNormalizeJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
			hasError: false,
		},
		{
			name:     "simple object",
			input:    `{"key": "value"}`,
			expected: `{"key":"value"}`,
			hasError: false,
		},
		{
			name:     "pretty formatted object",
			input:    "{\n  \"key\": \"value\",\n  \"number\": 123\n}",
			expected: `{"key":"value","number":123}`,
			hasError: false,
		},
		{
			name:     "array",
			input:    `[1, 2, 3]`,
			expected: `[1,2,3]`,
			hasError: false,
		},
		{
			name:     "complex nested object",
			input:    `{"outer": {"inner": {"deep": "value"}}}`,
			expected: `{"outer":{"inner":{"deep":"value"}}}`,
			hasError: false,
		},
		{
			name:     "invalid JSON",
			input:    `{"key": invalid}`,
			expected: "",
			hasError: true,
		},
		{
			name:     "incomplete JSON",
			input:    `{"key":`,
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := NormalizeJSON(tt.input)

			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestJSONNormalizePlanModifier_Description(t *testing.T) {
	modifier := jsonNormalizePlanModifier{}

	desc := modifier.Description(context.Background())
	if desc == "" {
		t.Error("Description should not be empty")
	}

	markdownDesc := modifier.MarkdownDescription(context.Background())
	if markdownDesc == "" {
		t.Error("MarkdownDescription should not be empty")
	}

	if desc != markdownDesc {
		t.Error("Description and MarkdownDescription should be the same")
	}
}

func TestJSONNormalize_OriginalIssue(t *testing.T) {
	t.Parallel()

	// This test reproduces the original issue where formatted JSON from user
	// input was being seen as different from minified JSON returned by server
	planValue := types.StringValue(`{
  "title": "dynamic",
  "description": "A dynamic schema constructed at runtime, used for data transformation.",
  "properties": {
    "entity": {
      "description": "The record of the entity",
      "type": "object"
    }
  },
  "type": "object"
}`)

	stateValue := types.StringValue(`{"description":"A dynamic schema constructed at runtime, used for data transformation.","properties":{"entity":{"description":"The record of the entity","type":"object"}},"title":"dynamic","type":"object"}`)

	modifier := JSONNormalize()

	req := planmodifier.StringRequest{
		PlanValue:  planValue,
		StateValue: stateValue,
	}
	resp := &planmodifier.StringResponse{
		PlanValue: planValue,
	}

	modifier.PlanModifyString(context.Background(), req, resp)

	// The plan modifier should suppress the diff by setting resp.PlanValue to stateValue
	if !resp.PlanValue.Equal(stateValue) {
		t.Error("Plan modifier should have suppressed the diff for semantically equivalent JSON")
		t.Errorf("Expected plan value to equal state value")
		t.Errorf("Plan value: %v", resp.PlanValue)
		t.Errorf("State value: %v", stateValue)
	}

	// Verify that the original plan value was actually changed
	if resp.PlanValue.Equal(planValue) {
		t.Error("Plan value should have been modified from original")
	}
}
