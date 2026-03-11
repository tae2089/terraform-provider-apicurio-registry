terraform {
  required_providers {
    apicurio = {
      source = "tae2089/apicurio-registry"
    }
  }
}
provider "apicurio" {
  endpoint = "http://localhost:8080/apis/registry/v3"
  keycloak_server_url = ""
  keycloak_client_id = ""
  keycloak_client_secret = ""
  keycloak_realm = ""
}

resource "apicurio_artifact" "example_avro" {
  group_id    = "default3"
  artifact_id = "user-schema"
  type        = "AVRO"
  content     = jsonencode({
    type = "record"
    name = "User"
    fields = [
      { name = "id", type = "string" },
      { name = "name", type = "string" },
      { name = "age", type = "int" }
    ]
  })
}

resource "apicurio_artifact" "example_avro2" {
  group_id    = "default2"
  artifact_id = "user-schema2"
  type        = "AVRO"
  content     = jsonencode({
    type = "record"
    name = "User2"
    fields = [
      { name = "id", type = "string" },
      { name = "name", type = "string" }
    ]
  })
}

# Rule to check for backward compatibility when versions are updated
resource "apicurio_artifact_rule" "backward_compatibility" {
  group_id    = apicurio_artifact.example_avro.group_id
  artifact_id = apicurio_artifact.example_avro.artifact_id
  type        = "COMPATIBILITY"
  config      = "BACKWARD"
  depends_on = [ apicurio_artifact.example_avro2 ]
}

# Rule to strictly enforce full syntax validation
resource "apicurio_artifact_rule" "full_validity" {
  group_id    = apicurio_artifact.example_avro.group_id
  artifact_id = apicurio_artifact.example_avro.artifact_id
  type        = "VALIDITY"
  config      = "FULL"
  
}
