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

# Example: Fetching an existing AVRO artifact
data "apicurio_artifact" "example" {
  group_id    = "default"
  artifact_id = "user-schema"
}

output "artifact_content" {
  value = data.apicurio_artifact.example.content
}

output "artifact_type" {
  value = data.apicurio_artifact.example.type
}

output "artifact_version" {
  value = data.apicurio_artifact.example.version
}
