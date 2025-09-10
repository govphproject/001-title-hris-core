Title: chore: migrate jwt tests and add payroll helper/tests

Description
-----------
This branch fixes a duplicated/corrupted payroll integration test, adds a small payroll helper and unit test, fixes a json tag in a recruitment test, and introduces a CI workflow that runs build/tests and golangci-lint.

Changes
-------
- backend/payroll_integration_test.go — cleaned duplicated content and ensured router-backed test asserts net-pay
- backend/payroll.go — added CalculateNet helper
- backend/payroll_test.go — unit test for CalculateNet
- backend/tests/integration/test_recruitment.go — fixed duplicate JSON tag for Email
- .github/workflows/ci.yml — CI workflow (build, test, golangci-lint)
- formatting commits applied to backend files (gofmt)

Testing performed
-----------------
- Ran `go test ./... -v` in `backend` — all tests passed locally including new unit and integration tests.
- Ran `gofmt` and `go vet` where applicable; formatting applied and committed.

How to create the PR (copy/paste)
---------------------------------
If you have collaborator access, run this from the repository root:

gh pr create --title "chore: migrate jwt tests and add payroll helper/tests" \
  --body "Fix duplicated payroll integration test; add payroll CalculateNet helper + unit test; fix recruitment test JSON tag; add CI workflow for build/test/lint." \
  --base main --head chore/migrate-jwt-tests --draft

Or open in browser:
https://github.com/govphproject/001-title-hris-core/compare/main...chore/migrate-jwt-tests?expand=1

Suggested reviewers
-------------------
- @your-team-lead
- @backend-maintainer

Labels (suggested)
------------------
- chore
- tests
- ci

Notes
-----
If you want I can attempt to open the PR via GitHub CLI; earlier the workspace attempted this and the account lacked permissions. This file is provided so you can open the PR manually or forward to a maintainer.
