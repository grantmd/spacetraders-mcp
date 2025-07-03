package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"

	"github.com/mark3labs/mcp-go/mcp"
)

// SystemsResource handles the systems resource
type SystemsResource struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewSystemsResource creates a new systems resource handler
func NewSystemsResource(client *spacetraders.Client, logger *logging.Logger) *SystemsResource {
	return &SystemsResource{
		client: client,
		logger: logger,
	}
}

// Resource returns the MCP resource definition
func (r *SystemsResource) Resource() mcp.Resource {
	return mcp.Resource{
		URI:         "spacetraders://systems/*",
		Name:        "Systems Data",
		Description: "Systems information - use 'spacetraders://systems' for all systems or 'spacetraders://systems/{systemSymbol}' for specific system details",
		MIMEType:    "application/json",
	}
}

// Handler returns the resource handler function
func (r *SystemsResource) Handler() func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// Set up context logger
		ctxLogger := r.logger.WithContext(ctx, "systems-resource")
		ctxLogger.Debug("Processing systems resource request")

		// Check if this is a request for a specific system or all systems
		if request.Params.URI == "spacetraders://systems" {
			return r.handleSystemsList(ctx, request, ctxLogger)
		} else if strings.HasPrefix(request.Params.URI, "spacetraders://systems/") {
			return r.handleSpecificSystem(ctx, request, ctxLogger)
		} else {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Invalid systems resource URI",
				},
			}, nil
		}
	}
}

// handleSystemsList handles requests for the full systems list
func (r *SystemsResource) handleSystemsList(ctx context.Context, request mcp.ReadResourceRequest, ctxLogger *logging.ContextLogger) ([]mcp.ResourceContents, error) {
	ctxLogger.Debug("Fetching systems list from API")

	// Get systems from the API
	start := time.Now()
	systems, err := r.client.GetSystems()
	duration := time.Since(start)

	if err != nil {
		ctxLogger.Error("Failed to fetch systems: %v", err)
		ctxLogger.APICall("/systems", 0, duration.String())
		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "text/plain",
				Text:     "Error fetching systems: " + err.Error(),
			},
		}, nil
	}

	ctxLogger.APICall("/systems", 200, duration.String())
	ctxLogger.Info("Successfully retrieved %d systems", len(systems))

	// Format the response
	result := map[string]interface{}{
		"systems": r.formatSystemsList(systems),
		"meta": map[string]interface{}{
			"total":     len(systems),
			"retrieved": time.Now().UTC().Format(time.RFC3339),
		},
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		ctxLogger.Error("Failed to marshal systems data to JSON: %v", err)
		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "text/plain",
				Text:     "Error formatting systems information",
			},
		}, nil
	}

	ctxLogger.ResourceRead(request.Params.URI, true)
	ctxLogger.Debug("Systems resource response size: %d bytes", len(jsonData))

	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}

// handleSpecificSystem handles requests for a specific system
func (r *SystemsResource) handleSpecificSystem(ctx context.Context, request mcp.ReadResourceRequest, ctxLogger *logging.ContextLogger) ([]mcp.ResourceContents, error) {
	// Extract system symbol from URI
	systemSymbol, err := r.parseSystemSymbol(request.Params.URI)
	if err != nil {
		ctxLogger.Error("Failed to parse system symbol from URI: %v", err)
		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "text/plain",
				Text:     "Invalid system URI format",
			},
		}, nil
	}

	ctxLogger.Debug("Fetching system details for: %s", systemSymbol)

	// Get system details from the API
	start := time.Now()
	system, err := r.client.GetSystem(systemSymbol)
	duration := time.Since(start)

	if err != nil {
		ctxLogger.Error("Failed to fetch system %s: %v", systemSymbol, err)
		ctxLogger.APICall(fmt.Sprintf("/systems/%s", systemSymbol), 0, duration.String())
		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "text/plain",
				Text:     fmt.Sprintf("Error fetching system %s: %s", systemSymbol, err.Error()),
			},
		}, nil
	}

	ctxLogger.APICall(fmt.Sprintf("/systems/%s", systemSymbol), 200, duration.String())
	ctxLogger.Info("Successfully retrieved system details for: %s", systemSymbol)

	// Format the response
	result := map[string]interface{}{
		"system": r.formatSystemDetails(system),
		"meta": map[string]interface{}{
			"retrieved": time.Now().UTC().Format(time.RFC3339),
		},
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		ctxLogger.Error("Failed to marshal system data to JSON: %v", err)
		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "text/plain",
				Text:     "Error formatting system information",
			},
		}, nil
	}

	ctxLogger.ResourceRead(request.Params.URI, true)
	ctxLogger.Debug("System resource response size: %d bytes", len(jsonData))

	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}

// parseSystemSymbol extracts the system symbol from the URI
func (r *SystemsResource) parseSystemSymbol(uri string) (string, error) {
	// Expected format: spacetraders://systems/{systemSymbol}
	if !strings.HasPrefix(uri, "spacetraders://systems/") {
		return "", fmt.Errorf("invalid URI format")
	}

	// Remove the protocol prefix
	systemSymbol := strings.TrimPrefix(uri, "spacetraders://systems/")

	// System symbol shouldn't be empty
	if systemSymbol == "" {
		return "", fmt.Errorf("system symbol cannot be empty")
	}

	return systemSymbol, nil
}

// formatSystemsList formats a list of systems for the response
func (r *SystemsResource) formatSystemsList(systems []spacetraders.System) []map[string]interface{} {
	result := make([]map[string]interface{}, len(systems))

	for i, system := range systems {
		result[i] = map[string]interface{}{
			"symbol":       system.Symbol,
			"sectorSymbol": system.SectorSymbol,
			"type":         system.Type,
			"coordinates": map[string]interface{}{
				"x": system.X,
				"y": system.Y,
			},
			"waypoints": len(system.Waypoints),
			"factions":  len(system.Factions),
		}
	}

	return result
}

// formatSystemDetails formats detailed system information
func (r *SystemsResource) formatSystemDetails(system *spacetraders.System) map[string]interface{} {
	// Format waypoints
	waypoints := make([]map[string]interface{}, len(system.Waypoints))
	for i, waypoint := range system.Waypoints {
		waypoints[i] = map[string]interface{}{
			"symbol": waypoint.Symbol,
			"type":   waypoint.Type,
			"coordinates": map[string]interface{}{
				"x": waypoint.X,
				"y": waypoint.Y,
			},
		}
	}

	// Format factions
	factions := make([]map[string]interface{}, len(system.Factions))
	for i, faction := range system.Factions {
		factions[i] = map[string]interface{}{
			"symbol": faction.Symbol,
		}
	}

	return map[string]interface{}{
		"symbol":       system.Symbol,
		"sectorSymbol": system.SectorSymbol,
		"type":         system.Type,
		"coordinates": map[string]interface{}{
			"x": system.X,
			"y": system.Y,
		},
		"waypoints": waypoints,
		"factions":  factions,
	}
}
