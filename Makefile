.PHONY: run build test migrate-up migrate-down migrate-status sqlc lint install-tools

GOOSE_VERSION := v3.29.0
SQLC_VERSION  := v1.31.1

install-tools:
	go install github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)

ENV_FILE ?= .env
LOAD_ENV = set -a; . ./$(ENV_FILE); set +a

run:
	$(LOAD_ENV); go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:
	go test ./...

lint:
	go build ./... && go vet ./...

migrate-up:
	$(LOAD_ENV); goose -dir db/migrations postgres "$$DATABASE_URL" up

migrate-down:
	$(LOAD_ENV); goose -dir db/migrations postgres "$$DATABASE_URL" down

migrate-status:
	$(LOAD_ENV); goose -dir db/migrations postgres "$$DATABASE_URL" status

sqlc:
	sqlc generate
