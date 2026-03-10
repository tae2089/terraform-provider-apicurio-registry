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

resource "apicurio_artifact" "example_openapi" {
  group_id    = "default"
  artifact_id = "petstore-api"
  type        = "OPENAPI"
  content     = <<EOT
openapi: 3.0.0
info:
  title: Petstore API
  version: 1.0.0
paths:
  /pets:
    get:
      responses:
        '200':
          description: OK
EOT
}
