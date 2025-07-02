# Understanding MCP Resources

This document explains how MCP (Model Context Protocol) resources work in the SpaceTraders MCP Server and how to effectively use them with Claude Desktop.

## What Are MCP Resources?

MCP resources are a way to provide **read-only access** to external data sources through a standardized URI-based system. In the context of SpaceTraders, resources allow Claude Desktop to automatically fetch and understand your current game state without you having to manually provide information.

## How MCP Resources Work

### Traditional Approach vs. MCP Resources

**Without MCP Resources:**
```
You: "I have 3 ships: GHOST-01 at X1-DF55-A, GHOST-02 at X1-DF55-B, and GHOST-03 at X1-DF55-C. 
     GHOST-01 has 50/100 cargo with iron ore, GHOST-02 is empty, and GHOST-03 has fuel..."
Claude: "Based on the information you provided..."
```

**With MCP Resources:**
```
You: "Check my fleet status"
Claude: [Automatically accesses spacetraders://ships/list]
Claude: "I can see you have 3 ships: GHOST-01 at X1-DF55-A with 50/100 cargo containing iron ore..."
```

### Automatic Data Fetching

When Claude encounters a `spacetraders://` URI, it automatically:
1. Requests the data from the MCP server
2. Receives the current, live data from your SpaceTraders account
3. Uses that data to provide informed responses
4. Updates its understanding of your situation

## Available Resource URIs

### Agent Information
- **URI**: `spacetraders://agent/info`
- **Contains**: Credits, headquarters, starting faction, ship count
- **Use case**: Understanding your overall status

### Fleet Information
- **URI**: `spacetraders://ships/list`
- **Contains**: All ships, their locations, cargo, fuel, status
- **Use case**: Fleet management and ship coordination

### System Data
- **URI**: `spacetraders://systems/{systemSymbol}/waypoints`
- **Contains**: All waypoints in a system, their types, traits, facilities
- **Use case**: System exploration and facility location

### Shipyard Information
- **URI**: `spacetraders://systems/{systemSymbol}/waypoints/{waypointSymbol}/shipyard`
- **Contains**: Available ships, prices, specifications
- **Use case**: Ship purchasing decisions

### Market Information
- **URI**: `spacetraders://systems/{systemSymbol}/waypoints/{waypointSymbol}/market`
- **Contains**: Trade goods, prices, import/export data
- **Use case**: Trading strategy and profit analysis

## How to Use Resources Effectively

### 1. Let Claude Access Resources Automatically

**Good**: "What's my current situation?"
```
Claude automatically accesses:
- spacetraders://agent/info
- spacetraders://ships/list
And provides a comprehensive overview
```

**Less Effective**: "I have 150,000 credits and 3 ships, what should I do?"
```
Claude relies only on the limited information you provided
```

### 2. Reference Resources Explicitly When Needed

**Explicit Reference**: "Check spacetraders://systems/X1-DF55/waypoints to find shipyards"
```
Claude will fetch the system data and identify all waypoints with shipyard facilities
```

### 3. Combine Multiple Resources for Complex Analysis

**Example**: "Analyze trading opportunities in my current system"
```
Claude automatically accesses:
- spacetraders://ships/list (to see current ship locations)
- spacetraders://systems/X1-DF55/waypoints (to find markets)
- spacetraders://systems/X1-DF55/waypoints/X1-DF55-MARKET/market (for trade data)
```

## Resource Refresh and Caching

### Automatic Updates
- Resources are fetched fresh each time they're accessed
- No stale data concerns - you always get current information
- Changes made through tools are reflected immediately in subsequent resource access

### Performance Considerations
- Resources are fetched on-demand
- Multiple references to the same resource in one conversation may use cached data
- Complex queries that require multiple resources are optimized automatically

## Resource vs. Tool Distinction

### Resources (Read-Only)
- **Purpose**: Provide information about current game state
- **Examples**: Ship locations, cargo contents, available contracts
- **Access**: Automatic through Claude's understanding or explicit URI reference
- **Updates**: Always current when accessed

### Tools (Actions)
- **Purpose**: Perform actions that change game state
- **Examples**: Navigate ship, extract resources, accept contracts
- **Access**: Through natural language commands to Claude
- **Effects**: Modify your game state and may trigger resource updates

## Best Practices

### 1. Start Conversations with Context
```
"Run a status check" or "What's my current situation?"
```
This allows Claude to access your current resources and provide informed assistance.

### 2. Be Specific About Systems
```
"Explore system X1-DF55 for trading opportunities"
```
Rather than "find trading opportunities" - specificity helps Claude access the right resources.

### 3. Trust the Automation
```
Instead of: "I'm at X1-DF55-A with 50 cargo units of iron ore, where should I sell?"
Try: "Where should I sell my current cargo?"
```
Claude will automatically check your ship locations and cargo to provide accurate advice.

### 4. Use Resource URIs for Verification
```
"Double-check spacetraders://agent/info to confirm my credits"
```
Useful when you want to verify specific information or troubleshoot.

## Troubleshooting Resources

### Resource Not Found
- Verify the URI syntax is correct
- Ensure system/waypoint symbols exist and are spelled correctly
- Check that your agent has access to the requested data

### Outdated Information
- Resources are always current when fetched
- If information seems outdated, try explicitly referencing the resource URI
- Restart Claude Desktop if issues persist

### Missing Context
- If Claude seems unaware of your situation, try starting with "Run a status check"
- Explicitly reference key resources at the beginning of conversations
- Use the status_check prompt for comprehensive situation awareness

## Integration with Prompts and Tools

Resources work seamlessly with the server's prompts and tools:

### Smart Prompts Use Resources
- `status_check` automatically accesses agent and ship resources
- `explore_system` fetches system waypoint data
- `contract_strategy` reviews your contracts and fleet capabilities

### Tools Update Resource Data
- After using navigation tools, ship location resources reflect new positions
- Resource extraction updates cargo resources
- Contract acceptance updates contract status resources

This integration creates a seamless experience where Claude maintains awareness of your current game state and can provide intelligent, context-aware assistance.

## Advanced Resource Usage

### Parameterized Resources
Some resources accept parameters in their URIs:
```
spacetraders://systems/X1-DF55/waypoints          # All waypoints in X1-DF55
spacetraders://systems/X1-GX37/waypoints          # All waypoints in X1-GX37
```

### Resource Chaining
Claude can automatically chain resource access:
```
1. Access spacetraders://ships/list to find ship locations
2. Extract system symbols from ship locations
3. Access spacetraders://systems/{system}/waypoints for each system
4. Provide comprehensive multi-system analysis
```

### Conditional Resource Access
Claude intelligently decides which resources to access based on context:
```
"Should I buy a new ship?" triggers access to:
- spacetraders://agent/info (for credits)
- spacetraders://ships/list (for current fleet)
- spacetraders://systems/{current-system}/waypoints (to find shipyards)
```

This intelligent resource management ensures you get comprehensive, accurate assistance without having to manually specify every piece of information needed.