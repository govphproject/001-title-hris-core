````markdown
# Research: HRIS Core — Phase 0

**Spec**: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/specs/001-title-hris-core/spec.md
**Date**: 2025-09-08

## Purpose
Resolve open questions from the spec and gather decisions/recommendations to feed Phase 1 design.

## Resolved Decisions (recommended defaults)
- Auth method: Recommend supporting SSO (SAML/OAuth with IdP) as primary, with email/password + optional MFA as fallback. Rationale: enterprises often prefer SSO; fallback enables small orgs.
- Payroll scope: Start by supporting payroll record-keeping and export (CSV/Excel) and defer full payroll calculation engine to a later phase. Rationale: reduces legal/tax risk early.
- Imports: Support CSV imports for timekeeping and attendance as v1. Provide a connector API for future integrations.
- Roles: Start with canonical roles: HR Admin, Manager, Payroll, Employee, Recruiter. Provide role-permission matrix in Phase 1.
- Data retention: Default to retain active employee data indefinitely; archive terminated employee data after 7 years (configurable). [NEEDS CLARIFICATION: legal/jurisdiction-specific retention rules]
- Scale target: initial target 5k employees, design for horizontal scaling to 100k+ with sharding/partitioning options. [NEEDS CLARIFICATION: exact SLAs]

## Open Research Questions (need follow-up or user decision)
1. Conflict resolution for concurrent edits (last-write-wins vs optimistic locking vs merge UI).
2. Retroactive payroll adjustments policy and desired UI/workflow.
3. Multi-country payroll/tax support: which countries/currencies must be supported at launch.
4. Exact role permission matrix and approval hierarchy (single vs multi-level approvals).
5. Legal retention policies per jurisdiction.
6. Required reporting metrics and frequency (daily/weekly/monthly/quarterly exports).

## Research Tasks (Phase 0 output)
- Task: Research best practices for enterprise SSO integration (SAML, OIDC) and how to offer fallback email/password + MFA.
- Task: Research CSV import formats for common timekeeping systems (e.g., TSheets, Kronos, Deputy) and propose a minimal import schema.
- Task: Research payroll compliance concerns and the implications of performing calculations vs exporting raw data.
- Task: Research concurrency control patterns for master data in web apps and recommended approach for HR data.
- Task: Identify data retention legal minimums for common jurisdictions (US, EU) and produce config options.
- Task: Draft initial role-permission matrix template for HR Admin, Manager, Payroll, Employee, Recruiter.

## Decisions to carry into Phase 1
- Implement SSO primary + email/password fallback (configurable). Document auth choices in Technical Context.
- Implement payroll export-first approach (no calculations) initially.
- Support CSV import for attendance/time as v1.
- Use optimistic locking with version numbers by default; confirm with stakeholders.

## Research Findings (short summaries)
- SSO recommendation: OIDC is widely supported and simpler to implement than SAML for modern IdPs; however SAML remains required for some enterprises. Support both via an abstraction layer.
- CSV import: A minimal schema with employee_id, timestamp, type (clock-in/clock-out), source_system is broadly compatible.
- Concurrency: Optimistic locking (version field + conflict error) is standard and enables a later merge UI.

## Output
- File produced: /Users/ronaldpalay/Sourcecode/Projects/WebDevelopment/hris/specs/001-title-hris-core/research.md
- Next: Phase 1 will consume `research.md` to generate `data-model.md`, `quickstart.md`, and `/contracts/` artifacts.

---
````
