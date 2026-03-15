package info

import (
	"context"
	"fmt"
	"strings"
	"time"

	"spacetraders-mcp/pkg/client"
	"spacetraders-mcp/pkg/logging"

	"github.com/mark3labs/mcp-go/mcp"
)

// ContractInfoTool provides detailed information about contracts
type ContractInfoTool struct {
	client *client.Client
	logger *logging.Logger
}

// NewContractInfoTool creates a new contract info tool
func NewContractInfoTool(client *client.Client, logger *logging.Logger) *ContractInfoTool {
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
		contracts, err := t.client.GetAllContracts()
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
		var filteredContracts []client.Contract
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
			fmt.Fprintf(&response, "📋 **Contract Details: %s**\n\n", contract.ID)
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

		hasMiningContract := false
		for _, contract := range filteredContracts {
			if !contract.Accepted {
				fmt.Fprintf(&response, "• Use `accept_contract` with contract_id=%s to accept this contract\n", contract.ID)
			}

			// Check if any contract requires mining
			for _, delivery := range contract.Terms.Deliver {
				if t.isMiningMaterial(delivery.TradeSymbol) {
					hasMiningContract = true
					break
				}
			}
		}

		if hasMiningContract {
			response.WriteString("• ⛏️ **Mining contracts detected!** You may need:\n")
			response.WriteString("  - A SHIP_MINING_DRONE for resource extraction\n")
			response.WriteString("  - Use `get_status_summary` to check your current fleet\n")
			response.WriteString("  - Use `purchase_ship` to buy a mining drone at a shipyard\n")
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
func (t *ContractInfoTool) formatContractDetails(contract client.Contract) string {
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

	fmt.Fprintf(&details, "**ID:** %s\n", contract.ID)
	fmt.Fprintf(&details, "**Status:** %s\n", status)
	fmt.Fprintf(&details, "**Type:** %s\n", contract.Type)
	fmt.Fprintf(&details, "**Faction:** %s\n", contract.FactionSymbol)

	// Payment information
	totalPayment := contract.Terms.Payment.OnAccepted + contract.Terms.Payment.OnFulfilled
	fmt.Fprintf(&details, "**Payment:** %d credits total\n", totalPayment)
	fmt.Fprintf(&details, "  • On Accept: %d credits\n", contract.Terms.Payment.OnAccepted)
	fmt.Fprintf(&details, "  • On Fulfill: %d credits\n", contract.Terms.Payment.OnFulfilled)

	// Deadlines
	if !contract.Accepted {
		fmt.Fprintf(&details, "**Accept By:** %s\n", contract.DeadlineToAccept)
	}
	fmt.Fprintf(&details, "**Complete By:** %s\n", contract.Terms.Deadline)
	fmt.Fprintf(&details, "**Expires:** %s\n", contract.Expiration)

	// Delivery requirements
	requiresMining := false
	miningMaterials := []string{}
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

			fmt.Fprintf(&details, "%d. %s **%s** (%d/%d units) → %s\n",
				i+1, progress, delivery.TradeSymbol,
				delivery.UnitsFulfilled, delivery.UnitsRequired,
				delivery.DestinationSymbol)

			if remaining > 0 {
				fmt.Fprintf(&details, "   *Need %d more units*\n", remaining)
			}

			// Check if this is a mining material
			if t.isMiningMaterial(delivery.TradeSymbol) {
				requiresMining = true
				miningMaterials = append(miningMaterials, delivery.TradeSymbol)
			}
		}
	}

	// Analysis
	details.WriteString("\n**Analysis:**\n")
	if !contract.Accepted {
		profitMargin := float64(contract.Terms.Payment.OnFulfilled) / float64(totalPayment) * 100
		fmt.Fprintf(&details, "• Profit margin: %.1f%% (%d of %d credits on completion)\n",
			profitMargin, contract.Terms.Payment.OnFulfilled, totalPayment)

		if len(contract.Terms.Deliver) > 0 {
			details.WriteString("• Requires cargo space and delivery logistics\n")
		}

		// Mining requirements analysis
		if requiresMining {
			fmt.Fprintf(&details, "• ⛏️ **MINING REQUIRED** for: %s\n", strings.Join(miningMaterials, ", "))
			details.WriteString("• You will need a SHIP_MINING_DRONE to extract these materials\n")
			details.WriteString("• Find asteroids or mining sites in systems to extract resources\n")
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
			fmt.Fprintf(&details, "• Progress: %.1f%% complete (%d/%d deliveries)\n",
				completionPercent, completed, total)
		}

		if requiresMining {
			fmt.Fprintf(&details, "• ⛏️ Mining needed for: %s\n", strings.Join(miningMaterials, ", "))
		}
	}

	return details.String()
}

// isMiningMaterial checks if a trade symbol represents a material that requires mining
func (t *ContractInfoTool) isMiningMaterial(tradeSymbol string) bool {
	miningMaterials := map[string]bool{
		"IRON_ORE":          true,
		"COPPER_ORE":        true,
		"ALUMINUM_ORE":      true,
		"SILVER_ORE":        true,
		"GOLD_ORE":          true,
		"PLATINUM_ORE":      true,
		"URANITE_ORE":       true,
		"MERITIUM_ORE":      true,
		"HYDROCARBON":       true,
		"QUARTZ_SAND":       true,
		"SILICON_CRYSTALS":  true,
		"AMMONIA_ICE":       true,
		"LIQUID_HYDROGEN":   true,
		"LIQUID_NITROGEN":   true,
		"ICE_WATER":         true,
		"EXOTIC_MATTER":     true,
		"GRAVITON_EMITTERS": true,
		"IRON":              true,
		"COPPER":            true,
		"ALUMINUM":          true,
		"SILVER":            true,
		"GOLD":              true,
		"PLATINUM":          true,
		"URANITE":           true,
		"MERITIUM":          true,
	}
	return miningMaterials[tradeSymbol]
}
