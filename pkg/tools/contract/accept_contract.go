package contract

import (
	"context"
	"encoding/json"
	"fmt"
	"spacetraders-mcp/pkg/spacetraders"

	"github.com/mark3labs/mcp-go/mcp"
)

// AcceptContractTool handles accepting contracts
type AcceptContractTool struct {
	client *spacetraders.Client
}

// NewAcceptContractTool creates a new AcceptContractTool
func NewAcceptContractTool(client *spacetraders.Client) *AcceptContractTool {
	return &AcceptContractTool{
		client: client,
	}
}

// Tool returns the MCP tool definition
func (t *AcceptContractTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "accept_contract",
		Description: "Accept a contract by its ID. This commits the agent to fulfilling the contract terms and provides an upfront payment.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"contract_id": map[string]interface{}{
					"type":        "string",
					"description": "The unique identifier of the contract to accept",
				},
			},
			Required: []string{"contract_id"},
		},
	}
}

// Handler returns the tool handler function
func (t *AcceptContractTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract contract ID from arguments using the helper method
		contractID, err := request.RequireString("contract_id")
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					mcp.NewTextContent("contract_id must be a valid string"),
				},
			}, nil
		}

		if contractID == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					mcp.NewTextContent("contract_id cannot be empty"),
				},
			}, nil
		}

		// Accept the contract
		contract, agent, err := t.client.AcceptContract(contractID)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("Failed to accept contract: %v", err)),
				},
			}, nil
		}

		// Format the response
		result := map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Successfully accepted contract %s", contractID),
			"contract": map[string]interface{}{
				"id":         contract.ID,
				"faction":    contract.FactionSymbol,
				"type":       contract.Type,
				"accepted":   contract.Accepted,
				"fulfilled":  contract.Fulfilled,
				"expiration": contract.Expiration,
				"terms": map[string]interface{}{
					"deadline": contract.Terms.Deadline,
					"payment": map[string]interface{}{
						"on_accepted":  contract.Terms.Payment.OnAccepted,
						"on_fulfilled": contract.Terms.Payment.OnFulfilled,
					},
					"deliver": contract.Terms.Deliver,
				},
			},
			"agent": map[string]interface{}{
				"symbol":  agent.Symbol,
				"credits": agent.Credits,
				"ships":   agent.ShipCount,
				"faction": agent.StartingFaction,
			},
		}

		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("Failed to format response: %v", err)),
				},
			}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(string(resultJSON)),
			},
		}, nil
	}
}
