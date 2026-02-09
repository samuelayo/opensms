.PHONY: help build run test clean migrate-up migrate-down migrate-create docker-up docker-down

# Default target
help:
	@echo "OpenSMS Backend - Makefile Commands"
	@echo "===================================="
	@echo "make build         - Build the application"
	@echo "make run           - Run the application"
	@echo "make test          - Run tests"
	@echo "make lint          - Run linter"
	@echo "make clean         - Clean build artifacts"
	@echo "make migrate-up    - Run database migrations up"
	@echo "make migrate-down  - Run database migrations down"
	@echo "make migrate-create NAME=<name> - Create new migration"
	@echo "make docker-up     - Start Docker services"
	@echo "make docker-down   - Stop Docker services"
	@echo "make dev           - Run in development mode with hot reload"

# Build the application
build:
	@echo "Building OpenSMS API..."
	go build -o bin/api cmd/api/main.go

# Run the application
run: build
	@echo "Running OpenSMS API..."
	./bin/api

# Run in development mode (with air for hot reload if installed)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Installing air for hot reload..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run --timeout=5m

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Database migrations
migrate-up:
	@echo "Running migrations up..."
	goose -dir migrations postgres "host=localhost port=5432 user=opensms password=opensms_dev_password dbname=opensms sslmode=disable" up

migrate-down:
	@echo "Running migrations down..."
	goose -dir migrations postgres "host=localhost port=5432 user=opensms password=opensms_dev_password dbname=opensms sslmode=disable" down

migrate-status:
	@echo "Migration status..."
	goose -dir migrations postgres "host=localhost port=5432 user=opensms password=opensms_dev_password dbname=opensms sslmode=disable" status

migrate-create:
	@echo "Creating migration: $(NAME)"
	goose -dir migrations create $(NAME) sql

# Docker commands
docker-up:
	@echo "Starting Docker services..."
	cd .. && docker-compose up -d

docker-down:
	@echo "Stopping Docker services..."
	cd .. && docker-compose down

docker-logs:
	@echo "Showing Docker logs..."
	cd .. && docker-compose logs -f

# Development setup
setup:
	@echo "Setting up development environment..."
	@echo "1. Installing Go dependencies..."
	go mod download
	@echo "2. Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "3. Copying .env.example to .env..."
	cp .env.example .env
	@echo "Setup complete! Edit .env file with your configuration."

# Generate code (if using code generation)
generate:
	@echo "Generating code..."
	go generate ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
