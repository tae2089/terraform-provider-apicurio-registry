resource "apicurio_artifact" "payment_schema" {
  group_id    = "com.example"
  artifact_id = "payment-schema"
  type        = "AVRO"
  content     = file("${path.module}/schemas/payment.avsc")
}

resource "apicurio_artifact_rule" "payment_compat" {
  group_id    = apicurio_artifact.payment_schema.group_id
  artifact_id = apicurio_artifact.payment_schema.artifact_id
  type        = "COMPATIBILITY"
  config      = "BACKWARD_TRANSITIVE"
}

resource "apicurio_artifact_rule" "payment_validity" {
  group_id    = apicurio_artifact.payment_schema.group_id
  artifact_id = apicurio_artifact.payment_schema.artifact_id
  type        = "VALIDITY"
  config      = "SYNTAX_ONLY"
}