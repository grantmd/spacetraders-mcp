# Available Tools

This document describes the MCP tools available in the SpaceTraders MCP Server. Tools allow you to perform actions and make changes to your SpaceTraders game state.

## How to Use Tools

Tools are automatically available to Claude Desktop through the MCP integration. Simply ask Claude to perform actions, and it will use the appropriate tools to execute your requests.

## Available Tools

### `get_status_summary`

**Purpose:** Get a comprehensive summary of your agent's current status.

**What it does:**
- Retrieves agent information (credits, headquarters, etc.)
- Lists all ships with their current status
- Shows active contracts and progress
- Provides a complete overview of your current situation

**Parameters:** None

**Example usage:**
"Show me my current status"

### `get_contract_info`

**Purpose:** Retrieve detailed information about contracts.

**Parameters:**
- `contract_id` (optional): Specific contract ID to retrieve. If not provided, returns all contracts.

**What it does:**
- Lists available and active contracts
- Shows contract requirements and rewards
- Displays progress on active contracts
- Provides contract deadlines and terms

**Example usage:**
"Get information about my contracts"
"Show me details for contract cl9s5c5yi0001js08v5h4x8mz"

### `analyze_fleet_capabilities`

**Purpose:** Analyze your fleet's current capabilities and composition.

**What it does:**
- Reviews all ships in your fleet
- Analyzes cargo capacity, fuel capacity, and specializations
- Identifies fleet strengths and weaknesses
- Suggests improvements or additions

**Parameters:** None

**Example usage:**
"Analyze my fleet capabilities"

### `purchase_ship`

**Purpose:** Purchase a new ship from a shipyard.

**Parameters:**
- `ship_type`: Type of ship to purchase (e.g., "SHIP_PROBE", "SHIP_MINING_DRONE")
- `waypoint_symbol`: Waypoint symbol where the shipyard is located

**What it does:**
- Purchases the specified ship type
- Automatically names the ship
- Returns details of the newly purchased ship

**Example usage:**
"Purchase a SHIP_PROBE at X1-DF55-20250Z"

### `refuel_ship`

**Purpose:** Refuel a ship at its current location.

**Parameters:**
- `ship_symbol`: Symbol of the ship to refuel
- `units` (optional): Number of fuel units to purchase. If not specified, refuels to full capacity
- `from_cargo` (optional): Whether to refuel from cargo instead of purchasing (default: false)

**What it does:**
- Refuels the specified ship
- Deducts credits for fuel purchase
- Updates ship's fuel status

**Example usage:**
"Refuel ship GHOST-01"
"Refuel GHOST-02 with 100 units"

### `extract_resources`

**Purpose:** Extract resources at the current waypoint using a mining ship.

**Parameters:**
- `ship_symbol`: Symbol of the ship to use for extraction

**What it does:**
- Extracts resources from the current waypoint
- Adds extracted materials to ship's cargo
- Consumes fuel and time for the extraction operation

**Requirements:**
- Ship must be at a waypoint with extractable resources
- Ship must have mining capabilities
- Ship must be in orbit

**Example usage:**
"Extract resources with GHOST-01"

### `jettison_cargo`

**Purpose:** Jettison (discard) cargo from a ship.

**Parameters:**
- `ship_symbol`: Symbol of the ship
- `cargo_symbol`: Symbol of the cargo item to jettison
- `units`: Number of units to jettison

**What it does:**
- Removes specified cargo from the ship
- Frees up cargo space
- Permanently destroys the jettisoned items

**Example usage:**
"Jettison 10 units of IRON_ORE from GHOST-01"

### `accept_contract`

**Purpose:** Accept a contract.

**Parameters:**
- `contract_id`: ID of the contract to accept

**What it does:**
- Accepts the specified contract
- Makes the contract active
- May provide initial resources or credits

**Example usage:**
"Accept contract cl9s5c5yi0001js08v5h4x8mz"

**📖 Detailed Documentation:** See [accept_contract.md](tools/accept_contract.md) for comprehensive examples, error handling, and JSON request/response formats.

### `orbit_ship`

**Purpose:** Put a ship into orbit around its current waypoint.

**Parameters:**
- `ship_symbol`: Symbol of the ship to orbit

**What it does:**
- Changes ship's navigation status to "IN_ORBIT"
- Allows the ship to extract resources, jump, or warp
- Required for many operations

**Example usage:**
"Put GHOST-01 into orbit"

### `dock_ship`

**Purpose:** Dock a ship at its current waypoint.

**Parameters:**
- `ship_symbol`: Symbol of the ship to dock

**What it does:**
- Changes ship's navigation status to "DOCKED"
- Allows the ship to trade, refuel, or purchase items
- Required for market operations

**Example usage:**
"Dock GHOST-01"

### `navigate_ship`

**Purpose:** Navigate a ship to a different waypoint within the same system.

