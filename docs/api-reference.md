# API Reference

This document provides a comprehensive reference for the SpaceTraders MCP Server API, including all resources, tools, and prompts available through the MCP protocol.

## Overview

The SpaceTraders MCP Server implements the Model Context Protocol (MCP) to provide seamless integration between Claude Desktop and the SpaceTraders API. It exposes three main types of capabilities:

- **Resources**: Read-only access to game data
- **Tools**: Actions that can modify game state
- **Prompts**: Intelligent assistance for common tasks

## MCP Protocol Compliance

The server implements MCP version 2024-11-05 and supports:

- Resource listing and reading
- Tool listing and execution
- Prompt listing and execution
- Proper error handling and status codes
- Structured logging and debugging

## Authentication

All API calls require valid SpaceTraders credentials:

```bash
export SPACETRADERS_TOKEN=your_api_token
export SPACETRADERS_AGENT_SYMBOL=your_agent_symbol
```

## Resources API

Resources provide read-only access to SpaceTraders data using URI-based addressing.

### Base URI Format

```
spacetraders://{resource_type}/{path}
```

### Available Resources

#### Agent Information

**URI**: `spacetraders://agent/info`

**Description**: Returns comprehensive information about your SpaceTraders agent.

**Response Schema**:
```json
{
  "agent": {
    "accountId": "string",
    "symbol": "string",
    "headquarters": "string",
    "credits": "number",
    "startingFaction": "string",
    "shipCount": "number"
  }
}
```

**Example Response**:
```json
{
  "agent": {
    "accountId": "clk3f8b9a0000mp08k5q9r4q3",
    "symbol": "GHOST",
    "headquarters": "X1-DF55-20250Z",
    "credits": 150000,
    "startingFaction": "COSMIC",
    "shipCount": 3
  }
}
```

#### Ship Fleet

**URI**: `spacetraders://ships/list`

**Description**: Lists all ships in your fleet with detailed information.

**Response Schema**:
```json
{
  "ships": [
    {
      "symbol": "string",
      "registration": {
        "name": "string",
        "factionSymbol": "string",
        "role": "string"
      },
      "nav": {
        "systemSymbol": "string",
        "waypointSymbol": "string",
        "route": {
          "destination": {
            "symbol": "string",
            "type": "string",
            "systemSymbol": "string",
            "x": "number",
            "y": "number"
          },
          "origin": {
            "symbol": "string",
            "type": "string",
            "systemSymbol": "string",
            "x": "number",
            "y": "number"
          },
          "departureTime": "string",
          "arrival": "string"
        },
        "status": "string",
        "flightMode": "string"
      },
      "crew": {
        "current": "number",
        "required": "number",
        "capacity": "number",
        "rotation": "string",
        "morale": "number",
        "wages": "number"
      },
      "frame": {
        "symbol": "string",
        "name": "string",
        "description": "string",
        "condition": "number",
        "moduleSlots": "number",
        "mountingPoints": "number",
        "fuelCapacity": "number",
        "requirements": {
          "power": "number",
          "crew": "number",
          "slots": "number"
        }
      },
      "reactor": {
        "symbol": "string",
        "name": "string",
        "description": "string",
        "condition": "number",
        "powerOutput": "number",
        "requirements": {
          "power": "number",
          "crew": "number",
          "slots": "number"
        }
      },
      "engine": {
        "symbol": "string",
        "name": "string",
        "description": "string",
        "condition": "number",
        "speed": "number",
        "requirements": {
          "power": "number",
          "crew": "number",
          "slots": "number"
        }
      },
      "modules": [
        {
          "symbol": "string",
          "capacity": "number",
          "range": "number",
          "name": "string",
          "description": "string",
          "requirements": {
            "power": "number",
            "crew": "number",
            "slots": "number"
          }
        }
      ],
      "mounts": [
        {
          "symbol": "string",
          "name": "string",
          "description": "string",
          "strength": "number",
          "deposits": ["string"],
          "requirements": {
            "power": "number",
            "crew": "number",
            "slots": "number"
          }
        }
      ],
      "cargo": {
        "capacity": "number",
        "units": "number",
        "inventory": [
          {
            "symbol": "string",
            "name": "string",
            "description": "string",
            "units": "number"
          }
        ]
      },
      "fuel": {
        "current": "number",
        "capacity": "number",
        "consumed": {
          "amount": "number",
          "timestamp": "string"
        }
      }
    }
  ],
  "meta": {
    "total": "number",
    "page": "number",
    "limit": "number"
  }
}
```

