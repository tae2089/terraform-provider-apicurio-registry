// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
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

// v3 API Structs
type CreateArtifactRequest struct {
	ArtifactId   string                `json:"artifactId,omitempty"`
	ArtifactType string                `json:"artifactType,omitempty"`
	FirstVersion *CreateVersionRequest `json:"firstVersion,omitempty"`
}

type CreateVersionRequest struct {
	Version string           `json:"version,omitempty"`
	Content *ArtifactContent `json:"content"`
}

type CreateVersionResponse struct {
	Version  string `json:"version"`
	GlobalId int64  `json:"globalId"`
	State    string `json:"state"`
}

type ArtifactContent struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}

type ArtifactMetaData struct {
	ArtifactId   string `json:"artifactId"`
	Id           string `json:"id"`
	GroupId      string `json:"groupId"`
	ArtifactType string `json:"artifactType"`
	Type         string `json:"type"`
}

type VersionMetaData struct {
	Version  string `json:"version"`
	GlobalId int64  `json:"globalId"`
	State    string `json:"state"`
}

func (r *ArtifactResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_artifact"
}

func (r *ArtifactResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Apicurio Registry Artifact resource (v3)",

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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"artifact_id": schema.StringAttribute{
				MarkdownDescription: "Artifact ID",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
				MarkdownDescription: "State of the latest version",
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

	groupId := data.GroupId.ValueString()
	if groupId == "" {
		groupId = "default"
	}

	// Construct v3 CreateArtifactRequest
	createReq := CreateArtifactRequest{
		FirstVersion: &CreateVersionRequest{
			Content: &ArtifactContent{
				Content:     data.Content.ValueString(),
				ContentType: "application/json",
			},
		},
	}

	if !data.ArtifactId.IsNull() && !data.ArtifactId.IsUnknown() {
		createReq.ArtifactId = data.ArtifactId.ValueString()
	}
	if !data.Type.IsNull() && !data.Type.IsUnknown() {
		createReq.ArtifactType = data.Type.ValueString()
	}

	url := fmt.Sprintf("%s/groups/%s/artifacts", r.client.Endpoint, groupId)
	payload, err := json.Marshal(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to marshal request, got error: %s", err))
		return
	}

	httpReq, err := r.client.NewRequest(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

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

	// v3 Create Artifact response returns ArtifactMetaData
	var meta ArtifactMetaData
	if err := json.NewDecoder(httpResp.Body).Decode(&meta); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to decode artifact metadata, got error: %s", err))
		return
	}

	// Fallback for ID
	artifactId := meta.ArtifactId
	if artifactId == "" {
		artifactId = meta.Id
	}
	if artifactId == "" {
		artifactId = data.ArtifactId.ValueString()
	}

	// Fallback for Type
	artifactType := meta.ArtifactType
	if artifactType == "" {
		artifactType = meta.Type
	}
	if artifactType == "" {
		artifactType = data.Type.ValueString()
	}

	// Fetch latest version metadata to get version/globalId/state
	vMetaUrl := fmt.Sprintf("%s/groups/%s/artifacts/%s/versions/branch=latest", r.client.Endpoint, groupId, artifactId)
	vMetaReq, _ := r.client.NewRequest(ctx, "GET", vMetaUrl, nil)
	vMetaResp, err := r.client.HttpClient.Do(vMetaReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch latest version metadata: %s", err))
		return
	}
	defer vMetaResp.Body.Close()

	if vMetaResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(vMetaResp.Body)
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to read latest version metadata after creation, got status: %d, body: %s", vMetaResp.StatusCode, body))
		return
	}

	var vMeta VersionMetaData
	if err := json.NewDecoder(vMetaResp.Body).Decode(&vMeta); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to decode version metadata: %s", err))
		return
	}

	data.Version = types.StringValue(vMeta.Version)
	data.GlobalId = types.Int64Value(vMeta.GlobalId)
	data.State = types.StringValue(vMeta.State)
	data.ArtifactId = types.StringValue(artifactId)
	data.GroupId = types.StringValue(groupId)
	data.Type = types.StringValue(artifactType)
	data.Id = types.StringValue(fmt.Sprintf("%s/%s", groupId, artifactId))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ArtifactResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId := data.GroupId.ValueString()
	artifactId := data.ArtifactId.ValueString()

	// Robust fallback parsing from data.Id (format: groupId/artifactId)
	if groupId == "" || artifactId == "" {
		parts := strings.Split(data.Id.ValueString(), "/")
		if len(parts) == 2 {
			if groupId == "" {
				groupId = parts[0]
			}
			if artifactId == "" {
				artifactId = parts[1]
			}
		}
	}

	if groupId == "" {
		groupId = "default"
	}

	if artifactId == "" {
		tflog.Warn(ctx, "Artifact ID is missing or empty in state, removing resource from state", map[string]any{"id": data.Id.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}

	// 1. Read Artifact Metadata
	url := fmt.Sprintf("%s/groups/%s/artifacts/%s", r.client.Endpoint, groupId, artifactId)
	httpReq, err := r.client.NewRequest(ctx, "GET", url, nil)
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

	// Sync back ID from server metadata if possible
	if meta.ArtifactId != "" {
		artifactId = meta.ArtifactId
	} else if meta.Id != "" {
		artifactId = meta.Id
	}

	// Fallback for Type
	artifactType := meta.ArtifactType
	if artifactType == "" {
		artifactType = meta.Type
	}
	if artifactType == "" {
		artifactType = data.Type.ValueString()
	}

	// 2. Read Latest Version Metadata
	vMetaUrl := fmt.Sprintf("%s/groups/%s/artifacts/%s/versions/branch=latest", r.client.Endpoint, groupId, artifactId)
	vMetaReq, _ := r.client.NewRequest(ctx, "GET", vMetaUrl, nil)
	vMetaResp, err := r.client.HttpClient.Do(vMetaReq)
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
	contentUrl := fmt.Sprintf("%s/groups/%s/artifacts/%s/versions/branch=latest/content", r.client.Endpoint, groupId, artifactId)
	contentReq, _ := r.client.NewRequest(ctx, "GET", contentUrl, nil)
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

func (r *ArtifactResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ArtifactResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId := data.GroupId.ValueString()
	if groupId == "" {
		groupId = "default"
	}

	// v3 Update means adding a new version: POST /groups/{groupId}/artifacts/{artifactId}/versions
	url := fmt.Sprintf("%s/groups/%s/artifacts/%s/versions", r.client.Endpoint, groupId, data.ArtifactId.ValueString())

	versionReq := CreateVersionRequest{
		Content: &ArtifactContent{
			Content:     data.Content.ValueString(),
			ContentType: "application/json",
		},
	}
	payload, _ := json.Marshal(versionReq)

	httpReq, err := r.client.NewRequest(ctx, "POST", url, bytes.NewReader(payload))
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

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to update artifact (create version), got status: %d, body: %s", httpResp.StatusCode, body))
		return
	}

	var vMeta CreateVersionResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&vMeta); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to decode version metadata, got error: %s", err))
		return
	}

	data.Version = types.StringValue(vMeta.Version)
	data.GlobalId = types.Int64Value(vMeta.GlobalId)
	data.State = types.StringValue(vMeta.State)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ArtifactResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId := data.GroupId.ValueString()
	if groupId == "" {
		groupId = "default"
	}

	// v3 Delete Artifact: DELETE /groups/{groupId}/artifacts/{artifactId}
	url := fmt.Sprintf("%s/groups/%s/artifacts/%s", r.client.Endpoint, groupId, data.ArtifactId.ValueString())
	httpReq, err := r.client.NewRequest(ctx, "DELETE", url, nil)
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
