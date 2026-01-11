.PHONY: help tidy test test-race lint build cover \
        migrate-up migrate-down migrate-status migrate-new sqlc

MIGRATIONS_DIR ?= $(MIGRATIONS_DIR)
DATABASE_URL ?= $(DATABASE_URL)

ifneq (,$(wildcard .env))
	include .env
	export
endif

help:  ## Вывод помощи
	@grep -E '^[0-9a-z.A-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# ===== Go =====

tidy:
	go mod tidy
	go mod verify

test:
	go test -v ./...

test-race:
	go test -race ./...

lint:
	golangci-lint run ./...

build:
	go build -trimpath -ldflags="-s -w" -o bin/urlshortener ./cmd/urlshortener

cover:
	go test -coverprofile=cover.out ./...
	go tool cover -html=cover.out
	rm cover.out

# ===== Goose (migrations) =====

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

migrate-new:
	@$(if $(name), \
		goose -dir $(MIGRATIONS_DIR) create $(name) sql, \
		$(error Usage: make migrate-new name=migration_name))

# ===== sqlc =====

sqlc:
	sqlc generate -f ./sqlc.yaml