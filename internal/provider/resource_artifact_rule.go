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

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ArtifactRuleResource{}
var _ resource.ResourceWithImportState = &ArtifactRuleResource{}

func NewArtifactRuleResource() resource.Resource {
	return &ArtifactRuleResource{}
}

type ArtifactRuleResource struct {
	client *ApicurioClient
}

type ArtifactRuleResourceModel struct {
	Id         types.String `tfsdk:"id"`
	GroupId    types.String `tfsdk:"group_id"`
	ArtifactId types.String `tfsdk:"artifact_id"`
	Type       types.String `tfsdk:"type"`
	Config     types.String `tfsdk:"config"`
}

type RulePayload struct {
	Type   string `json:"type,omitempty"`
	Config string `json:"config"`
}

func (r *ArtifactRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_artifact_rule"
}

func (r *ArtifactRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Apicurio Registry Artifact Rule resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource ID (group_id/artifact_id/type)",
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
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Rule type (VALIDITY, COMPATIBILITY, INTEGRITY)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("COMPATIBILITY", "VALIDITY"),
				},
			},
			"config": schema.StringAttribute{
				MarkdownDescription: "Rule configuration (e.g., FULL, BACKWARD, SYNTAX_ONLY)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						// COMPATIBILITY values
						"BACKWARD", "BACKWARD_TRANSITIVE",
						"FORWARD", "FORWARD_TRANSITIVE",
						"FULL", "FULL_TRANSITIVE",
						// Shared
						"NONE",
						// VALIDITY values
						"SYNTAX_ONLY",
					),
				},
			},
		},
	}
}

func (r *ArtifactRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ArtifactRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ArtifactRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId := data.GroupId.ValueString()
	if groupId == "" {
		groupId = "default"
	}

	url := fmt.Sprintf("%s/groups/%s/artifacts/%s/rules", r.client.Endpoint, groupId, data.ArtifactId.ValueString())

	payload := RulePayload{
		Type:   data.Type.ValueString(),
		Config: data.Config.ValueString(),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("JSON Encoding Error", fmt.Sprintf("Unable to encode payload: %s", err))
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.HttpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create artifact rule, got error: %s", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent && httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to create artifact rule, got status: %d, body: %s", httpResp.StatusCode, body))
		return
	}

	data.Id = types.StringValue(fmt.Sprintf("%s/%s/%s", groupId, data.ArtifactId.ValueString(), data.Type.ValueString()))

	tflog.Trace(ctx, "created an artifact rule resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ArtifactRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId := data.GroupId.ValueString()
	if groupId == "" {
		groupId = "default"
	}

	url := fmt.Sprintf("%s/groups/%s/artifacts/%s/rules/%s", r.client.Endpoint, groupId, data.ArtifactId.ValueString(), data.Type.ValueString())
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}

	httpResp, err := r.client.HttpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read artifact rule, got error: %s", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to read artifact rule, got status: %d, body: %s", httpResp.StatusCode, body))
		return
	}

	var ruleResp RulePayload
	if err := json.NewDecoder(httpResp.Body).Decode(&ruleResp); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to decode artifact rule, got error: %s", err))
		return
	}

	data.Config = types.StringValue(ruleResp.Config)
	data.Id = types.StringValue(fmt.Sprintf("%s/%s/%s", groupId, data.ArtifactId.ValueString(), data.Type.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ArtifactRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId := data.GroupId.ValueString()
	if groupId == "" {
		groupId = "default"
	}

	url := fmt.Sprintf("%s/groups/%s/artifacts/%s/rules/%s", r.client.Endpoint, groupId, data.ArtifactId.ValueString(), data.Type.ValueString())

	payload := RulePayload{
		Config: data.Config.ValueString(),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("JSON Encoding Error", fmt.Sprintf("Unable to encode payload: %s", err))
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(payloadBytes))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create update request, got error: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.HttpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update artifact rule, got error: %s", err))
		return
	}
	defer httpResp.Body.Close()

	// Update endpoint for Apicurio Registry rules might return the updated rule object or empty body (200 OK)
	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to update artifact rule, got status: %d, body: %s", httpResp.StatusCode, body))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ArtifactRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId := data.GroupId.ValueString()
	if groupId == "" {
		groupId = "default"
	}

	url := fmt.Sprintf("%s/groups/%s/artifacts/%s/rules/%s", r.client.Endpoint, groupId, data.ArtifactId.ValueString(), data.Type.ValueString())
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create delete request, got error: %s", err))
		return
	}

	httpResp, err := r.client.HttpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete artifact rule, got error: %s", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent && httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to delete artifact rule, got status: %d, body: %s", httpResp.StatusCode, body))
		return
	}
}

func (r *ArtifactRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'groupId/artifactId/type', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("artifact_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
