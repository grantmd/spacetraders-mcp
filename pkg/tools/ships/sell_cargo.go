package ships

import (
	"context"
	"fmt"
	"strings"
	"time"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"
	"spacetraders-mcp/pkg/tools/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

// SellCargoTool handles selling cargo from ships at markets
type SellCargoTool struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewSellCargoTool creates a new sell cargo tool
func NewSellCargoTool(client *spacetraders.Client, logger *logging.Logger) *SellCargoTool {
	return &SellCargoTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *SellCargoTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "sell_cargo",
		Description: "Sell cargo from a ship at a marketplace. Ship must be docked at a waypoint with a marketplace that accepts the cargo type.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"ship_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the ship to sell cargo from (e.g., 'SHIP_1234')",
				},
				"cargo_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the cargo item to sell (e.g., 'IRON_ORE', 'ALUMINUM_ORE', 'FUEL')",
				},
				"units": map[string]interface{}{
					"type":        "integer",
					"description": "Number of units to sell",
					"minimum":     1,
				},
			},
			Required: []string{"ship_symbol", "cargo_symbol", "units"},
		},
	}
}

// Handler returns the tool handler function
func (t *SellCargoTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Set up context logger
		ctxLogger := t.logger.WithContext(ctx, "sell-cargo-tool")
		ctxLogger.Debug("Processing cargo sell request")

		// Parse arguments
		shipSymbol := ""
		cargoSymbol := ""
		units := 0

		if request.Params.Arguments == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("❌ Missing required arguments: ship_symbol, cargo_symbol, and units"),
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
			if cs, exists := argsMap["cargo_symbol"]; exists {
				if csStr, ok := cs.(string); ok {
					cargoSymbol = strings.TrimSpace(strings.ToUpper(csStr))
				}
			}
			if u, exists := argsMap["units"]; exists {
				if uFloat, ok := u.(float64); ok {
					units = int(uFloat)
				} else if uInt, ok := u.(int); ok {
					units = uInt
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

		if cargoSymbol == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("❌ cargo_symbol is required and must be a non-empty string"),
				},
				IsError: true,
			}, nil
		}

		if units <= 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("❌ units must be a positive integer"),
				},
				IsError: true,
			}, nil
		}

		ctxLogger.Info("Attempting to sell %d units of %s from ship %s", units, cargoSymbol, shipSymbol)

		// Sell the cargo
		start := time.Now()
		agent, cargo, transaction, err := t.client.SellCargo(shipSymbol, cargoSymbol, units)
		duration := time.Since(start)

		if err != nil {
			ctxLogger.Error("Failed to sell cargo: %v", err)
			ctxLogger.APICall(fmt.Sprintf("/my/ships/%s/sell", shipSymbol), 0, duration.String())
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("❌ Failed to sell cargo: %s", err.Error())),
				},
				IsError: true,
			}, nil
		}

		ctxLogger.APICall(fmt.Sprintf("/my/ships/%s/sell", shipSymbol), 201, duration.String())
		ctxLogger.Info("Successfully sold %d units of %s from ship %s for %d credits", units, cargoSymbol, shipSymbol, transaction.TotalPrice)

		// Format the response
		result := map[string]interface{}{
			"success":      true,
			"message":      fmt.Sprintf("Successfully sold %d units of %s from ship %s", units, cargoSymbol, shipSymbol),
			"ship_symbol":  shipSymbol,
			"cargo_symbol": cargoSymbol,
			"units_sold":   units,
			"transaction": map[string]interface{}{
				"waypoint_symbol": transaction.WaypointSymbol,
				"ship_symbol":     transaction.ShipSymbol,
				"trade_symbol":    transaction.TradeSymbol,
				"type":            transaction.Type,
				"units":           transaction.Units,
				"price_per_unit":  transaction.PricePerUnit,
				"total_price":     transaction.TotalPrice,
				"timestamp":       transaction.Timestamp,
			},
			"cargo": map[string]interface{}{
				"capacity": cargo.Capacity,
				"units":    cargo.Units,
				"inventory": func() []map[string]interface{} {
					inventory := make([]map[string]interface{}, len(cargo.Inventory))
					for i, item := range cargo.Inventory {
						inventory[i] = map[string]interface{}{
							"symbol":      item.Symbol,
							"name":        item.Name,
							"description": item.Description,
							"units":       item.Units,
						}
					}
					return inventory
				}(),
			},
			"agent": map[string]interface{}{
				"credits": agent.Credits,
			},
		}

		jsonData := utils.FormatJSON(result)

		// Calculate cargo utilization and profit
		cargoPercent := float64(cargo.Units) / float64(cargo.Capacity) * 100
		freedSpace := cargo.Capacity - cargo.Units
		profitPerUnit := transaction.PricePerUnit

		// Find the sold item name
		soldItemName := cargoSymbol
		for _, item := range cargo.Inventory {
			if item.Symbol == cargoSymbol {
				soldItemName = item.Name
				break
			}
		}

		// Create formatted text summary
		textSummary := "💰 **Cargo Sale Successful!**\n\n"
		textSummary += fmt.Sprintf("**Ship:** %s\n", shipSymbol)
		textSummary += fmt.Sprintf("**Sold:** %d units of %s\n", units, soldItemName)
		textSummary += fmt.Sprintf("**Price per Unit:** %d credits\n", profitPerUnit)
		textSummary += fmt.Sprintf("**Total Revenue:** %d credits\n", transaction.TotalPrice)
		textSummary += fmt.Sprintf("**Current Credits:** %d\n", agent.Credits)
		textSummary += fmt.Sprintf("**Location:** %s\n\n", transaction.WaypointSymbol)

		// Cargo status
		textSummary += fmt.Sprintf("**Cargo Status:** %d/%d units (%.1f%% full)\n", cargo.Units, cargo.Capacity, cargoPercent)
		textSummary += fmt.Sprintf("**Available Space:** %d units\n\n", freedSpace)

		// Show current cargo inventory
		if len(cargo.Inventory) > 0 {
			textSummary += "**Remaining Cargo Inventory:**\n"
			for _, item := range cargo.Inventory {
				textSummary += fmt.Sprintf("- %s: %d units\n", item.Name, item.Units)
			}
		} else {
			textSummary += "**Cargo Hold:** Empty - ready for new cargo!\n"
		}

		// Add helpful tips based on cargo status and profit
		textSummary += "\n💡 **Next Steps:**\n"
		if profitPerUnit >= 100 {
			textSummary += "• 🎉 **Excellent profit!** - Great trading choice\n"
		} else if profitPerUnit >= 50 {
			textSummary += "• ✅ **Good profit** - Solid trading performance\n"
		} else {
			textSummary += "• 💭 **Consider** higher-value trade routes for better margins\n"
		}

		if freedSpace >= cargo.Capacity/2 {
			textSummary += "• 📦 **Plenty of space** - ready for more cargo\n"
			textSummary += "• ⛏️ Use `extract_resources` to mine valuable materials\n"
			textSummary += "• 🛒 Use `buy_cargo` to purchase goods for resale\n"
		} else if cargo.Units > 0 {
			textSummary += "• 💼 Consider selling more cargo to free up space\n"
		}

		textSummary += "• 📊 Use `get_status_summary` to check your fleet status\n"
		textSummary += "• 🗺️ Use `find_waypoints` to find more markets\n"

		// Add trading tips
		if transaction.TotalPrice >= 1000 {
			textSummary += "\n🚀 **Pro Trading Tip:** High-value sales like this indicate profitable trade routes!\n"
		}

		ctxLogger.ToolCall("sell_cargo", true)
		ctxLogger.Debug("Sell cargo response size: %d bytes", len(jsonData))

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%s\n```", jsonData)),
			},
		}, nil
	}
}
