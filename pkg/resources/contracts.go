package resources

import (
	"context"
	"encoding/json"
	"time"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/client"

	"github.com/mark3labs/mcp-go/mcp"
)

// ContractsResource handles the contracts information resource
type ContractsResource struct {
	client *client.Client
	logger *logging.Logger
}

// NewContractsResource creates a new contracts resource handler
func NewContractsResource(client *client.Client, logger *logging.Logger) *ContractsResource {
	return &ContractsResource{
		client: client,
		logger: logger,
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

		// Set up context logger
		ctxLogger := r.logger.WithContext(ctx, "contracts-resource")
		ctxLogger.Debug("Fetching contracts list from API")

		// Get contracts information from the API
		start := time.Now()
		contracts, err := r.client.GetAllContracts()
		duration := time.Since(start)

		if err != nil {
			ctxLogger.Error("Failed to fetch contracts info: %v", err)
			ctxLogger.APICall("/my/contracts", 0, duration.String())
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Error fetching contracts info: " + err.Error(),
				},
			}, nil
		}

		ctxLogger.APICall("/my/contracts", 200, duration.String())
		ctxLogger.Info("Successfully retrieved %d contracts", len(contracts))

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
			ctxLogger.Error("Failed to marshal contracts data to JSON: %v", err)
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Error formatting contracts information",
				},
			}, nil
		}

		ctxLogger.ResourceRead(request.Params.URI, true)
		ctxLogger.Debug("Contracts resource response size: %d bytes", len(jsonData))

		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		}, nil
	}
}
