````markdown
# Contract Tests (placeholders) — HRIS Core

These tests are generated from API contracts and should fail until the backend implements the API.

1. GET /employees returns 200 and JSON with `items` array and `total` integer.
2. POST /employees with valid payload returns 201 and created employee with `employee_id`.
3. GET /employees/{employee_id} returns 200 for existing employee.
4. PUT /employees/{employee_id} with correct `version` updates record and returns 200.
5. DELETE /employees/{employee_id} returns 204 and archives employee.

Implementations should provide automated tests mirroring these scenarios.

---
````
