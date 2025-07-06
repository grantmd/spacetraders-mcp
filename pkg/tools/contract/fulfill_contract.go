package contract

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

// FulfillContractTool handles fulfilling contracts
type FulfillContractTool struct {
	client *client.Client
	logger *logging.Logger
}

// NewFulfillContractTool creates a new fulfill contract tool
func NewFulfillContractTool(client *client.Client, logger *logging.Logger) *FulfillContractTool {
	return &FulfillContractTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *FulfillContractTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "fulfill_contract",
		Description: "Fulfill a contract by delivering all required cargo. The contract must be accepted and all delivery requirements must be met before fulfillment.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"contract_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the contract to fulfill (e.g., 'CONTRACT_123')",
				},
			},
			Required: []string{"contract_id"},
		},
	}
}

// Handler returns the tool handler function
func (t *FulfillContractTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Set up context logger
		ctxLogger := t.logger.WithContext(ctx, "fulfill-contract-tool")
		ctxLogger.Debug("Processing contract fulfillment request")

		// Parse arguments
		contractID := ""

		if request.Params.Arguments == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("❌ Missing required argument: contract_id"),
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
		}

		if contractID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("❌ contract_id is required and must be a non-empty string"),
				},
				IsError: true,
			}, nil
		}

		ctxLogger.Info("Attempting to fulfill contract %s", contractID)

		// Fulfill the contract
		start := time.Now()
		resp, err := t.client.FulfillContract(contractID)
		duration := time.Since(start)

		if err != nil {
			ctxLogger.Error("Failed to fulfill contract: %v", err)
			ctxLogger.APICall(fmt.Sprintf("/my/contracts/%s/fulfill", contractID), 0, duration.String())
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("❌ Failed to fulfill contract: %s", err.Error())),
				},
				IsError: true,
			}, nil
		}

		ctxLogger.APICall(fmt.Sprintf("/my/contracts/%s/fulfill", contractID), 200, duration.String())
		ctxLogger.Info("Successfully fulfilled contract %s, received %d credits", contractID, resp.Data.Contract.Terms.Payment.OnFulfilled)

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

		// Format the response
		result := map[string]interface{}{
			"success":     true,
			"message":     fmt.Sprintf("Successfully fulfilled contract %s", contractID),
			"contract_id": contractID,
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
			"agent": map[string]interface{}{
				"credits": resp.Data.Agent.Credits,
			},
		}

		jsonData := utils.FormatJSON(result)

		// Calculate total payment received
		totalPayment := resp.Data.Contract.Terms.Payment.OnAccepted + resp.Data.Contract.Terms.Payment.OnFulfilled
		fulfillmentPayment := resp.Data.Contract.Terms.Payment.OnFulfilled

		// Create formatted text summary
		textSummary := "🎉 **Contract Fulfilled Successfully!**\n\n"
		textSummary += fmt.Sprintf("**Contract ID:** %s\n", contractID)
		textSummary += fmt.Sprintf("**Faction:** %s\n", resp.Data.Contract.FactionSymbol)
		textSummary += fmt.Sprintf("**Type:** %s\n", resp.Data.Contract.Type)
		textSummary += "**Status:** ✅ Fulfilled\n\n"

		// Payment information
		textSummary += "💰 **Payment Details:**\n"
		textSummary += fmt.Sprintf("• Fulfillment Bonus: **%d credits**\n", fulfillmentPayment)
		textSummary += fmt.Sprintf("• Total Contract Value: %d credits\n", totalPayment)
		textSummary += fmt.Sprintf("• Your Current Credits: **%d**\n\n", resp.Data.Agent.Credits)

		// Delivery information
		textSummary += "📦 **Delivery Summary:**\n"
		for _, good := range resp.Data.Contract.Terms.Deliver {
			textSummary += fmt.Sprintf("• %s: %d/%d units → %s ✅\n",
				good.TradeSymbol,
				good.UnitsFulfilled,
				good.UnitsRequired,
				good.DestinationSymbol)
		}

		// Contract timeline
		textSummary += "\n📅 **Contract Timeline:**\n"
		textSummary += fmt.Sprintf("• Deadline: %s\n", resp.Data.Contract.Terms.Deadline)
		textSummary += fmt.Sprintf("• Expiration: %s\n", resp.Data.Contract.Expiration)

		// Success celebration and next steps
		textSummary += "\n🎊 **Congratulations!**\n"
		textSummary += "You have successfully completed this contract and earned the full payment!\n\n"

		textSummary += "💡 **Next Steps:**\n"
		textSummary += "• 📋 Use `get_status_summary` to see available contracts\n"
		textSummary += "• 🤝 Look for new contracts to accept and complete\n"
		textSummary += "• 🚀 Use your earned credits to expand your fleet\n"
		textSummary += "• 📈 Build reputation with factions for better contracts\n"

		// Add reputation building tip
		textSummary += fmt.Sprintf("\n🌟 **Reputation Boost:** Completing contracts with %s improves your standing with this faction!\n", resp.Data.Contract.FactionSymbol)

		// Payment celebration based on amount
		if fulfillmentPayment >= 50000 {
			textSummary += "\n💎 **Exceptional Payout!** This was a high-value contract - well done!\n"
		} else if fulfillmentPayment >= 20000 {
			textSummary += "\n💰 **Great Earnings!** Solid contract completion!\n"
		} else if fulfillmentPayment >= 5000 {
			textSummary += "\n✅ **Good Work!** Building your trading empire step by step!\n"
		}

		ctxLogger.ToolCall("fulfill_contract", true)
		ctxLogger.Debug("Fulfill contract response size: %d bytes", len(jsonData))

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%s\n```", jsonData)),
			},
		}, nil
	}
}
