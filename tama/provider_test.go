// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tama

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestProvider_DataSources(t *testing.T) {
	provider := &TamaProvider{}
	dataSources := provider.DataSources(context.Background())

	// Check that we have the expected number of data sources
	if len(dataSources) == 0 {
		t.Fatal("Expected at least one data source, got 0")
	}

	// Verify that the action data source is registered
	found := false
	for _, dsFunc := range dataSources {
		ds := dsFunc()
		var resp datasource.MetadataResponse
		ds.Metadata(context.Background(), datasource.MetadataRequest{
			ProviderTypeName: "tama",
		}, &resp)

		if resp.TypeName == "tama_action" {
			found = true
			break
		}
	}

	if !found {
		t.Error("tama_action data source not found in provider registration")
	}
}

func TestProvider_Resources(t *testing.T) {
	provider := &TamaProvider{}
	resources := provider.Resources(context.Background())
	want := map[string]bool{
		"tama_action_modifier":       false,
		"tama_thought_tool_modifier": false,
	}

	for _, resourceFunc := range resources {
		providerResource := resourceFunc()
		var resp resource.MetadataResponse
		providerResource.Metadata(context.Background(), resource.MetadataRequest{
			ProviderTypeName: "tama",
		}, &resp)

		if _, ok := want[resp.TypeName]; ok {
			want[resp.TypeName] = true
		}
	}

	for resourceName, found := range want {
		if !found {
			t.Errorf("%s resource not found in provider registration", resourceName)
		}
	}
}
