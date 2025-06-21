package resources

import (
	"context"
	"encoding/json"

	"spacetraders-mcp/pkg/spacetraders"

	"github.com/mark3labs/mcp-go/mcp"
)

// ShipsResource handles the ships information resource
type ShipsResource struct {
	client *spacetraders.Client
}

// NewShipsResource creates a new ships resource handler
func NewShipsResource(client *spacetraders.Client) *ShipsResource {
	return &ShipsResource{
		client: client,
	}
}

// Resource returns the MCP resource definition
func (r *ShipsResource) Resource() mcp.Resource {
	return mcp.Resource{
		URI:         "spacetraders://ships/list",
		Name:        "Ships List",
		Description: "List of all ships owned by the agent with their status, location, and cargo information",
		MIMEType:    "application/json",
	}
}

// Handler returns the resource handler function
func (r *ShipsResource) Handler() func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// Validate the resource URI
		if request.Params.URI != "spacetraders://ships/list" {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Invalid resource URI",
				},
			}, nil
		}

		// Get ships information from the API
		ships, err := r.client.GetShips()
		if err != nil {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Error fetching ships info: " + err.Error(),
				},
			}, nil
		}

		// Format the response as structured JSON
		result := map[string]interface{}{
			"ships": ships,
			"meta": map[string]interface{}{
				"count": len(ships),
			},
		}

		// Convert to JSON for response
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Error formatting ships information",
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
