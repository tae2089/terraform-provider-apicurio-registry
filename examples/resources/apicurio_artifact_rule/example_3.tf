resource "apicurio_artifact_rule" "syntax_check" {
  group_id    = "com.example"
  artifact_id = "user-schema"
  type        = "VALIDITY"
  config      = "SYNTAX_ONLY"
}