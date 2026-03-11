terraform {
  required_providers {
    apicurio = {
      source = "tae2089/apicurio-registry"
    }
  }
}

provider "apicurio" {
  endpoint = "http://localhost:8080/apis/registry/v3"
}

resource "apicurio_artifact" "example_avro" {
  group_id    = "default"
  artifact_id = "user-schema"
  type        = "AVRO"
  content = jsonencode({
    type = "record"
    name = "User"
    fields = [
      { name = "id", type = "string" },
      { name = "name", type = "string" },
    ]
  })
}