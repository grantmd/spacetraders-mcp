package resources

import (
	"context"
	"encoding/json"

	"spacetraders-mcp/pkg/spacetraders"

	"github.com/mark3labs/mcp-go/mcp"
)

// ContractsResource handles the contracts information resource
type ContractsResource struct {
	client *spacetraders.Client
}

// NewContractsResource creates a new contracts resource handler
func NewContractsResource(client *spacetraders.Client) *ContractsResource {
	return &ContractsResource{
		client: client,
	}
}

// Resource returns the MCP resource definition
func (r *ContractsResource) Resource() mcp.Resource {
	return mcp.Resource{
		URI:         "spacetraders://contracts/list",
		Name:        "Contracts List",
		Description: "List of all available contracts including terms, payments, and delivery requirements",
		MIMEType:    "application/json",
	}
}

// Handler returns the resource handler function
func (r *ContractsResource) Handler() func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// Validate the resource URI
		if request.Params.URI != "spacetraders://contracts/list" {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Invalid resource URI",
				},
			}, nil
		}

		// Get contracts information from the API
		contracts, err := r.client.GetContracts()
		if err != nil {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Error fetching contracts info: " + err.Error(),
				},
			}, nil
		}

		// Format the response as structured JSON
		result := map[string]interface{}{
			"contracts": contracts,
			"meta": map[string]interface{}{
				"count": len(contracts),
			},
		}

		// Convert to JSON for response
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Error formatting contracts information",
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
