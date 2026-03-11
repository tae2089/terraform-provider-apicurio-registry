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
