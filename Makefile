# variables
GOOSE := $(shell command -v goose 2> /dev/null || echo ~/go/bin/goose)
LOCAL_DB_URL := "host=localhost port=5432 user=developer dbname=core sslmode=disable"

.PHONY: run-api generate-migration migrate-up-local local-db-rebuild test-db-rebuild migrate-up-test local-db-seed dep generate-mocks

build:
	@go build ./cmd/api

run-api: build
	./api

generate-migration:
	@read -p "Enter migration name: " name; \
	$(GOOSE) -dir db/migrations create $$name sql

# Local Database Management
migrate-up-local:
	@$(GOOSE) -dir db/migrations postgres "host=localhost port=5432 user=developer dbname=core sslmode=disable" up

local-db-clean:
	@psql -U postgres -d core -c "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = 'core' AND pid <> pg_backend_pid();" >& /dev/null || true
	@dropdb -U postgres core || true
	@createdb -U postgres core

local-db-seed:
	@go run ./scripts/seed.go

local-db-rebuild: local-db-clean migrate-up-local local-db-seed

# Test Database Management
migrate-up-test:
	@$(GOOSE) -dir db/migrations postgres "host=localhost port=5432 user=developer dbname=core_test sslmode=disable" up

test-db-clean:
	@psql -U postgres -d core_test -c "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = 'core_test' AND pid <> pg_backend_pid();" >& /dev/null || true
	@dropdb -U postgres core_test || true
	@createdb -U postgres core_test

test-db-rebuild: test-db-clean migrate-up-test

deps:
	@go mod download
	@go install "github.com/vektra/mockery/v3@v3.7.1"

generate-mocks: deps
	@mockery --log-level error