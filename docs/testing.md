# Testing Guide

This document provides comprehensive information about testing the SpaceTraders MCP Server, including unit tests, integration tests, and manual testing procedures.

## Overview

The SpaceTraders MCP Server includes multiple levels of testing to ensure reliability and correctness:

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test interactions with the SpaceTraders API
- **Manual Tests**: Interactive testing procedures
- **End-to-End Tests**: Full workflow testing with Claude Desktop

## Running Tests

### Quick Test Commands

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run only unit tests
make test-unit

# Run only integration tests (requires API token)
make test-integration

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./pkg/client/
```

### Environment Setup for Testing

```bash
# Required for integration tests
export SPACETRADERS_TOKEN=your_test_token_here
export SPACETRADERS_AGENT_SYMBOL=your_test_agent_symbol

# Optional: Use test-specific agent
export SPACETRADERS_TEST_MODE=true
```

## Test Categories

### Unit Tests

Unit tests focus on individual components and functions without external dependencies.

**Location**: `*_test.go` files alongside source code

**Coverage includes:**
- API client request/response handling
- Resource URI parsing and validation
- Tool parameter validation
- Error handling logic
- Data transformation functions

**Example test structure:**
```go
func TestClientGetAgent(t *testing.T) {
    // Setup
    server := httptest.NewServer(mockHandler)
    defer server.Close()
    
    client := NewClient(server.URL, "test-token")
    
    // Execute
    agent, err := client.GetAgent(context.Background())
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "TEST-AGENT", agent.Symbol)
}
```

### Integration Tests

Integration tests verify interactions with the actual SpaceTraders API.

**Location**: `test/integration/`

**Requirements:**
- Valid SpaceTraders API token
- Active agent with ships and contracts
- Network connectivity

**Test scenarios:**
- Agent information retrieval
- Ship listing and details
- System exploration
- Navigation operations
- Contract management
- Tool registration and execution

**Running integration tests:**
```bash
# Set up test environment
export SPACETRADERS_TOKEN=your_token
export SPACETRADERS_AGENT_SYMBOL=your_agent

# Run integration tests
make test-integration

# Run specific integration test
go test ./test/integration/ -run TestAgentInfo
```

### Manual Testing

Manual testing procedures for interactive verification.

**Test checklist:**

1. **Server Startup**
   ```bash
   ./spacetraders-mcp
   # Should start without errors
   # Should log "Server ready" message
   ```

2. **Resource Access**
   ```bash
   # Test agent info resource
   curl -X POST http://localhost:8080/mcp \
     -d '{"method":"resources/read","params":{"uri":"spacetraders://agent/info"}}'
   ```

3. **Tool Execution**
   ```bash
   # Test status summary tool
   curl -X POST http://localhost:8080/mcp \
     -d '{"method":"tools/call","params":{"name":"get_status_summary","arguments":{}}}'
   ```

4. **Claude Desktop Integration**
   - Configure Claude Desktop with MCP server
   - Test basic commands: "Show my agent status"
   - Test resource access: Reference `spacetraders://agent/info`
   - Test tools: "Navigate GHOST-01 to X1-DF55-20250Z"

## Test Coverage

### Current Coverage Targets

- **Overall**: 80%+ line coverage
- **Critical paths**: 95%+ coverage (authentication, navigation, trading)
- **Error handling**: 90%+ coverage
- **API client**: 85%+ coverage

### Generating Coverage Reports

```bash
# Generate HTML coverage report
make coverage-html

# View coverage in browser
open coverage.html

# Generate coverage summary
go test -cover ./...

# Detailed coverage by function
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Coverage Analysis

```bash
# Find uncovered code
go tool cover -html=coverage.out

# Coverage by package
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -E "pkg/"
```

## Test Data Management

### Mock Data

Mock data for unit tests is stored in `test/fixtures/`:

```
test/fixtures/
├── agent.json          # Sample agent data
├── ships.json          # Sample ship listings
├── systems.json        # Sample system data
├── contracts.json      # Sample contract data
└── markets.json        # Sample market data
```

### Test Utilities

Common test utilities in `test/utils/`:

- **Mock HTTP server**: For API client testing
- **Test data generators**: Create realistic test data
- **Assertion helpers**: Common test assertions
- **Setup/teardown helpers**: Test environment management

### API Mocking

```go
// Example mock setup
func setupMockServer() *httptest.Server {
    mux := http.NewServeMux()
    
    mux.HandleFunc("/my/agent", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(mockAgent)
    })
    
    return httptest.NewServer(mux)
}
```

## Test Environment Configuration

### Development Testing

```bash
# Use development API endpoint
export SPACETRADERS_API_URL=https://api.spacetraders.io/v2

