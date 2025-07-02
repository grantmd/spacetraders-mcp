# Available Prompts

This document describes the predefined prompts available in the SpaceTraders MCP Server. These prompts provide intelligent, context-aware assistance for common SpaceTraders tasks.

## How to Use Prompts

Prompts are accessed through Claude Desktop's MCP integration. Simply reference the prompt name in your conversation, and Claude will use the predefined logic to help you with specific tasks.

## Available Prompts

### `status_check`

**Purpose:** Provides a comprehensive overview of your current SpaceTraders situation.

**What it does:**
- Analyzes your agent's current status
- Reviews your fleet composition and locations
- Checks active contracts and their progress
- Identifies immediate opportunities or issues
- Suggests next steps based on your current situation

**When to use:**
- Starting a new session
- When you're unsure what to do next
- After completing major tasks
- For a quick situation assessment

### `explore_system`

**Purpose:** Intelligently explores and analyzes a star system.

**What it does:**
- Maps all waypoints in the specified system
- Identifies key facilities (shipyards, markets, jump gates)
- Analyzes trading opportunities
- Suggests strategic locations for operations
- Provides a comprehensive system overview

**When to use:**
- Entering a new system
- Looking for specific facilities
- Planning trading routes
- Scouting for expansion opportunities

**Usage:** Mention the system symbol you want to explore.

### `contract_strategy`

**Purpose:** Analyzes your contracts and provides strategic recommendations.

**What it does:**
- Reviews all available and active contracts
- Analyzes contract requirements vs. your capabilities
- Calculates potential profits and risks
- Suggests optimal contract combinations
- Provides step-by-step completion strategies

**When to use:**
- Deciding which contracts to accept
- Planning contract completion routes
- Optimizing profit from multiple contracts
- When stuck on contract requirements

### `fleet_optimization`

**Purpose:** Analyzes your fleet composition and suggests improvements.

**What it does:**
- Reviews your current ships and their configurations
- Identifies gaps in your fleet capabilities
- Suggests ship purchases or modifications
- Optimizes ship roles and assignments
- Plans fleet expansion strategies

**When to use:**
- Planning fleet expansion
- Optimizing existing ships
- Deciding on new ship purchases
- Balancing fleet capabilities

## Smart Workflow for Contract Management

The prompts work together to create an intelligent workflow:

1. **Start with `status_check`** - Get your bearings
2. **Use `contract_strategy`** - Plan your contract approach
3. **Apply `explore_system`** - Scout target systems
4. **Implement `fleet_optimization`** - Ensure you have the right ships
5. **Return to `status_check`** - Monitor progress and adapt

## Tips for Using Prompts

- **Be specific:** When using system-related prompts, provide system symbols
- **Context matters:** The more information you provide, the better the recommendations
- **Combine prompts:** Use multiple prompts together for comprehensive planning
- **Regular check-ins:** Use `status_check` frequently to stay oriented
- **Ask follow-ups:** Prompts are starting points - ask for clarification or deeper analysis

## Example Usage

**Starting a session:**
```
"Run a status_check to see what my current situation is"
```

**Exploring a new system:**
```
"Use explore_system for X1-DF55 to find the best trading opportunities"
```

**Contract planning:**
```
"Apply contract_strategy to help me decide which contracts to focus on"
```

**Fleet planning:**
```
"Run fleet_optimization to see if I should buy any new ships"
```
