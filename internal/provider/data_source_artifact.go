// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ArtifactDataSource{}

func NewArtifactDataSource() datasource.DataSource {
	return &ArtifactDataSource{}
}

type ArtifactDataSource struct {
	client *ApicurioClient
}

type ArtifactDataSourceModel struct {
	Id         types.String `tfsdk:"id"`
	GroupId    types.String `tfsdk:"group_id"`
	ArtifactId types.String `tfsdk:"artifact_id"`
	Content    types.String `tfsdk:"content"`
	Type       types.String `tfsdk:"type"`
	Version    types.String `tfsdk:"version"`
	GlobalId   types.Int64  `tfsdk:"global_id"`
	State      types.String `tfsdk:"state"`
}

func (d *ArtifactDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_artifact"
}

func (d *ArtifactDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `apicurio_artifact` data source allows you to retrieve metadata and content for an existing artifact in the Apicurio Registry.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The composite ID of the artifact, formatted as `group_id/artifact_id`.",
			},
			"group_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the artifact group. If not provided, it defaults to `default`.",
				Optional:            true,
				Computed:            true,
			},
			"artifact_id": schema.StringAttribute{
				MarkdownDescription: "The unique ID of the artifact within the group.",
				Required:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The actual content of the latest version of the artifact (e.g., the JSON or YAML of a schema or API definition).",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the artifact (e.g., `AVRO`, `JSON`, `OPENAPI`, `ASYNCAPI`, `GRAPHQL`, `KCONNECT`, `WSDL`, `XSD`, `XML`).",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "The version number of the latest version of the artifact.",
				Computed:            true,
			},
			"global_id": schema.Int64Attribute{
				MarkdownDescription: "The globally unique ID of the latest version of the artifact.",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The state of the latest version of the artifact (e.g., `ENABLED`, `DISABLED`, `DEPRECATED`).",
				Computed:            true,
			},
		},
	}
}

func (d *ArtifactDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ApicurioClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ApicurioClient, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *ArtifactDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ArtifactDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId := data.GroupId.ValueString()
	if groupId == "" {
		groupId = "default"
	}
	artifactId := data.ArtifactId.ValueString()

	// 1. Read Artifact Metadata
	url := fmt.Sprintf("%s/groups/%s/artifacts/%s", d.client.Endpoint, groupId, artifactId)
	httpReq, err := d.client.NewRequest(ctx, "GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}

	httpResp, err := d.client.HttpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read artifact, got error: %s", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to read artifact metadata, got status: %d, body: %s", httpResp.StatusCode, body))
		return
	}

	var meta ArtifactMetaData
	if err := json.NewDecoder(httpResp.Body).Decode(&meta); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to decode artifact metadata, got error: %s", err))
		return
	}

	// 2. Read Latest Version Metadata
	vMetaUrl := fmt.Sprintf("%s/groups/%s/artifacts/%s/versions/branch=latest", d.client.Endpoint, groupId, artifactId)
	vMetaReq, _ := d.client.NewRequest(ctx, "GET", vMetaUrl, nil)
	vMetaResp, err := d.client.HttpClient.Do(vMetaReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read latest version metadata, got error: %s", err))
		return
	}
	defer vMetaResp.Body.Close()

	if vMetaResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(vMetaResp.Body)
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to read latest version metadata, got status: %d, body: %s", vMetaResp.StatusCode, body))
		return
	}

	var vMeta VersionMetaData
	if err := json.NewDecoder(vMetaResp.Body).Decode(&vMeta); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to decode version metadata, got error: %s", err))
		return
	}

	// 3. Read Content (latest version)
	contentUrl := fmt.Sprintf("%s/groups/%s/artifacts/%s/versions/branch=latest/content", d.client.Endpoint, groupId, artifactId)
	contentReq, _ := d.client.NewRequest(ctx, "GET", contentUrl, nil)
	contentResp, err := d.client.HttpClient.Do(contentReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read artifact content, got error: %s", err))
		return
	}
	defer contentResp.Body.Close()

	if contentResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(contentResp.Body)
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to read artifact content, got status: %d, body: %s", contentResp.StatusCode, body))
		return
	}

	content, err := io.ReadAll(contentResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read artifact content body, got error: %s", err))
		return
	}

	// Fallback for Type
	artifactType := meta.ArtifactType
	if artifactType == "" {
		artifactType = meta.Type
	}

	data.ArtifactId = types.StringValue(artifactId)
	data.GroupId = types.StringValue(groupId)
	data.Type = types.StringValue(artifactType)
	data.Version = types.StringValue(vMeta.Version)
	data.GlobalId = types.Int64Value(vMeta.GlobalId)
	data.State = types.StringValue(vMeta.State)
	data.Content = types.StringValue(string(content))
	data.Id = types.StringValue(fmt.Sprintf("%s/%s", groupId, artifactId))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
