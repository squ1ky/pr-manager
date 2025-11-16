include .env
export

COMPOSE       ?= docker compose
MIGRATE       ?= migrate
MIGRATIONS_DIR ?= migrations

DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# Cross-platform file remove
ifeq ($(OS),Windows_NT)
    RM := del /Q
else
    RM := rm -f
endif

##@ Help

.PHONY: help
help: ## Show this help message
	@echo ""
	@echo "Available Make targets:"
	@echo ""
	@echo "  compose-up           Run docker-compose in background"
	@echo "  compose-down         Stop and remove containers (no volumes)"
	@echo "  compose-restart      Restart docker-compose (down + up)"
	@echo "  docker-rm-volume     Remove containers and postgres_data volume"
	@echo ""
	@echo "  migrate-create       Create new migration (use: make migrate-create name=...)"
	@echo "  migrate-up           Apply all pending migrations"
	@echo "  migrate-down         Rollback all migrations"
	@echo "  migrate-step-up      Apply one migration step up"
	@echo "  migrate-step-down    Rollback one migration step down"
	@echo ""
	@echo "  linter-golangci      Run golangci-lint"
	@echo ""
	@echo "  test                 Run unit tests"
	@echo "  cover                Run tests and print coverage"
	@echo "  cover-html           Run tests and generate HTML coverage report"
	@echo ""

##@ Docker

compose-up: ## Run docker-compose in background
	$(COMPOSE) up -d

compose-down: ## Stop and remove containers (without volumes)
	$(COMPOSE) down

compose-restart: ## Restart docker-compose (down + up)
	$(COMPOSE) down
	$(COMPOSE) up -d

docker-rm-volume: ## Remove containers and postgres_data volume
	$(COMPOSE) down -v || true
	docker volume rm postgres_data || true

.PHONY: compose-up compose-down compose-restart docker-rm-volume

##@ Migrations

migrate-create: ## Create new migration: make migrate-create name=create_users_table
ifndef name
	$(error name is not set, use: make migrate-create name=create_users_table)
endif
	$(MIGRATE) create -seq -ext sql -dir $(MIGRATIONS_DIR) $(name)

migrate-up: ## Apply all pending migrations up
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down: ## Rollback all migrations down
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down

migrate-step-up: ## Apply one migration step up
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" steps 1

migrate-step-down: ## Rollback one migration step down
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" steps -1

.PHONY: migrate-create migrate-up migrate-down migrate-step-up migrate-step-down

##@ Lint

linter-golangci: ## Run golangci-lint
	golangci-lint run
.PHONY: linter-golangci

##@ Tests & Coverage

test: ## Run tests
	go test -v ./...
.PHONY: test

cover: ## Run tests and print coverage to console
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	$(RM) coverage.out
.PHONY: cover

cover-html: ## Run tests and generate HTML coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"
.PHONY: cover-html