package ships

import (
	"context"
	"fmt"
	"strings"
	"time"

	"spacetraders-mcp/pkg/client"
	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/tools/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

// JettisonCargoTool handles jettisoning cargo from ships
type JettisonCargoTool struct {
	client *client.Client
	logger *logging.Logger
}

// NewJettisonCargoTool creates a new jettison cargo tool
func NewJettisonCargoTool(client *client.Client, logger *logging.Logger) *JettisonCargoTool {
	return &JettisonCargoTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *JettisonCargoTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "jettison_cargo",
		Description: "Jettison (dump) cargo from a ship to free up space. The cargo will be lost permanently. Ship must be in orbit to jettison cargo.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"ship_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the ship to jettison cargo from (e.g., 'SHIP_1234')",
				},
				"cargo_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the cargo item to jettison (e.g., 'IRON_ORE', 'ALUMINUM_ORE')",
				},
				"units": map[string]interface{}{
					"type":        "integer",
					"description": "Number of units to jettison",
					"minimum":     1,
				},
			},
			Required: []string{"ship_symbol", "cargo_symbol", "units"},
		},
	}
}

// Handler returns the tool handler function
func (t *JettisonCargoTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Set up context logger
		ctxLogger := t.logger.WithContext(ctx, "jettison-cargo-tool")
		ctxLogger.Debug("Processing cargo jettison request")

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

		ctxLogger.Info("Attempting to jettison %d units of %s from ship %s", units, cargoSymbol, shipSymbol)

		// Jettison the cargo
		start := time.Now()
		resp, err := t.client.JettisonCargo(shipSymbol, cargoSymbol, units)
		duration := time.Since(start)

		if err != nil {
			ctxLogger.Error("Failed to jettison cargo: %v", err)
			ctxLogger.APICall(fmt.Sprintf("/my/ships/%s/jettison", shipSymbol), 0, duration.String())
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("❌ Failed to jettison cargo: %s", err.Error())),
				},
				IsError: true,
			}, nil
		}

		cargo := resp.Data.Cargo

		ctxLogger.APICall(fmt.Sprintf("/my/ships/%s/jettison", shipSymbol), 200, duration.String())
		ctxLogger.Info("Successfully jettisoned %d units of %s from ship %s", units, cargoSymbol, shipSymbol)

		// Format the response
		result := map[string]interface{}{
			"success":          true,
			"message":          fmt.Sprintf("Successfully jettisoned %d units of %s from ship %s", units, cargoSymbol, shipSymbol),
			"ship_symbol":      shipSymbol,
			"cargo_symbol":     cargoSymbol,
			"units_jettisoned": units,
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
		}

		jsonData := utils.FormatJSON(result)

		// Calculate cargo utilization
		cargoPercent := float64(cargo.Units) / float64(cargo.Capacity) * 100
		freedSpace := cargo.Capacity - cargo.Units

		// Find the jettisoned item in inventory to get its name
		jettisonedItemName := cargoSymbol
		for _, item := range cargo.Inventory {
			if item.Symbol == cargoSymbol {
				jettisonedItemName = item.Name
				break
			}
		}

		// Create formatted text summary
		textSummary := "🗑️ **Cargo Jettisoned Successfully!**\n\n"
		textSummary += fmt.Sprintf("**Ship:** %s\n", shipSymbol)
		textSummary += fmt.Sprintf("**Jettisoned:** %d units of %s\n", units, jettisonedItemName)
		textSummary += fmt.Sprintf("**Cargo Status:** %d/%d units (%.1f%% full)\n", cargo.Units, cargo.Capacity, cargoPercent)
		textSummary += fmt.Sprintf("**Free Space:** %d units available\n\n", freedSpace)

		// Show warning about permanent loss
		textSummary += "⚠️ **Warning:** The jettisoned cargo is permanently lost and cannot be recovered!\n\n"

		// Show current cargo inventory
		if len(cargo.Inventory) > 0 {
			textSummary += "**Remaining Cargo Inventory:**\n"
			for _, item := range cargo.Inventory {
				textSummary += fmt.Sprintf("- %s: %d units\n", item.Name, item.Units)
			}
		} else {
			textSummary += "**Cargo Hold:** Empty - ready for new cargo!\n"
		}

		// Add helpful tips based on cargo status
		textSummary += "\n💡 **Next Steps:**\n"
		if freedSpace >= cargo.Capacity/2 {
			textSummary += "• ✅ **Plenty of space** available for mining or trading\n"
		} else if freedSpace >= cargo.Capacity/4 {
			textSummary += "• 🟡 **Moderate space** available - consider what to do next\n"
		} else {
			textSummary += "• 🟠 **Limited space** remaining - may need to jettison more or dock to sell\n"
		}

		textSummary += "• ⛏️ Use `extract_resources` to mine more materials\n"
		textSummary += "• 🏪 Dock at a marketplace to sell valuable cargo instead of jettisoning\n"
		textSummary += "• 📊 Use `get_status_summary` to check your ship status\n"

		// Add recommendation about selling vs jettisoning
		if cargoPercent < 50 {
			textSummary += "\n💰 **Pro Tip:** Consider selling cargo at a marketplace instead of jettisoning to earn credits!\n"
		}

		ctxLogger.ToolCall("jettison_cargo", true)
		ctxLogger.Debug("Jettison cargo response size: %d bytes", len(jsonData))

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%s\n```", jsonData)),
			},
		}, nil
	}
}
