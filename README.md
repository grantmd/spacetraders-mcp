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

#### `spacetraders://systems/{systemSymbol}/waypoints`

Provides information about all waypoints in a system, including their types, traits, and orbital relationships.

**URI:** `spacetraders://systems/{systemSymbol}/waypoints`  
**MIME Type:** `application/json`  
**Description:** List of all waypoints in a system with detailed information about markets, shipyards, and other facilities

**Example URI:** `spacetraders://systems/X1-FM66/waypoints`

**Content:**
```json
{
  "system": "X1-FM66",
  "waypoints": [
    {
      "symbol": "X1-FM66-A1",
      "type": "PLANET",
      "x": 10,
      "y": 20,
      "orbitals": [
        {
          "symbol": "X1-FM66-A1-M1"
        }
      ],
      "traits": [
        {
          "symbol": "MARKETPLACE",
          "name": "Marketplace",
          "description": "A thriving center of commerce where traders can buy and sell goods"
        }
      ]
    },
    {
      "symbol": "X1-FM66-B2",
      "type": "MOON",
      "x": 15,
      "y": 25,
      "traits": [
        {
          "symbol": "SHIPYARD",
          "name": "Shipyard",
          "description": "Facility for constructing and repairing starships"
        }
      ]
    }
  ],
  "summary": {
    "total": 12,
    "byType": {
      "PLANET": 3,
      "MOON": 5,
      "ASTEROID": 4
    },
    "shipyards": ["X1-FM66-B2", "X1-FM66-C3"],
    "markets": ["X1-FM66-A1", "X1-FM66-D4"]
  }
}
```

#### `spacetraders://systems/{systemSymbol}/waypoints/{waypointSymbol}/shipyard`

Provides detailed information about a shipyard, including available ships, prices, and recent transactions.

**URI:** `spacetraders://systems/{systemSymbol}/waypoints/{waypointSymbol}/shipyard`  
**MIME Type:** `application/json`  
**Description:** Information about ships available for purchase, their specifications, and pricing

**Example URI:** `spacetraders://systems/X1-FM66/waypoints/X1-FM66-B2/shipyard`

**Content:**
```json
{
  "system": "X1-FM66",
  "waypoint": "X1-FM66-B2",
  "shipyard": {
    "symbol": "X1-FM66-B2",
    "shipTypes": [
      {
        "type": "SHIP_PROBE"
      },
      {
        "type": "SHIP_MINING_DRONE"
      }
    ],
    "ships": [
      {
        "type": "SHIP_PROBE",
        "name": "Probe",
        "description": "A small, fast exploration vessel",
        "supply": "ABUNDANT",
        "purchasePrice": 65000,
        "frame": {
          "symbol": "FRAME_PROBE",
          "name": "Probe Frame",
          "moduleSlots": 2,
          "mountingPoints": 1,
          "fuelCapacity": 400
        },
        "reactor": {
          "symbol": "REACTOR_FISSION_I",
          "name": "Fission Reactor I",
          "powerOutput": 31
        },
        "engine": {
          "symbol": "ENGINE_IMPULSE_DRIVE_I",
          "name": "Impulse Drive I",
          "speed": 30
        },
        "crew": {
          "required": 1,
          "capacity": 3
        }
      }
    ],
    "modificationsFee": 5000
  },
  "summary": {
    "availableShipTypes": ["SHIP_PROBE", "SHIP_MINING_DRONE"],
    "totalShipsAvailable": 3,
    "priceRange": {
      "min": 65000,
      "max": 250000
    },
    "shipsByType": {
      "SHIP_PROBE": 2,
      "SHIP_MINING_DRONE": 1
    },
    "shipsBySupply": {
      "ABUNDANT": 2,
      "MODERATE": 1
    },
    "modificationsFee": 5000,
    "recentTransactions": 5
  }
}
```

## Important: How MCP Resources Work

**Key Understanding:** MCP resources are **NOT automatically loaded** by clients like Claude Desktop. They work like web pages - they exist at specific URIs but are only fetched when explicitly requested.

### What This Means

- ❌ **Claude doesn't automatically know about your agent, ships, contracts, etc.**
- ✅ **You must explicitly ask Claude to read resources**
- ✅ **Use the status tool for quick overviews**

### How to Use Resources

Instead of asking "What ships do I have?", try:
- "Get my status summary"
- "Read the spacetraders://ships/list resource"
- "Show me my agent information"
- "Read my contracts list"

### Available Prompts

Prompts are **conversation starters** that automatically guide Claude to read relevant resources and provide strategic analysis. Instead of manually asking Claude to read specific resources, use these prompts for guided interactions:

#### `status_check`

**Description:** Get comprehensive status of your SpaceTraders agent including ships, contracts, and opportunities

**Parameters:**
- `detail_level` (optional): "basic", "detailed", or "full"

**What it does:**
- Automatically calls `get_status_summary` tool
- Reads your ships list and contracts
- For detailed/full: checks waypoints and facilities at your locations
- For full: suggests concrete next actions and identifies opportunities