**Parameters:**
- `ship_symbol`: Symbol of the ship to navigate
- `waypoint_symbol`: Destination waypoint symbol

**What it does:**
- Moves the ship to the specified waypoint
- Consumes fuel based on distance
- Takes time to complete the journey

**Requirements:**
- Ship must be in orbit
- Destination must be in the same system

**Example usage:**
"Navigate GHOST-01 to X1-DF55-20250Z"

### `patch_ship_nav`

**Purpose:** Update a ship's navigation configuration.

**Parameters:**
- `ship_symbol`: Symbol of the ship
- `flight_mode`: New flight mode ("CRUISE", "BURN", "DRIFT", "STEALTH")

**What it does:**
- Changes the ship's flight mode
- Affects speed, fuel consumption, and detectability
- Different modes have different trade-offs

**Flight modes:**
- **CRUISE**: Balanced speed and fuel consumption
- **BURN**: Maximum speed, high fuel consumption
- **DRIFT**: Minimum fuel consumption, slow speed
- **STEALTH**: Reduced detectability, moderate speed

**Example usage:**
"Set GHOST-01 flight mode to BURN"

### `warp_ship`

**Purpose:** Warp a ship to a different system.

**Parameters:**
- `ship_symbol`: Symbol of the ship to warp
- `waypoint_symbol`: Destination waypoint symbol (must be in a different system)

**What it does:**
- Moves the ship to a waypoint in another system
- Consumes significant fuel
- Requires a warp drive

**Requirements:**
- Ship must be in orbit
- Ship must have a warp drive
- Sufficient fuel for the jump

**Example usage:**
"Warp GHOST-01 to X1-GX37-40410B"

### `jump_ship`

**Purpose:** Jump a ship through a jump gate to another system.

**Parameters:**
- `ship_symbol`: Symbol of the ship to jump
- `waypoint_symbol`: Destination waypoint symbol (connected via jump gate)

**What it does:**
- Instantly moves the ship to another system via jump gate
- More fuel-efficient than warping
- Limited to connected systems

**Requirements:**
- Ship must be at a jump gate waypoint
- Destination must be connected via jump gate network

**Example usage:**
"Jump GHOST-01 to X1-GX37-40410B"

### `find_waypoints`

**Purpose:** Find waypoints in a system that match specific criteria.

**Parameters:**
- `system_symbol`: System to search in
- `waypoint_type` (optional): Filter by waypoint type
- `trait` (optional): Filter by waypoint trait

**What it does:**
- Searches for waypoints matching the specified criteria
- Returns a list of matching waypoints with their properties
- Useful for finding specific facilities or resources

**Example usage:**
"Find all shipyards in system X1-DF55"
"Find waypoints with MARKETPLACE trait in X1-DF55"

### `system_overview`

**Purpose:** Get a comprehensive overview of a star system.

**Parameters:**
- `system_symbol`: System to analyze

**What it does:**
- Provides detailed information about the system
- Lists all waypoints with their types and traits
- Identifies key facilities and resources
- Summarizes trading and strategic opportunities

**Example usage:**
"Get an overview of system X1-DF55"

### `current_location`

**Purpose:** Get detailed information about your ships' current locations.

**What it does:**
- Shows where each of your ships is currently located
- Provides waypoint details for each location
- Includes available facilities and resources at each location

**Parameters:** None

**Example usage:**
"Show me where all my ships are currently located"

## Tool Usage Tips

- **Check requirements:** Some tools require specific ship states (docked vs. orbiting)
- **Fuel management:** Navigation tools consume fuel - monitor your fuel levels
- **System boundaries:** Some operations are limited to the current system
- **Error handling:** Tools will provide clear error messages if requirements aren't met
- **Combine tools:** Use multiple tools together for complex operations

## Common Workflows

**Resource Extraction:**
1. `orbit_ship` - Put ship in orbit
2. `extract_resources` - Extract materials
3. `dock_ship` - Dock to sell resources

**System Exploration:**
1. `system_overview` - Get system information
2. `find_waypoints` - Locate specific facilities
3. `navigate_ship` - Move to points of interest

**Contract Completion:**
1. `get_contract_info` - Review contract requirements
2. `navigate_ship` - Move to required locations
3. `extract_resources` or trade as needed
4. `fulfill_contract` - Complete the contract for rewards

### `sell_cargo`

**Purpose:** Sell cargo from a ship at a marketplace.

**Parameters:**
- `ship_symbol`: Symbol of the ship to sell cargo from
- `cargo_symbol`: Symbol of the cargo item to sell (e.g., "IRON_ORE", "ALUMINUM_ORE", "FUEL")
- `units`: Number of units to sell

**What it does:**
- Sells the specified cargo at the current marketplace
- Adds credits to your account
- Removes cargo from the ship's inventory
- Frees up cargo space

