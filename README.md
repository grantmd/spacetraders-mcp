# SpaceTraders MCP Server

A Model Context Protocol (MCP) server for interacting with the SpaceTraders API. This server provides tools to interact with your SpaceTraders agent and game data through MCP-compatible clients like Claude Desktop.

## Features

- **Agent Information Resource**: Access your agent's current information including credits, headquarters, faction, and ship count via the resource `spacetraders://agent/info`
- **Ships List Resource**: View all your ships with detailed information including location, status, cargo, and equipment via the resource `spacetraders://ships/list`
- **Contracts List Resource**: View available contracts with terms, payments, and delivery requirements via the resource `spacetraders://contracts/list`
- **Comprehensive Logging**: Built-in structured logging with API call timing, performance metrics, and component-specific debugging
- **Modular Architecture**: Clean, extensible codebase that makes it easy to add new SpaceTraders API resources and tools

## Prerequisites

- Go 1.24.4 or later
- A SpaceTraders API token (get one by registering at [SpaceTraders](https://spacetraders.io/))
- An MCP-compatible client (like Claude Desktop) that supports resources

## Setup

1. **Clone and build the project:**
   ```bash
   git clone <your-repo-url>
   cd spacetraders-mcp
   go mod tidy
   go build -o spacetraders-mcp .
   ```

2. **Configure your API token:**
   The server uses Viper for configuration management. You can configure your API token in multiple ways:
   
   **Option 1: Create a `.env` file (recommended):**
   ```
   SPACETRADERS_API_TOKEN="your_api_token_here"
   ```
   
   **Option 2: Set environment variable:**
   ```bash
   export SPACETRADERS_API_TOKEN="your_api_token_here"
   ```

   You can get your API token by:
   - Registering a new agent at https://spacetraders.io/
   - Or using an existing token if you already have an agent

3. **Test the server:**
   ```bash
   ./test_mcp.sh
   ```

## Usage

### Running the Server

The MCP server communicates via stdin/stdout:

```bash
./spacetraders-mcp
```

### Available Resources

#### `spacetraders://agent/info`

Provides comprehensive information about your SpaceTraders agent as a readable resource.

**URI:** `spacetraders://agent/info`  
**MIME Type:** `application/json`  
**Description:** Current agent information including credits, headquarters, faction, and ship count

**Content:**
```json
{
  "agent": {
    "accountId": "your_account_id",
    "credits": 175000,
    "headquarters": "X1-FM66-A1",
    "shipCount": 2,
    "startingFaction": "ASTRO",
    "symbol": "YOUR_CALLSIGN"
  }
}
```

#### `spacetraders://ships/list`

Provides detailed information about all ships owned by your agent.

**URI:** `spacetraders://ships/list`  
**MIME Type:** `application/json`  
**Description:** List of all ships with their status, location, cargo, crew, and equipment information

**Content:**
```json
{
  "ships": [
    {
      "symbol": "YOUR_SHIP-1",
      "registration": {
        "name": "YOUR_SHIP-1",
        "factionSymbol": "ASTRO",
        "role": "COMMAND"
      },
      "nav": {
        "systemSymbol": "X1-FM66",
        "waypointSymbol": "X1-FM66-A1",
        "status": "DOCKED",
        "flightMode": "CRUISE"
      },
      "cargo": {
        "capacity": 40,
        "units": 0,
        "inventory": []
      },
      "fuel": {
        "current": 400,
        "capacity": 400
      }
    }
  ],
  "meta": {
    "count": 2
  }
}
```

### Manual Testing

You can test the server manually using JSON-RPC 2.0 messages:

1. **List available resources:**
   ```bash
   echo '{"jsonrpc": "2.0", "id": 1, "method": "resources/list"}' | ./spacetraders-mcp
   ```

2. **Read agent info resource:**
   ```bash
   echo '{"jsonrpc": "2.0", "id": 2, "method": "resources/read", "params": {"uri": "spacetraders://agent/info"}}' | ./spacetraders-mcp
   ```

## Integration with Claude Desktop

To use this MCP server with Claude Desktop, add it to your Claude Desktop configuration:

### macOS
Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "spacetraders": {
      "command": "/path/to/your/spacetraders-mcp/spacetraders-mcp",
      "args": []
    }
  }
}
```

### Windows
Edit `%APPDATA%\Claude\claude_desktop_config.json` with the same structure, using Windows paths.

## Development

### Project Structure

```
spacetraders-mcp/
├── main.go                           # Application entry point
├── .env                              # API token configuration (not in git)  
├── pkg/                              # Package directory
│   ├── config/                       # Configuration management
│   │   └── config.go                 # Viper-based config loading
│   ├── logging/                      # Structured logging system
│   │   └── logger.go                 # Logger with MCP integration
│   ├── spacetraders/                 # SpaceTraders API client
│   │   └── client.go                 # API client and data types
│   ├── resources/                    # MCP resource handlers
│   │   ├── registry.go               # Resource registry
│   │   ├── agent.go                  # Agent info resource
│   │   ├── ships.go                  # Ships list resource
│   │   └── contracts.go              # Contracts list resource
│   └── tools/                        # MCP tool handlers (future)
│       └── registry.go               # Tool registry placeholder
├── test_mcp.sh                       # Test script
├── test_shutdown.sh                  # Shutdown test script
├── claude_desktop_config_example.json # Claude Desktop config example
├── go.mod                            # Go module definition
├── go.sum                            # Go module checksums
└── README.md                         # This file
```

### Adding New Resources

To add new SpaceTraders API resources:

1. **Add API method**: Extend `pkg/spacetraders/client.go` with the new API endpoint method and response types
2. **Create resource handler**: Add a new file in `pkg/resources/` (e.g., `contracts.go`) implementing the `ResourceHandler` interface
3. **Register resource**: Add your new resource to the `registerResources()` function in `pkg/resources/registry.go`
4. **Test**: Run the test script to verify your new resource works

**Example structure for a new resource:**
```go
// pkg/resources/systems.go
type SystemsResource struct {
    client *spacetraders.Client
    logger *logging.Logger
}

