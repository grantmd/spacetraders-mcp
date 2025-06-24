package info

import (
	"context"
	"fmt"
	"strings"
	"time"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"

	"github.com/mark3labs/mcp-go/mcp"
)

// ContractInfoTool provides detailed information about contracts
type ContractInfoTool struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewContractInfoTool creates a new contract info tool
func NewContractInfoTool(client *spacetraders.Client, logger *logging.Logger) *ContractInfoTool {
	return &ContractInfoTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *ContractInfoTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "get_contract_info",
		Description: "Get detailed information about contracts, including specific contract details by ID or all contracts overview",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"contract_id": map[string]interface{}{
					"type":        "string",
					"description": "Specific contract ID to get details for (optional - if not provided, shows all contracts)",
				},
				"include_fulfilled": map[string]interface{}{
					"type":        "boolean",
					"description": "Include fulfilled contracts in the results",
					"default":     false,
				},
			},
		},
	}
}

// Handler returns the tool handler function
func (t *ContractInfoTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Set up context logger
		ctxLogger := t.logger.WithContext(ctx, "contract-info-tool")
		ctxLogger.Debug("Getting contract information")

		// Parse arguments
		var contractID string
		includeFulfilled := false

		if request.Params.Arguments != nil {
			if argsMap, ok := request.Params.Arguments.(map[string]interface{}); ok {
				if id, exists := argsMap["contract_id"]; exists {
					if idStr, ok := id.(string); ok {
						contractID = strings.TrimSpace(idStr)
					}
				}
				if include, exists := argsMap["include_fulfilled"]; exists {
					if includeBool, ok := include.(bool); ok {
						includeFulfilled = includeBool
					}
				}
			}
		}

		// Get contracts from API
		start := time.Now()
		contracts, err := t.client.GetContracts()
		duration := time.Since(start)

		if err != nil {
			ctxLogger.Error("Failed to fetch contracts: %v", err)
			ctxLogger.APICall("/my/contracts", 0, duration.String())
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("❌ Error fetching contracts: %s", err.Error())),
				},
				IsError: true,
			}, nil
		}

		ctxLogger.APICall("/my/contracts", 200, duration.String())
		ctxLogger.Info("Successfully retrieved %d contracts", len(contracts))

		// Filter contracts if needed
		var filteredContracts []spacetraders.Contract
		for _, contract := range contracts {
			// Skip fulfilled contracts unless explicitly requested
			if contract.Fulfilled && !includeFulfilled {
				continue
			}

			// If specific contract ID requested, only include that one
			if contractID != "" && contract.ID != contractID {
				continue
			}

			filteredContracts = append(filteredContracts, contract)
		}

		// If specific contract ID was requested but not found
		if contractID != "" && len(filteredContracts) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("❌ Contract with ID '%s' not found or is fulfilled", contractID)),
				},
				IsError: true,
			}, nil
		}

		// Build response
		var response strings.Builder

		if contractID != "" {
			// Detailed view for specific contract
			contract := filteredContracts[0]
			response.WriteString(fmt.Sprintf("📋 **Contract Details: %s**\n\n", contract.ID))
			response.WriteString(t.formatContractDetails(contract))
		} else {
			// Overview of all contracts
			response.WriteString("📋 **Contracts Overview**\n\n")

			if len(filteredContracts) == 0 {
				response.WriteString("No contracts available")
				if !includeFulfilled {
					response.WriteString(" (use include_fulfilled=true to see completed contracts)")
				}
				response.WriteString(".\n")
			} else {
				for i, contract := range filteredContracts {
					if i > 0 {
						response.WriteString("\n---\n\n")
					}
					response.WriteString(t.formatContractDetails(contract))
				}
			}
		}

		// Add helpful actions
		response.WriteString("\n💡 **Available Actions:**\n")
		if contractID == "" && len(filteredContracts) > 0 {
			response.WriteString("• Use `get_contract_info` with a specific contract_id for detailed analysis\n")
		}

		for _, contract := range filteredContracts {
			if !contract.Accepted {
				response.WriteString(fmt.Sprintf("• Use `accept_contract` with contract_id=%s to accept this contract\n", contract.ID))
			}
		}

		ctxLogger.ToolCall("get_contract_info", true)
		ctxLogger.Debug("Contract info response size: %d bytes", len(response.String()))

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(response.String()),
			},
		}, nil
	}
}

