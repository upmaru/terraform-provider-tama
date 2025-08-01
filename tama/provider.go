// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tama

import (
	"context"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	tama "github.com/upmaru/tama-go"

	"github.com/upmaru/terraform-provider-tama/tama/memory/prompt"
	"github.com/upmaru/terraform-provider-tama/tama/neural/bridge"
	"github.com/upmaru/terraform-provider-tama/tama/neural/class"
	"github.com/upmaru/terraform-provider-tama/tama/neural/corpus"
	"github.com/upmaru/terraform-provider-tama/tama/neural/node"
	space_processor "github.com/upmaru/terraform-provider-tama/tama/neural/processor"
	"github.com/upmaru/terraform-provider-tama/tama/neural/space"
	"github.com/upmaru/terraform-provider-tama/tama/perception/chain"
	perception_context "github.com/upmaru/terraform-provider-tama/tama/perception/context"
	"github.com/upmaru/terraform-provider-tama/tama/perception/modular_thought"
	"github.com/upmaru/terraform-provider-tama/tama/perception/delegated_thought"
	"github.com/upmaru/terraform-provider-tama/tama/perception/path"
	thought_processor "github.com/upmaru/terraform-provider-tama/tama/perception/processor"
	source_identity "github.com/upmaru/terraform-provider-tama/tama/sensory/identity"
	"github.com/upmaru/terraform-provider-tama/tama/sensory/limit"
	"github.com/upmaru/terraform-provider-tama/tama/sensory/model"
	"github.com/upmaru/terraform-provider-tama/tama/sensory/source"
	"github.com/upmaru/terraform-provider-tama/tama/sensory/specification"
)

// Ensure TamaProvider satisfies various provider interfaces.
var _ provider.Provider = &TamaProvider{}
var _ provider.ProviderWithFunctions = &TamaProvider{}
var _ provider.ProviderWithEphemeralResources = &TamaProvider{}

// TamaProvider defines the provider implementation.
type TamaProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// TamaProviderModel describes the provider data model.
type TamaProviderModel struct {
	BaseURL types.String `tfsdk:"base_url"`
	APIKey  types.String `tfsdk:"api_key"`
	Timeout types.Int64  `tfsdk:"timeout"`
}

func (p *TamaProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "tama"
	resp.Version = p.version
}

func (p *TamaProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Terraform provider for Tama API resources",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				MarkdownDescription: "The base URL for the Tama API. Can also be set via the TAMA_BASE_URL environment variable.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API key for authenticating with the Tama API. Can also be set via the TAMA_API_KEY environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout for API requests in seconds. Defaults to 30.",
				Optional:            true,
			},
		},
	}
}

func (p *TamaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data TamaProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values
	baseURL := "https://api.tama.io"
	apiKey := ""
	timeout := int64(30)

	// Override with configuration values
	if !data.BaseURL.IsNull() {
		baseURL = data.BaseURL.ValueString()
	}

	if !data.APIKey.IsNull() {
		apiKey = data.APIKey.ValueString()
	}

	if !data.Timeout.IsNull() {
		timeout = data.Timeout.ValueInt64()
	}

	// Override with environment variables
	if envBaseURL := os.Getenv("TAMA_BASE_URL"); envBaseURL != "" {
		baseURL = envBaseURL
	}

	if envAPIKey := os.Getenv("TAMA_API_KEY"); envAPIKey != "" {
		apiKey = envAPIKey
	}

	// Validate required configuration
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key Configuration",
			"The provider cannot create the Tama API client as there is a missing or empty value for the API key. "+
				"Set the api_key value in the configuration or use the TAMA_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
		return
	}

	if baseURL == "" {
		resp.Diagnostics.AddError(
			"Missing Base URL Configuration",
			"The provider cannot create the Tama API client as there is a missing or empty value for the base URL. "+
				"Set the base_url value in the configuration or use the TAMA_BASE_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
		return
	}

	ctx = tflog.SetField(ctx, "tama_base_url", baseURL)
	ctx = tflog.SetField(ctx, "tama_timeout", timeout)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "tama_api_key")

	tflog.Debug(ctx, "Creating Tama API client")

	// Create Tama client configuration
	config := tama.Config{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Timeout: time.Duration(timeout) * time.Second,
	}

	// Create Tama client
	client := tama.NewClient(config)

	// Make the client available during DataSource and Resource type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Tama API client", map[string]any{"success": true})
}

func (p *TamaProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		space.NewResource,
		bridge.NewResource,
		class.NewResource,
		corpus.NewResource,
		node.NewResource,
		space_processor.NewResource,
		source.NewResource,
		source_identity.NewResource,
		model.NewResource,
		limit.NewResource,
		specification.NewResource,
		prompt.NewResource,
		chain.NewResource,
		modular_thought.NewResource,
		delegated_thought.NewResource,
		thought_processor.NewResource,
		perception_context.NewResource,
		path.NewResource,
	}
}

func (p *TamaProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *TamaProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		space.NewDataSource,
		bridge.NewDataSource,
		class.NewDataSource,
		corpus.NewDataSource,
		node.NewDataSource,
		space_processor.NewDataSource,
		source.NewDataSource,
		source_identity.NewDataSource,
		model.NewDataSource,
		limit.NewDataSource,
		specification.NewDataSource,
		prompt.NewDataSource,
		chain.NewDataSource,
		modular_thought.NewDataSource,
		thought_processor.NewDataSource,
		perception_context.NewDataSource,
		path.NewDataSource,
	}
}

func (p *TamaProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TamaProvider{
			version: version,
		}
	}
}
