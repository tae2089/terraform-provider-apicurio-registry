# Tasks: Implement apicurio_artifact resource

- [x] Create project tracking files (task.md, implementation.md, workthrough.md)
- [x] Write a failing test for `apicurio_artifact` resource (Create & Read).
- [x] Implement `apicurio_artifact` skeleton and minimal `Schema`, `Create`, `Read` to make test pass.
- [x] Implement actual API calls to Apicurio Registry for Create/Read.
- [x] Implement `Update` and `Delete` logic with actual API calls.
- [x] Add `examples/resources/apicurio_artifact/resource.tf`.
- [x] Refactor resource implementation (if needed).

# Tasks: Implement apicurio_artifact_rule resource

- [x] Create `internal/provider/resource_artifact_rule.go` skeleton.
- [x] Create `internal/provider/resource_artifact_rule_test.go` with a failing test (Red).
- [x] Implement actual API calls for `apicurio_artifact_rule` (Green).
- [x] Register `NewArtifactRuleResource` in `provider.go`.
- [x] Run Acceptance Tests to verify rule management with Docker Compose.
- [x] Add example to `examples/resources/apicurio_artifact_rule/resource.tf`.

# Tasks: Implement apicurio_artifact data source

- [x] Create `internal/provider/data_source_artifact.go`.
- [x] Register `NewArtifactDataSource` in `provider.go`.
- [x] Create `internal/provider/data_source_artifact_test.go` and pass acceptance tests.
- [x] Add example to `examples/data-sources/apicurio_artifact/data-source.tf`.
