// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package modifier

import (
	"errors"
	"net/http"
	"testing"

	"github.com/upmaru/tama-go/tools"
)

func TestModelFromModifier(t *testing.T) {
	modifier := &tools.Modifier{
		ID:              "modifier-id",
		ThoughtToolID:   "thought-tool-id",
		Index:           0,
		Target:          "/body/search/scope/user_id",
		OnMissingParent: tools.ModifierMissingPolicySkip,
		OnMissingSource: tools.ModifierMissingPolicyError,
		Source: tools.ModifierSource{
			Type: tools.ModifierSourceTypeMetadata,
			Path: tools.ModifierSourcePathActorIdentifier,
		},
		ProvisionState: "active",
	}

	model := modelFromModifier(modifier)

	if model.Id.ValueString() != modifier.ID ||
		model.ThoughtToolId.ValueString() != modifier.ThoughtToolID ||
		model.Index.ValueInt64() != int64(modifier.Index) ||
		model.Target.ValueString() != modifier.Target ||
		model.OnMissingParent.ValueString() != modifier.OnMissingParent ||
		model.OnMissingSource.ValueString() != modifier.OnMissingSource ||
		model.Source == nil ||
		model.Source.Type.ValueString() != modifier.Source.Type ||
		model.Source.Path.ValueString() != modifier.Source.Path ||
		model.ProvisionState.ValueString() != modifier.ProvisionState {
		t.Fatalf("model did not preserve the complete modifier response: %#v", model)
	}
}

func TestIsNotFound(t *testing.T) {
	tests := map[string]struct {
		err  error
		want bool
	}{
		"typed not found": {
			err:  &tools.Error{StatusCode: http.StatusNotFound},
			want: true,
		},
		"wrapped typed not found": {
			err:  errors.Join(errors.New("request failed"), &tools.Error{StatusCode: http.StatusNotFound}),
			want: true,
		},
		"other API error": {
			err:  &tools.Error{StatusCode: http.StatusUnprocessableEntity},
			want: false,
		},
		"ordinary error": {
			err:  errors.New("request failed"),
			want: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := isNotFound(test.err); got != test.want {
				t.Fatalf("isNotFound() = %t, want %t", got, test.want)
			}
		})
	}
}
