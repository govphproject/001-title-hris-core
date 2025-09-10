````markdown
# Quickstart: HRIS Core — Phase 1

**Purpose**: Provide a short path to validate the feature spec with failing contract tests and a manual smoke check.

Prerequisites:
- MongoDB running locally (or a test instance)
- Tooling: curl or HTTP client

Steps:
1. Review `data-model.md` and ensure entity fields meet business needs.
2. Open `/contracts/` and inspect API contracts (endpoints, request/response schemas).
3. Run contract tests (these should fail until backend is implemented).
4. Manually exercise the directory endpoint with a mocked response to validate integration test expectations.

Expected outcome:
- Contract tests fail (placeholders). Data model and contracts align with user scenarios.

---
````