**Usage Examples:**
- "Use the status_check prompt"
- "Run status_check with detail_level=full"

#### `explore_system`

**Description:** Explore a specific system to find trading opportunities, shipyards, and points of interest

**Parameters:**
- `system_symbol` (required): System to explore (e.g., "X1-FM66")

**What it does:**
- Reads all waypoints in the specified system
- Identifies marketplaces, shipyards, and mining sites
- Checks available ships at any shipyards
- Provides strategic analysis and recommendations

**Usage Examples:**
- "Use explore_system prompt for X1-FM66"
- "Run the explore_system prompt with system_symbol=X1-ABC123"

#### `contract_strategy`

**Description:** Analyze available contracts and suggest the best ones to accept based on current capabilities

**What it does:**
- Reads your current contracts and agent status
- Analyzes profitability, feasibility, and logistics for each contract
- Recommends which contracts to accept and why
- Provides execution plan including ship movements and cargo requirements

**Usage Examples:**
- "Use the contract_strategy prompt"
- "Run contract analysis"

#### `fleet_optimization`

**Description:** Analyze current fleet and suggest optimizations for better efficiency and profit

**What it does:**
- Reads your ships and agent status
- Analyzes fleet composition and utilization
- Checks shipyards in your current systems
- Recommends fleet improvements with cost-benefit analysis

**Usage Examples:**
- "Use the fleet_optimization prompt"
- "Run fleet analysis"

### Available Tools

#### `get_status_summary`

Provides a comprehensive status overview by automatically fetching and summarizing your agent information, ships, and contracts.

**Parameters:**
- `include_ships` (boolean, default: true): Include detailed ship information
- `include_contracts` (boolean, default: true): Include contract information

**Usage Examples:**
- "Get my status summary"
- "Show me a status overview"
- "Get status summary without contracts"

**Sample Output:**
```
🚀 SpaceTraders Status Summary

👤 Agent: PLAYER_123
💰 Credits: 175,000
🏠 Headquarters: X1-FM66-A1
🏴 Faction: ASTRO
🚢 Ships: 3

🚢 Fleet Status:
  • Total Ships: 3
  • Cargo Usage: 15/120 (12.5%)
  • Ship Status:
    - DOCKED: 2
    - IN_ORBIT: 1
  • Ships by System:
    - X1-FM66: 3

📋 Contracts:
  • Total: 2
  • Accepted: 1
  • Fulfilled: 0
  • Pending: 1
  • Total Value: 50,000 credits
```

#### `accept_contract`

Accept a contract by its ID. This commits the agent to fulfilling the contract terms and provides an upfront payment.

**Parameters:**
- `contract_id` (string, required): The unique identifier of the contract to accept

**Example:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "accept_contract",
    "arguments": {
      "contract_id": "clm0n4k8q0001js08g2h1k9v8"
    }
  }
}
```

**Response:** Returns contract details and updated agent information including the acceptance payment.

## Testing

The project includes comprehensive tests covering both unit tests and integration tests. Tests have been converted from shell scripts to native Go tests for better reliability and maintainability.

### Running Tests

The easiest way to run tests is using the provided test runner or Makefile:

```bash
# Run all tests (unit + integration tests that don't require API access)
make test

# Run unit tests only
make test-unit

# Run integration tests only  
make test-integration

# Run full integration tests with real API calls (requires SPACETRADERS_API_TOKEN)
make test-full

# Or use the Go test runner directly
go run ./cmd/test_runner.go
go run ./cmd/test_runner.go --integration  # For full API integration tests
```

### Test Categories

1. **Unit Tests** (`./pkg/...`): Test individual packages and functions without external dependencies
2. **Integration Tests** (`./test/...`): Test the complete MCP server functionality
   - Protocol compliance tests (no API token required)
   - Resource structure validation (no API token required)
   - Server startup/shutdown tests (no API token required)
   - API-dependent tests (require `SPACETRADERS_API_TOKEN`, will be skipped if not provided)

### Manual Testing

You can also test the server manually using JSON-RPC 2.0 messages:

1. **List available resources:**
   ```bash
   echo '{"jsonrpc": "2.0", "id": 1, "method": "resources/list"}' | ./spacetraders-mcp
   ```

2. **Read agent info resource:**
   ```bash
   echo '{"jsonrpc": "2.0", "id": 2, "method": "resources/read", "params": {"uri": "spacetraders://agent/info"}}' | ./spacetraders-mcp
   ```

### Test Coverage

Tests cover:
- MCP protocol compliance (JSON-RPC 2.0)
- Resource listing and reading
- Error handling for invalid resources
- API authentication error handling
- Server graceful shutdown
- Multiple request handling
- All package functionality

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
├── test/                             # Integration tests
│   └── integration_test.go           # Comprehensive integration tests
├── cmd/                              # Command line tools
│   └── test_runner.go                # Go test runner (replaces shell scripts)
├── .github/                          # GitHub Actions workflows
│   ├── workflows/                    # CI/CD pipelines
│   │   ├── ci.yml                    # Main CI workflow
│   │   ├── integration.yml           # Integration tests with real API
│   │   └── release.yml               # Release automation
│   ├── ISSUE_TEMPLATE/               # Issue templates
│   └── dependabot.yml               # Dependency updates
├── claude_desktop_config_example.json # Claude Desktop config example
├── Dockerfile                        # Container image for deployment
├── go.mod                            # Go module definition
├── go.sum                            # Go module checksums
└── README.md                         # This file
```

