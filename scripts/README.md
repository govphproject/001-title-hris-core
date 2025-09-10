Migration scripts
=================

This folder contains utilities to migrate or normalize data in the development MongoDB used by the project.

Current recommended migration (Go):

- `backend/cmd/migrate` is a Go CLI that normalizes legacy employee documents (rename `employeeid` -> `employee_id`, etc.), fills missing `employee_id` values, and attempts to create a unique index on `employee_id`.

Run it locally with the Makefile target or directly:

```bash
# using Makefile
make migrate

# or directly (from repo root)
cd backend
go build ./cmd/migrate
MONGO_URI="mongodb://localhost:27017" ./migrate
rm -f migrate
```

Notes:
- A previous JS migration (`mongo_migrate_normalize_employees.js`) existed as a fallback. The Go migration is now the canonical tool and the JS file has been removed to avoid duplication.
- The migration is idempotent and safe to run multiple times, but run it against a test/backup of production data if you're unsure.
