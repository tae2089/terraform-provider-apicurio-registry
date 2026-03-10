// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ArtifactResource{}
var _ resource.ResourceWithImportState = &ArtifactResource{}

func NewArtifactResource() resource.Resource {
	return &ArtifactResource{}
}

type ArtifactResource struct {
	client *ApicurioClient
}

type ArtifactResourceModel struct {
	Id         types.String `tfsdk:"id"`
	GroupId    types.String `tfsdk:"group_id"`
	ArtifactId types.String `tfsdk:"artifact_id"`
	Content    types.String `tfsdk:"content"`
	Type       types.String `tfsdk:"type"`
	Version    types.String `tfsdk:"version"`
	GlobalId   types.Int64  `tfsdk:"global_id"`
	State      types.String `tfsdk:"state"`
}

type ArtifactMetaData struct {
	Id       string `json:"id"`
	GroupId  string `json:"groupId"`
	Type     string `json:"type"`
	Version  string `json:"version"`
	State    string `json:"state"`
	GlobalId int64  `json:"globalId"`
}

func (r *ArtifactResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_artifact"
}

func (r *ArtifactResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Apicurio Registry Artifact resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource ID (group_id/artifact_id)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_id": schema.StringAttribute{
				MarkdownDescription: "Artifact group ID",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("default"),
			},
			"artifact_id": schema.StringAttribute{
				MarkdownDescription: "Artifact ID",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "Artifact content",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Artifact type (e.g. AVRO, JSON, OPENAPI)",
				Optional:            true,
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Artifact version",
				Computed:            true,
			},
			"global_id": schema.Int64Attribute{
				MarkdownDescription: "Global artifact ID",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "State of the artifact",
				Computed:            true,
			},
		},
	}
}

func (r *ArtifactResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ApicurioClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ApicurioClient, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *ArtifactResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ArtifactResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/groups/%s/artifacts", r.client.Endpoint, data.GroupId.ValueString())
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(data.Content.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if !data.ArtifactId.IsNull() && !data.ArtifactId.IsUnknown() {
		httpReq.Header.Set("X-Registry-ArtifactId", data.ArtifactId.ValueString())
	}
	if !data.Type.IsNull() && !data.Type.IsUnknown() {
		httpReq.Header.Set("X-Registry-ArtifactType", data.Type.ValueString())
	}

	httpResp, err := r.client.HttpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create artifact, got error: %s", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to create artifact, got status: %d, body: %s", httpResp.StatusCode, body))
		return
	}

	var meta ArtifactMetaData
	if err := json.NewDecoder(httpResp.Body).Decode(&meta); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to decode artifact metadata, got error: %s", err))
		return
	}

	groupId := meta.GroupId
	if groupId == "" {
		groupId = data.GroupId.ValueString()
	}

	data.ArtifactId = types.StringValue(meta.Id)
	data.GroupId = types.StringValue(groupId)
	data.Type = types.StringValue(meta.Type)
	data.Version = types.StringValue(meta.Version)
	data.GlobalId = types.Int64Value(meta.GlobalId)
	data.State = types.StringValue(meta.State)
	data.Id = types.StringValue(fmt.Sprintf("%s/%s", groupId, meta.Id))

	tflog.Trace(ctx, "created an artifact resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ArtifactResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read metadata
	url := fmt.Sprintf("%s/groups/%s/artifacts/%s/meta", r.client.Endpoint, data.GroupId.ValueString(), data.ArtifactId.ValueString())
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}

	httpResp, err := r.client.HttpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read artifact, got error: %s", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

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

	groupId := meta.GroupId
	if groupId == "" {
		groupId = data.GroupId.ValueString()
	}

	// Read content
	contentUrl := fmt.Sprintf("%s/groups/%s/artifacts/%s/versions/%s", r.client.Endpoint, groupId, meta.Id, meta.Version)
	contentReq, err := http.NewRequestWithContext(ctx, "GET", contentUrl, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create content request, got error: %s", err))
		return
	}

	contentResp, err := r.client.HttpClient.Do(contentReq)
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

	data.ArtifactId = types.StringValue(meta.Id)
	data.GroupId = types.StringValue(groupId)
	data.Type = types.StringValue(meta.Type)
	data.Version = types.StringValue(meta.Version)
	data.GlobalId = types.Int64Value(meta.GlobalId)
	data.State = types.StringValue(meta.State)
	data.Content = types.StringValue(string(content))
	data.Id = types.StringValue(fmt.Sprintf("%s/%s", groupId, meta.Id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ArtifactResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/groups/%s/artifacts/%s", r.client.Endpoint, data.GroupId.ValueString(), data.ArtifactId.ValueString())
	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, strings.NewReader(data.Content.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create update request, got error: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.HttpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update artifact, got error: %s", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to update artifact, got status: %d, body: %s", httpResp.StatusCode, body))
		return
	}

	var meta ArtifactMetaData
	if err := json.NewDecoder(httpResp.Body).Decode(&meta); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to decode artifact metadata, got error: %s", err))
		return
	}

	data.Version = types.StringValue(meta.Version)
	data.GlobalId = types.Int64Value(meta.GlobalId)
	data.State = types.StringValue(meta.State)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ArtifactResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/groups/%s/artifacts/%s", r.client.Endpoint, data.GroupId.ValueString(), data.ArtifactId.ValueString())
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create delete request, got error: %s", err))
		return
	}

	httpResp, err := r.client.HttpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete artifact, got error: %s", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent && httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to delete artifact, got status: %d, body: %s", httpResp.StatusCode, body))
		return
	}
}

func (r *ArtifactResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'groupId/artifactId', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("artifact_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
