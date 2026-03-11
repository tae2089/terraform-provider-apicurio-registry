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
var _ datasource.DataSource = &ArtifactRuleDataSource{}

func NewArtifactRuleDataSource() datasource.DataSource {
	return &ArtifactRuleDataSource{}
}

type ArtifactRuleDataSource struct {
	client *ApicurioClient
}

type ArtifactRuleDataSourceModel struct {
	Id         types.String `tfsdk:"id"`
	GroupId    types.String `tfsdk:"group_id"`
	ArtifactId types.String `tfsdk:"artifact_id"`
	Type       types.String `tfsdk:"type"`
	Config     types.String `tfsdk:"config"`
}

func (d *ArtifactRuleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_artifact_rule"
}

func (d *ArtifactRuleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `apicurio_artifact_rule` data source allows you to retrieve the configuration of a specific rule applied to an artifact in the Apicurio Registry.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The composite ID of the artifact rule, formatted as `group_id/artifact_id/type`.",
			},
			"group_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the artifact group. If not provided, it defaults to `default`.",
				Optional:            true,
				Computed:            true,
			},
			"artifact_id": schema.StringAttribute{
				MarkdownDescription: "The unique ID of the artifact.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the rule (e.g., `VALIDITY`, `COMPATIBILITY`).",
				Required:            true,
			},
			"config": schema.StringAttribute{
				MarkdownDescription: "The configuration value for the rule. Valid values depend on the rule `type`:\n" +
					"  - For `COMPATIBILITY`: `BACKWARD`, `BACKWARD_TRANSITIVE`, `FORWARD`, `FORWARD_TRANSITIVE`, `FULL`, `FULL_TRANSITIVE`, `NONE`\n" +
					"  - For `VALIDITY`: `FULL`, `SYNTAX_ONLY`, `NONE`",
				Computed: true,
			},
		},
	}
}

func (d *ArtifactRuleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ArtifactRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ArtifactRuleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId := data.GroupId.ValueString()
	if groupId == "" {
		groupId = "default"
	}
	artifactId := data.ArtifactId.ValueString()
	ruleType := data.Type.ValueString()

	// v3 Artifact Rule Read: GET /groups/{groupId}/artifacts/{artifactId}/rules/{ruleType}
	url := fmt.Sprintf("%s/groups/%s/artifacts/%s/rules/%s", d.client.Endpoint, groupId, artifactId, ruleType)
	httpReq, err := d.client.NewRequest(ctx, "GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}

	httpResp, err := d.client.HttpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read artifact rule, got error: %s", err))
		return
	}
	defer httpResp.Body.Close()

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
	data.Id = types.StringValue(fmt.Sprintf("%s/%s/%s", groupId, artifactId, ruleType))
	data.GroupId = types.StringValue(groupId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
