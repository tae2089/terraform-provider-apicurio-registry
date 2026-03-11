resource "apicurio_artifact_rule" "full_compat" {
  group_id    = "com.example"
  artifact_id = "user-schema"
  type        = "COMPATIBILITY"
  config      = "FULL"
}