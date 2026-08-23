include config/.env
export

MIGRATIONS_DIR := db/migrations
DB_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)

.PHONY: wire migrate-up migrate-down migrate-force migrate-version migrate-create run build seed

## Regenerate wire_gen.go from wire.go injectors
wire:
	wire ./cmd/api/

## Create a new migration file: make migrate-create name=create_users_table
migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

## Apply all up migrations
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

## Rollback the last migration
migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

## Force set migration version (use if the schema is dirty): make migrate-force version=1
migrate-force:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(version)

## Show current migration version
migrate-version:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

## Run the API locally
run:
	go run ./cmd/api

## Build the API binary
build:
	go build -o bin/api ./cmd/api

## Seed the owner user (idempotent)
seed:
	go run ./cmd/seed
