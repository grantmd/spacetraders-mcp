package status

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"

	"github.com/mark3labs/mcp-go/mcp"
)

// StatusTool provides a comprehensive status summary by aggregating multiple resources
type StatusTool struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewStatusTool creates a new status summary tool
func NewStatusTool(client *spacetraders.Client, logger *logging.Logger) *StatusTool {
	return &StatusTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *StatusTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "get_status_summary",
		Description: "Get a comprehensive status summary including agent info, ships, and contracts",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"include_ships": map[string]interface{}{
					"type":        "boolean",
					"description": "Include detailed ship information in the summary",
					"default":     true,
				},
				"include_contracts": map[string]interface{}{
					"type":        "boolean",
					"description": "Include contract information in the summary",
					"default":     true,
				},
			},
		},
	}
}

// Handler returns the tool handler function
func (t *StatusTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Set up context logger
		ctxLogger := t.logger.WithContext(ctx, "status-tool")
		ctxLogger.Debug("Getting comprehensive status summary")

		// Parse arguments
		includeShips := true
		includeContracts := true

		if request.Params.Arguments != nil {
			if argsMap, ok := request.Params.Arguments.(map[string]interface{}); ok {
				if val, exists := argsMap["include_ships"]; exists {
					if b, ok := val.(bool); ok {
						includeShips = b
					}
				}
				if val, exists := argsMap["include_contracts"]; exists {
					if b, ok := val.(bool); ok {
						includeContracts = b
					}
				}
			}
		}

		// Build status summary
		summary := map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"summary":   "SpaceTraders Agent Status Summary",
		}

		// Get agent information
		ctxLogger.Debug("Fetching agent information")
		agent, err := t.client.GetAgent()
		if err != nil {
			ctxLogger.Error("Failed to fetch agent info: %v", err)
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("❌ Error fetching agent information: %s", err.Error())),
				},
				IsError: true,
			}, nil
		}

		summary["agent"] = map[string]interface{}{
			"symbol":          agent.Symbol,
			"credits":         agent.Credits,
			"headquarters":    agent.Headquarters,
			"startingFaction": agent.StartingFaction,
			"shipCount":       agent.ShipCount,
		}

		ctxLogger.Info("Successfully retrieved agent info for: %s", agent.Symbol)

		// Get ships if requested
		if includeShips {
			ctxLogger.Debug("Fetching ships information")
			ships, err := t.client.GetShips()
			if err != nil {
				ctxLogger.Error("Failed to fetch ships: %v", err)
				summary["ships"] = map[string]interface{}{
					"error": fmt.Sprintf("Failed to fetch ships: %s", err.Error()),
				}
			} else {
				// Create ship summary
				shipsByStatus := make(map[string]int)
				shipsBySystem := make(map[string]int)
				totalCargo := 0
				totalCargoCapacity := 0

				for _, ship := range ships {
					shipsByStatus[ship.Nav.Status]++
					shipsBySystem[ship.Nav.SystemSymbol]++
					totalCargo += ship.Cargo.Units
					totalCargoCapacity += ship.Cargo.Capacity
				}

				summary["ships"] = map[string]interface{}{
					"total":        len(ships),
					"byStatus":     shipsByStatus,
					"bySystem":     shipsBySystem,
					"cargoUsage":   fmt.Sprintf("%d/%d", totalCargo, totalCargoCapacity),
					"cargoPercent": float64(totalCargo) / float64(totalCargoCapacity) * 100,
				}

				ctxLogger.Info("Successfully retrieved %d ships", len(ships))
			}
		}

		// Get contracts if requested
		if includeContracts {
			ctxLogger.Debug("Fetching contracts information")
			contracts, err := t.client.GetContracts()
			if err != nil {
				ctxLogger.Error("Failed to fetch contracts: %v", err)
				summary["contracts"] = map[string]interface{}{
					"error": fmt.Sprintf("Failed to fetch contracts: %s", err.Error()),
				}
			} else {
				// Create contract summary
				acceptedCount := 0
				fulfilledCount := 0
				totalValue := 0

				for _, contract := range contracts {
					if contract.Accepted {
						acceptedCount++
					}
					if contract.Fulfilled {
						fulfilledCount++
					}
					totalValue += contract.Terms.Payment.OnAccepted + contract.Terms.Payment.OnFulfilled
				}

				summary["contracts"] = map[string]interface{}{
					"total":      len(contracts),
					"accepted":   acceptedCount,
					"fulfilled":  fulfilledCount,
					"pending":    len(contracts) - acceptedCount,
					"totalValue": totalValue,
				}

				ctxLogger.Info("Successfully retrieved %d contracts", len(contracts))
			}
		}

		// Format the response
		jsonData, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			ctxLogger.Error("Failed to marshal status summary: %v", err)
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("❌ Error formatting status summary"),
				},
				IsError: true,
			}, nil
		}

		ctxLogger.ToolCall("get_status_summary", true)
		ctxLogger.Debug("Status summary response size: %d bytes", len(jsonData))

		// Create formatted text summary
		textSummary := t.formatTextSummary(summary)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("**Raw JSON Data:**\n```json\n%s\n```", string(jsonData))),
			},
		}, nil
	}
}

