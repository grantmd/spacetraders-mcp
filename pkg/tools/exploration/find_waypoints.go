package exploration

import (
	"context"
	"fmt"
	"strings"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"
	"spacetraders-mcp/pkg/tools/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

// FindWaypointsTool helps find waypoints by traits and facilities
type FindWaypointsTool struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewFindWaypointsTool creates a new waypoint search tool
func NewFindWaypointsTool(client *spacetraders.Client, logger *logging.Logger) *FindWaypointsTool {
	return &FindWaypointsTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *FindWaypointsTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "find_waypoints",
		Description: "Find waypoints in a system by specific traits or facilities (SHIPYARD, MARKETPLACE, etc.)",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"system_symbol": map[string]interface{}{
					"type":        "string",
					"description": "System symbol to search in (e.g., 'X1-FM66')",
				},
				"trait": map[string]interface{}{
					"type":        "string",
					"description": "Trait to search for (e.g., 'SHIPYARD', 'MARKETPLACE', 'ASTEROID_FIELD', 'JUMP_GATE')",
				},
				"waypoint_type": map[string]interface{}{
					"type":        "string",
					"description": "Optional: Filter by waypoint type (e.g., 'PLANET', 'MOON', 'ASTEROID')",
				},
			},
			Required: []string{"system_symbol", "trait"},
		},
	}
}

