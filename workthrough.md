# Workthrough Log

- Initialized tracking files (`task.md`, `implementation.md`, `workthrough.md`).
- Will now write an acceptance test for `apicurio_artifact` resource to begin TDD cycle (Red -> Green -> Refactor).
- [2026년  3월 10일 화요일 21시 19분 11초 KST] Wrote failing test for `apicurio_artifact`.
- [2026년  3월 10일 화요일 21시 19분 11초 KST] Created basic skeleton for `apicurio_artifact` to pass the acceptance test (Red -> Green).
- [2026년  3월 10일 화요일 21시 19분 11초 KST] Next: Refactor to add real API calls, which requires passing the Endpoint from the Provider to the Resource.
- [2026년  3월 10일 화요일 21시 23분 54초 KST] Implemented real REST API calls (POST, GET, PUT, DELETE) in `artifact_resource.go`.
- [2026년  3월 10일 화요일 21시 23분 54초 KST] Added `ApicurioClient` in `client.go` and updated `provider.go` to pass it to resources.
- [2026년  3월 10일 화요일 21시 23분 54초 KST] Added example HCL in `examples/resources/apicurio_artifact/resource.tf`.
- [2026년  3월 10일 화요일 21시 44분 00초 KST] Started implementation of `apicurio_artifact_rule` resource based on Apicurio Registry v2 API.
- [2026년  3월 10일 화요일 21시 47분 09초 KST] Successfully implemented `apicurio_artifact_rule` resource. Checked Create/Read/Update/Delete operations and verified integration with Apicurio Registry via Acceptance Tests.
