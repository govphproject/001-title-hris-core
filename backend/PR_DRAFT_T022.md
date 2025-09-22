Title: feat(recruitment): implement RecruitmentService (T022) + API routes and tests

Summary

This branch implements T022 (RecruitmentService) including:

- In-memory recruitment repository (`src/services/recruitment_repo.go`) supporting jobs and candidates with thread-safe maps.
- Business service (`src/services/recruitment_service.go`) with methods to post jobs, apply candidates, list/get/update/delete resources and basic validation.
- API routes (`src/api/recruitment.go`) registering endpoints under `/api/jobs` and `/api/applicants` for create/list operations (basic handlers).
- Unit tests (`src/services/recruitment_service_test.go`) for the service.
- Integration test (`tests/integration/test_recruitment_end_to_end_test.go`) exercising the recruitment flow.
- Minor fix: added GET `/api/departments/:id/employees` to return department employee list (integration test requirement).

Files changed (high level)

- Added: `src/services/recruitment_repo.go`
- Added: `src/services/recruitment_service.go`
- Added: `src/services/recruitment_service_test.go`
- Added: `src/api/recruitment.go`
- Modified: `main.go` (register recruitment routes)
- Added: `tests/integration/test_recruitment_end_to_end_test.go`
- Modified: `src/api/departments.go` (GET employees endpoint)

Testing

- Run `cd backend && go test ./... -v` — all backend tests (unit, integration, contract) pass.

Notes for reviewers

- This implements an in-memory repo for recruitment; if persistence is required, add a Mongo-backed `RecruitmentRepo` similar to other repos.
- Endpoints are intentionally minimal (create/list). Additional endpoints (PUT/DELETE, query filters, paging) can be added in follow-ups.
- The APIs return simple JSON responses consistent with other API handlers in the project.

Suggested PR title

"feat(recruitment): implement RecruitmentService (T022) with API routes and tests"

Suggested reviewers

- Backend owners and anyone responsible for services and API contracts.

Merge notes

- No DB migrations required for in-memory defaults.
- After merge consider opening follow-ups to implement persistent storage and richer API behavior.
