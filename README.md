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
    "recentTransactions": 0
   }

   #### `spacetraders://systems/{systemSymbol}/waypoints/{waypointSymbol}/market`

   Provides detailed market information including trade goods, prices, and trading opportunities.

   **URI:** `spacetraders://systems/{systemSymbol}/waypoints/{waypointSymbol}/market`  
   **MIME Type:** `application/json`  
   **Description:** Market data including exports, imports, current prices, and recent transactions

   **Example URI:** `spacetraders://systems/X1-FM66/waypoints/X1-FM66-A1/market`

   **Example Response:**
   ```json
   {
     "system": "X1-FM66",
     "waypoint": "X1-FM66-A1",
     "market": {
       "symbol": "X1-FM66-A1",
       "exports": [
         {
           "symbol": "FOOD",
           "name": "Food",
           "description": "Essential nutrition for crew"
         }
       ],
       "imports": [
         {
           "symbol": "MACHINERY",
           "name": "Machinery",
           "description": "Industrial equipment"
         }
       ],
       "trade_goods": [
         {
           "symbol": "FOOD",
           "type": "EXPORT",
           "purchase_price": 45,
           "sell_price": 55,
           "supply": "ABUNDANT",
           "activity": "STRONG"
         }
       ]
     },
     "analysis": {
       "total_trade_goods": 12,
       "high_value_goods": ["PLATINUM", "RARE_METALS"],
       "activity_level": "HIGH"
     }
   }
}
```

## 🚀 Quick Start: System Exploration Workflow

For the best experience with Claude Desktop, follow this recommended workflow:

### 1. **Understand Your Current Situation**
```
current_location
```
This shows where all your ships are and what facilities are nearby.

### 2. **Explore Your Current System**
```
system_overview system_symbol=X1-YOUR-SYSTEM
```
Gets a complete overview of facilities, shipyards, markets, and opportunities.

### 3. **Find Specific Facilities**
```
find_waypoints system_symbol=X1-YOUR-SYSTEM trait=SHIPYARD
find_waypoints system_symbol=X1-YOUR-SYSTEM trait=MARKETPLACE
```
Quickly locate exactly what you need.

### 4. **Navigate and Take Action**
```
navigate_ship ship_symbol=YOUR_SHIP waypoint_symbol=TARGET_WAYPOINT
purchase_ship ship_type=SHIP_MINING_DRONE waypoint_symbol=SHIPYARD_WAYPOINT
```

### 5. **Check Market Opportunities**
Use resources like: `spacetraders://systems/X1-SYSTEM/waypoints/X1-MARKETPLACE/market`

**This workflow eliminates guessing and provides Claude with complete system knowledge for strategic decision-making.**

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

### Smart Workflow for Contract Management

1. **Check Status**: "Get my status summary" - See credits, ships, contracts
2. **Analyze Contracts**: "Get contract info" - See requirements and profitability  
3. **Check Fleet**: "Analyze fleet capabilities" - See if you have the right ships
4. **Buy Ships**: "Purchase a SHIP_MINING_DRONE at X1-SYSTEM-SHIPYARD" - Get needed ships
5. **Accept Contract**: "Accept contract [contract-id]" - Commit to the contract

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

  📄 Contract Details:
    • PROCUREMENT (ID: clq123abc)
      - Faction: COSMIC
      - Status: ⏳ Available (Accept by: 2025-12-25T23:59:59.000Z)
      - Payment: 25000 credits (5000 on accept, 20000 on completion)
      - Deliveries:
        * 100 units of IRON_ORE to X1-FM66-B2 (0/100 completed)
```

#### `get_contract_info`

Get detailed information about contracts, including specific contract analysis and acceptance recommendations.

**Parameters:**
- `contract_id` (string, optional): Specific contract ID to analyze. If not provided, shows all contracts
- `include_fulfilled` (boolean, default: false): Include completed contracts in results

**Usage Examples:**
- "Get contract info"
- "Get contract info for clq123abc"
- "Show me contract details with include_fulfilled=true"

**Sample Output:**
```
📋 Contract Details: clq123abc

