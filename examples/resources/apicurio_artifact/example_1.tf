resource "apicurio_artifact" "order_schema" {
  group_id    = "com.example"
  artifact_id = "order-schema"
  type        = "JSON"
  content = jsonencode({
    "$schema" = "http://json-schema.org/draft-07/schema#"
    title     = "Order"
    type      = "object"
    properties = {
      id = {
        type = "integer"
      }
      product = {
        type = "string"
      }
    }
    required = ["id", "product"]
  })
}