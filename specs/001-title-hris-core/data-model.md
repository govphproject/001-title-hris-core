````markdown
# Data Model: HRIS Core — Phase 1

**Source Spec**: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/specs/001-title-hris-core/spec.md
**Decisions**: Auth: SSO primary + email/password fallback; Payroll: export-first; Concurrency: optimistic locking (version field)

## Entities

### Employee
- employee_id: string (canonical)
- legal_name: {first, middle?, last}
- preferred_name: string?
- email: string
- phone: string?
- address: object
- national_identifiers: array (type, value) [sensitive]
- hire_date: date
- termination_date: date?
- employment_status: enum (active, on_leave, terminated)
- job_history: array of {title, department_id, start_date, end_date}
- compensation_records: array of {effective_date, salary_amount, currency, pay_grade}
- manager_id: employee_id?
- metadata: map
- version: integer (for optimistic locking)

### UserAccount
- account_id: string
- employee_id: string?
- username/email: string
- roles: array of Role identifiers
- last_login: timestamp
- status: enum (active, suspended)

### Role
- role_id: string
- name: string
- permissions: array of permission keys

### PayrollRecord
- payroll_id: string
- employee_id: string
- pay_period_start: date
- pay_period_end: date
- gross_pay: number
- deductions: array
- taxes: array
- net_pay: number
- adjustments: array
- exported: boolean

### LeaveRequest
- request_id: string
- employee_id: string
- type: enum (vacation, sick, unpaid, etc.)
- start_date: date
- end_date: date
- duration_days: number
- status: enum (pending, approved, rejected, canceled)
- approver_history: array of {approver_id, action, timestamp, comment}

### AttendanceRecord
- attendance_id: string
- employee_id: string
- timestamp: datetime
- type: enum (clock_in, clock_out, imported)
- source_system: string?

### Candidate
- candidate_id: string
- name, contacts, resume_ref, stage_history, outcome

### PerformanceReview
- review_id: string
- employee_id, reviewer_id, cycle, goals, ratings, comments, status

### TrainingRecord
- training_id, employee_id, course_id, completion_date, certificate_ref, expiry_date

### BenefitPlan
- plan_id, name, coverage_summary, eligibility_criteria

### OrganizationUnit
- org_id, name, parent_id?, location, cost_center

## Indexing & Queries (suggested)
- Index employee_id, email, department, manager_id for directory queries.
- TTL/archival indices for terminated employees based on retention policy.

## Validation Rules (high level)
- employee_id required and unique.
- email must be RFC-compliant when provided.
- hire_date <= termination_date when termination_date present.

---
````
