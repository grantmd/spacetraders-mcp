package contract

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"spacetraders-mcp/pkg/client"
	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/tools/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

// DeliverContractTool handles delivering goods to contracts
type DeliverContractTool struct {
	client *client.Client
	logger *logging.Logger
}

// NewDeliverContractTool creates a new deliver contract tool
func NewDeliverContractTool(client *client.Client, logger *logging.Logger) *DeliverContractTool {
	return &DeliverContractTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *DeliverContractTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "deliver_contract",
		Description: "Deliver goods to a contract. Ship must be docked at the delivery location and have the required cargo. This is used to incrementally deliver goods to contracts before final fulfillment.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"contract_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the contract to deliver goods to (e.g., 'CONTRACT_123')",
				},
				"ship_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the ship that will deliver the goods (e.g., 'MYSHIP-1')",
				},
				"trade_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the trade good to deliver (e.g., 'IRON_ORE', 'COPPER')",
				},
				"units": map[string]interface{}{
					"type":        "integer",
					"description": "Number of units to deliver",
					"minimum":     1,
				},
			},
			Required: []string{"contract_id", "ship_symbol", "trade_symbol", "units"},
		},
	}
}

// Handler returns the tool handler function
func (t *DeliverContractTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Set up context logger
		ctxLogger := t.logger.WithContext(ctx, "deliver-contract-tool")
		ctxLogger.Debug("Processing contract delivery request")

		// Parse arguments
		contractID := ""
		shipSymbol := ""
		tradeSymbol := ""
		units := 0

		if request.Params.Arguments == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("❌ Missing required arguments: contract_id, ship_symbol, trade_symbol, units"),
				},
				IsError: true,
			}, nil
		}

		if argsMap, ok := request.Params.Arguments.(map[string]interface{}); ok {
			if cid, exists := argsMap["contract_id"]; exists {
				if cidStr, ok := cid.(string); ok {
					contractID = strings.TrimSpace(cidStr)
				}
			}
			if ss, exists := argsMap["ship_symbol"]; exists {
				if ssStr, ok := ss.(string); ok {
					shipSymbol = strings.TrimSpace(ssStr)
				}
			}
			if ts, exists := argsMap["trade_symbol"]; exists {
				if tsStr, ok := ts.(string); ok {
					tradeSymbol = strings.TrimSpace(tsStr)
				}
			}
			if u, exists := argsMap["units"]; exists {
				switch v := u.(type) {
				case int:
					units = v
				case float64:
					units = int(v)
				case string:
					if parsed, err := strconv.Atoi(v); err == nil {
						units = parsed
					}
				}
			}
		}

		// Validate required arguments
		if contractID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("❌ contract_id is required and must be a non-empty string"),
				},
				IsError: true,
			}, nil
		}

		if shipSymbol == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("❌ ship_symbol is required and must be a non-empty string"),
				},
				IsError: true,
			}, nil
		}

		if tradeSymbol == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("❌ trade_symbol is required and must be a non-empty string"),
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

		ctxLogger.Info("Attempting to deliver %d units of %s from ship %s to contract %s", units, tradeSymbol, shipSymbol, contractID)

		// Deliver goods to contract
		start := time.Now()
		resp, err := t.client.DeliverContract(contractID, shipSymbol, tradeSymbol, units)
		duration := time.Since(start)

		if err != nil {
			ctxLogger.Error("Failed to deliver contract goods: %v", err)
			ctxLogger.APICall(fmt.Sprintf("/my/contracts/%s/deliver", contractID), 0, duration.String())
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("❌ Failed to deliver contract goods: %s", err.Error())),
				},
				IsError: true,
			}, nil
		}

		ctxLogger.APICall(fmt.Sprintf("/my/contracts/%s/deliver", contractID), 200, duration.String())
		ctxLogger.Info("Successfully delivered %d units of %s to contract %s", units, tradeSymbol, contractID)

		// Format delivery goods
		deliveryGoods := make([]map[string]interface{}, len(resp.Data.Contract.Terms.Deliver))
		for i, good := range resp.Data.Contract.Terms.Deliver {
			deliveryGoods[i] = map[string]interface{}{
				"trade_symbol":       good.TradeSymbol,
				"destination_symbol": good.DestinationSymbol,
				"units_required":     good.UnitsRequired,
				"units_fulfilled":    good.UnitsFulfilled,
			}
		}

		// Format cargo items
		cargoItems := make([]map[string]interface{}, len(resp.Data.Cargo.Inventory))
		for i, item := range resp.Data.Cargo.Inventory {
			cargoItems[i] = map[string]interface{}{
				"symbol":      item.Symbol,
				"name":        item.Name,
				"description": item.Description,
				"units":       item.Units,
			}
		}

		// Format the response
		result := map[string]interface{}{
			"success":      true,
			"message":      fmt.Sprintf("Successfully delivered %d units of %s to contract %s", units, tradeSymbol, contractID),
			"contract_id":  contractID,
			"ship_symbol":  shipSymbol,
			"trade_symbol": tradeSymbol,
			"units":        units,
			"contract": map[string]interface{}{
				"id":                 resp.Data.Contract.ID,
				"faction_symbol":     resp.Data.Contract.FactionSymbol,
				"type":               resp.Data.Contract.Type,
				"accepted":           resp.Data.Contract.Accepted,
				"fulfilled":          resp.Data.Contract.Fulfilled,
				"expiration":         resp.Data.Contract.Expiration,
				"deadline_to_accept": resp.Data.Contract.DeadlineToAccept,
				"terms": map[string]interface{}{
					"deadline": resp.Data.Contract.Terms.Deadline,
					"payment": map[string]interface{}{
						"on_accepted":  resp.Data.Contract.Terms.Payment.OnAccepted,
						"on_fulfilled": resp.Data.Contract.Terms.Payment.OnFulfilled,
					},
					"deliver": deliveryGoods,
				},
			},
			"cargo": map[string]interface{}{
				"capacity":  resp.Data.Cargo.Capacity,
				"units":     resp.Data.Cargo.Units,
				"inventory": cargoItems,
			},
		}

		jsonData := utils.FormatJSON(result)

		// Create formatted text summary
		textSummary := "📦 **Contract Goods Delivered Successfully!**\n\n"
		textSummary += fmt.Sprintf("**Contract ID:** %s\n", contractID)
		textSummary += fmt.Sprintf("**Ship:** %s\n", shipSymbol)
		textSummary += fmt.Sprintf("**Delivered:** %d units of %s\n", units, tradeSymbol)
		textSummary += fmt.Sprintf("**Faction:** %s\n", resp.Data.Contract.FactionSymbol)
		textSummary += fmt.Sprintf("**Contract Type:** %s\n\n", resp.Data.Contract.Type)

		// Delivery progress
		textSummary += "📊 **Delivery Progress:**\n"
		for _, good := range resp.Data.Contract.Terms.Deliver {
			progress := float64(good.UnitsFulfilled) / float64(good.UnitsRequired) * 100
			statusIcon := "🟡"
			if good.UnitsFulfilled >= good.UnitsRequired {
				statusIcon = "✅"
			}
			textSummary += fmt.Sprintf("• %s %s: **%d/%d units** (%.1f%%) → %s\n",
				statusIcon, good.TradeSymbol, good.UnitsFulfilled, good.UnitsRequired, progress, good.DestinationSymbol)
		}

		// Contract status
		textSummary += "\n📋 **Contract Status:**\n"
		if resp.Data.Contract.Fulfilled {
			textSummary += "• Status: ✅ **FULFILLED** (ready to claim rewards!)\n"
		} else {
			textSummary += "• Status: 🔄 **IN PROGRESS** (continue delivering)\n"
		}
		textSummary += fmt.Sprintf("• Accepted: %v\n", resp.Data.Contract.Accepted)
		textSummary += fmt.Sprintf("• Deadline: %s\n", resp.Data.Contract.Terms.Deadline)
		textSummary += fmt.Sprintf("• Expiration: %s\n", resp.Data.Contract.Expiration)

		// Cargo status
		textSummary += "\n🚢 **Ship Cargo Status:**\n"
		cargoPercent := float64(resp.Data.Cargo.Units) / float64(resp.Data.Cargo.Capacity) * 100
		textSummary += fmt.Sprintf("• Capacity: **%d/%d units** (%.1f%% full)\n", resp.Data.Cargo.Units, resp.Data.Cargo.Capacity, cargoPercent)

		if len(resp.Data.Cargo.Inventory) > 0 {
			textSummary += "• Current Inventory:\n"
			for _, item := range resp.Data.Cargo.Inventory {
				textSummary += fmt.Sprintf("  - %s: %d units\n", item.Symbol, item.Units)
			}
		}

		// Payment information
		textSummary += "\n💰 **Payment Information:**\n"
		textSummary += fmt.Sprintf("• On Acceptance: %d credits\n", resp.Data.Contract.Terms.Payment.OnAccepted)
		textSummary += fmt.Sprintf("• On Fulfillment: **%d credits**\n", resp.Data.Contract.Terms.Payment.OnFulfilled)
		totalPayment := resp.Data.Contract.Terms.Payment.OnAccepted + resp.Data.Contract.Terms.Payment.OnFulfilled
		textSummary += fmt.Sprintf("• Total Contract Value: **%d credits**\n", totalPayment)

		// Next steps
		textSummary += "\n🎯 **Next Steps:**\n"
		if resp.Data.Contract.Fulfilled {
			textSummary += "• 🎉 **Contract fully completed!** Use `fulfill_contract` to claim rewards\n"
			textSummary += "• 💰 This will award you the full payment amount\n"
			textSummary += "• 🌟 You'll gain reputation with the faction\n"
		} else {
			textSummary += "• 🔄 **Continue delivering** remaining goods to complete the contract\n"
			textSummary += "• 🛒 **Purchase more cargo** if needed from markets\n"
			textSummary += "• 🚀 **Navigate to delivery locations** as required\n"
			textSummary += "• ⏰ **Monitor deadline** to avoid contract expiration\n"
		}

		// Progress celebration
		var allCompleted = true
		for _, good := range resp.Data.Contract.Terms.Deliver {
			if good.UnitsFulfilled < good.UnitsRequired {
				allCompleted = false
				break
			}
		}

		if allCompleted {
			textSummary += "\n🎊 **All deliveries completed!** Ready to fulfill the contract!\n"
		} else {
			completedCount := 0
			for _, good := range resp.Data.Contract.Terms.Deliver {
				if good.UnitsFulfilled >= good.UnitsRequired {
					completedCount++
				}
			}
			if completedCount > 0 {
				textSummary += fmt.Sprintf("\n✨ **Great progress!** %d/%d delivery requirements completed!\n", completedCount, len(resp.Data.Contract.Terms.Deliver))
			}
		}

		ctxLogger.ToolCall("deliver_contract", true)
		ctxLogger.Debug("Deliver contract response size: %d bytes", len(jsonData))

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%s\n```", jsonData)),
			},
		}, nil
	}
}
