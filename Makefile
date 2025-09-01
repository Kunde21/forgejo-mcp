# Makefile for Forgejo MCP Server
# Standard build, test, and deployment targets

.PHONY: help build test clean install lint vet fmt mod-tidy coverage docker-build docker-run release

# Default target
help: ## Show this help message
	@echo "Forgejo MCP Server - Build and Development Commands"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the application for current platform
	@echo "Building Forgejo MCP server..."
	@go build -o bin/forgejo-mcp ./cmd
	@echo "Build complete: bin/forgejo-mcp"

build-all: ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	@mkdir -p bin
	@GOOS=linux GOARCH=amd64 go build -o bin/forgejo-mcp-linux-amd64 ./cmd
	@GOOS=darwin GOARCH=amd64 go build -o bin/forgejo-mcp-darwin-amd64 ./cmd
	@GOOS=darwin GOARCH=arm64 go build -o bin/forgejo-mcp-darwin-arm64 ./cmd
	@GOOS=windows GOARCH=amd64 go build -o bin/forgejo-mcp-windows-amd64.exe ./cmd
	@echo "Multi-platform builds complete in bin/"

# Test targets
test: ## Run all tests
	@echo "Running all tests..."
	@go test ./...

test-verbose: ## Run all tests with verbose output
	@echo "Running all tests (verbose)..."
	@go test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-unit: ## Run only unit tests (skip integration/e2e)
	@echo "Running unit tests..."
	@SKIP_DOCKER_TESTS=true go test ./... -short

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test ./server -run TestAuthIntegration
	@go test ./test/e2e -run TestCleanupProcedures

test-e2e: ## Run end-to-end tests (requires Docker)
	@echo "Running E2E tests..."
	@go test ./test/e2e -v

# Code quality targets
lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@golangci-lint run

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

fmt: ## Format code with goimports
	@echo "Formatting code..."
	@goimports -w .

mod-tidy: ## Clean up go.mod and go.sum
	@echo "Tidying modules..."
	@go mod tidy

# Development targets
install: ## Install the binary to $GOPATH/bin
	@echo "Installing binary..."
	@go install ./cmd

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@find . -name "*.test" -delete

dev: ## Start development server
	@echo "Starting development server..."
	@go run ./cmd serve

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t forgejo-mcp:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run --rm -p 3000:3000 forgejo-mcp:latest

# Release targets
release: ## Create a new release
	@echo "Creating release..."
	@./scripts/release.sh

release-dry-run: ## Preview release without publishing
	@echo "Dry run release..."
	@./scripts/release.sh --dry-run

# CI/CD targets
ci: ## Run full CI pipeline locally
	@echo "Running CI pipeline..."
	@make mod-tidy
	@make fmt
	@make vet
	@make lint
	@make test-coverage
	@make build

# Utility targets
deps: ## Install development dependencies
	@echo "Installing development dependencies..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

version: ## Show version information
	@echo "Forgejo MCP Server"
	@git describe --tags --always --dirty
	@echo "Go version: $(shell go version)"
	@echo "Build time: $(shell date)"

# Security scanning
security-scan: ## Run security vulnerability scan
	@echo "Running security scan..."
	@gosec ./...

# Performance testing
bench: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Help target (automatically generated above)