ID: clq123abc
Status: ⏳ Available
Type: PROCUREMENT
Faction: COSMIC
Payment: 25000 credits total
  • On Accept: 5000 credits
  • On Fulfill: 20000 credits
Accept By: 2025-12-25T23:59:59.000Z
Complete By: 2025-12-30T23:59:59.000Z

Delivery Requirements:
1. 🔴 IRON_ORE (0/100 units) → X1-FM66-B2
   Need 100 more units

Analysis:
• Profit margin: 80.0% (20000 of 25000 credits on completion)
• Requires cargo space and delivery logistics

💡 Available Actions:
• Use accept_contract with contract_id=clq123abc to accept this contract
```

#### `analyze_fleet_capabilities`

Analyze your current fleet's capabilities against contract requirements and suggest needed ships.

**Parameters:**
- `include_recommendations` (boolean, default: true): Include ship purchase recommendations

**Usage Examples:**
- "Analyze my fleet capabilities"
- "Check if my fleet can handle my contracts"
- "Analyze fleet capabilities with include_recommendations=false"

**Sample Output:**
```
🚢 Fleet Capability Analysis

Current Fleet:
• Total Ships: 2
• Total Cargo Capacity: 70 units
• Mining Capable Ships: 0
• Hauling Capable Ships: 1
• Combat Capable Ships: 1

Fleet Composition:
• COMMAND: 1 ship(s) - Multi-purpose command vessel
• HAULER: 1 ship(s) - Large cargo capacity transport

Contract Requirements Analysis:

📋 Contract clq123abc:
• Status: Available
• Required Materials:
  - IRON_ORE: 100 units
    *Requires mining*
• Required Cargo Space: 100 units

🔍 Gap Analysis:
❌ CRITICAL GAP: No mining ships available but contracts require mining
   *You need a SHIP_MINING_DRONE or similar mining vessel*
✅ Sufficient cargo capacity available

💡 Recommendations:
🔥 URGENT: Purchase a SHIP_MINING_DRONE
   • Required for mining contracts
   • Use purchase_ship with ship_type=SHIP_MINING_DRONE
   • Must be at a shipyard that sells mining drones
```

#### `purchase_ship`

Purchase a ship at a shipyard. Requires being docked at the shipyard and having sufficient credits.

**Parameters:**
- `ship_type` (string, required): Type of ship to purchase (e.g., SHIP_MINING_DRONE, SHIP_PROBE, SHIP_LIGHT_HAULER)
- `waypoint_symbol` (string, required): Waypoint symbol of the shipyard where you want to purchase the ship

**Usage Examples:**
- "Purchase a SHIP_MINING_DRONE at X1-FM66-B2"
- "Buy ship with ship_type=SHIP_PROBE and waypoint_symbol=X1-ABC-SHIPYARD"

**Sample Output:**
```
🚢 Ship Purchase Successful!

New Ship: TEST_SHIP_NEW (Test Mining Drone)
Type: SHIP_MINING_DRONE
Role: EXCAVATOR
Location: X1-FM66-B2 (Status: DOCKED)
Cost: 75000 credits
Remaining Credits: 125000
Total Ships: 3

Ship Specifications:
• Cargo Capacity: 30 units
• Fuel Capacity: 400 units
• Crew Capacity: 1/3

💡 Next Steps:
• Use get_status_summary to see your updated fleet
• Your new ship is ready for missions!
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

#### `orbit_ship`

Puts a ship into orbit around its current waypoint. Ship must be docked to use this command.

**Parameters:**
- `ship_symbol` (string, required): Symbol of the ship to put into orbit (e.g., 'SHIP_1234')

**Usage Examples:**
- "Put ship SHIP_1234 into orbit"
- "Orbit my command ship"

#### `dock_ship`

Docks a ship at its current waypoint. Ship must be in orbit to use this command.