// Handler returns the tool handler function
func (t *FindWaypointsTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contextLogger := t.logger.WithContext(ctx, "find-waypoints-tool")

		// Extract parameters
		var systemSymbol, trait, waypointType string
		if request.Params.Arguments != nil {
			if argsMap, ok := request.Params.Arguments.(map[string]interface{}); ok {
				if val, exists := argsMap["system_symbol"]; exists {
					if s, ok := val.(string); ok {
						systemSymbol = strings.ToUpper(s)
					}
				}
				if val, exists := argsMap["trait"]; exists {
					if s, ok := val.(string); ok {
						trait = strings.ToUpper(s)
					}
				}
				if val, exists := argsMap["waypoint_type"]; exists {
					if s, ok := val.(string); ok {
						waypointType = strings.ToUpper(s)
					}
				}
			}
		}

		if systemSymbol == "" {
			contextLogger.Error("Missing system_symbol parameter")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("Error: system_symbol parameter is required"),
				},
				IsError: true,
			}, nil
		}

		if trait == "" {
			contextLogger.Error("Missing trait parameter")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("Error: trait parameter is required"),
				},
				IsError: true,
			}, nil
		}

		contextLogger.Info(fmt.Sprintf("Searching for waypoints with trait '%s' in system %s", trait, systemSymbol))

		// Get waypoints from the system
		waypoints, err := t.client.GetAllSystemWaypoints(systemSymbol)
		if err != nil {
			contextLogger.Error(fmt.Sprintf("Failed to get waypoints for system %s: %v", systemSymbol, err))
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("Failed to retrieve waypoints for system %s: %v", systemSymbol, err)),
				},
				IsError: true,
			}, nil
		}

		// Filter waypoints by trait and optionally by type
		var matchingWaypoints []spacetraders.SystemWaypoint
		for _, waypoint := range waypoints {
			// Check waypoint type filter
			if waypointType != "" && waypoint.Type != waypointType {
				continue
			}

			// Check if waypoint has the requested trait
			hasTrait := false
			for _, waypointTrait := range waypoint.Traits {
				if waypointTrait.Symbol == trait {
					hasTrait = true
					break
				}
			}

			if hasTrait {
				matchingWaypoints = append(matchingWaypoints, waypoint)
			}
		}

		contextLogger.ToolCall("find_waypoints", true)
		contextLogger.Info(fmt.Sprintf("Found %d waypoints with trait '%s' in system %s", len(matchingWaypoints), trait, systemSymbol))

		// Create structured response
		result := map[string]interface{}{
			"system_symbol":        systemSymbol,
			"searched_trait":       trait,
			"waypoint_type_filter": waypointType,
			"total_found":          len(matchingWaypoints),
			"waypoints":            []map[string]interface{}{},
		}

		// Build waypoints data
		for _, waypoint := range matchingWaypoints {
			waypointData := map[string]interface{}{
				"symbol": waypoint.Symbol,
				"type":   waypoint.Type,
				"x":      waypoint.X,
				"y":      waypoint.Y,
				"traits": []map[string]interface{}{},
			}

			// Add all traits for context
			for _, t := range waypoint.Traits {
				waypointData["traits"] = append(waypointData["traits"].([]map[string]interface{}), map[string]interface{}{
					"symbol":      t.Symbol,
					"name":        t.Name,
					"description": t.Description,
				})
			}

			// Add orbital information if available
			if len(waypoint.Orbitals) > 0 {
				orbitals := []string{}
				for _, orbital := range waypoint.Orbitals {
					orbitals = append(orbitals, orbital.Symbol)
				}
				waypointData["orbitals"] = orbitals
			}

			result["waypoints"] = append(result["waypoints"].([]map[string]interface{}), waypointData)
		}

		// Create text summary
		textSummary := fmt.Sprintf("## Waypoints with %s in %s\n\n", trait, systemSymbol)

		if len(matchingWaypoints) == 0 {
			textSummary += fmt.Sprintf("❌ **No waypoints found** with trait '%s'", trait)
			if waypointType != "" {
				textSummary += fmt.Sprintf(" and type '%s'", waypointType)
			}
			textSummary += fmt.Sprintf(" in system %s.\n\n", systemSymbol)
			textSummary += "**Common traits to search for:**\n"
			textSummary += "- `SHIPYARD` - Build and buy ships\n"
			textSummary += "- `MARKETPLACE` - Trade goods\n"
			textSummary += "- `ASTEROID_FIELD` - Mine resources\n"
			textSummary += "- `JUMP_GATE` - Travel to other systems\n"
			textSummary += "- `FUEL_STATION` - Refuel ships\n"
		} else {
			textSummary += fmt.Sprintf("✅ **Found %d waypoint(s)** with trait '%s'", len(matchingWaypoints), trait)
			if waypointType != "" {
				textSummary += fmt.Sprintf(" and type '%s'", waypointType)
			}
			textSummary += ":\n\n"

			for i, waypoint := range matchingWaypoints {
				textSummary += fmt.Sprintf("### %d. %s (%s)\n", i+1, waypoint.Symbol, waypoint.Type)
				textSummary += fmt.Sprintf("**Location:** (%d, %d)\n", waypoint.X, waypoint.Y)

				if len(waypoint.Traits) > 0 {
					textSummary += "**Traits:**\n"
					for _, t := range waypoint.Traits {
						icon := "•"
						if t.Symbol == trait {
							icon = "🎯"
						}
						textSummary += fmt.Sprintf("%s %s - %s\n", icon, t.Name, t.Description)
					}
				}

				if len(waypoint.Orbitals) > 0 {
					textSummary += "**Orbitals:** "
					orbitalNames := []string{}
					for _, orbital := range waypoint.Orbitals {
						orbitalNames = append(orbitalNames, orbital.Symbol)
					}
					textSummary += strings.Join(orbitalNames, ", ") + "\n"
				}

				textSummary += "\n"
			}

			// Add next steps
			textSummary += "## 🚀 Next Steps\n\n"
			switch trait {
			case "SHIPYARD":
				textSummary += "To see available ships at a shipyard, use:\n"
				for _, waypoint := range matchingWaypoints {
					textSummary += fmt.Sprintf("- Check ships at %s: `spacetraders://systems/%s/waypoints/%s/shipyard`\n", waypoint.Symbol, systemSymbol, waypoint.Symbol)
				}
			case "MARKETPLACE":
				textSummary += "To see market prices and trade opportunities:\n"
				for _, waypoint := range matchingWaypoints {
					textSummary += fmt.Sprintf("- Check market at %s: `spacetraders://systems/%s/waypoints/%s/market`\n", waypoint.Symbol, systemSymbol, waypoint.Symbol)
				}
			}
			textSummary += "\nTo navigate to a waypoint, use: `navigate_ship` tool with your ship symbol and chosen waypoint.\n"
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%s\n```", utils.FormatJSON(result))),
			},
		}, nil
	}
}
