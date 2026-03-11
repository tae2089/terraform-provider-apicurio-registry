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

# Example: Fetching an existing artifact rule
data "apicurio_artifact_rule" "example" {
  group_id    = "default"
  artifact_id = "user-schema"
  type        = "VALIDITY"
}

output "rule_config" {
  value = data.apicurio_artifact_rule.example.config
}
