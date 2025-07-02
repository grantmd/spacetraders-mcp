package ships

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"

	"github.com/mark3labs/mcp-go/mcp"
)

// RefuelShipTool handles refueling ships at fuel stations
type RefuelShipTool struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewRefuelShipTool creates a new refuel ship tool
func NewRefuelShipTool(client *spacetraders.Client, logger *logging.Logger) *RefuelShipTool {
	return &RefuelShipTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *RefuelShipTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "refuel_ship",
		Description: "Refuel a ship at the current waypoint. The ship must be docked at a waypoint with a fuel station or market that sells fuel.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"ship_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the ship to refuel (e.g., 'SHIP_1234')",
				},
				"units": map[string]interface{}{
					"type":        "integer",
					"description": "Optional: Specific amount of fuel units to purchase. If not specified, refuels to full capacity.",
					"minimum":     1,
				},
				"from_cargo": map[string]interface{}{
					"type":        "boolean",
					"description": "Optional: Whether to refuel from cargo instead of purchasing from marketplace. Defaults to false.",
					"default":     false,
				},
			},
			Required: []string{"ship_symbol"},
		},
	}
}

// Handler returns the tool handler function
func (t *RefuelShipTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Set up context logger
		ctxLogger := t.logger.WithContext(ctx, "refuel-ship-tool")
		ctxLogger.Debug("Processing ship refuel request")

		// Parse arguments
		shipSymbol := ""
		units := 0
		fromCargo := false

		if request.Params.Arguments == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("❌ Missing required argument: ship_symbol"),
				},
				IsError: true,
			}, nil
		}

		if argsMap, ok := request.Params.Arguments.(map[string]interface{}); ok {
			if ss, exists := argsMap["ship_symbol"]; exists {
				if ssStr, ok := ss.(string); ok {
					shipSymbol = strings.TrimSpace(ssStr)
				}
			}
			if u, exists := argsMap["units"]; exists {
				if uFloat, ok := u.(float64); ok {
					units = int(uFloat)
				} else if uInt, ok := u.(int); ok {
					units = uInt
				}
			}
			if fc, exists := argsMap["from_cargo"]; exists {
				if fcBool, ok := fc.(bool); ok {
					fromCargo = fcBool
				}
			}
		}

		if shipSymbol == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("❌ ship_symbol is required and must be a non-empty string"),
				},
				IsError: true,
			}, nil
		}

		if units < 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("❌ units must be a positive integer if specified"),
				},
				IsError: true,
			}, nil
		}

		ctxLogger.Info("Attempting to refuel ship %s", shipSymbol)
		if units > 0 {
			ctxLogger.Info("Refueling %d units", units)
		} else {
			ctxLogger.Info("Refueling to full capacity")
		}
		if fromCargo {
			ctxLogger.Info("Refueling from cargo")
		}

		// Refuel the ship
		start := time.Now()
		agent, fuel, transaction, err := t.client.RefuelShip(shipSymbol, units, fromCargo)
		duration := time.Since(start)

		if err != nil {
			ctxLogger.Error("Failed to refuel ship: %v", err)
			ctxLogger.APICall(fmt.Sprintf("/my/ships/%s/refuel", shipSymbol), 0, duration.String())
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("❌ Failed to refuel ship: %s", err.Error())),
				},
				IsError: true,
			}, nil
		}

		ctxLogger.APICall(fmt.Sprintf("/my/ships/%s/refuel", shipSymbol), 200, duration.String())
		ctxLogger.Info("Successfully refueled ship %s", shipSymbol)

		// Format the response
		result := map[string]interface{}{
			"success":     true,
			"message":     fmt.Sprintf("Successfully refueled ship %s", shipSymbol),
			"ship_symbol": shipSymbol,
			"fuel": map[string]interface{}{
				"current":  fuel.Current,
				"capacity": fuel.Capacity,
			},
			"transaction": map[string]interface{}{
				"waypoint_symbol": transaction.WaypointSymbol,
				"price":           transaction.Price,
				"timestamp":       transaction.Timestamp,
			},
			"agent": map[string]interface{}{
				"credits": agent.Credits,
			},
		}

		// Add fuel consumption details if available
		if fuel.Consumed.Amount > 0 {
			result["fuel_consumed"] = map[string]interface{}{
				"amount":    fuel.Consumed.Amount,
				"timestamp": fuel.Consumed.Timestamp,
			}
		}

		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			ctxLogger.Error("Failed to marshal refuel result: %v", err)
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("✅ Ship refueled successfully, but failed to format response"),
				},
			}, nil
		}

		// Calculate fuel purchased and cost per unit
		fuelPurchased := fuel.Current - (fuel.Capacity - transaction.Price) // This is an approximation
		if units > 0 {
			fuelPurchased = units
		}

		costPerUnit := 0
		if fuelPurchased > 0 {
			costPerUnit = transaction.Price / fuelPurchased
		}

		// Create formatted text summary
		textSummary := "⛽ **Ship Refuel Successful!**\n\n"
		textSummary += fmt.Sprintf("**Ship:** %s\n", shipSymbol)
		textSummary += fmt.Sprintf("**Location:** %s\n", transaction.WaypointSymbol)
		textSummary += fmt.Sprintf("**Fuel Status:** %d/%d units", fuel.Current, fuel.Capacity)

		if fuel.Current == fuel.Capacity {
			textSummary += " (Full tank! ⛽)\n"
		} else {
			fuelPercentage := float64(fuel.Current) / float64(fuel.Capacity) * 100
			textSummary += fmt.Sprintf(" (%.1f%% full)\n", fuelPercentage)
		}

		textSummary += fmt.Sprintf("**Cost:** %d credits", transaction.Price)
		if fuelPurchased > 0 && costPerUnit > 0 {
			textSummary += fmt.Sprintf(" (%d credits per unit)", costPerUnit)
		}
		textSummary += "\n"

		textSummary += fmt.Sprintf("**Remaining Credits:** %d\n", agent.Credits)

		if fromCargo {
			textSummary += "\n**Source:** Refueled from ship's cargo inventory\n"
		} else {
			textSummary += "\n**Source:** Purchased fuel from local marketplace\n"
		}

		if fuel.Current < fuel.Capacity {
			remainingCapacity := fuel.Capacity - fuel.Current
			textSummary += fmt.Sprintf("\n💡 **Note:** Ship can still hold %d more fuel units if needed.\n", remainingCapacity)
		}

		textSummary += "\n🚀 **Ready for Travel:** Your ship is now fueled and ready for navigation!\n"

		ctxLogger.ToolCall("refuel_ship", true)
		ctxLogger.Debug("Refuel ship response size: %d bytes", len(jsonData))

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("**Raw JSON Data:**\n```json\n%s\n```", string(jsonData))),
			},
		}, nil
	}
}
