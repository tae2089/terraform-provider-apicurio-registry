// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package artifact_test

import (
	providertesting "github.com/tae2089/terraform-provider-apicurio-registry/internal/testing"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccArtifactDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { providertesting.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: providertesting.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArtifactDataSourceConfig("ds-artifact", "AVRO", "{\"type\":\"record\",\"name\":\"User\",\"fields\":[]}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apicurio_artifact.test", "artifact_id", "ds-artifact"),
					resource.TestCheckResourceAttr("data.apicurio_artifact.test", "type", "AVRO"),
					resource.TestCheckResourceAttrSet("data.apicurio_artifact.test", "content"),
					resource.TestCheckResourceAttrSet("data.apicurio_artifact.test", "version"),
				),
			},
		},
	})
}

func testAccArtifactDataSourceConfig(artifactId, artifactType, content string) string {
	return `
provider "apicurio" {
  endpoint = "http://localhost:8080/apis/registry/v3"
}

resource "apicurio_artifact" "test" {
  group_id    = "default"
  artifact_id = "` + artifactId + `"
  type        = "` + artifactType + `"
  content     = jsonencode(` + content + `)
}

data "apicurio_artifact" "test" {
  group_id    = apicurio_artifact.test.group_id
  artifact_id = apicurio_artifact.test.artifact_id
}
`
}
