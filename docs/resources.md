# Available Resources

This document describes the MCP resources available in the SpaceTraders MCP Server. Resources provide read-only access to your SpaceTraders game data and are automatically refreshed.

## How to Use Resources

Resources are accessed using the format `spacetraders://resource/path`. Claude Desktop will automatically fetch and display this data when you reference these URIs in your prompts.

## Available Resources

### `spacetraders://agent/info`

Provides information about your SpaceTraders agent.

**Response Structure:**
```
agent
├── accountId
├── credits
├── headquarters
├── shipCount
├── startingFaction
└── symbol
```

### `spacetraders://ships/list`

Lists all ships in your fleet with detailed information.

**Response Structure:**
```
ships
├── symbol
├── registration
│   ├── name
│   ├── factionSymbol
│   └── role
├── nav
│   ├── systemSymbol
│   ├── waypointSymbol
│   ├── status
│   └── flightMode
├── cargo
│   ├── capacity
│   ├── units
│   └── inventory
└── fuel
    ├── current
    └── capacity

meta
└── count
```

### `spacetraders://systems/{systemSymbol}/waypoints`

Lists all waypoints in a specific system with their properties.

**Usage:** Replace `{systemSymbol}` with the actual system symbol (e.g., `spacetraders://systems/X1-DF55/waypoints`)

**Response Structure:**
```
system

waypoints
├── symbol
├── type
├── x
├── y
├── orbitals
│   └── symbol
└── traits
    ├── symbol
    ├── name
    └── description

summary
├── total
├── byType
│   ├── PLANET
│   ├── MOON
│   └── ASTEROID
├── shipyards
└── markets
```

### `spacetraders://systems/{systemSymbol}/waypoints/{waypointSymbol}/shipyard`

Provides detailed information about a shipyard at a specific waypoint.

**Usage:** Replace both `{systemSymbol}` and `{waypointSymbol}` with actual values

**Response Structure:**
```
system
waypoint

shipyard
├── symbol
├── shipTypes
│   └── type
├── ships
│   ├── type
│   ├── name
│   ├── description
│   ├── supply
│   ├── purchasePrice
│   ├── frame
│   │   ├── symbol
│   │   ├── name
│   │   ├── moduleSlots
│   │   ├── mountingPoints
│   │   └── fuelCapacity
│   ├── reactor
│   │   ├── symbol
│   │   ├── name
│   │   └── powerOutput
│   ├── engine
│   │   ├── symbol
│   │   ├── name
│   │   └── speed
│   └── crew
│       ├── required
│       └── capacity
└── modificationsFee

summary
├── availableShipTypes
├── totalShipsAvailable
├── priceRange
│   ├── min
│   └── max
├── shipsByType
│   ├── SHIP_PROBE
│   └── SHIP_MINING_DRONE
├── shipsBySupply
│   ├── ABUNDANT
│   └── MODERATE
├── modificationsFee
└── recentTransactions
```

### `spacetraders://systems/{systemSymbol}/waypoints/{waypointSymbol}/market`

Provides market information for a specific waypoint.

**Usage:** Replace both `{systemSymbol}` and `{waypointSymbol}` with actual values

**Response Structure:**
```
system
waypoint

market
├── symbol
├── exports
│   ├── symbol
│   ├── name
│   └── description
├── imports
│   ├── symbol
│   ├── name
│   └── description
└── trade_goods
    ├── symbol
    ├── type
    ├── purchase_price
    ├── sell_price
    ├── supply
    └── activity

analysis
├── total_trade_goods
├── high_value_goods
└── activity_level
```

## Important Notes

- Resources are **read-only** - they provide information but cannot be used to make changes
- Data is automatically refreshed when accessed
- Some resources require specific parameters (system symbols, waypoint symbols)
- Resources work seamlessly with Claude Desktop's MCP integration