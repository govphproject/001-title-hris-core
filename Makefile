.PHONY: up down test-contract test-unit

up:
	docker-compose up -d mongo

down:
	docker-compose down

test-contract: up
	# ensure DB normalized before running external contract tests
	cd backend && go build ./cmd/migrate && MONGO_URI="mongodb://localhost:27017" ./migrate && rm -f migrate
	HRIS_TEST_EXTERNAL=1 MONGO_URI="mongodb://localhost:27017" make _test-contract-run

migrate:
	cd backend && go build ./cmd/migrate && MONGO_URI="mongodb://localhost:27017" ./migrate && rm -f migrate

_test-contract-run:
	cd backend && go test ./tests/contract -v

test-unit:
	cd backend && go test ./... 
