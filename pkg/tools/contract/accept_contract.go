package contract

import (
	"context"
	"encoding/json"
	"fmt"
	"spacetraders-mcp/pkg/client"

	"github.com/mark3labs/mcp-go/mcp"
)

// AcceptContractTool handles accepting contracts
type AcceptContractTool struct {
	client *client.Client
}

// NewAcceptContractTool creates a new AcceptContractTool
func NewAcceptContractTool(client *client.Client) *AcceptContractTool {
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
		resp, err := t.client.AcceptContract(contractID)
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
				"id":         resp.Data.Contract.ID,
				"faction":    resp.Data.Contract.FactionSymbol,
				"type":       resp.Data.Contract.Type,
				"accepted":   resp.Data.Contract.Accepted,
				"fulfilled":  resp.Data.Contract.Fulfilled,
				"expiration": resp.Data.Contract.Expiration,
				"deadline":   resp.Data.Contract.DeadlineToAccept,
				"terms": map[string]interface{}{
					"deadline": resp.Data.Contract.Terms.Deadline,
					"payment": map[string]interface{}{
						"on_accepted":  resp.Data.Contract.Terms.Payment.OnAccepted,
						"on_fulfilled": resp.Data.Contract.Terms.Payment.OnFulfilled,
					},
					"deliver": resp.Data.Contract.Terms.Deliver,
				},
			},
			"agent": map[string]interface{}{
				"symbol":  resp.Data.Agent.Symbol,
				"credits": resp.Data.Agent.Credits,
				"ships":   resp.Data.Agent.ShipCount,
				"faction": resp.Data.Agent.StartingFaction,
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
