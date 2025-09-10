#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

# if port 27017 is already accepting TCP connections, assume local mongod is running and use it
if nc -z localhost 27017 >/dev/null 2>&1; then
  echo "Found existing process listening on 27017; using local Mongo instance"
  STARTED_MONGO=0
else
  echo "Starting dockerized Mongo"
  docker-compose up -d mongo

  echo "Waiting for Mongo to be ready..."
  for i in {1..30}; do
    if docker exec hris-mongo mongo --eval "db.adminCommand('ping')" &>/dev/null; then
      echo "Mongo is ready"
      break
    fi
    sleep 1
  done
  STARTED_MONGO=1
fi

# run backend in background
cd "$REPO_ROOT/backend"
HRIS_JWT_SECRET="dev-jwt-secret" MONGO_URI="mongodb://localhost:27017" MONGO_DB="hris" MONGO_COLLECTION="employees" go run main.go &
BACK_PID=$!

# wait for backend health
for i in {1..20}; do
  if curl -sS http://localhost:8080/health | grep -q "ok"; then
    echo "Backend is ready"
    break
  fi
  sleep 1
done

# run contract tests against the real server
export HRIS_TEST_EXTERNAL=1
cd "$REPO_ROOT/backend"
go test ./tests/contract -v

# cleanup
kill $BACK_PID || true

if [ "${STARTED_MONGO:-0}" -eq 1 ]; then
  echo "Stopping docker mongo"
  docker-compose down
fi

echo "Done"
