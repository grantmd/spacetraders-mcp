# Development Guide

This document provides information for developers who want to contribute to or extend the SpaceTraders MCP Server.

## Project Structure

```
spacetraders-mcp/
├── cmd/                    # Command-line interface
├── pkg/                    # Core packages
│   ├── client/            # SpaceTraders API client
│   ├── mcp/               # MCP server implementation
│   ├── resources/         # MCP resource handlers
│   ├── tools/             # MCP tool handlers
│   └── prompts/           # MCP prompt handlers
├── docs/                  # Documentation
├── test/                  # Test files
├── .github/               # GitHub workflows
├── main.go               # Main entry point
├── go.mod                # Go module definition
├── go.sum                # Go dependency checksums
└── Makefile              # Build automation
```

### Key Components

#### `pkg/client/`
Contains the SpaceTraders API client implementation:
- HTTP client with authentication
- API endpoint definitions
- Request/response structures
- Error handling

#### `pkg/mcp/`
Core MCP server implementation:
- Server initialization and configuration
- Protocol handling
- Resource and tool registration
- Logging and error management

#### `pkg/resources/`
MCP resource handlers that provide read-only access to game data:
- Agent information
- Ship listings
- System and waypoint data
- Market and shipyard information

#### `pkg/tools/`
MCP tool handlers for performing actions:
- Ship navigation and management
- Resource extraction and trading
- Contract management
- Fleet operations

#### `generated/`
Auto-generated SpaceTraders API client code:
- Generated from the official SpaceTraders OpenAPI specification
- Not checked into version control (regenerated during builds)
- Contains Go client code for all API endpoints

## Code Generation

This project uses OpenAPI Generator to create the SpaceTraders API client from the official API specification.

### Prerequisites

- Java 8+ (required for OpenAPI Generator)
- OpenAPI Generator CLI (automatically downloaded by Make targets)

### Generating the Client

The API client is generated from the SpaceTraders OpenAPI specification:

```bash
# Generate the client (downloads spec and generates Go code)
make generate-client

# Clean generated files
make clean-generated
```

### How Code Generation Works

1. **Download Specification**: The Makefile downloads the latest OpenAPI spec from `https://spacetraders.io/SpaceTraders.json`
2. **Generate Client**: OpenAPI Generator creates Go client code in `./generated/spacetraders/`
3. **Configuration**: Generation is configured via `openapi-generator-config.yaml`
4. **Integration**: The wrapper client in `pkg/client/` uses the generated code

### Generated Code Structure

```
generated/spacetraders/
├── api/                   # API endpoint implementations
├── docs/                  # Generated documentation
├── model_*.go            # Data models
├── api_*.go              # API clients
├── client.go             # Main client
├── configuration.go      # Client configuration
└── go.mod                # Module definition
```

### CI/CD Integration

The GitHub Actions workflows automatically:
1. Install OpenAPI Generator
2. Generate the client code
3. Build and test the application

This ensures that:
- Builds always use the latest API specification
- Generated code is never stale
- CI environments match local development

### Configuration

Generation behavior is controlled by `openapi-generator-config.yaml`:

```yaml
generatorName: go
packageName: spacetraders
additionalProperties:
  generateInterfaces: true
  structPrefix: true
  enumClassPrefix: true
```

### Troubleshooting Generation

If generation fails:

1. **Check Java Installation**:
   ```bash
   java -version
   ```

2. **Verify Network Access**:
   ```bash
   curl -I https://spacetraders.io/SpaceTraders.json
   ```

3. **Clean and Regenerate**:
   ```bash
   make clean-generated
   make generate-client
   ```

4. **Check OpenAPI Generator Version**:
   ```bash
   openapi-generator version
   ```
- Resource extraction
- Trading operations
- Contract handling

#### `pkg/prompts/`
Intelligent prompts for common tasks:
- Status checking
- System exploration
- Contract strategy
- Fleet optimization

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional, for build automation)
- A SpaceTraders API token for testing

### Local Development

1. **Clone the repository:**
```bash
git clone https://github.com/grantmd/spacetraders-mcp.git
cd spacetraders-mcp
```

2. **Install dependencies:**
```bash
go mod download
```

3. **Set up environment variables:**
```bash
export SPACETRADERS_TOKEN=your_token_here
export SPACETRADERS_AGENT_SYMBOL=your_agent_symbol
```

4. **Run the server:**
```bash
go run main.go
```

### Building

```bash
# Build for current platform
go build -o spacetraders-mcp

# Build for all platforms
make build-all

# Build with version info
make build VERSION=v1.0.0
```

## Adding New Resources

Resources provide read-only access to SpaceTraders data. Here's how to add a new resource:

### 1. Create the Resource Handler

Create a new file in `pkg/resources/`:

```go
package resources

import (
    "context"
    "encoding/json"
    "fmt"
    "net/url"

    "github.com/grantmd/spacetraders-mcp/pkg/client"
    "github.com/grantmd/spacetraders-mcp/pkg/mcp"
)

type ExampleResource struct {
    client *client.Client
    logger mcp.Logger
}

func NewExampleResource(client *client.Client, logger mcp.Logger) *ExampleResource {
    return &ExampleResource{
        client: client,
        logger: logger,
    }
}

func (r *ExampleResource) Resource() mcp.Resource {
    return mcp.Resource{
        URI:         "spacetraders://example/{param}",
        Name:        "Example Resource",
        Description: "Provides example data",
        MimeType:    "application/json",
    }
}

func (r *ExampleResource) Handler(ctx context.Context, uri *url.URL) ([]byte, error) {
    // Extract parameters from URI
    param := extractParam(uri.Path)

    // Fetch data from SpaceTraders API
    data, err := r.client.GetExampleData(ctx, param)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch example data: %w", err)
    }

    // Return JSON response
    return json.Marshal(data)
}
```

