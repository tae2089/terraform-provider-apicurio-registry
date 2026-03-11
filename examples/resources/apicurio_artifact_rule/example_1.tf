resource "apicurio_artifact_rule" "backward_compat" {
  group_id    = "com.example"
  artifact_id = "user-schema"
  type        = "COMPATIBILITY"
  config      = "BACKWARD"
}