// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tama

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
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
