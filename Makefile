# ================================================================
# OPTRION Makefile
# ================================================================
# Commands:
#   make build       — Build the binary
#   make run         — Run locally
#   make test        — Run all tests
#   make test-cover  — Run tests with coverage
#   make lint        — Run linter
#   make docker-up   — Start Docker Compose stack
#   make docker-down — Stop Docker Compose stack
#   make clean       — Remove build artifacts
# ================================================================

.PHONY: build run test test-cover lint docker-up docker-down clean help

# Variables
APP_NAME := optrion
BUILD_DIR := ./bin
MAIN_PATH := ./cmd/optrion
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"
DOCKER_COMPOSE := docker compose -f deploy/docker/docker-compose.yml

## help: Display this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/ /'

## build: Build the application binary
build:
	@echo "Building $(APP_NAME)..."
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "Binary: $(BUILD_DIR)/$(APP_NAME)"

## run: Run the application locally (requires .env.local sourced)
run: build
	@echo "Starting $(APP_NAME)..."
	@$(BUILD_DIR)/$(APP_NAME)

## test: Run all tests
test:
	@echo "Running tests..."
	@go test ./... -v -race -count=1

## test-cover: Run tests with coverage report
test-cover:
	@echo "Running tests with coverage..."
	@go test ./... -race -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out
	@echo "Coverage report: coverage.out"

## test-short: Run unit tests only (skip integration)
test-short:
	@echo "Running unit tests..."
	@go test ./... -short -v -race -count=1

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

## fmt: Format all Go files
fmt:
	@echo "Formatting..."
	@gofmt -s -w .
	@goimports -w .

## tidy: Tidy Go modules
tidy:
	@go mod tidy

## docker-up: Start all services with Docker Compose
docker-up:
	@echo "Starting Docker Compose stack..."
	@$(DOCKER_COMPOSE) up -d
	@echo "Stack is running. App at http://localhost:8080"

## docker-down: Stop and remove Docker Compose stack
docker-down:
	@echo "Stopping Docker Compose stack..."
	@$(DOCKER_COMPOSE) down

## docker-logs: Follow application logs
docker-logs:
	@$(DOCKER_COMPOSE) logs -f app

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -f deploy/docker/Dockerfile -t $(APP_NAME):$(VERSION) .

## infra-up: Start only infrastructure (postgres, redis) for local development
infra-up:
	@echo "Starting infrastructure..."
	@$(DOCKER_COMPOSE) up -d postgres redis
	@echo "PostgreSQL: localhost:5432 | Redis: localhost:6379"

## infra-down: Stop infrastructure
infra-down:
	@$(DOCKER_COMPOSE) down postgres redis

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out
	@echo "Clean complete"
