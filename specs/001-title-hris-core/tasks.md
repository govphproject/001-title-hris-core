````markdown
# Tasks: HRIS Core (MVP)

**Feature Dir**: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/specs/001-title-hris-core
**Input docs used**: plan.md, research.md, data-model.md, contracts/, quickstart.md

## Execution rules
- Tests MUST be created first and must fail (TDD).
- Use optimistic locking (`version` field) on update endpoints (per research).
- Use JWT auth for MVP (backend) and role-based checks (Admin, HR, Employee).

## Phase A: Setup
- T001 Initialize backend Go module and create project skeleton
  - Files to create:
    - /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/go.mod
    - /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/main.go
    - /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/src/
  - Notes: include Gin, MongoDB Go driver, JWT libs, bcrypt, zap, viper in go.mod

- T002 Initialize frontend React app with Tailwind and router
  - Files to create:
    - /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/frontend/package.json
    - /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/frontend/src/index.jsx
    - /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/frontend/src/App.jsx
    - /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/frontend/tailwind.config.js
  - Notes: minimal CRA/Vite scaffold acceptable. Include Axios, React Router.

- T003 [P] Create `docker-compose.yml` and local MongoDB dev config
  - Files to create:
    - /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/docker-compose.yml
    - /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/.env.example
  - Purpose: make it simple to run MongoDB locally for testers.

## Phase B: Tests First (TDD) — Contract & Integration tests (must fail)
Note: Each contract file generates a contract test task [P].

- T004 [P] Contract test for employees API (based on contracts/employees.openapi.yaml)
  - Test file: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/tests/contract/test_employees_contract.go
  - Asserts: GET /employees returns 200 & JSON with `items` and `total`; POST /employees returns 201; GET/PUT/DELETE /employees/{id} behave per contract.

- T005 [P] Integration auth flow test (JWT login → access protected endpoint)
  - Test file: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/tests/integration/test_auth.go
  - Scenario: create seeded admin user, login, receive JWT, call GET /employees and expect 200.

- T006 [P] Integration test: employee CRUD (end-to-end scenario)
  - Test file: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/tests/integration/test_employee_crud.go
  - Scenario: as HR/Admin create employee, fetch employee, update (with version), delete/archive; verify audit entry present.

- T007 [P] Integration test: department CRUD and employee-department linking
  - Test file: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/tests/integration/test_department_crud.go
  - Scenario: create department, assign employee, fetch employees by department.

- T008 [P] Integration test: recruitment flow (create job posting, add applicant)
  - Test file: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/tests/integration/test_recruitment.go

- T009 [P] Integration test: payroll view for employee (employee can view their payroll history)
  - Test file: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/tests/integration/test_payroll_view.go

## Phase C: Core Implementation (after tests fail)
Ordering rules: models -> services -> endpoints -> frontend wiring

Models (create data models according to data-model.md):
- T010 [P] Create backend model: Employee
  - File: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/src/models/employee.go
  - Fields: employee_id, legal_name, preferred_name, email, phone, hire_date, termination_date, employment_status, job_history, compensation_records, manager_id, version

- T011 [P] Create backend model: UserAccount
  - File: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/src/models/user_account.go

- T012 [P] Create backend model: Role
  - File: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/src/models/role.go

- T013 [P] Create backend model: PayrollRecord
  - File: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/src/models/payroll_record.go

- T014 [P] Create backend model: LeaveRequest
  - File: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/src/models/leave_request.go

- T015 [P] Create backend model: AttendanceRecord
  - File: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/src/models/attendance_record.go

- T016 [P] Create backend model: Candidate and JobPosting
  - Files:
    - /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/src/models/candidate.go
    - /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/src/models/job_posting.go

- T017 [P] Create backend model: PerformanceReview, TrainingRecord, BenefitPlan, OrganizationUnit
  - Files under backend/src/models/