**Parameters:**
- `ship_symbol` (string, required): Symbol of the ship to dock (e.g., 'SHIP_1234')

**Usage Examples:**
- "Dock ship SHIP_1234"
- "Dock my hauler at the station"

#### `navigate_ship`

Navigate a ship to a waypoint within the same system. Ship must be in orbit to navigate.

**Parameters:**
- `ship_symbol` (string, required): Symbol of the ship to navigate (e.g., 'SHIP_1234')
- `waypoint_symbol` (string, required): Symbol of the destination waypoint (e.g., 'X1-DF55-20250Z')

**Usage Examples:**
- "Navigate ship SHIP_1234 to X1-DF55-20250Z"
- "Send my mining ship to the asteroid field"

#### `patch_ship_nav`

Change a ship's navigation settings, particularly the flight mode.

**Parameters:**
- `ship_symbol` (string, required): Symbol of the ship to modify (e.g., 'SHIP_1234')
- `flight_mode` (string, required): Flight mode to set. Options: DRIFT (slowest, most fuel efficient), STEALTH (slow, hard to detect), CRUISE (balanced), BURN (fastest, most fuel consumption)

**Usage Examples:**
- "Set ship SHIP_1234 flight mode to BURN"
- "Change my ship to stealth mode"

#### `warp_ship`

Warp a ship to a waypoint in a different system. Ship must have a warp drive and be in orbit to warp.

**Parameters:**
- `ship_symbol` (string, required): Symbol of the ship to warp (e.g., 'SHIP_1234')
- `waypoint_symbol` (string, required): Symbol of the destination waypoint in another system (e.g., 'X1-AB12-34567Z')

**Usage Examples:**
- "Warp ship SHIP_1234 to X1-AB12-34567Z"
- "Send my explorer to the next system"

#### `jump_ship`

Jump a ship to a different system using a jump gate. Ship must have a jump drive and be in orbit to jump. Creates a cooldown period.

**Parameters:**
- `ship_symbol` (string, required): Symbol of the ship to jump (e.g., 'SHIP_1234')
- `system_symbol` (string, required): Symbol of the destination system (e.g., 'X1-AB12')

**Usage Examples:**
- "Jump ship SHIP_1234 to system X1-AB12"
- "Use jump gate to send my ship to the distant system"

#### `find_waypoints`

Find waypoints in a system by specific traits or facilities. Perfect for discovering shipyards, marketplaces, mining sites, and other key locations.

**Parameters:**
- `system_symbol` (string, required): System symbol to search in (e.g., 'X1-FM66')
- `trait` (string, required): Trait to search for (e.g., 'SHIPYARD', 'MARKETPLACE', 'ASTEROID_FIELD', 'JUMP_GATE')
- `waypoint_type` (string, optional): Filter by waypoint type (e.g., 'PLANET', 'MOON', 'ASTEROID')

**Usage Examples:**
- "Find all shipyards in system X1-FM66"
- "find_waypoints system_symbol=X1-FM66 trait=MARKETPLACE"
- "Show me asteroid fields for mining in X1-ABC123"

#### `system_overview`

Get a comprehensive overview of a system including all facilities, waypoints, and strategic opportunities. Includes shipyard details if available.

**Parameters:**
- `system_symbol` (string, required): System symbol to analyze (e.g., 'X1-FM66')
- `include_shipyards` (boolean, optional): Whether to include detailed shipyard information (default: true)

**Usage Examples:**
- "Give me an overview of system X1-FM66"
- "system_overview system_symbol=X1-ABC123"
- "Analyze the facilities and opportunities in my current system"

#### `current_location`

Analyze where your ships are currently located and what facilities are nearby. Provides strategic recommendations based on your fleet's positions.

**Parameters:**
- `include_nearby` (boolean, optional): Include nearby waypoints and facilities in the same system (default: true)
- `ship_symbol` (string, optional): Analyze specific ship only (otherwise analyzes all ships)

**Usage Examples:**
- "Where are my ships and what's nearby?"
- "current_location"
- "Analyze the location of ship SHIP_1234"

