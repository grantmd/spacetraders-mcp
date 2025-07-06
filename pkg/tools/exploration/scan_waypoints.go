package exploration

import (
	"context"
	"fmt"
	"strings"

	"spacetraders-mcp/pkg/client"
	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/tools/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

// ScanWaypointsTool allows scanning for waypoints around a ship
type ScanWaypointsTool struct {
	client *client.Client
	logger *logging.Logger
}

// NewScanWaypointsTool creates a new scan waypoints tool
func NewScanWaypointsTool(client *client.Client, logger *logging.Logger) *ScanWaypointsTool {
	return &ScanWaypointsTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *ScanWaypointsTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "scan_waypoints",
		Description: "Scan for waypoints around a ship using its sensors. Reveals hidden waypoints and asteroid fields.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"ship_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the ship to scan with (e.g., 'MYSHIP-1')",
				},
			},
			Required: []string{"ship_symbol"},
		},
	}
}

// Handler returns the tool handler function
func (t *ScanWaypointsTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contextLogger := t.logger.WithContext(ctx, "scan-waypoints-tool")

		// Extract parameters
		var shipSymbol string
		if request.Params.Arguments != nil {
			if argsMap, ok := request.Params.Arguments.(map[string]interface{}); ok {
				if val, exists := argsMap["ship_symbol"]; exists {
					if s, ok := val.(string); ok {
						shipSymbol = strings.ToUpper(s)
					}
				}
			}
		}

		if shipSymbol == "" {
			contextLogger.Error("Missing ship_symbol parameter")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("Error: ship_symbol parameter is required"),
				},
				IsError: true,
			}, nil
		}

		contextLogger.Info(fmt.Sprintf("Scanning for waypoints using ship %s", shipSymbol))

		// Perform the scan
		scanData, err := t.client.ScanWaypoints(shipSymbol)
		if err != nil {
			contextLogger.Error(fmt.Sprintf("Failed to scan waypoints with ship %s: %v", shipSymbol, err))
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("Failed to scan waypoints with ship %s: %v", shipSymbol, err)),
				},
				IsError: true,
			}, nil
		}

		contextLogger.ToolCall("scan_waypoints", true)
		contextLogger.Info(fmt.Sprintf("Successfully scanned %d waypoints with ship %s", len(scanData.Data.Waypoints), shipSymbol))

		// Create structured response
		result := map[string]interface{}{
			"ship_symbol":     shipSymbol,
			"waypoints_found": len(scanData.Data.Waypoints),
			"cooldown":        scanData.Data.Cooldown,
			"waypoints":       []map[string]interface{}{},
		}

		// Build waypoints data
		for _, waypoint := range scanData.Data.Waypoints {
			waypointData := map[string]interface{}{
				"symbol": waypoint.Symbol,
				"type":   waypoint.Type,
				"x":      waypoint.X,
				"y":      waypoint.Y,
				"traits": []map[string]interface{}{},
			}

			// Add traits
			for _, trait := range waypoint.Traits {
				waypointData["traits"] = append(waypointData["traits"].([]map[string]interface{}), map[string]interface{}{
					"symbol":      trait.Symbol,
					"name":        trait.Name,
					"description": trait.Description,
				})
			}

			// ScannedWaypoint doesn't include modifiers field

			// Add orbital information if available
			if len(waypoint.Orbitals) > 0 {
				orbitals := []string{}
				for _, orbital := range waypoint.Orbitals {
					orbitals = append(orbitals, orbital.Symbol)
				}
				waypointData["orbitals"] = orbitals
			}

			// Add faction information if available
			if waypoint.Faction.Symbol != "" {
				waypointData["faction"] = waypoint.Faction.Symbol
			}

			result["waypoints"] = append(result["waypoints"].([]map[string]interface{}), waypointData)
		}

		// Create text summary
		textSummary := fmt.Sprintf("## 🔍 Waypoint Scan Results for %s\n\n", shipSymbol)

		if len(scanData.Data.Waypoints) == 0 {
			textSummary += "❌ **No waypoints detected** in scanning range.\n\n"
			textSummary += "This could mean:\n"
			textSummary += "- Your ship doesn't have appropriate scanning equipment\n"
			textSummary += "- No undiscovered waypoints are within scanning range\n"
			textSummary += "- The ship is on cooldown from previous scans\n\n"
		} else {
			textSummary += fmt.Sprintf("✅ **Detected %d waypoint(s)** in scanning range:\n\n", len(scanData.Data.Waypoints))

			for i, waypoint := range scanData.Data.Waypoints {
				textSummary += fmt.Sprintf("### %d. %s (%s)\n", i+1, waypoint.Symbol, waypoint.Type)
				textSummary += fmt.Sprintf("**Location:** (%d, %d)\n", waypoint.X, waypoint.Y)

				if len(waypoint.Traits) > 0 {
					textSummary += "**Traits:**\n"
					for _, trait := range waypoint.Traits {
						// Add icons for common traits
						icon := "•"
						switch trait.Symbol {
						case "ASTEROID_FIELD":
							icon = "⛏️"
						case "MARKETPLACE":
							icon = "🏪"
						case "SHIPYARD":
							icon = "🏭"
						case "JUMP_GATE":
							icon = "🌀"
						case "FUEL_STATION":
							icon = "⛽"
						}
						textSummary += fmt.Sprintf("%s %s - %s\n", icon, trait.Name, trait.Description)
					}
				}

				// ScannedWaypoint doesn't include modifiers field

				if len(waypoint.Orbitals) > 0 {
					textSummary += "**Orbitals:** "
					orbitalNames := []string{}
					for _, orbital := range waypoint.Orbitals {
						orbitalNames = append(orbitalNames, orbital.Symbol)
					}
					textSummary += strings.Join(orbitalNames, ", ") + "\n"
				}

				if waypoint.Faction != nil && waypoint.Faction.Symbol != "" {
					textSummary += fmt.Sprintf("**Faction:** %s\n", waypoint.Faction.Symbol)
				}

				textSummary += "\n"
			}
		}

		// Add cooldown information
		if scanData.Data.Cooldown.TotalSeconds > 0 {
			textSummary += "## ⏳ Cooldown Information\n\n"
			textSummary += fmt.Sprintf("**Total Cooldown:** %d seconds\n", scanData.Data.Cooldown.TotalSeconds)
			textSummary += fmt.Sprintf("**Remaining:** %d seconds\n", scanData.Data.Cooldown.RemainingSeconds)
			if scanData.Data.Cooldown.Expiration != "" {
				textSummary += fmt.Sprintf("**Expires:** %s\n", scanData.Data.Cooldown.Expiration)
			}
		}

		// Add next steps
		textSummary += "## 🚀 Next Steps\n\n"
		if len(scanData.Data.Waypoints) > 0 {
			textSummary += "**Explore discovered waypoints:**\n"
			for _, waypoint := range scanData.Data.Waypoints {
				textSummary += fmt.Sprintf("- Navigate to %s: `navigate_ship` tool\n", waypoint.Symbol)

				// Add specific recommendations based on traits
				for _, trait := range waypoint.Traits {
					switch trait.Symbol {
					case "ASTEROID_FIELD":
						textSummary += fmt.Sprintf("  - Mine resources at %s: `extract_resources` tool\n", waypoint.Symbol)
					case "MARKETPLACE":
						textSummary += fmt.Sprintf("  - Check market prices at %s: Use waypoint info tools\n", waypoint.Symbol)
					case "SHIPYARD":
						textSummary += fmt.Sprintf("  - View available ships at %s: `purchase_ship` tool\n", waypoint.Symbol)
					case "FUEL_STATION":
						textSummary += fmt.Sprintf("  - Refuel at %s: `refuel_ship` tool\n", waypoint.Symbol)
					}
				}
			}
		} else {
			textSummary += "- Try scanning again after cooldown expires\n"
			textSummary += "- Move to a different location and scan again\n"
			textSummary += "- Ensure your ship has appropriate scanning equipment\n"
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%s\n```", utils.FormatJSON(result))),
			},
		}, nil
	}
}
