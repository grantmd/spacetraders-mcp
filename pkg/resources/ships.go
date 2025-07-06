package resources

import (
	"context"
	"encoding/json"
	"time"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"

	"github.com/mark3labs/mcp-go/mcp"
)

// ShipsResource handles the ships information resource
type ShipsResource struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewShipsResource creates a new ships resource handler
func NewShipsResource(client *spacetraders.Client, logger *logging.Logger) *ShipsResource {
	return &ShipsResource{
		client: client,
		logger: logger,
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

		// Set up context logger
		ctxLogger := r.logger.WithContext(ctx, "ships-resource")
		ctxLogger.Debug("Fetching ships list from API")

		// Get ships information from the API
		start := time.Now()
		ships, err := r.client.GetAllShips()
		duration := time.Since(start)

		if err != nil {
			ctxLogger.Error("Failed to fetch ships info: %v", err)
			ctxLogger.APICall("/my/ships", 0, duration.String())
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Error fetching ships info: " + err.Error(),
				},
			}, nil
		}

		ctxLogger.APICall("/my/ships", 200, duration.String())
		ctxLogger.Info("Successfully retrieved %d ships", len(ships))

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
			ctxLogger.Error("Failed to marshal ships data to JSON: %v", err)
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Error formatting ships information",
				},
			}, nil
		}

		ctxLogger.ResourceRead(request.Params.URI, true)
		ctxLogger.Debug("Ships resource response size: %d bytes", len(jsonData))

		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		}, nil
	}
}
