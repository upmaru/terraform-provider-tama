// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acceptance

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/upmaru/terraform-provider-tama/tama"
)

// TestAccProtoV6ProviderFactories is used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"tama": providerserver.NewProtocol6WithError(tama.New("test")()),
}

// ProviderConfig is a shared configuration to combine with the actual
// test configuration so the Tama client is properly configured.
const ProviderConfig = `
	provider "tama" {}
	`

// TestAccPreCheck validates that all the required environment variables
// are set before running acceptance tests.
func TestAccPreCheck(t *testing.T) {
	if v := os.Getenv("TAMA_BASE_URL"); v == "" {
		t.Fatal("TAMA_BASE_URL must be set for acceptance tests")
	}
	if v := os.Getenv("TAMA_API_KEY"); v == "" {
		t.Fatal("TAMA_API_KEY must be set for acceptance tests")
	}
}
