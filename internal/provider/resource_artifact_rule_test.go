// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccArtifactRuleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccArtifactRuleResourceConfig("my-rule-artifact", "VALIDITY", "FULL"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apicurio_artifact_rule.test", "artifact_id", "my-rule-artifact"),
					resource.TestCheckResourceAttr("apicurio_artifact_rule.test", "type", "VALIDITY"),
					resource.TestCheckResourceAttr("apicurio_artifact_rule.test", "config", "FULL"),
				),
			},
			// Update testing
			{
				Config: testAccArtifactRuleResourceConfig("my-rule-artifact", "VALIDITY", "SYNTAX_ONLY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apicurio_artifact_rule.test", "config", "SYNTAX_ONLY"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "apicurio_artifact_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccArtifactRuleResourceConfig(artifactId string, ruleType string, ruleConfig string) string {
	return `
provider "apicurio" {
  endpoint = "http://localhost:8080/apis/registry/v2"
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
`
}
