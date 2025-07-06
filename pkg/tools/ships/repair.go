package ships

import (
	"context"
	"fmt"
	"strings"

	"spacetraders-mcp/pkg/client"
	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/tools/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

// RepairShipTool allows repairing a ship at a shipyard
type RepairShipTool struct {
	client *client.Client
	logger *logging.Logger
}

// NewRepairShipTool creates a new repair ship tool
func NewRepairShipTool(client *client.Client, logger *logging.Logger) *RepairShipTool {
	return &RepairShipTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *RepairShipTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "repair_ship",
		Description: "Repair a ship at a shipyard. Ship must be docked at a waypoint with a shipyard.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"ship_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the ship to repair (e.g., 'MYSHIP-1')",
				},
			},
			Required: []string{"ship_symbol"},
		},
	}
}

// Handler returns the tool handler function
func (t *RepairShipTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contextLogger := t.logger.WithContext(ctx, "repair-ship-tool")

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

		contextLogger.Info(fmt.Sprintf("Repairing ship %s", shipSymbol))

		// Perform the repair
		resp, err := t.client.RepairShip(shipSymbol)
		if err != nil {
			contextLogger.Error(fmt.Sprintf("Failed to repair ship %s: %v", shipSymbol, err))
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("Failed to repair ship %s: %v", shipSymbol, err)),
				},
				IsError: true,
			}, nil
		}

		contextLogger.ToolCall("repair_ship", true)
		contextLogger.Info(fmt.Sprintf("Successfully repaired ship %s for %d credits", shipSymbol, resp.Data.Transaction.TotalPrice))

		// Create structured response
		result := map[string]interface{}{
			"ship_symbol": shipSymbol,
			"repair_cost": resp.Data.Transaction.TotalPrice,
			"agent": map[string]interface{}{
				"symbol":  resp.Data.Agent.Symbol,
				"credits": resp.Data.Agent.Credits,
			},
			"ship_condition": map[string]interface{}{
				"frame_integrity":   resp.Data.Ship.Frame.Integrity,
				"reactor_integrity": resp.Data.Ship.Reactor.Integrity,
				"engine_integrity":  resp.Data.Ship.Engine.Integrity,
			},
			"transaction": map[string]interface{}{
				"waypoint_symbol": resp.Data.Transaction.WaypointSymbol,
				"ship_symbol":     resp.Data.Transaction.ShipSymbol,
				"price":           resp.Data.Transaction.TotalPrice,
				"timestamp":       resp.Data.Transaction.Timestamp,
			},
		}

		// Add module and mount integrity information
		if len(resp.Data.Ship.Modules) > 0 {
			modules := []map[string]interface{}{}
			for _, module := range resp.Data.Ship.Modules {
				modules = append(modules, map[string]interface{}{
					"symbol":    module.Symbol,
					"condition": "OPERATIONAL", // Assume operational after repair
				})
			}
			result["modules"] = modules
		}

		if len(resp.Data.Ship.Mounts) > 0 {
			mounts := []map[string]interface{}{}
			for _, mount := range resp.Data.Ship.Mounts {
				mounts = append(mounts, map[string]interface{}{
					"symbol":    mount.Symbol,
					"condition": "OPERATIONAL", // Assume operational after repair
				})
			}
			result["mounts"] = mounts
		}

		// Create text summary
		textSummary := fmt.Sprintf("## 🔧 Ship Repair Complete for %s\n\n", shipSymbol)
		textSummary += fmt.Sprintf("✅ **Ship successfully repaired** at %s\n\n", resp.Data.Transaction.WaypointSymbol)

		// Financial summary
		textSummary += "## 💰 Financial Summary\n\n"
		textSummary += fmt.Sprintf("**Repair Cost:** %d credits\n", resp.Data.Transaction.TotalPrice)
		textSummary += fmt.Sprintf("**Remaining Credits:** %d credits\n", resp.Data.Agent.Credits)
		textSummary += fmt.Sprintf("**Agent:** %s\n\n", resp.Data.Agent.Symbol)

		// Ship condition summary
		textSummary += "## 🛠️ Ship Condition After Repair\n\n"
		textSummary += fmt.Sprintf("- **Frame Integrity:** %d%%\n", resp.Data.Ship.Frame.Integrity)
		textSummary += fmt.Sprintf("- **Reactor Integrity:** %d%%\n", resp.Data.Ship.Reactor.Integrity)
		textSummary += fmt.Sprintf("- **Engine Integrity:** %d%%\n", resp.Data.Ship.Engine.Integrity)

		// Determine overall condition
		minIntegrity := resp.Data.Ship.Frame.Integrity
		if resp.Data.Ship.Reactor.Integrity < minIntegrity {
			minIntegrity = resp.Data.Ship.Reactor.Integrity
		}
		if resp.Data.Ship.Engine.Integrity < minIntegrity {
			minIntegrity = resp.Data.Ship.Engine.Integrity
		}

		conditionIcon := "🟢"
		conditionText := "Excellent"
		if minIntegrity < 100 {
			conditionIcon = "🟡"
			conditionText = "Good"
		}
		if minIntegrity < 75 {
			conditionIcon = "🟠"
			conditionText = "Fair"
		}
		if minIntegrity < 50 {
			conditionIcon = "🔴"
			conditionText = "Poor"
		}

		textSummary += fmt.Sprintf("**Overall Condition:** %s %s (%d%% minimum)\n\n", conditionIcon, conditionText, minIntegrity)

		// Add module and mount status
		if len(resp.Data.Ship.Modules) > 0 {
			textSummary += "**Modules:**\n"
			for _, module := range resp.Data.Ship.Modules {
				textSummary += fmt.Sprintf("✅ %s - Operational\n", module.Symbol)
			}
			textSummary += "\n"
		}

		if len(resp.Data.Ship.Mounts) > 0 {
			textSummary += "**Mounts:**\n"
			for _, mount := range resp.Data.Ship.Mounts {
				textSummary += fmt.Sprintf("✅ %s - Operational\n", mount.Symbol)
			}
			textSummary += "\n"
		}

		// Transaction details
		textSummary += "## 📋 Transaction Details\n\n"
		textSummary += fmt.Sprintf("**Location:** %s\n", resp.Data.Transaction.WaypointSymbol)
		textSummary += fmt.Sprintf("**Timestamp:** %s\n", resp.Data.Transaction.Timestamp)
		textSummary += fmt.Sprintf("**Agent:** %s\n\n", resp.Data.Agent.Symbol)

		// Next steps
		textSummary += "## 🚀 Next Steps\n\n"
		textSummary += "Your ship is now fully repaired and ready for action:\n"
		textSummary += "- **Navigate** to new destinations without performance penalties\n"
		textSummary += "- **Extract resources** at maximum efficiency\n"
		textSummary += "- **Engage in combat** with full capabilities\n"
		textSummary += "- **Haul cargo** without structural concerns\n\n"

		textSummary += "**Maintenance Tips:**\n"
		textSummary += "- Monitor ship integrity regularly\n"
		textSummary += "- Repair before integrity drops below 50%\n"
		textSummary += "- Consider upgrading components for better durability\n"

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%s\n```", utils.FormatJSON(result))),
			},
		}, nil
	}
}
