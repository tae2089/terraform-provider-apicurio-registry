// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tae2089/terraform-provider-apicurio-registry/internal/client"
	"github.com/tae2089/terraform-provider-apicurio-registry/internal/provider/artifact"
	"github.com/tae2089/terraform-provider-apicurio-registry/internal/provider/artifact_rule"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure ApicurioProvider satisfies various provider interfaces.
var _ provider.Provider = &ApicurioProvider{}

// ApicurioProvider defines the provider implementation.
type ApicurioProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ApicurioProviderModel describes the provider data model.
type ApicurioProviderModel struct {
	Endpoint             types.String `tfsdk:"endpoint"`
	KeycloakServerUrl    types.String `tfsdk:"keycloak_server_url"`
	KeycloakRealm        types.String `tfsdk:"keycloak_realm"`
	KeycloakClientId     types.String `tfsdk:"keycloak_client_id"`
	KeycloakClientSecret types.String `tfsdk:"keycloak_client_secret"`
}

func (p *ApicurioProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "apicurio"
	resp.Version = p.version
}

func (p *ApicurioProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The **Apicurio** provider is used to interact with [Apicurio Registry](https://www.apicur.io/registry/), an open-source API and schema registry. It allows you to manage artifacts (schemas) and their validation/compatibility rules as Terraform resources.",

		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The URL of the Apicurio Registry API endpoint. Defaults to `http://localhost:8080/apis/registry/v3`. May also be provided via the `APICURIO_REGISTRY_URL` environment variable.",
				Optional:            true,
			},
			"keycloak_server_url": schema.StringAttribute{
				MarkdownDescription: "Keycloak server URL. Required if using Keycloak authentication.",
				Optional:            true,
			},
			"keycloak_realm": schema.StringAttribute{
				MarkdownDescription: "Keycloak realm name.",
				Optional:            true,
			},
			"keycloak_client_id": schema.StringAttribute{
				MarkdownDescription: "Keycloak client ID.",
				Optional:            true,
			},
			"keycloak_client_secret": schema.StringAttribute{
				MarkdownDescription: "Keycloak client secret.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *ApicurioProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ApicurioProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to get config", map[string]any{"success": false})
		return
	}
	// Configuration values are now available.
	endpoint := os.Getenv("APICURIO_REGISTRY_URL")
	if endpoint == "" {
		endpoint = "http://localhost:8080/apis/registry/v3" // Default fallback
	}
	if !data.Endpoint.IsNull() && !data.Endpoint.IsUnknown() {
		endpoint = data.Endpoint.ValueString()
	}

	// Keycloak Configuration
	kcUrl := os.Getenv("APICURIO_KEYCLOAK_SERVER_URL")
	if !data.KeycloakServerUrl.IsNull() && !data.KeycloakServerUrl.IsUnknown() {
		kcUrl = data.KeycloakServerUrl.ValueString()
	}

	kcRealm := os.Getenv("APICURIO_KEYCLOAK_REALM")
	if !data.KeycloakRealm.IsNull() && !data.KeycloakRealm.IsUnknown() {
		kcRealm = data.KeycloakRealm.ValueString()
	}

	kcClientId := os.Getenv("APICURIO_KEYCLOAK_CLIENT_ID")
	if !data.KeycloakClientId.IsNull() && !data.KeycloakClientId.IsUnknown() {
		kcClientId = data.KeycloakClientId.ValueString()
	}

	kcClientSecret := os.Getenv("APICURIO_KEYCLOAK_CLIENT_SECRET")
	if !data.KeycloakClientSecret.IsNull() && !data.KeycloakClientSecret.IsUnknown() {
		kcClientSecret = data.KeycloakClientSecret.ValueString()
	}

	accessToken := ""
	if kcUrl != "" && kcRealm != "" && kcClientId != "" {

		tokenEndpoint := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", strings.TrimSuffix(kcUrl, "/"), kcRealm)

		formData := url.Values{}
		formData.Set("grant_type", "client_credentials")
		formData.Set("client_id", kcClientId)
		if kcClientSecret != "" {
			formData.Set("client_secret", kcClientSecret)
		}

		tokenReq, err := http.NewRequestWithContext(ctx, "POST", tokenEndpoint, strings.NewReader(formData.Encode()))
		if err != nil {
			resp.Diagnostics.AddError("Auth Error", fmt.Sprintf("Unable to create token request: %s", err))
			return
		}
		tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		tokenResp, err := http.DefaultClient.Do(tokenReq)
		if err != nil {
			resp.Diagnostics.AddError("Auth Error", fmt.Sprintf("Unable to fetch access token from Keycloak: %s", err))
			return
		}
		defer tokenResp.Body.Close()

		if tokenResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(tokenResp.Body)
			resp.Diagnostics.AddError("Auth Error", fmt.Sprintf("Keycloak returned status %d: %s", tokenResp.StatusCode, body))
			return
		}

		var tokenData struct {
			AccessToken string `json:"access_token"`
		}
		if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
			resp.Diagnostics.AddError("Auth Error", fmt.Sprintf("Unable to decode Keycloak response: %s", err))
			return
		}
		accessToken = tokenData.AccessToken
	}

	// Example client configuration for data sources and resources
	client := &client.ApicurioClient{
		HttpClient: http.DefaultClient,
		Endpoint:   endpoint,
		Token:      accessToken,
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ApicurioProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		artifact.NewArtifactResource,
		artifact_rule.NewArtifactRuleResource,
	}
}

func (p *ApicurioProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		artifact.NewArtifactDataSource,
		artifact_rule.NewArtifactRuleDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ApicurioProvider{
			version: version,
		}
	}
}
