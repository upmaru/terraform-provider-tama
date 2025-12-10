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
	"github.com/upmaru/terraform-provider-tama/tama/neural/filter"

	"github.com/upmaru/terraform-provider-tama/tama/contexts/input"
	"github.com/upmaru/terraform-provider-tama/tama/memory/prompt"
	"github.com/upmaru/terraform-provider-tama/tama/memory/topic"
	"github.com/upmaru/terraform-provider-tama/tama/motor/action"
	"github.com/upmaru/terraform-provider-tama/tama/motor/modifier"
	"github.com/upmaru/terraform-provider-tama/tama/neural/bridge"
	"github.com/upmaru/terraform-provider-tama/tama/neural/class"
	class_operation "github.com/upmaru/terraform-provider-tama/tama/neural/class/operation"
	"github.com/upmaru/terraform-provider-tama/tama/neural/corpus"
	"github.com/upmaru/terraform-provider-tama/tama/neural/listener"
	"github.com/upmaru/terraform-provider-tama/tama/neural/node"
	space_processor "github.com/upmaru/terraform-provider-tama/tama/neural/processor"
	"github.com/upmaru/terraform-provider-tama/tama/neural/space"
	"github.com/upmaru/terraform-provider-tama/tama/perception/activation"
	"github.com/upmaru/terraform-provider-tama/tama/perception/chain"
	perception_context "github.com/upmaru/terraform-provider-tama/tama/perception/context"

	"github.com/upmaru/terraform-provider-tama/tama/perception/delegated_thought"
	"github.com/upmaru/terraform-provider-tama/tama/perception/directive"
	thought_initializer "github.com/upmaru/terraform-provider-tama/tama/perception/initializer"
	"github.com/upmaru/terraform-provider-tama/tama/perception/modular_thought"
	module_input "github.com/upmaru/terraform-provider-tama/tama/perception/module/input"
	"github.com/upmaru/terraform-provider-tama/tama/perception/path"
	thought_processor "github.com/upmaru/terraform-provider-tama/tama/perception/processor"
	"github.com/upmaru/terraform-provider-tama/tama/perception/tool"
	source_identity "github.com/upmaru/terraform-provider-tama/tama/sensory/identity"
	"github.com/upmaru/terraform-provider-tama/tama/sensory/limit"
	"github.com/upmaru/terraform-provider-tama/tama/sensory/model"
	"github.com/upmaru/terraform-provider-tama/tama/sensory/source"
	"github.com/upmaru/terraform-provider-tama/tama/sensory/specification"
	system_queue "github.com/upmaru/terraform-provider-tama/tama/system/queue"
	tool_initializer "github.com/upmaru/terraform-provider-tama/tama/tools/initializer"
	tool_input "github.com/upmaru/terraform-provider-tama/tama/tools/input"
	tool_option "github.com/upmaru/terraform-provider-tama/tama/tools/option"
	tool_output "github.com/upmaru/terraform-provider-tama/tama/tools/output"
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
	BaseURL      types.String `tfsdk:"base_url"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	Scopes       types.List   `tfsdk:"scopes"`
	Timeout      types.Int64  `tfsdk:"timeout"`
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
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The OAuth2 Client ID for authenticating with the Tama API. Can also be set via the TAMA_CLIENT_ID environment variable.",
				Optional:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "The OAuth2 Client Secret for authenticating with the Tama API. Can also be set via the TAMA_CLIENT_SECRET environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"scopes": schema.ListAttribute{
				MarkdownDescription: "OAuth2 scopes to request for the Tama API. Defaults to [\"provision.all\"].",
				Optional:            true,
				ElementType:         types.StringType,
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
	clientID := ""
	clientSecret := ""
	scopes := []string{"provision.all"}
	timeout := int64(30)

	// Override with configuration values
	if !data.BaseURL.IsNull() {
		baseURL = data.BaseURL.ValueString()
	}

	if !data.ClientID.IsNull() {
		clientID = data.ClientID.ValueString()
	}

	if !data.ClientSecret.IsNull() {
		clientSecret = data.ClientSecret.ValueString()
	}

	if !data.Timeout.IsNull() {
		timeout = data.Timeout.ValueInt64()
	}

	if !data.Scopes.IsNull() && !data.Scopes.IsUnknown() {
		var providedScopes []string
		resp.Diagnostics.Append(data.Scopes.ElementsAs(ctx, &providedScopes, false)...)
		if !resp.Diagnostics.HasError() {
			scopes = providedScopes
		}
	}

	// Override with environment variables
	if envBaseURL := os.Getenv("TAMA_BASE_URL"); envBaseURL != "" {
		baseURL = envBaseURL
	}

	if envClientID := os.Getenv("TAMA_CLIENT_ID"); envClientID != "" {
		clientID = envClientID
	}

	if envClientSecret := os.Getenv("TAMA_CLIENT_SECRET"); envClientSecret != "" {
		clientSecret = envClientSecret
	}

	// Validate required configuration
	if clientID == "" {
		resp.Diagnostics.AddError(
			"Missing Client ID Configuration",
			"The provider cannot create the Tama API client as there is a missing or empty value for the client ID. "+
				"Set the client_id value in the configuration or use the TAMA_CLIENT_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
		return
	}

	if clientSecret == "" {
		resp.Diagnostics.AddError(
			"Missing Client Secret Configuration",
			"The provider cannot create the Tama API client as there is a missing or empty value for the client secret. "+
				"Set the client_secret value in the configuration or use the TAMA_CLIENT_SECRET environment variable. "+
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
	ctx = tflog.SetField(ctx, "tama_scopes", scopes)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "tama_client_secret")

	tflog.Debug(ctx, "Creating Tama API client")

	// Create Tama client configuration
	config := tama.Config{
		BaseURL:      baseURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Timeout:      time.Duration(timeout) * time.Second,
		Scopes:       scopes,
	}

	// Create Tama client
	client, err := tama.NewClient(config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create Tama API client",
			"An error occurred while creating the Tama API client: "+err.Error(),
		)
		return
	}

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
		class_operation.NewResource,
		corpus.NewResource,
		node.NewResource,
		space_processor.NewResource,
		source.NewResource,
		source_identity.NewResource,
		model.NewResource,
		limit.NewResource,
		specification.NewResource,
		prompt.NewResource,
		topic.NewResource,
		listener.NewResource,
		filter.NewResource,
		chain.NewResource,
		modular_thought.NewResource,
		delegated_thought.NewResource,
		thought_processor.NewResource,
		perception_context.NewResource,
		path.NewResource,
		module_input.NewResource,
		directive.NewResource,
		thought_initializer.NewResource,
		tool.NewResource,
		activation.NewResource,
		input.NewResource,
		tool_input.NewResource,
		tool_initializer.NewResource,
		tool_output.NewResource,
		tool_option.NewResource,
		modifier.NewResource,
		system_queue.NewResource,
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
		source.NewDataSource,
		source_identity.NewDataSource,
		model.NewDataSource,
		specification.NewDataSource,
		prompt.NewDataSource,
		chain.NewDataSource,
		modular_thought.NewDataSource,
		perception_context.NewDataSource,
		path.NewDataSource,
		action.NewDataSource,
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