# Enable debug logging
export DEBUG=1

# Use test database
export TEST_MODE=true
```

### CI/CD Testing

Configuration for automated testing in CI environments:

```yaml
# GitHub Actions example
env:
  SPACETRADERS_TOKEN: ${{ secrets.SPACETRADERS_TOKEN }}
  SPACETRADERS_AGENT_SYMBOL: ${{ secrets.SPACETRADERS_AGENT_SYMBOL }}
  TEST_TIMEOUT: "300s"
```

## Performance Testing

### Benchmarks

```bash
# Run benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkGetAgent ./pkg/client/

# Memory profiling
go test -bench=. -memprofile=mem.prof ./...
go tool pprof mem.prof
```

### Load Testing

```bash
# Test with multiple concurrent requests
go test -race ./...

# Stress testing with high concurrency
GOMAXPROCS=8 go test -parallel 4 ./...
```

## Error Testing

### Error Scenarios

Tests cover various error conditions:

- Network failures
- API authentication errors
- Invalid parameters
- Rate limiting
- Server errors (5xx responses)
- Malformed responses

### Error Injection

```go
func TestNetworkFailure(t *testing.T) {
    // Create client with invalid URL
    client := NewClient("http://invalid-url", "token")
    
    // Test should handle network error gracefully
    _, err := client.GetAgent(context.Background())
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "network error")
}
```

## Test Best Practices

### Writing Tests

1. **Use descriptive test names**
   ```go
   func TestNavigateShip_InvalidWaypoint_ReturnsError(t *testing.T)
   ```

2. **Follow AAA pattern** (Arrange, Act, Assert)
   ```go
   func TestExample(t *testing.T) {
       // Arrange
       client := setupTestClient()
       
       // Act
       result, err := client.DoSomething()
       
       // Assert
       assert.NoError(t, err)
       assert.Equal(t, expected, result)
   }
   ```

3. **Test error conditions**
   ```go
   func TestErrorHandling(t *testing.T) {
       // Test various error scenarios
       testCases := []struct {
           name        string
           input       string
           expectError bool
           errorMsg    string
       }{
           {"invalid input", "", true, "required parameter"},
           {"valid input", "valid", false, ""},
       }
       
       for _, tc := range testCases {
           t.Run(tc.name, func(t *testing.T) {
               // Test implementation
           })
       }
   }
   ```

### Test Organization

- **One test file per source file** (`client.go` → `client_test.go`)
- **Group related tests** using subtests
- **Use setup/teardown functions** for common initialization
- **Keep tests independent** - no shared state between tests

### Mock Guidelines

- **Mock external dependencies** (HTTP calls, databases)
- **Use interfaces** to enable easy mocking
- **Verify mock expectations** are met
- **Keep mocks simple** and focused

## Test Automation

### Pre-commit Hooks

```bash
# Install pre-commit hooks
make install-hooks

# Hooks run automatically on git commit:
# - go fmt
# - go vet
# - golint
# - tests
```

### Continuous Integration

Tests run automatically on:
- Pull requests
- Commits to main branch
- Release tags
- Scheduled runs (daily)

### Test Reporting

- **Coverage reports** uploaded to code coverage services
- **Test results** reported in pull requests
- **Performance benchmarks** tracked over time
- **Flaky test detection** and reporting

## Debugging Tests

### Common Issues

1. **Tests pass locally but fail in CI**
   - Check environment differences
   - Verify test isolation
   - Check for race conditions

2. **Flaky tests**
   - Add proper waiting/retries
   - Avoid time-dependent assertions
   - Use deterministic test data

3. **Slow tests**
   - Profile test execution
   - Optimize setup/teardown
   - Use parallel testing where appropriate

### Debug Commands

```bash
# Run tests with verbose output
go test -v ./...

# Run specific test with debugging
go test -v -run TestSpecificFunction ./pkg/client/

# Race condition detection
go test -race ./...

# Test with timeout
go test -timeout 30s ./...
```

## Test Maintenance

### Regular Tasks

- **Update test data** when API changes
- **Review test coverage** and add missing tests
- **Clean up obsolete tests** when refactoring
- **Update mock expectations** with API changes
- **Performance test regression** checking

### Test Review Checklist

- [ ] Tests cover new functionality
- [ ] Error conditions are tested
- [ ] Tests are deterministic
- [ ] No external dependencies in unit tests
- [ ] Integration tests use appropriate test data
- [ ] Tests run in reasonable time
- [ ] Test names are descriptive
- [ ] Code coverage remains high