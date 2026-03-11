provider "apicurio" {
  endpoint               = "http://localhost:8080/apis/registry/v3"
  keycloak_server_url    = "" #optional
  keycloak_client_id     = "" #optional
  keycloak_client_secret = "" #optional
  keycloak_realm         = "" #optional
}