## 🎯 Recommended Claude Desktop Prompts

### For New Players
- "Show me my current situation and recommend next steps"
- "Find the nearest shipyard and tell me what ships are available"
- "What trading opportunities are available in my current system?"

### For System Exploration
- "Explore system X1-[SYSTEM] and identify the best opportunities"
- "Find all mining sites in my current system"
- "Show me a complete overview of system X1-[SYSTEM]"

### For Fleet Management
- "Analyze my fleet locations and suggest optimizations"
- "Which of my ships need fuel or are at capacity?"
- "What should I do with my docked ships?"

### For Trading and Commerce
- "Find the best trading routes in my current system"
- "Show me market prices at all marketplaces in X1-[SYSTEM]"
- "Where can I sell goods for the highest profit?"

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

### Tool Error Handling

#### Problem: Getting "status: 422" without details when using navigation tools

**Solution:** This has been fixed in the latest version. Navigation tools (orbit, dock, navigate, warp, jump, patch nav) now properly expose detailed error messages from the SpaceTraders API.

**What was fixed:**
- Previously, HTTP error responses only showed status codes (e.g., "failed to orbit ship, status: 422")
- Now, error responses include the full API error message with details about what went wrong
- Fixed Content-Type header issue where POST requests with empty bodies were rejected by the API
- This applies to all navigation methods: `orbit_ship`, `dock_ship`, `navigate_ship`, `patch_ship_nav`, `warp_ship`, `jump_ship`

**Example error messages you'll now see:**
```
API request failed with status 422: {"error":{"message":"Ship must be docked to enter orbit","code":4214,"data":{"shipSymbol":"SHIP_1234","shipStatus":"IN_TRANSIT"}}}
```

Instead of just:
```
failed to orbit ship, status: 422
```

Or the Content-Type error:
```
You specified a 'Content-Type' header of 'application/json', but the request body is an empty string (which can't be parsed as valid JSON). Send an empty object (e.g. {}) instead.
```

**Common 422 errors and their meanings:**
- **Orbit:** "Ship must be docked to enter orbit" - Ship is not at a waypoint or is in transit
- **Dock:** "Ship must be in orbit to dock" - Ship is not in orbit around a waypoint
- **Navigate:** "Ship must be in orbit to navigate" - Ship is docked or in transit
- **Warp:** "Ship does not have warp drive installed" - Ship lacks warp capability
- **Jump:** "Ship does not have jump drive installed" - Ship lacks jump capability

### System Exploration and Discovery

#### Problem: Claude can't find waypoints with specific facilities

**Solution:** Use the new exploration tools to efficiently discover and navigate to facilities.

**New tools available:**

1. **`find_waypoints`** - Search for specific facilities by trait:
   ```
   find_waypoints system_symbol=X1-FM66 trait=SHIPYARD
   ```

2. **`system_overview`** - Get comprehensive system analysis:
   ```
   system_overview system_symbol=X1-FM66
   ```

3. **`current_location`** - Analyze your fleet's current positions:
   ```
   current_location
   ```

**Common traits to search for:**
- `SHIPYARD` - Build and buy ships
- `MARKETPLACE` - Trade goods and check prices
- `ASTEROID_FIELD` - Mine resources
- `JUMP_GATE` - Travel to other systems
- `FUEL_STATION` - Refuel ships

**Example workflow:**
```bash
# 1. Find all shipyards in current system
find_waypoints system_symbol=X1-FM66 trait=SHIPYARD

# 2. Check what ships are available
# Use resource: spacetraders://systems/X1-FM66/waypoints/X1-FM66-SHIPYARD/shipyard

# 3. Navigate to the shipyard
navigate_ship ship_symbol=SHIP_1234 waypoint_symbol=X1-FM66-SHIPYARD

# 4. Purchase a ship
purchase_ship ship_type=SHIP_MINING_DRONE waypoint_symbol=X1-FM66-SHIPYARD
```

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