### 2. Register the Resource

Add the resource to the server in `pkg/mcp/server.go`:

```go
func (s *Server) registerResources() {
    // ... existing resources ...

    exampleResource := resources.NewExampleResource(s.client, s.logger)
    s.resources["example"] = exampleResource
}
```

## Adding New Tools

Tools allow actions to be performed on SpaceTraders data. Here's how to add a new tool:

### 1. Create the Tool Handler

Create a new file in `pkg/tools/`:

```go
package tools

import (
    "context"
    "encoding/json"

    "github.com/grantmd/spacetraders-mcp/pkg/client"
    "github.com/grantmd/spacetraders-mcp/pkg/mcp"
)

type ExampleTool struct {
    client *client.Client
    logger mcp.Logger
}

func NewExampleTool(client *client.Client, logger mcp.Logger) *ExampleTool {
    return &ExampleTool{
        client: client,
        logger: logger,
    }
}

func (t *ExampleTool) Tool() mcp.Tool {
    return mcp.Tool{
        Name:        "example_action",
        Description: "Performs an example action",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]interface{}{
                "param1": map[string]interface{}{
                    "type":        "string",
                    "description": "First parameter",
                },
                "param2": map[string]interface{}{
                    "type":        "number",
                    "description": "Second parameter",
                    "optional":    true,
                },
            },
            Required: []string{"param1"},
        },
    }
}

func (t *ExampleTool) Handler(ctx context.Context, arguments map[string]interface{}) ([]byte, error) {
    // Extract parameters
    param1, ok := arguments["param1"].(string)
    if !ok {
        return nil, fmt.Errorf("param1 is required and must be a string")
    }

    param2, _ := arguments["param2"].(float64)

    // Perform action via SpaceTraders API
    result, err := t.client.ExampleAction(ctx, param1, int(param2))
    if err != nil {
        return nil, fmt.Errorf("failed to perform example action: %w", err)
    }

    return json.Marshal(result)
}
```

### 2. Register the Tool

Add the tool to the server in `pkg/mcp/server.go`:

```go
func (s *Server) registerTools() {
    // ... existing tools ...

    exampleTool := tools.NewExampleTool(s.client, s.logger)
    s.tools["example_action"] = exampleTool
}
```

## Architecture Benefits

The current architecture provides several benefits:

### Separation of Concerns
- **Client**: Handles SpaceTraders API communication
- **Resources**: Provide read-only data access
- **Tools**: Handle actions and state changes
- **Prompts**: Offer intelligent assistance

### Extensibility
- Easy to add new resources and tools
- Modular design allows independent development
- Clear interfaces for new functionality

### Maintainability
- Each component has a single responsibility
- Well-defined interfaces between components
- Comprehensive error handling and logging

### Testability
- Components can be tested independently
- Mock implementations for external dependencies
- Clear separation between business logic and I/O

## Error Handling

The project uses structured error handling:

```go
// Wrap errors with context
return fmt.Errorf("failed to fetch ship data: %w", err)

// Log errors with structured data
t.logger.Error("Operation failed", "error", err, "ship", shipSymbol)

// Return appropriate HTTP status codes
if isNotFound(err) {
    return nil, &mcp.Error{Code: 404, Message: "Ship not found"}
}
```

## Testing

### Unit Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/client/
```

### Integration Tests
```bash
# Run integration tests (requires API token)
make test-integration
```

### Manual Testing
```bash
# Test with debug output
DEBUG=1 go run main.go

# Test specific functionality
go run main.go -test-resource ships
```

## Code Style

### Go Conventions
- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Write comprehensive comments for exported functions
- Handle errors explicitly

### Project Conventions
- Use structured logging with contextual information
- Wrap errors with additional context
- Validate input parameters thoroughly
- Return appropriate error types for MCP responses

## Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/new-feature`
3. **Make your changes** following the coding standards
4. **Add tests** for new functionality
5. **Update documentation** as needed
6. **Submit a pull request** with a clear description

### Pull Request Guidelines

- Ensure all tests pass
- Include relevant documentation updates
- Follow the existing code style
- Provide clear commit messages
- Reference any related issues

## Debugging

### Enable Debug Logging
```bash
export DEBUG=1
go run main.go
```

### Common Debugging Techniques
- Add structured logging to trace execution flow
- Use the `-test` flag for isolated component testing
- Check API responses by examining raw HTTP traffic
- Validate MCP protocol compliance with debug output

## Performance Considerations

- **Caching**: Consider caching frequently accessed data
- **Rate Limiting**: Respect SpaceTraders API rate limits
- **Concurrency**: Use goroutines for independent operations
- **Memory Usage**: Monitor memory usage for large datasets

## Security

- **API Tokens**: Never hardcode or log API tokens
- **Input Validation**: Validate all user inputs
- **Error Information**: Don't expose internal details in error messages
- **Dependencies**: Keep dependencies updated for security patches
