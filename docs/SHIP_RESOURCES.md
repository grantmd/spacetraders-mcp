# Ship Resources Documentation

This document describes the individual ship resources and cooldown functionality implemented in the SpaceTraders MCP server.

## Overview

The SpaceTraders MCP server now supports detailed, real-time information about individual ships through dedicated resources. These resources provide enhanced analysis and operational status beyond the basic ship list.

## New Resources

### 1. Individual Ship Resource

**URI Pattern:** `spacetraders://ships/{shipSymbol}`

**Description:** Provides comprehensive information about a specific ship including detailed analysis of its operational status, capabilities, and recommendations.

**Example URIs:**
- `spacetraders://ships/MYSHIP-1`
- `spacetraders://ships/TRADER-A2`
- `spacetraders://ships/MINER-01`

**Response Structure:**
```json
{
  "ship": {
    "symbol": "MYSHIP-1",
    "registration": { ... },
    "nav": { ... },
    "crew": { ... },
    "frame": { ... },
    "reactor": { ... },
    "engine": { ... },
    "cooldown": { ... },
    "modules": [ ... ],
    "mounts": [ ... ],
    "cargo": { ... },
    "fuel": { ... }
  },
  "analysis": {
    "status": {
      "status": "operational",
      "message": "Ship is operational",
      "conditions": ["docked"],
      "nav_status": "DOCKED",
      "location": "X1-FM66-A1",
      "system": "X1-FM66"
    },
    "cooldown_status": {
      "active": false,
      "remaining_seconds": 0,
      "status": "ready",
      "message": "Ship is ready for actions"
    },
    "cargo_utilization": {
      "capacity": 100,
      "units": 25,
      "available_space": 75,
      "utilization_percent": 25.0,
      "status": "low",
      "message": "Plenty of cargo space available",
      "composition": { ... },
      "item_count": 2
    },
    "location_analysis": {
      "system": "X1-FM66",
      "waypoint": "X1-FM66-A1",
      "status": "DOCKED"
    },
    "operational_status": {
      "can_act": true,
      "can_dock": false,
      "can_orbit": true,
      "can_navigate": false,
      "can_extract": false,
      "can_trade": true,
      "can_refuel": true
    }
  },
  "capabilities": {
    "primary_capabilities": {
      "mining": true,
      "scanning": false,
      "trading": true,
      "combat": false,
      "surveying": false
    },
    "mount_capabilities": ["mining"],
    "cargo_capacity": 100,
    "fuel_capacity": 400,
    "crew_capacity": 10
  },
  "recommendations": [
    "Ship is docked - can access market, shipyard, and refueling",
    "Cargo hold has good capacity remaining"
  ],
  "meta": {
    "last_updated": "2024-01-15T10:30:00Z",
    "ship_symbol": "MYSHIP-1"
  }
}
```

### 2. Ship Cooldown Resource

**URI Pattern:** `spacetraders://ships/{shipSymbol}/cooldown`

**Description:** Provides real-time cooldown status for a specific ship, including remaining time, operational availability, and recommendations.

**Example URIs:**
- `spacetraders://ships/MYSHIP-1/cooldown`
- `spacetraders://ships/TRADER-A2/cooldown`

**Response Structure (Active Cooldown):**
```json
{
  "ship_symbol": "MYSHIP-1",
  "cooldown": {
    "active": true,
    "remaining_seconds": 127,
    "total_seconds": 300,
    "expiration": "2024-01-15T10:32:07Z"
  },
  "status": {
    "operational": false,
    "message": "Short cooldown active - almost ready",
    "priority": "short",
    "icon": "🟠"
  },
  "timing": {
    "ready_at": "2024-01-15T10:32:07Z",
    "expiration_time": "2024-01-15T10:32:07Z",
    "time_display": "2m 7s",
    "elapsed_seconds": 173,
    "progress_percent": 57.7
  },
  "actions": {
    "can_extract": false,
    "can_scan": false,
    "can_jump": false,
    "can_navigate": true,
    "can_survey": false,
    "blocked_actions": [
      "extract_resources",
      "scan_systems",
      "scan_waypoints",
      "scan_ships",
      "create_survey",
      "jump_ship",
      "siphon_resources"
    ]
  },
  "recommendations": [
    "Short cooldown - prepare for next action",
    "Ship will be ready shortly",
    "Navigation and trading are still available during cooldown",
    "Use 'get_ship_details' to check full ship status"
  ],
  "meta": {
    "last_checked": "2024-01-15T10:30:00Z",
    "ship_symbol": "MYSHIP-1"
  }
}
```

**Response Structure (No Cooldown):**
```json
{
  "ship_symbol": "MYSHIP-1",
  "cooldown": {
    "active": false,
    "remaining_seconds": 0,
    "total_seconds": 0,
    "expiration": null
  },
  "status": {
    "operational": true,
    "message": "Ship is ready for all actions",
    "priority": "ready",
    "icon": "🟢"
  },
  "timing": {
    "ready_at": "2024-01-15T10:30:00Z",
    "time_display": "Ready now"
  },
  "actions": {
    "can_extract": true,
    "can_scan": true,
    "can_jump": true,
    "can_navigate": true,
    "can_survey": true
  },
  "recommendations": [
    "Ship is ready for all actions",
    "Good time to plan next operation"
  ],
  "meta": {
    "last_checked": "2024-01-15T10:30:00Z",
    "ship_symbol": "MYSHIP-1"
  }
}
```

## Key Features

### Enhanced Analysis

**Ship Status Analysis:**
- Operational status determination
- Condition tracking (docked, in_orbit, in_transit, etc.)
- Fuel level monitoring
- Cargo capacity analysis

**Cooldown Intelligence:**
- Real-time remaining time calculation
- Human-readable time display
- Priority classification (ready, short, medium, long)
- Progress tracking with percentage completion

