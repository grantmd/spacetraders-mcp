# SpaceTraders MCP Server Makefile

.PHONY: build test test-unit test-integration clean help dev

# Default target
all: build test

# Build the server binary
build:
	@echo "Building SpaceTraders MCP Server..."
	go build -o spacetraders-mcp .

# Run all tests using the Go test runner
test:
	@echo "Running all tests with Go test runner..."
	go run ./cmd/test_runner.go

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	go test -v ./pkg/...

# Run integration tests (requires SPACETRADERS_API_TOKEN)
test-integration:
	@echo "Running integration tests..."
	@if [ -z "$(SPACETRADERS_API_TOKEN)" ]; then \
		echo "Warning: SPACETRADERS_API_TOKEN not set, integration tests will be skipped"; \
	fi
	go test -v ./test/...

# Run full integration tests with real API calls
test-full:
	@echo "Running full integration tests (including real API calls)..."
	go run ./cmd/test_runner.go --integration

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./pkg/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests in short mode (skip long-running tests)
test-short:
	@echo "Running short tests..."
	go test -short -v ./pkg/...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f spacetraders-mcp
	rm -f spacetraders-mcp-test
	rm -f coverage.out
	rm -f coverage.html

# Development mode - build and run with example request
dev: build
	@echo "Starting development server..."
	@echo "Send a test request in 2 seconds..."
	@sleep 2 && echo '{"jsonrpc": "2.0", "id": 1, "method": "resources/list"}' | ./spacetraders-mcp &

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Check for security vulnerabilities
security:
	@echo "Checking for security vulnerabilities..."
	gosec ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./pkg/...

# Install development dependencies
install-dev-deps:
	@echo "Installing development dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Quick test with real API (requires valid token)
quick-test: build
	@if [ -z "$(SPACETRADERS_API_TOKEN)" ]; then \
		echo "Error: SPACETRADERS_API_TOKEN environment variable is required"; \
		exit 1; \
	fi
	@echo "Testing agent info resource..."
	@echo '{"jsonrpc": "2.0", "id": 1, "method": "resources/read", "params": {"uri": "spacetraders://agent/info"}}' | ./spacetraders-mcp 2>/dev/null | jq .

# Show help
help:
	@echo "SpaceTraders MCP Server - Available targets:"
	@echo ""
	@echo "  build              Build the server binary"
	@echo "  test               Run all tests using Go test runner"
	@echo "  test-unit          Run unit tests only"
	@echo "  test-integration   Run integration tests (requires SPACETRADERS_API_TOKEN)"
	@echo "  test-full          Run full integration tests with real API calls"
	@echo "  test-coverage      Run tests with coverage report"
	@echo "  test-short         Run short tests only"
	@echo "  clean              Clean build artifacts"
	@echo "  dev                Build and run development server"
	@echo "  fmt                Format code"
	@echo "  lint               Lint code (requires golangci-lint)"
	@echo "  tidy               Tidy dependencies"
	@echo "  security           Check for security vulnerabilities (requires gosec)"
	@echo "  bench              Run benchmarks"
	@echo "  install-dev-deps   Install development dependencies"
	@echo "  quick-test         Quick test with real API (requires SPACETRADERS_API_TOKEN)"
	@echo "  help               Show this help message"
	@echo ""
	@echo "Environment variables:"
	@echo "  SPACETRADERS_API_TOKEN   Required for test-full and quick-test"
	@echo ""
	@echo "Examples:"
	@echo "  make build                           # Build the server"
	@echo "  make test                            # Run all tests with Go test runner"
	@echo "  make test-unit                       # Run unit tests only"
	@echo "  SPACETRADERS_API_TOKEN=xyz make test-full # Run tests with real API"
	@echo "  make quick-test                      # Quick API test"