// formatContractDetails formats detailed information for a single contract
func (t *ContractInfoTool) formatContractDetails(contract spacetraders.Contract) string {
	var details strings.Builder

	// Header with status
	status := "⏳ Available"
	if contract.Accepted {
		if contract.Fulfilled {
			status = "✅ Completed"
		} else {
			status = "🔄 In Progress"
		}
	}

	details.WriteString(fmt.Sprintf("**ID:** %s\n", contract.ID))
	details.WriteString(fmt.Sprintf("**Status:** %s\n", status))
	details.WriteString(fmt.Sprintf("**Type:** %s\n", contract.Type))
	details.WriteString(fmt.Sprintf("**Faction:** %s\n", contract.FactionSymbol))

	// Payment information
	totalPayment := contract.Terms.Payment.OnAccepted + contract.Terms.Payment.OnFulfilled
	details.WriteString(fmt.Sprintf("**Payment:** %d credits total\n", totalPayment))
	details.WriteString(fmt.Sprintf("  • On Accept: %d credits\n", contract.Terms.Payment.OnAccepted))
	details.WriteString(fmt.Sprintf("  • On Fulfill: %d credits\n", contract.Terms.Payment.OnFulfilled))

	// Deadlines
	if !contract.Accepted {
		details.WriteString(fmt.Sprintf("**Accept By:** %s\n", contract.DeadlineToAccept))
	}
	details.WriteString(fmt.Sprintf("**Complete By:** %s\n", contract.Terms.Deadline))
	details.WriteString(fmt.Sprintf("**Expires:** %s\n", contract.Expiration))

	// Delivery requirements
	if len(contract.Terms.Deliver) > 0 {
		details.WriteString("\n**Delivery Requirements:**\n")
		for i, delivery := range contract.Terms.Deliver {
			remaining := delivery.UnitsRequired - delivery.UnitsFulfilled
			progress := "🔴"
			if delivery.UnitsFulfilled == delivery.UnitsRequired {
				progress = "✅"
			} else if delivery.UnitsFulfilled > 0 {
				progress = "🟡"
			}

			details.WriteString(fmt.Sprintf("%d. %s **%s** (%d/%d units) → %s\n",
				i+1, progress, delivery.TradeSymbol,
				delivery.UnitsFulfilled, delivery.UnitsRequired,
				delivery.DestinationSymbol))

			if remaining > 0 {
				details.WriteString(fmt.Sprintf("   *Need %d more units*\n", remaining))
			}
		}
	}

	// Analysis
	details.WriteString("\n**Analysis:**\n")
	if !contract.Accepted {
		profitMargin := float64(contract.Terms.Payment.OnFulfilled) / float64(totalPayment) * 100
		details.WriteString(fmt.Sprintf("• Profit margin: %.1f%% (%d of %d credits on completion)\n",
			profitMargin, contract.Terms.Payment.OnFulfilled, totalPayment))

		if len(contract.Terms.Deliver) > 0 {
			details.WriteString("• Requires cargo space and delivery logistics\n")
		}
	} else if !contract.Fulfilled {
		completed := 0
		total := len(contract.Terms.Deliver)
		for _, delivery := range contract.Terms.Deliver {
			if delivery.UnitsFulfilled == delivery.UnitsRequired {
				completed++
			}
		}

		if total > 0 {
			completionPercent := float64(completed) / float64(total) * 100
			details.WriteString(fmt.Sprintf("• Progress: %.1f%% complete (%d/%d deliveries)\n",
				completionPercent, completed, total))
		}
	}

	return details.String()
}