Services & Repositories:
- T018 [P] Implement MongoDB connection and repository layer (common DB client)
  - File: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/src/db/mongo.go
  - Notes: provide helper to get collection and basic CRUD with context and timeouts.

- T019 Implement Auth service (JWT issuance, middleware)
  - Files:
    - /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/src/services/auth_service.go
    - /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/src/middleware/auth_middleware.go
  - Notes: seed an Admin user during startup for tests.

- T020 [P] Implement EmployeeService with CRUD operations and optimistic locking
  - File: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/src/services/employee_service.go

- T021 Implement DepartmentService (OrganizationUnit CRUD and linking)
  - File: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/src/services/department_service.go

- T022 Implement RecruitmentService (job postings, applicants)

- T023 Implement PayrollService (CRUD and export endpoint)

- T024 Implement BenefitsService (CRUD, assign to employee)

- T025 Implement TrainingService (events, registration)

API Endpoints (Gin routes)
- T026 Implement routes for employees per contract
  - File: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/src/api/employees.go
  - Endpoints: GET /employees, POST /employees, GET/PUT/DELETE /employees/{employee_id}

- T027 Implement auth endpoints: POST /auth/login, POST /auth/logout
  - Files: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/src/api/auth.go

- T028 Implement department endpoints: /departments CRUD

- T029 Implement recruitment endpoints: /jobs, /applicants

- T030 Implement payroll endpoints: /payroll, /payroll/{id}, /payroll/export

- T031 Implement benefits endpoints: /benefits, /benefits/{id}, /employees/{id}/benefits

- T032 Implement training endpoints: /trainings, /trainings/{id}/register

Frontend minimal wiring (role-based pages)
- T033 Create frontend auth page and JWT storage
  - Files: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/frontend/src/pages/Login.jsx

- T034 Create frontend dashboard shell with sidebar and route placeholders
  - Files: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/frontend/src/components/Sidebar.jsx

- T035 Implement employees list and detail pages consuming backend endpoints
  - Files: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/frontend/src/pages/Employees.jsx

- T036 Implement department pages and linking UI

- T037 Implement job postings and applicant submission UI

- T038 Implement payroll view for employees

## Phase D: Integration & Infra
- T039 Integrate logging (zap) and structured request logging
  - File: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/src/logging/logger.go

- T040 Add CORS and security headers middleware

- T041 Seed data script for dev (employees, admin user, departments)
  - File: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/scripts/seed-dev-data.sh

## Phase E: Polish & Tests
- T042 [P] Unit tests for models and validation
  - Files: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/tests/unit/

- T043 Integration test runner script to run all integration tests locally
  - File: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/scripts/run-integration-tests.sh

- T044 Create README with local run instructions and quickstart
  - File: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/README.md

## Parallel groups (examples)
- Group 1 (can run together): T004, T005, T006, T007, T008, T009 (all tests creation) [P]
- Group 2 (models can be created in parallel): T010, T011, T012, T013, T014, T015, T016, T017 [P]
- Group 3 (model -> service): T018 (DB) must finish before Group 2 services: T019, T020, T021

## Dependency notes
- Setup tasks T001-T003 must complete before running tests or implementation.
- All contract tests (T004) must exist and fail before implementing corresponding endpoints (T026 etc.).
- Models tasks (T010-T017) must be implemented before services that use them.

## Agent commands examples
- To implement a task using an LLM agent, provide: `Implement <file path> with <behavior>` e.g.
  - Implement /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/backend/src/api/employees.go to satisfy contracts/employees.openapi.yaml
  - Create tests in /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/tests/integration/test_employee_crud.go to assert the full CRUD flow (use local MongoDB seeded data)

## Validation checklist (final gate)
- [ ] All contract files have corresponding tests (T004 covers employees contract; add more if new contracts created)
- [ ] All entities in data-model.md have model tasks (T010-T017)
- [ ] Tests exist and fail before implementation
- [ ] Each task specifies exact file paths

---
````
