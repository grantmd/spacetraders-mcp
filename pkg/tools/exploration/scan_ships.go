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

// ScanShipsTool allows scanning for ships around a ship
type ScanShipsTool struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewScanShipsTool creates a new scan ships tool
func NewScanShipsTool(client *spacetraders.Client, logger *logging.Logger) *ScanShipsTool {
	return &ScanShipsTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *ScanShipsTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "scan_ships",
		Description: "Scan for ships around a ship using its sensors. Reveals nearby ships and their basic information.",
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
func (t *ScanShipsTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contextLogger := t.logger.WithContext(ctx, "scan-ships-tool")

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

		contextLogger.Info(fmt.Sprintf("Scanning for ships using ship %s", shipSymbol))

		// Perform the scan
		scanData, err := t.client.ScanShips(shipSymbol)
		if err != nil {
			contextLogger.Error(fmt.Sprintf("Failed to scan ships with ship %s: %v", shipSymbol, err))
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("Failed to scan ships with ship %s: %v", shipSymbol, err)),
				},
				IsError: true,
			}, nil
		}

		contextLogger.ToolCall("scan_ships", true)
		contextLogger.Info(fmt.Sprintf("Successfully scanned %d ships with ship %s", len(scanData.Ships), shipSymbol))

		// Create structured response
		result := map[string]interface{}{
			"ship_symbol": shipSymbol,
			"ships_found": len(scanData.Ships),
			"cooldown":    scanData.Cooldown,
			"ships":       []map[string]interface{}{},
		}

		// Build ships data
		for _, ship := range scanData.Ships {
			shipData := map[string]interface{}{
				"symbol":     ship.Symbol,
				"faction":    ship.Registration.FactionSymbol,
				"role":       ship.Registration.Role,
				"frame_type": ship.Frame.Symbol,
				"current_location": map[string]interface{}{
					"system":   ship.Nav.SystemSymbol,
					"waypoint": ship.Nav.WaypointSymbol,
					"status":   ship.Nav.Status,
				},
			}

			// Add engine information
			if ship.Engine.Symbol != "" {
				shipData["engine"] = map[string]interface{}{
					"symbol": ship.Engine.Symbol,
					"speed":  ship.Engine.Speed,
				}
			}

			// Add reactor information
			if ship.Reactor.Symbol != "" {
				shipData["reactor"] = map[string]interface{}{
					"symbol":       ship.Reactor.Symbol,
					"power_output": ship.Reactor.PowerOutput,
				}
			}

			// Add mounts information
			if len(ship.Mounts) > 0 {
				mounts := []map[string]interface{}{}
				for _, mount := range ship.Mounts {
					mounts = append(mounts, map[string]interface{}{
						"symbol":   mount.Symbol,
						"strength": mount.Strength,
					})
				}
				shipData["mounts"] = mounts
			}

			result["ships"] = append(result["ships"].([]map[string]interface{}), shipData)
		}

		// Create text summary
		textSummary := fmt.Sprintf("## 🔍 Ship Scan Results for %s\n\n", shipSymbol)

		if len(scanData.Ships) == 0 {
			textSummary += "❌ **No ships detected** in scanning range.\n\n"
			textSummary += "This could mean:\n"
			textSummary += "- Your ship doesn't have appropriate scanning equipment\n"
			textSummary += "- No other ships are within scanning range\n"
			textSummary += "- The ship is on cooldown from previous scans\n\n"
		} else {
			textSummary += fmt.Sprintf("✅ **Detected %d ship(s)** in scanning range:\n\n", len(scanData.Ships))

			for i, ship := range scanData.Ships {
				textSummary += fmt.Sprintf("### %d. %s\n", i+1, ship.Symbol)
				textSummary += fmt.Sprintf("**Faction:** %s\n", ship.Registration.FactionSymbol)
				textSummary += fmt.Sprintf("**Role:** %s\n", ship.Registration.Role)
				textSummary += fmt.Sprintf("**Frame:** %s\n", ship.Frame.Symbol)
				textSummary += fmt.Sprintf("**Location:** %s at %s (%s)\n", ship.Nav.SystemSymbol, ship.Nav.WaypointSymbol, ship.Nav.Status)

				if ship.Engine.Symbol != "" {
					textSummary += fmt.Sprintf("**Engine:** %s (Speed: %d)\n", ship.Engine.Symbol, ship.Engine.Speed)
				}

				if ship.Reactor.Symbol != "" {
					textSummary += fmt.Sprintf("**Reactor:** %s (Power: %d)\n", ship.Reactor.Symbol, ship.Reactor.PowerOutput)
				}

				if len(ship.Mounts) > 0 {
					textSummary += "**Mounts:**\n"
					for _, mount := range ship.Mounts {
						mountIcon := "🔧"
						switch {
						case strings.Contains(mount.Symbol, "MINING"):
							mountIcon = "⛏️"
						case strings.Contains(mount.Symbol, "LASER"):
							mountIcon = "🔫"
						case strings.Contains(mount.Symbol, "SURVEYOR"):
							mountIcon = "📡"
						}
						textSummary += fmt.Sprintf("%s %s (Strength: %d)\n", mountIcon, mount.Symbol, mount.Strength)
					}
				}

				// Add tactical assessment
				textSummary += "\n**Tactical Assessment:**\n"
				switch ship.Registration.Role {
				case "COMBAT":
					textSummary += "⚔️ **Combat vessel** - Potentially hostile\n"
				case "HAULER":
					textSummary += "📦 **Hauler** - Likely carrying cargo\n"
				case "EXCAVATOR":
					textSummary += "⛏️ **Mining vessel** - Focused on resource extraction\n"
				case "EXPLORER":
					textSummary += "🔍 **Explorer** - Likely scouting or surveying\n"
				}

				switch ship.Nav.Status {
				case "DOCKED":
					textSummary += "🏭 Currently docked - Not immediately mobile\n"
				case "IN_ORBIT":
					textSummary += "🌌 Currently in orbit - Ready to move\n"
				case "IN_TRANSIT":
					textSummary += "🚀 Currently in transit - Moving between waypoints\n"
				}

				textSummary += "\n"
			}
		}

		// Add cooldown information
		if scanData.Cooldown.TotalSeconds > 0 {
			textSummary += "## ⏳ Cooldown Information\n\n"
			textSummary += fmt.Sprintf("**Total Cooldown:** %d seconds\n", scanData.Cooldown.TotalSeconds)
			textSummary += fmt.Sprintf("**Remaining:** %d seconds\n", scanData.Cooldown.RemainingSeconds)
			if scanData.Cooldown.Expiration != "" {
				textSummary += fmt.Sprintf("**Expires:** %s\n", scanData.Cooldown.Expiration)
			}
			textSummary += "\n"
		}

		// Add next steps and tactical recommendations
		textSummary += "## 🚀 Next Steps\n\n"
		if len(scanData.Ships) > 0 {
			textSummary += "**Intelligence Gathering:**\n"
			textSummary += "- Monitor ship movements and patterns\n"
			textSummary += "- Identify potential trading partners or threats\n"
			textSummary += "- Track faction presence in the area\n\n"

			textSummary += "**Tactical Considerations:**\n"
			combatShips := 0
			for _, ship := range scanData.Ships {
				if ship.Registration.Role == "COMBAT" {
					combatShips++
				}
			}
			if combatShips > 0 {
				textSummary += fmt.Sprintf("- ⚠️ **%d combat vessel(s) detected** - Exercise caution\n", combatShips)
			}
			textSummary += "- Consider diplomatic approach if same faction\n"
			textSummary += "- Maintain safe distance from unknown vessels\n"
		} else {
			textSummary += "- Try scanning again after cooldown expires\n"
			textSummary += "- Move to a busier location (markets, jump gates)\n"
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
