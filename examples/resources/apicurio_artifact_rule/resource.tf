provider "apicurio" {
  endpoint = "http://localhost:8080/apis/registry/v2"
}

resource "apicurio_artifact" "example_avro" {
  group_id    = "default"
  artifact_id = "user-schema"
  type        = "AVRO"
  content     = jsonencode({
    type = "record"
    name = "User"
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
}

# Rule to strictly enforce full syntax validation
resource "apicurio_artifact_rule" "full_validity" {
  group_id    = apicurio_artifact.example_avro.group_id
  artifact_id = apicurio_artifact.example_avro.artifact_id
  type        = "VALIDITY"
  config      = "FULL"
}