func NewSystemsResource(client *spacetraders.Client, logger *logging.Logger) *SystemsResource {
    return &SystemsResource{
        client: client,
        logger: logger,
    }
}

func (r *SystemsResource) Resource() mcp.Resource {
    return mcp.Resource{
        URI:         "spacetraders://systems/list",
        Name:        "Systems List",
        Description: "List of all star systems",
        MIMEType:    "application/json",
    }
}

func (r *SystemsResource) Handler() func(...) (...) {
    return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
        ctxLogger := r.logger.WithContext(ctx, "systems-resource")
        ctxLogger.Debug("Fetching systems from API")
        
        start := time.Now()
        systems, err := r.client.GetSystems()
        duration := time.Since(start)
        
        if err != nil {
            ctxLogger.Error("Failed to fetch systems: %v", err)
            ctxLogger.APICall("/systems", 0, duration.String())
            // Handle error...
        }
        
        ctxLogger.APICall("/systems", 200, duration.String())
        ctxLogger.Info("Successfully retrieved %d systems", len(systems))
        ctxLogger.ResourceRead(request.Params.URI, true)
        
        // Format and return response...
    }
}
```

### Adding Tools

For interactive SpaceTraders API actions (like trading, navigation):

1. **Add API method**: Extend `pkg/spacetraders/client.go` with the action method
2. **Create tool handler**: Add a new file in `pkg/tools/` implementing the `ToolHandler` interface  
3. **Register tool**: Add your new tool to the `registerTools()` function in `pkg/tools/registry.go`
4. **Enable tools**: Update `main.go` to enable tool capabilities: `server.WithToolCapabilities(true, false)`
5. **Test**: Verify your new tool works correctly

### Architecture Benefits

The modular structure provides:
- **Separation of concerns**: Each package has a clear responsibility
- **Easy testing**: Individual components can be tested in isolation
- **Scalability**: Adding new resources/tools is straightforward
- **Maintainability**: Code is organized and easy to navigate
- **Reusability**: SpaceTraders client can be used independently
- **Comprehensive logging**: Built-in structured logging with performance metrics
- **Debugging support**: Component-specific logging with context and timing information

### Error Handling

The server uses proper MCP error handling:
- API errors are returned as resource contents with error messages
- Network errors are handled gracefully
- Invalid resource URIs return appropriate error responses
- Tools (when implemented) use proper tool error handling
- Graceful shutdown on Ctrl+C with no error messages
- Comprehensive logging for debugging and monitoring
- Performance tracking for all API calls

## API Reference

This server uses the SpaceTraders v2 API. For full API documentation, visit:
- [SpaceTraders API Docs](https://spacetraders.stoplight.io/docs/spacetraders)
- [SpaceTraders Getting Started](https://docs.spacetraders.io/)

## License

[Add your license here]

## Contributing

[Add contribution guidelines here]

## Troubleshooting

### Common Issues

1. **"SPACETRADERS_API_TOKEN not found in configuration"**
   - Make sure your `.env` file exists and contains the token, or set the environment variable
   - Check that the token is properly quoted in the `.env` file
   - Verify that Viper can read your configuration (check file permissions)

2. **"API request failed with status 401"**
   - Your API token is invalid or expired
   - Register a new agent or check your existing token

3. **"Failed to make request"**
   - Check your internet connection
   - Verify the SpaceTraders API is accessible

4. **Build errors**
   - Ensure you have Go 1.24.4 or later
   - Run `go mod tidy` to resolve dependencies
   - Check that all required packages (including Viper) are properly installed

5. **Configuration file issues**
   - Viper supports multiple configuration formats (.env, .yaml, .json, etc.)
   - If using environment variables, make sure they're properly exported
   - The server will work with environment variables even if no .env file exists

### Server Management

The server supports graceful shutdown:
- Press `Ctrl+C` to stop the server cleanly
- No error messages are displayed on normal shutdown
- The server automatically handles signal cleanup

### Logging

The server provides comprehensive logging capabilities:

#### Built-in Logging Levels
- **ERROR**: Server errors, API failures, configuration issues
- **INFO**: Server startup, API call timing, successful operations, resource metrics
- **DEBUG**: Detailed operation tracing, response sizes, component-specific debugging

#### Log Output Examples
```
[INFO] 2025/06/20 18:00:16 Starting SpaceTraders MCP Server
[DEBUG] 2025/06/20 18:00:16 [agent-resource] Fetching agent information from API
[INFO] 2025/06/20 18:00:16 [agent-resource] API call: /my/agent -> 200 (513ms)
[INFO] 2025/06/20 18:00:16 [agent-resource] Successfully retrieved agent info for: CORNTAB
[DEBUG] 2025/06/20 18:00:16 [agent-resource] Agent resource response size: 199 bytes
```

#### Log Features
- **Component Identification**: Each log shows which resource/component generated it
- **API Performance Tracking**: HTTP status codes and response times for all SpaceTraders API calls
- **Resource Metrics**: Response sizes, item counts, and operation success rates
- **Context Awareness**: Logs include relevant context like agent symbols, ship counts, etc.

#### MCP Client Logging
The server supports MCP logging notifications that can be sent to compatible clients:
- Configurable log levels via MCP `logging/setLevel` requests
- Structured log messages sent as MCP notifications
- Integration with Claude Desktop and other MCP clients that support logging

#### Viewing Logs
- **stderr**: All logs are written to stderr for easy redirection
- **Silent JSON**: stdout only contains clean JSON-RPC responses
- **Log Filtering**: Redirect stderr to `/dev/null` to suppress logs: `./spacetraders-mcp 2>/dev/null`

### Getting Help

- Join the [SpaceTraders Discord](https://discord.gg/UpEfRRjsCT)
- Check the SpaceTraders documentation
- Open an issue in this repository