// formatTextSummary creates a human-readable text summary
func (t *StatusTool) formatTextSummary(summary map[string]interface{}) string {
	var text string

	text += "🚀 **SpaceTraders Status Summary**\n\n"

	// Agent info
	if agent, ok := summary["agent"].(map[string]interface{}); ok {
		text += fmt.Sprintf("👤 **Agent: %s**\n", agent["symbol"])
		text += fmt.Sprintf("💰 Credits: %v\n", agent["credits"])
		text += fmt.Sprintf("🏠 Headquarters: %s\n", agent["headquarters"])
		text += fmt.Sprintf("🏴 Faction: %s\n", agent["startingFaction"])
		text += fmt.Sprintf("🚢 Ships: %v\n\n", agent["shipCount"])
	}

	// Ships info
	if ships, ok := summary["ships"].(map[string]interface{}); ok {
		if errorMsg, hasError := ships["error"]; hasError {
			text += fmt.Sprintf("🚢 **Ships:** ❌ %s\n\n", errorMsg)
		} else {
			text += "🚢 **Fleet Status:**\n"
			text += fmt.Sprintf("  • Total Ships: %v\n", ships["total"])
			text += fmt.Sprintf("  • Cargo Usage: %s (%.1f%%)\n", ships["cargoUsage"], ships["cargoPercent"])

			if byStatus, ok := ships["byStatus"].(map[string]int); ok {
				text += "  • Ship Status:\n"
				for status, count := range byStatus {
					text += fmt.Sprintf("    - %s: %d\n", status, count)
				}
			}

			if bySystem, ok := ships["bySystem"].(map[string]int); ok && len(bySystem) > 0 {
				text += "  • Ships by System:\n"
				for system, count := range bySystem {
					text += fmt.Sprintf("    - %s: %d\n", system, count)
				}
			}
			text += "\n"
		}
	}

	// Contracts info
	if contracts, ok := summary["contracts"].(map[string]interface{}); ok {
		if errorMsg, hasError := contracts["error"]; hasError {
			text += fmt.Sprintf("📋 **Contracts:** ❌ %s\n\n", errorMsg)
		} else {
			text += "📋 **Contracts:**\n"
			text += fmt.Sprintf("  • Total: %v\n", contracts["total"])
			text += fmt.Sprintf("  • Accepted: %v\n", contracts["accepted"])
			text += fmt.Sprintf("  • Fulfilled: %v\n", contracts["fulfilled"])
			text += fmt.Sprintf("  • Pending: %v\n", contracts["pending"])
			text += fmt.Sprintf("  • Total Value: %v credits\n\n", contracts["totalValue"])
		}
	}

	text += "💡 **Quick Actions:**\n"
	text += "• Use `get_status_summary` for updated status\n"
	text += "• Read specific resources for detailed info:\n"
	text += "  - `spacetraders://agent/info`\n"
	text += "  - `spacetraders://ships/list`\n"
	text += "  - `spacetraders://contracts/list`\n"

	return text
}
