# PDV Backend Makefile

.PHONY: test test-unit test-e2e e2e-up e2e-down sqlc vet

# Run the full test suite (unit + e2e). Requires a PostgreSQL reachable at
# DATABASE_URL (defaults to postgres://pdv:pdv@localhost:5432/pdv?sslmode=disable).
test:
	go test -count=1 ./...

test-unit:
	go test -count=1 ./tests/unit/...

# Start the test Postgres stack (port 5433) and run the e2e suite.
# The e2e bootstrap auto-connects on localhost:5432; export DATABASE_URL to
# point at the composed instance if needed.
test-e2e:
	go test -count=1 ./tests/e2e/...

e2e-up:
	docker compose -f docker-compose.test.yml up -d

e2e-down:
	docker compose -f docker-compose.test.yml down

sqlc:
	sqlc generate

vet:
	go vet ./...