#### System Waypoints

**URI**: `spacetraders://systems/{systemSymbol}/waypoints`

**Parameters**:
- `systemSymbol`: The symbol of the system to query (e.g., "X1-DF55")

**Description**: Lists all waypoints in a specific system with their properties and traits.

**Response Schema**:
```json
{
  "data": [
    {
      "symbol": "string",
      "type": "string",
      "systemSymbol": "string",
      "x": "number",
      "y": "number",
      "orbitals": [
        {
          "symbol": "string"
        }
      ],
      "traits": [
        {
          "symbol": "string",
          "name": "string",
          "description": "string"
        }
      ],
      "modifiers": [
        {
          "symbol": "string",
          "name": "string",
          "description": "string"
        }
      ],
      "chart": {
        "waypointSymbol": "string",
        "submittedBy": "string",
        "submittedOn": "string"
      },
      "faction": {
        "symbol": "string"
      }
    }
  ],
  "meta": {
    "total": "number",
    "page": "number",
    "limit": "number"
  }
}
```

#### Shipyard Information

**URI**: `spacetraders://systems/{systemSymbol}/waypoints/{waypointSymbol}/shipyard`

**Parameters**:
- `systemSymbol`: The system symbol
- `waypointSymbol`: The waypoint symbol with a shipyard

**Description**: Provides detailed information about a shipyard's available ships and services.

#### Market Information

**URI**: `spacetraders://systems/{systemSymbol}/waypoints/{waypointSymbol}/market`

**Parameters**:
- `systemSymbol`: The system symbol
- `waypointSymbol`: The waypoint symbol with a market

**Description**: Returns market data including imports, exports, and current trade goods.

## Tools API

Tools allow you to perform actions that modify your SpaceTraders game state.

### Tool Execution Format

Tools are called using the MCP tools protocol:

```json
{
  "method": "tools/call",
  "params": {
    "name": "tool_name",
    "arguments": {
      "param1": "value1",
      "param2": "value2"
    }
  }
}
```

### Available Tools

#### get_status_summary

**Description**: Get a comprehensive summary of your agent's current status.

**Parameters**: None

**Returns**: Complete status including agent info, ships, contracts, and recommendations.

#### get_contract_info

**Description**: Retrieve detailed information about contracts.

**Parameters**:
- `contract_id` (optional, string): Specific contract ID to retrieve

**Returns**: Contract information including terms, progress, and deadlines.

#### analyze_fleet_capabilities

**Description**: Analyze your fleet's capabilities and suggest improvements.

**Parameters**: None

**Returns**: Fleet analysis with recommendations for optimization.

#### purchase_ship

**Description**: Purchase a new ship from a shipyard.

**Parameters**:
- `ship_type` (required, string): Type of ship to purchase
- `waypoint_symbol` (required, string): Waypoint with the shipyard

**Returns**: Details of the newly purchased ship.

#### refuel_ship

**Description**: Refuel a ship at its current location.

**Parameters**:
- `ship_symbol` (required, string): Symbol of the ship to refuel
- `units` (optional, number): Number of fuel units to purchase
- `from_cargo` (optional, boolean): Whether to refuel from cargo

**Returns**: Updated ship fuel status and transaction details.

#### extract_resources

**Description**: Extract resources at the current waypoint.

**Parameters**:
- `ship_symbol` (required, string): Symbol of the mining ship

**Returns**: Extraction results and updated cargo status.

#### jettison_cargo

**Description**: Jettison cargo from a ship.

**Parameters**:
- `ship_symbol` (required, string): Symbol of the ship
- `cargo_symbol` (required, string): Symbol of the cargo to jettison
- `units` (required, number): Number of units to jettison

**Returns**: Updated cargo status.

#### accept_contract

**Description**: Accept a contract.

**Parameters**:
- `contract_id` (required, string): ID of the contract to accept

**Returns**: Updated contract status.

#### orbit_ship

**Description**: Put a ship into orbit.

**Parameters**:
- `ship_symbol` (required, string): Symbol of the ship

**Returns**: Updated ship navigation status.

#### dock_ship

**Description**: Dock a ship at its current waypoint.

**Parameters**:
- `ship_symbol` (required, string): Symbol of the ship

**Returns**: Updated ship navigation status.

