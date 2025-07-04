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

// BuyCargoTool handles purchasing cargo for ships at markets
type BuyCargoTool struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewBuyCargoTool creates a new buy cargo tool
func NewBuyCargoTool(client *spacetraders.Client, logger *logging.Logger) *BuyCargoTool {
	return &BuyCargoTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *BuyCargoTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "buy_cargo",
		Description: "Purchase cargo for a ship at a marketplace. Ship must be docked at a waypoint with a marketplace that sells the cargo type and you must have sufficient credits and cargo space.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"ship_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the ship to buy cargo for (e.g., 'SHIP_1234')",
				},
				"cargo_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the cargo item to buy (e.g., 'FUEL', 'FOOD', 'MACHINERY')",
				},
				"units": map[string]interface{}{
					"type":        "integer",
					"description": "Number of units to buy",
					"minimum":     1,
				},
			},
			Required: []string{"ship_symbol", "cargo_symbol", "units"},
		},
	}
}

// Handler returns the tool handler function
func (t *BuyCargoTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Set up context logger
		ctxLogger := t.logger.WithContext(ctx, "buy-cargo-tool")
		ctxLogger.Debug("Processing cargo purchase request")

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

		ctxLogger.Info("Attempting to buy %d units of %s for ship %s", units, cargoSymbol, shipSymbol)

		// Buy the cargo
		start := time.Now()
		agent, cargo, transaction, err := t.client.BuyCargo(shipSymbol, cargoSymbol, units)
		duration := time.Since(start)

		if err != nil {
			ctxLogger.Error("Failed to buy cargo: %v", err)
			ctxLogger.APICall(fmt.Sprintf("/my/ships/%s/purchase", shipSymbol), 0, duration.String())
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("❌ Failed to buy cargo: %s", err.Error())),
				},
				IsError: true,
			}, nil
		}

		ctxLogger.APICall(fmt.Sprintf("/my/ships/%s/purchase", shipSymbol), 201, duration.String())
		ctxLogger.Info("Successfully bought %d units of %s for ship %s, spent %d credits", units, cargoSymbol, shipSymbol, transaction.TotalPrice)

		// Format the response
		result := map[string]interface{}{
			"success":      true,
			"message":      fmt.Sprintf("Successfully bought %d units of %s for ship %s", units, cargoSymbol, shipSymbol),
			"ship_symbol":  shipSymbol,
			"cargo_symbol": cargoSymbol,
			"units_bought": units,
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

		// Calculate cargo utilization and cost
		cargoPercent := float64(cargo.Units) / float64(cargo.Capacity) * 100
		remainingSpace := cargo.Capacity - cargo.Units
		costPerUnit := transaction.PricePerUnit

		// Find the bought item name
		boughtItemName := cargoSymbol
		for _, item := range cargo.Inventory {
			if item.Symbol == cargoSymbol {
				boughtItemName = item.Name
				break
			}
		}

		// Create formatted text summary
		textSummary := "🛒 **Cargo Purchase Successful!**\n\n"
		textSummary += fmt.Sprintf("**Ship:** %s\n", shipSymbol)
		textSummary += fmt.Sprintf("**Purchased:** %d units of %s\n", units, boughtItemName)
		textSummary += fmt.Sprintf("**Price per Unit:** %d credits\n", costPerUnit)
		textSummary += fmt.Sprintf("**Total Cost:** %d credits\n", transaction.TotalPrice)
		textSummary += fmt.Sprintf("**Remaining Credits:** %d\n", agent.Credits)
		textSummary += fmt.Sprintf("**Location:** %s\n\n", transaction.WaypointSymbol)

		// Cargo status
		textSummary += fmt.Sprintf("**Cargo Status:** %d/%d units (%.1f%% full)\n", cargo.Units, cargo.Capacity, cargoPercent)
		textSummary += fmt.Sprintf("**Remaining Space:** %d units\n\n", remainingSpace)

		// Show current cargo inventory
		if len(cargo.Inventory) > 0 {
			textSummary += "**Current Cargo Inventory:**\n"
			for _, item := range cargo.Inventory {
				textSummary += fmt.Sprintf("- %s: %d units\n", item.Name, item.Units)
			}
		}

		// Add helpful tips based on cargo status and cost
		textSummary += "\n💡 **Next Steps:**\n"
		if costPerUnit <= 50 {
			textSummary += "• 💰 **Great deal!** - Low-cost cargo with good profit potential\n"
		} else if costPerUnit <= 100 {
			textSummary += "• ✅ **Fair price** - Reasonable investment\n"
		} else {
			textSummary += "• 💰 **Premium goods** - Ensure good resale opportunities\n"
		}

		if remainingSpace <= 10 {
			textSummary += "• ⚠️ **Cargo nearly full** - find a market to sell soon\n"
			textSummary += "• 🏪 Use `sell_cargo` at profitable markets\n"
		} else if remainingSpace <= cargo.Capacity/4 {
			textSummary += "• 🟡 **Limited space** remaining - plan your next moves\n"
			textSummary += "• 🏪 Consider selling some cargo or find delivery contracts\n"
		} else {
			textSummary += "• 📦 **Space available** for more cargo\n"
			textSummary += "• 🛒 Continue trading or mining operations\n"
		}

		textSummary += "• 📊 Use `get_status_summary` to check your fleet status\n"
		textSummary += "• 🗺️ Use `find_waypoints` to find profitable markets\n"
		textSummary += "• 📋 Check contracts to see if this cargo fulfills any requirements\n"

		// Add trading strategy tips
		if transaction.TotalPrice >= 1000 {
			textSummary += "\n📈 **Trading Strategy:** This is a significant investment - track market prices for optimal resale timing!\n"
		}

		// Warn about cargo space
		if cargoPercent >= 90 {
			textSummary += "\n⚠️ **Warning:** Cargo hold is nearly full! Consider selling cargo soon to avoid being unable to mine or trade.\n"
		}

		ctxLogger.ToolCall("buy_cargo", true)
		ctxLogger.Debug("Buy cargo response size: %d bytes", len(jsonData))

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%s\n```", jsonData)),
			},
		}, nil
	}
}