**Capability Detection:**
- Automatic detection of ship capabilities based on mounts
- Mining, scanning, trading, combat, and surveying capabilities
- Mount-specific capability mapping

### Operational Status

**Action Availability:**
- Real-time determination of what actions the ship can perform
- Cooldown-aware action blocking
- Location-dependent action availability
- Clear indication of blocked actions during cooldown

**Smart Recommendations:**
- Context-aware suggestions based on ship status
- Cooldown-specific recommendations
- Cargo and fuel management suggestions
- Location-based opportunity identification

### Time Display

**Human-Readable Formats:**
- Cooldown times in hours, minutes, and seconds
- Progress percentages
- Ready-at timestamps
- Time remaining displays

## Usage Examples

### Check Ship Status
```bash
# Get comprehensive ship information
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "resources/read",
    "params": {
      "uri": "spacetraders://ships/MYSHIP-1"
    }
  }'
```

### Monitor Cooldown
```bash
# Get real-time cooldown status
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "resources/read",
    "params": {
      "uri": "spacetraders://ships/MYSHIP-1/cooldown"
    }
  }'
```

### Batch Ship Monitoring
```bash
# Check multiple ships
for ship in MYSHIP-1 MYSHIP-2 MYSHIP-3; do
  echo "Checking $ship cooldown..."
  curl -X POST http://localhost:8080/mcp \
    -H "Content-Type: application/json" \
    -d "{
      \"jsonrpc\": \"2.0\",
      \"id\": 1,
      \"method\": \"resources/read\",
      \"params\": {
        \"uri\": \"spacetraders://ships/$ship/cooldown\"
      }
    }"
done
```

## Implementation Details

### Client Methods

**New Client Methods:**
- `GetShip(shipSymbol string) (*Ship, error)` - Get individual ship details
- `GetShipCooldown(shipSymbol string) (*Cooldown, error)` - Get ship cooldown status

**Special Handling:**
- The GetShipCooldown method properly handles 204 No Content responses when no cooldown is active
- Returns nil for cooldown when ship is ready
- Includes comprehensive error handling for invalid ship symbols

### Resource Handlers

**ShipResource:**
- Extracts ship symbol from URI using regex pattern matching
- Combines ship data with cooldown information
- Performs comprehensive analysis and generates recommendations
- Handles invalid ship symbols gracefully

**ShipCooldownResource:**
- Dedicated cooldown monitoring with real-time status
- Priority-based status classification
- Action availability determination
- Time formatting and progress calculation

### Performance Considerations

**Caching:**
- Resources fetch fresh data on each request for real-time accuracy
- Cooldown information is time-sensitive and not cached
- Individual ship data includes cooldown for efficiency

**API Calls:**
- Individual ship resource makes 1-2 API calls (ship + optional cooldown)
- Cooldown resource makes 1 API call
- Error handling prevents cascading failures

## Error Handling

### Invalid Ship Symbols
- Returns descriptive error messages for invalid ship symbols
- Maintains consistent error response format
- Logs errors for debugging

### Network Errors
- Graceful handling of API failures
- Informative error messages
- Proper HTTP status code handling

### Cooldown Edge Cases
- Handles 204 No Content responses properly
- Manages expired cooldowns
- Deals with missing cooldown data

## Integration with Existing Tools

### Enhanced Tool Responses
The ship resources complement existing tools by providing:
- Pre-action ship status checks
- Post-action cooldown monitoring
- Operational readiness verification

### Workflow Integration
1. **Pre-Action Check:** Use ship resource to verify operational status
2. **Execute Action:** Use existing tools (extract, scan, navigate)
3. **Post-Action Monitor:** Use cooldown resource to track readiness
4. **Plan Next Action:** Use recommendations for decision making

## Best Practices

### Monitoring Strategy
- Use cooldown resources for real-time monitoring
- Use ship resources for comprehensive status checks
- Monitor multiple ships in parallel for fleet management

### Action Planning
- Check cooldown status before planning ship actions
- Use recommendations for optimal timing
- Monitor cargo and fuel levels for logistics planning

### Fleet Management
- Use ship resources to assess fleet readiness
- Monitor cooldowns across all ships
- Plan actions based on ship availability

## Future Enhancements

### Planned Features
- Fleet-wide cooldown summary resource
- Cooldown notification system
- Automated fleet scheduling based on cooldown status
- Historical cooldown analysis

### Potential Integrations
- Integration with contract delivery workflows
- Mining operation optimization
- Trading route planning
- Fleet coordination tools

## Troubleshooting

### Common Issues

**Invalid Ship Symbol:**
```
Error: Invalid ship cooldown resource URI. 
Expected format: spacetraders://ships/{shipSymbol}/cooldown
```
Solution: Verify ship symbol exists and URI format is correct.

**Network Timeouts:**
```
Error: Failed to fetch cooldown for ship MYSHIP-1: context deadline exceeded
```
Solution: Check network connectivity and API availability.

**Ship Not Found:**
```
Error: Failed to get ship: ship not found
```
Solution: Verify ship symbol exists and belongs to your agent.

### Debug Mode
Enable debug logging to see detailed API call information:
```bash
SPACETRADERS_DEBUG=true ./spacetraders-mcp
```

## Summary

The new ship resources provide comprehensive, real-time ship monitoring capabilities that enhance the SpaceTraders MCP server's functionality. They offer detailed analysis, operational status, and actionable recommendations for effective fleet management.

Key benefits:
- Real-time cooldown monitoring
- Enhanced operational intelligence
- Actionable recommendations
- Comprehensive ship analysis
- Seamless integration with existing tools

These resources form the foundation for advanced fleet management and automated ship operations in the SpaceTraders universe.