### Adding New Resources

To add new SpaceTraders API resources:

1. **Add API method**: Extend `pkg/spacetraders/client.go` with the new API endpoint method and response types
2. **Create resource handler**: Add a new file in `pkg/resources/` (e.g., `contracts.go`) implementing the `ResourceHandler` interface
3. **Register resource**: Add your new resource to the `registerResources()` function in `pkg/resources/registry.go`
4. **Add tests**: Create unit tests for your resource handler
5. **Test**: Run `make test` to verify your new resource works

## CI/CD and Automation

This project uses GitHub Actions for comprehensive CI/CD automation:

### Workflows

#### 1. **Main CI Pipeline** (`.github/workflows/ci.yml`)
Runs on every push and pull request:
- **Multi-version testing**: Tests against Go 1.21 and 1.22
- **Code quality**: `go vet`, `gofmt`, linting with `golangci-lint`
- **Comprehensive testing**: Unit tests, integration tests, and test runner
- **Security scanning**: `gosec` and `govulncheck`
- **Cross-platform builds**: Linux, macOS, Windows (AMD64, ARM64)
- **Coverage reporting**: Uploads to Codecov

#### 2. **Integration Tests** (`.github/workflows/integration.yml`)
Scheduled daily and manually triggered:
- **Real API testing**: Uses `SPACETRADERS_API_TOKEN` secret for live API tests
- **Automated issue creation**: Creates GitHub issues on test failures
- **Fallback testing**: Runs protocol tests when no API token is available
- **Test reporting**: Generates comprehensive test reports

#### 3. **Release Automation** (`.github/workflows/release.yml`)
Triggered on version tags:
- **Pre-release testing**: Full test suite before building release
- **Multi-platform binaries**: Builds for all supported platforms
- **Docker images**: Pushes to Docker Hub and GitHub Container Registry
- **GitHub releases**: Automatic release creation with changelog
- **Checksums**: SHA256 checksums for all binaries

### Secrets Configuration

Configure these secrets in your GitHub repository:

```bash
# Required for integration tests
SPACETRADERS_API_TOKEN=your_spacetraders_token_here

# Optional for Docker releases
DOCKER_USERNAME=your_dockerhub_username
DOCKER_PASSWORD=your_dockerhub_password

# Optional for notifications
SLACK_WEBHOOK_URL=your_slack_webhook_url
```

### Automated Dependency Updates

**Dependabot** automatically:
- Updates Go modules weekly
- Updates GitHub Actions weekly
- Updates Docker base images weekly
- Groups minor/patch updates
- Prioritizes security updates

### Quality Gates

All pull requests must pass:
- [ ] Unit tests (all packages)
- [ ] Integration tests (protocol compliance)
- [ ] Security scans (gosec, govulncheck)
- [ ] Code formatting (gofmt)
- [ ] Linting (golangci-lint)
- [ ] Cross-platform builds

### Release Process

1. **Create a version tag**: `git tag v1.0.0 && git push origin v1.0.0`
2. **Automated release**: GitHub Actions will:
   - Run full test suite
   - Build multi-platform binaries
   - Create GitHub release with changelog
   - Push Docker images
   - Generate checksums

### Development Workflow

```bash
# Run tests locally
make test

# Run with real API (requires token)
make test-full

# Run specific test categories
make test-unit
make test-integration

# Format and lint code
make fmt
make lint

# Security scan
make security
```

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

### MCP Resource Issues

#### Problem: Claude doesn't know about my agent/ships/contracts

**Cause:** MCP resources are not automatically loaded. They're only fetched when explicitly requested.

**Solutions:**
1. Use the status summary tool: "Get my status summary"
2. Explicitly request resources:
   - "Read the spacetraders://agent/info resource"
   - "Show me the spacetraders://ships/list resource"
   - "Read my contracts from spacetraders://contracts/list"

#### Problem: "Resource not found" errors

**Cause:** Usually incorrect URI format or missing system/waypoint symbols.

**Solutions:**
1. Check URI format exactly matches the patterns:
   - `spacetraders://systems/X1-SYSTEM/waypoints`
   - `spacetraders://systems/X1-SYSTEM/waypoints/X1-WAYPOINT/shipyard`
2. Ensure system and waypoint symbols are valid
3. Use the status tool first to get valid system names

#### Problem: Debug logs show no resource requests

**Cause:** Claude Desktop isn't calling your resources automatically.

**Solutions:**
1. Ask Claude explicitly: "What MCP resources are available?"
2. Request specific resources by name or URI
3. Use the status summary tool for automated data fetching

### Other Common Issues

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