.PHONY: help build run stop clean logs test docker-build docker-up docker-down docker-logs migrate

# Variables
APP_NAME=binance-live
DOCKER_COMPOSE=docker-compose
GO=go

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Go application
	@echo "Building $(APP_NAME)..."
	$(GO) build -o $(APP_NAME) ./cmd/server

run: ## Run the application locally
	@echo "Running $(APP_NAME)..."
	$(GO) run ./cmd/server -config config/config.yaml

test: ## Run tests
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(APP_NAME)
	rm -f coverage.out

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	$(DOCKER_COMPOSE) build

docker-up: ## Start all services with Docker Compose
	@echo "Starting services..."
	$(DOCKER_COMPOSE) up -d

docker-up-tools: ## Start all services including monitoring tools
	@echo "Starting services with tools..."
	$(DOCKER_COMPOSE) --profile tools up -d

docker-down: ## Stop all services
	@echo "Stopping services..."
	$(DOCKER_COMPOSE) down

docker-down-volumes: ## Stop all services and remove volumes
	@echo "Stopping services and removing volumes..."
	$(DOCKER_COMPOSE) down -v

docker-logs: ## View logs from all services
	$(DOCKER_COMPOSE) logs -f

docker-logs-app: ## View application logs
	$(DOCKER_COMPOSE) logs -f app

docker-restart: ## Restart all services
	@echo "Restarting services..."
	$(DOCKER_COMPOSE) restart

docker-ps: ## Show running containers
	$(DOCKER_COMPOSE) ps

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

fmt: ## Format code
	@echo "Formatting code..."
	$(GO) fmt ./...

lint: ## Run linters
	@echo "Running linters..."
	golangci-lint run

init: ## Initialize project (create .env file)
	@if [ ! -f .env ]; then \
		echo "Creating .env file from template..."; \
		cp .env.template .env; \
		echo ".env file created. Please update with your configuration."; \
	else \
		echo ".env file already exists."; \
	fi

setup: init deps ## Setup project for development

dev: docker-up-tools ## Start development environment with all tools

prod: docker-up ## Start production environment

