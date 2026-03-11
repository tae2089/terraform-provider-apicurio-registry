// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package artifact

import (
	providertesting "github.com/tae2089/terraform-provider-apicurio-registry/internal/testing"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccArtifactResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { providertesting.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: providertesting.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccArtifactResourceConfig("my-artifact", "AVRO", "{\"type\":\"record\",\"name\":\"User\",\"fields\":[]}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apicurio_artifact.test", "artifact_id", "my-artifact"),
					resource.TestCheckResourceAttr("apicurio_artifact.test", "type", "AVRO"),
					resource.TestCheckResourceAttr("apicurio_artifact.test", "id", "default/my-artifact"),
					resource.TestCheckResourceAttr("apicurio_artifact.test", "version", "1"),
				),
			},
			// Update testing: change content and check if version increments
			{
				Config: testAccArtifactResourceConfig("my-artifact", "AVRO", "{\"type\":\"record\",\"name\":\"UserV2\",\"fields\":[]}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apicurio_artifact.test", "artifact_id", "my-artifact"),
					resource.TestCheckResourceAttr("apicurio_artifact.test", "type", "AVRO"),
					resource.TestCheckResourceAttr("apicurio_artifact.test", "id", "default/my-artifact"),
					resource.TestCheckResourceAttr("apicurio_artifact.test", "version", "2"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "apicurio_artifact.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func testAccArtifactResourceConfig(artifactId string, artifactType string, content string) string {
	return `
provider "apicurio" {
  endpoint = "http://localhost:8080/apis/registry/v2"
}

resource "apicurio_artifact" "test" {
  group_id    = "default"
  artifact_id = "` + artifactId + `"
  type        = "` + artifactType + `"
  content     = jsonencode(` + content + `)
}
`
}
