.PHONY: help build run-gateway run-driver-service test lint docker-up docker-down docker-build clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build both services
	@echo "Building driver-service..."
	cd driver-service && go build -o bin/driver-service ./cmd/driver-service
	@echo "Building gateway..."
	cd gateway && go build -o bin/gateway ./cmd/gateway

run-gateway: ## Run the gateway service
	cd gateway && go run ./cmd/gateway

run-driver-service: ## Run the driver service
	cd driver-service && go run ./cmd/driver-service

test: ## Run tests
	@echo "Running driver-service tests..."
	cd driver-service && go test ./... -v
	@echo "Running gateway tests..."
	cd gateway && go test ./... -v

test-coverage: ## Run tests with coverage
	@echo "Running driver-service tests with coverage..."
	cd driver-service && go test ./... -coverprofile=coverage.out
	@echo "Running gateway tests with coverage..."
	cd gateway && go test ./... -coverprofile=coverage.out

lint: ## Run linter (requires golangci-lint)
	@echo "Linting driver-service..."
	cd driver-service && golangci-lint run
	@echo "Linting gateway..."
	cd gateway && golangci-lint run

docker-build: ## Build Docker images
	docker-compose build

docker-up: ## Start all services with Docker Compose
	docker-compose up -d

docker-down: ## Stop all services
	docker-compose down

docker-logs: ## View logs from all services
	docker-compose logs -f

docker-clean: ## Remove containers, volumes, and images
	docker-compose down -v --rmi all

clean: ## Clean build artifacts
	rm -rf driver-service/bin gateway/bin
	rm -f driver-service/coverage.out gateway/coverage.out

swagger-driver: ## Generate Swagger docs for driver-service
	cd driver-service && swag init -g cmd/driver-service/main.go -o docs --parseDependency --parseInternal

swagger-gateway: ## Generate Swagger docs for gateway
	cd gateway && swag init -g cmd/gateway/main.go -o docs --parseDependency --parseInternal

swagger: swagger-driver swagger-gateway ## Generate Swagger docs for both services

mod-tidy: ## Tidy go modules
	cd driver-service && go mod tidy
	cd gateway && go mod tidy

