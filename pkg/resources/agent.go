package resources

import (
	"context"
	"encoding/json"
	"time"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"

	"github.com/mark3labs/mcp-go/mcp"
)

// AgentResource handles the agent information resource
type AgentResource struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewAgentResource creates a new agent resource handler
func NewAgentResource(client *spacetraders.Client, logger *logging.Logger) *AgentResource {
	return &AgentResource{
		client: client,
		logger: logger,
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

		// Set up context logger
		ctxLogger := r.logger.WithContext(ctx, "agent-resource")
		ctxLogger.Debug("Fetching agent information from API")

		// Get agent information from the API
		start := time.Now()
		agent, err := r.client.GetAgent()
		duration := time.Since(start)

		if err != nil {
			ctxLogger.Error("Failed to fetch agent info: %v", err)
			ctxLogger.APICall("/my/agent", 0, duration.String())
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Error fetching agent info: " + err.Error(),
				},
			}, nil
		}

		ctxLogger.APICall("/my/agent", 200, duration.String())
		ctxLogger.Info("Successfully retrieved agent info for: %s", agent.Symbol)

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
			ctxLogger.Error("Failed to marshal agent data to JSON: %v", err)
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Error formatting agent information",
				},
			}, nil
		}

		ctxLogger.ResourceRead(request.Params.URI, true)
		ctxLogger.Debug("Agent resource response size: %d bytes", len(jsonData))

		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		}, nil
	}
}
