#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MONGO_URI="${MONGO_URI:-mongodb://localhost:27017}"

echo "Running employee normalization migration against $MONGO_URI"
# mongosh requires connection string and a --file argument
mongosh "$MONGO_URI" --quiet --file "$SCRIPT_DIR/mongo_migrate_normalize_employees.js"

echo "Migration finished."
