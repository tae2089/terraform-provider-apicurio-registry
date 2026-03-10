# Implementation Details: `apicurio_artifact`

## Objective
A Terraform resource for creating and managing artifacts in Apicurio Registry.

## Technical Design
- The resource will interact with the Apicurio Registry REST API using the standard HTTP client configured in `provider.go`.
- Endpoint: `POST /groups/{groupId}/artifacts` (for create), `GET /groups/{groupId}/artifacts/{artifactId}` (for read), etc.
- If the `provider` does not pass down `endpoint` configuration, we will need to update the `Configure` method of the provider to capture and provide it to the client, or pass a custom struct holding the endpoint and HTTP client.

## Data Model
- `group_id` (String): The ID of the group in which to create the artifact. Default: "default".
- `artifact_id` (String): The ID of the artifact. Optional, computed if omitted.
- `content` (String): The contents of the artifact (e.g., Avro schema, JSON schema).
- `type` (String): The type of the artifact.
- `version` (String, Computed): The version of the artifact created.
