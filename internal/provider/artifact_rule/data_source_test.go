// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package artifact_rule

import (
	providertesting "github.com/tae2089/terraform-provider-apicurio-registry/internal/testing"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccArtifactRuleDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { providertesting.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: providertesting.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArtifactRuleDataSourceConfig("ds-rule-artifact", "VALIDITY", "FULL"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apicurio_artifact_rule.test", "artifact_id", "ds-rule-artifact"),
					resource.TestCheckResourceAttr("data.apicurio_artifact_rule.test", "type", "VALIDITY"),
					resource.TestCheckResourceAttr("data.apicurio_artifact_rule.test", "config", "FULL"),
				),
			},
		},
	})
}

func testAccArtifactRuleDataSourceConfig(artifactId, ruleType, ruleConfig string) string {
	return `
provider "apicurio" {
  endpoint = "http://localhost:8080/apis/registry/v3"
}

resource "apicurio_artifact" "test" {
  group_id    = "default"
  artifact_id = "` + artifactId + `"
  type        = "AVRO"
  content     = jsonencode({"type":"record","name":"User","fields":[]})
}

resource "apicurio_artifact_rule" "test" {
  group_id    = apicurio_artifact.test.group_id
  artifact_id = apicurio_artifact.test.artifact_id
  type        = "` + ruleType + `"
  config      = "` + ruleConfig + `"
}

data "apicurio_artifact_rule" "test" {
  group_id    = apicurio_artifact_rule.test.group_id
  artifact_id = apicurio_artifact_rule.test.artifact_id
  type        = apicurio_artifact_rule.test.type
}
`
}
