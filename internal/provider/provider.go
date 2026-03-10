// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ApicurioProvider satisfies various provider interfaces.
var _ provider.Provider = &ApicurioProvider{}
var _ provider.ProviderWithFunctions = &ApicurioProvider{}
var _ provider.ProviderWithEphemeralResources = &ApicurioProvider{}
var _ provider.ProviderWithActions = &ApicurioProvider{}

// ApicurioProvider defines the provider implementation.
type ApicurioProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ApicurioProviderModel describes the provider data model.
type ApicurioProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
}

func (p *ApicurioProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "apicurio"
	resp.Version = p.version
}

func (p *ApicurioProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Apicurio Registry provider manages schemas and artifacts in Apicurio Registry.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The URL of the Apicurio Registry API endpoint. Defaults to `http://localhost:8080/apis/registry/v2`.",
				Optional:            true,
			},
		},
	}
}

func (p *ApicurioProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ApicurioProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	endpoint := "http://localhost:8080/apis/registry/v2" // Default fallback
	if !data.Endpoint.IsNull() && !data.Endpoint.IsUnknown() {
		endpoint = data.Endpoint.ValueString()
	}

	// Example client configuration for data sources and resources
	client := &ApicurioClient{
		HttpClient: http.DefaultClient,
		Endpoint:   endpoint,
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ApicurioProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
		NewArtifactResource,
		NewArtifactRuleResource,
	}
}

func (p *ApicurioProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		NewExampleEphemeralResource,
	}
}

func (p *ApicurioProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
	}
}

func (p *ApicurioProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewExampleFunction,
	}
}

func (p *ApicurioProvider) Actions(ctx context.Context) []func() action.Action {
	return []func() action.Action{
		NewExampleAction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ApicurioProvider{
			version: version,
		}
	}
}