**Requirements:**
- Ship must be docked at a waypoint with a marketplace
- Ship must have the specified cargo in inventory
- Marketplace must accept the cargo type

**Example usage:**
"Sell 50 units of IRON_ORE from GHOST-01"

### `buy_cargo`

**Purpose:** Purchase cargo for a ship at a marketplace.

**Parameters:**
- `ship_symbol`: Symbol of the ship to buy cargo for
- `cargo_symbol`: Symbol of the cargo item to buy (e.g., "FUEL", "FOOD", "MACHINERY")
- `units`: Number of units to buy

**What it does:**
- Purchases the specified cargo from the current marketplace
- Deducts credits from your account
- Adds cargo to the ship's inventory
- Consumes cargo space

**Requirements:**
- Ship must be docked at a waypoint with a marketplace
- You must have sufficient credits
- Ship must have sufficient cargo space
- Marketplace must sell the cargo type

**Example usage:**
"Buy 25 units of FUEL for GHOST-01"

### `fulfill_contract`

**Purpose:** Fulfill a contract by delivering all required cargo.

**Parameters:**
- `contract_id`: ID of the contract to fulfill

**What it does:**
- Completes the contract if all requirements are met
- Awards the fulfillment payment
- Marks the contract as fulfilled
- Improves faction reputation

**Requirements:**
- Contract must be accepted
- All delivery requirements must be satisfied
- Required cargo must be delivered to specified destinations

**Example usage:**
"Fulfill contract cl9s5c5yi0001js08v5h4x8mz"

## Trading Workflows

**Basic Trading Loop:**
1. `dock_ship` - Dock at a marketplace
2. `buy_cargo` - Purchase goods at low prices
3. `navigate_ship` - Travel to another marketplace
4. `dock_ship` - Dock at destination
5. `sell_cargo` - Sell goods at higher prices

**Contract Trading:**
1. `get_contract_info` - Review contract requirements
2. `buy_cargo` - Purchase required goods
3. `navigate_ship` - Travel to delivery location
4. `dock_ship` - Dock at delivery destination
5. `fulfill_contract` - Complete contract for rewards

### `scan_systems`

**Purpose:** Scan for systems around a ship using its sensors.

**Parameters:**
- `ship_symbol`: Symbol of the ship to scan with

**What it does:**
- Scans for nearby systems using ship's sensors
- Reveals undiscovered systems within range
- Provides system information including type, location, and factions
- Has a cooldown period after use

**Requirements:**
- Ship must have appropriate scanning equipment
- Ship must not be on scan cooldown

**Example usage:**
"Scan for systems with GHOST-01"

### `scan_waypoints`

**Purpose:** Scan for waypoints around a ship using its sensors.

**Parameters:**
- `ship_symbol`: Symbol of the ship to scan with

**What it does:**
- Scans for nearby waypoints using ship's sensors
- Reveals hidden waypoints and asteroid fields
- Provides waypoint information including traits and resources
- Has a cooldown period after use

**Requirements:**
- Ship must have appropriate scanning equipment
- Ship must not be on scan cooldown

**Example usage:**
"Scan for waypoints with GHOST-01"

### `scan_ships`

**Purpose:** Scan for ships around a ship using its sensors.

**Parameters:**
- `ship_symbol`: Symbol of the ship to scan with

**What it does:**
- Scans for nearby ships using ship's sensors
- Reveals ship information including faction, role, and equipment
- Provides tactical assessment of detected vessels
- Has a cooldown period after use

**Requirements:**
- Ship must have appropriate scanning equipment
- Ship must not be on scan cooldown

**Example usage:**
"Scan for ships with GHOST-01"

### `repair_ship`

**Purpose:** Repair a ship at a shipyard.

**Parameters:**
- `ship_symbol`: Symbol of the ship to repair

**What it does:**
- Repairs all ship components to full integrity
- Restores frame, reactor, engine, modules, and mounts
- Costs credits based on damage amount
- Returns ship to optimal operational condition

**Requirements:**
- Ship must be docked at a waypoint with a shipyard
- You must have sufficient credits for repairs

**Example usage:**
"Repair GHOST-01"

## Advanced Exploration Workflows

**System Reconnaissance:**
1. `scan_systems` - Discover nearby systems
2. `jump_ship` or `warp_ship` - Travel to new systems
3. `system_overview` - Get detailed system information
4. `scan_waypoints` - Discover hidden resources

**Tactical Intelligence:**
1. `scan_ships` - Identify nearby vessels
2. Assess threat levels and opportunities
3. Plan movement based on detected ships
4. Monitor faction presence and activity

**Fleet Maintenance:**
1. Monitor ship integrity regularly
2. `repair_ship` - Maintain ships at optimal condition
3. Plan repair schedules to minimize downtime
4. Budget for maintenance costs