#### navigate_ship

**Description**: Navigate a ship to a different waypoint.

**Parameters**:
- `ship_symbol` (required, string): Symbol of the ship
- `waypoint_symbol` (required, string): Destination waypoint

**Returns**: Navigation details and fuel consumption.

#### patch_ship_nav

**Description**: Update ship navigation settings.

**Parameters**:
- `ship_symbol` (required, string): Symbol of the ship
- `flight_mode` (required, string): New flight mode ("CRUISE", "BURN", "DRIFT", "STEALTH")

**Returns**: Updated navigation configuration.

#### warp_ship

**Description**: Warp a ship to another system.

**Parameters**:
- `ship_symbol` (required, string): Symbol of the ship
- `waypoint_symbol` (required, string): Destination waypoint in another system

**Returns**: Warp results and fuel consumption.

#### jump_ship

**Description**: Jump a ship through a jump gate.

**Parameters**:
- `ship_symbol` (required, string): Symbol of the ship
- `waypoint_symbol` (required, string): Destination waypoint

**Returns**: Jump results and updated ship status.

#### find_waypoints

**Description**: Find waypoints matching specific criteria.

**Parameters**:
- `system_symbol` (required, string): System to search
- `waypoint_type` (optional, string): Filter by waypoint type
- `trait` (optional, string): Filter by waypoint trait

**Returns**: List of matching waypoints.

#### system_overview

**Description**: Get a comprehensive overview of a system.

**Parameters**:
- `system_symbol` (required, string): System to analyze

**Returns**: Complete system analysis with strategic information.

#### current_location

**Description**: Get information about your ships' current locations.

**Parameters**: None

**Returns**: Location details for all ships in your fleet.

## Prompts API

Prompts provide intelligent assistance for common SpaceTraders tasks.

### Available Prompts

#### status_check

**Description**: Comprehensive overview of your current SpaceTraders situation.

**Usage**: Provides analysis of agent status, fleet, contracts, and recommendations.

#### explore_system

**Description**: Intelligent system exploration and analysis.

**Usage**: Maps systems, identifies facilities, analyzes opportunities.

#### contract_strategy

**Description**: Strategic contract analysis and recommendations.

**Usage**: Analyzes contracts, calculates profits, suggests strategies.

#### fleet_optimization

**Description**: Fleet composition analysis and optimization suggestions.

**Usage**: Reviews fleet capabilities, suggests improvements and purchases.

## Error Handling

The API uses standard HTTP status codes and structured error responses:

### Error Response Format

```json
{
  "error": {
    "code": "number",
    "message": "string",
    "details": "object"
  }
}
```

### Common Error Codes

- **400 Bad Request**: Invalid parameters or request format
- **401 Unauthorized**: Invalid or missing authentication
- **403 Forbidden**: Insufficient permissions
- **404 Not Found**: Resource not found
- **422 Unprocessable Entity**: Request validation failed
- **429 Too Many Requests**: Rate limit exceeded
- **500 Internal Server Error**: Server error

### SpaceTraders-Specific Errors

- **4000**: Cooldown period active
- **4001**: Insufficient credits
- **4002**: Ship not found
- **4003**: Waypoint not found
- **4004**: Ship already at destination
- **4005**: Contract already accepted

## Rate Limits

The server respects SpaceTraders API rate limits:

- **Burst**: 10 requests per second
- **Sustained**: 2 requests per second

Rate limit headers are included in responses:
- `X-RateLimit-Limit`: Request limit per time window
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Time when rate limit resets

## Logging and Debugging

### Log Levels

- **ERROR**: Critical errors and failures
- **WARN**: Warnings and recoverable issues
- **INFO**: General operational information
- **DEBUG**: Detailed debugging information

### Debug Mode

Enable debug logging:

```bash
export DEBUG=1
```

Debug mode provides:
- Detailed request/response logging
- MCP protocol message tracing
- Performance timing information
- Internal state debugging

## Version Information

Current API version: **v1.0.0**

The API follows semantic versioning:
- Major version changes indicate breaking changes
- Minor version changes add new functionality
- Patch version changes include bug fixes and improvements

## Support

For API support and questions:
- Check the [Troubleshooting Guide](troubleshooting.md)
- Review [GitHub Issues](https://github.com/your-repo/issues)
- Consult the [SpaceTraders API Documentation](https://docs.spacetraders.io/)