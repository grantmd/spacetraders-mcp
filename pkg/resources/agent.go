package resources

import (
	"context"
	"encoding/json"

	"spacetraders-mcp/pkg/spacetraders"

	"github.com/mark3labs/mcp-go/mcp"
)

// AgentResource handles the agent information resource
type AgentResource struct {
	client *spacetraders.Client
}

// NewAgentResource creates a new agent resource handler
func NewAgentResource(client *spacetraders.Client) *AgentResource {
	return &AgentResource{
		client: client,
	}
}

// Resource returns the MCP resource definition
func (r *AgentResource) Resource() mcp.Resource {
	return mcp.Resource{
		URI:         "spacetraders://agent/info",
		Name:        "Agent Information",
		Description: "Current agent information including credits, headquarters, faction, and ship count",
		MIMEType:    "application/json",
	}
}

// Handler returns the resource handler function
func (r *AgentResource) Handler() func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// Validate the resource URI
		if request.Params.URI != "spacetraders://agent/info" {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Invalid resource URI",
				},
			}, nil
		}

		// Get agent information from the API
		agent, err := r.client.GetAgent()
		if err != nil {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Error fetching agent info: " + err.Error(),
				},
			}, nil
		}

		// Format the response as structured JSON
		result := map[string]interface{}{
			"agent": map[string]interface{}{
				"accountId":       agent.AccountID,
				"symbol":          agent.Symbol,
				"headquarters":    agent.Headquarters,
				"credits":         agent.Credits,
				"startingFaction": agent.StartingFaction,
				"shipCount":       agent.ShipCount,
			},
		}

		// Convert to JSON for response
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Error formatting agent information",
				},
			}, nil
		}

		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		}, nil
	}
}
