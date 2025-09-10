Release notes — chore/migrate-jwt-tests
=====================================

Changes included in the recent work (already present on `main` / current branch):

- Add Go migration CLI: `backend/cmd/migrate`
  - Normalizes legacy employee documents (rename `employeeid` -> `employee_id`, `legalname` -> `legal_name`, `hiredate` -> `hire_date`, `preferredname` -> `preferred_name`).
  - Fills missing `employee_id` values and attempts to create a unique index.

- Remove JS migration fallback (scripts/mongo_migrate_normalize_employees.js removed). A short README in `scripts/README.md` documents the migration approach.

- Harden JWT handling in `backend/main.go`:
  - Validate signing method when parsing tokens (reject non-HMAC algs).
  - Robust parsing of `roles` claim (accept []string, []interface{}, or single string).

- Add unit tests for JWT middleware: `backend/jwt_middleware_test.go`.

- CI workflow `.github/workflows/ci-test-contract.yml`:
  - Run migration before starting backend for contract tests.
  - Add Go module caching and cleanup of migrate binary.

- Makefile: added `migrate` target and updated `test-contract` to run migrations before executing contract tests.

Verification performed locally:
- Ran migration locally and normalized legacy documents.
- Ran full backend test suite and contract tests — all passing locally.

Suggested next steps:
- (Optional) Open a PR if you'd like a formal review — current commits are already on `main`.
- (Optional) Remove any remaining legacy scripts or add CI seeding for deterministic runs.
Release notes — chore/migrate-jwt-tests
=====================================

Changes included in the recent work (already present on `main` / current branch):

- Add Go migration CLI: `backend/cmd/migrate`
  - Normalizes legacy employee documents (rename `employeeid` -> `employee_id`, `legalname` -> `legal_name`, `hiredate` -> `hire_date`, `preferredname` -> `preferred_name`).
  - Fills missing `employee_id` values and attempts to create a unique index.

- Remove JS migration fallback (scripts/mongo_migrate_normalize_employees.js removed). A short README in `scripts/README.md` documents the migration approach.

- Harden JWT handling in `backend/main.go`:
  - Validate signing method when parsing tokens (reject non-HMAC algs).
  - Robust parsing of `roles` claim (accept []string, []interface{}, or single string).

- Add unit tests for JWT middleware: `backend/jwt_middleware_test.go`.

- CI workflow `.github/workflows/ci-test-contract.yml`:
  - Run migration before starting backend for contract tests.
  - Add Go module caching and cleanup of migrate binary.

- Makefile: added `migrate` target and updated `test-contract` to run the migration before executing contract tests.

Verification performed locally:
- Ran migration locally and normalized legacy documents.
- Ran full backend test suite and contract tests — all passing locally.

Suggested next steps:
- (Optional) Open a PR if you'd like a formal review — current commits are already on `main`.
- (Optional) Remove any remaining legacy scripts or add CI seeding for deterministic runs.
