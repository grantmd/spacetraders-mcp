package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"

	"github.com/mark3labs/mcp-go/mcp"
)

// WaypointsResource handles the system waypoints information resource
type WaypointsResource struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewWaypointsResource creates a new waypoints resource handler
func NewWaypointsResource(client *spacetraders.Client, logger *logging.Logger) *WaypointsResource {
	return &WaypointsResource{
		client: client,
		logger: logger,
	}
}

// Resource returns the MCP resource definition
func (r *WaypointsResource) Resource() mcp.Resource {
	return mcp.Resource{
		URI:         "spacetraders://systems/{systemSymbol}/waypoints",
		Name:        "System Waypoints",
		Description: "List of all waypoints in a system with their types, traits, and orbital information",
		MIMEType:    "application/json",
	}
}

// Handler returns the resource handler function
func (r *WaypointsResource) Handler() func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// Parse the system symbol from the URI
		systemSymbol, err := r.parseSystemSymbol(request.Params.URI)
		if err != nil {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     fmt.Sprintf("Invalid resource URI: %s", err.Error()),
				},
			}, nil
		}

		// Set up context logger
		ctxLogger := r.logger.WithContext(ctx, "waypoints-resource")
		ctxLogger.Debug("Fetching waypoints for system %s from API", systemSymbol)

		// Get waypoints information from the API
		start := time.Now()
		waypoints, err := r.client.GetSystemWaypoints(systemSymbol)
		duration := time.Since(start)

		if err != nil {
			ctxLogger.Error("Failed to fetch waypoints for system %s: %v", systemSymbol, err)
			ctxLogger.APICall(fmt.Sprintf("/systems/%s/waypoints", systemSymbol), 0, duration.String())
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     fmt.Sprintf("Error fetching waypoints for system %s: %s", systemSymbol, err.Error()),
				},
			}, nil
		}

		ctxLogger.APICall(fmt.Sprintf("/systems/%s/waypoints", systemSymbol), 200, duration.String())
		ctxLogger.Info("Successfully retrieved %d waypoints for system %s", len(waypoints), systemSymbol)

		// Group waypoints by type for better organization
		waypointsByType := make(map[string][]spacetraders.SystemWaypoint)
		for _, waypoint := range waypoints {
			waypointsByType[waypoint.Type] = append(waypointsByType[waypoint.Type], waypoint)
		}

		// Format the response as structured JSON
		result := map[string]interface{}{
			"system":    systemSymbol,
			"waypoints": waypoints,
			"summary": map[string]interface{}{
				"total":     len(waypoints),
				"byType":    r.getWaypointTypeCounts(waypoints),
				"shipyards": r.getShipyardWaypoints(waypoints),
				"markets":   r.getMarketWaypoints(waypoints),
			},
		}

		// Convert to JSON for response
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			ctxLogger.Error("Failed to marshal waypoints data to JSON: %v", err)
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Error formatting waypoints information",
				},
			}, nil
		}

		ctxLogger.ResourceRead(request.Params.URI, true)
		ctxLogger.Debug("Waypoints resource response size: %d bytes", len(jsonData))

		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		}, nil
	}
}

// parseSystemSymbol extracts the system symbol from the resource URI
func (r *WaypointsResource) parseSystemSymbol(uri string) (string, error) {
	// Expected format: spacetraders://systems/{systemSymbol}/waypoints
	if !strings.HasPrefix(uri, "spacetraders://systems/") {
		return "", fmt.Errorf("invalid URI format")
	}

	// Remove the protocol prefix
	path := strings.TrimPrefix(uri, "spacetraders://systems/")

	// Split by '/' and expect at least 2 parts: systemSymbol and 'waypoints'
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "waypoints" {
		return "", fmt.Errorf("invalid URI format, expected spacetraders://systems/{systemSymbol}/waypoints")
	}

	systemSymbol := parts[0]
	if systemSymbol == "" {
		return "", fmt.Errorf("system symbol cannot be empty")
	}

	// URL decode the system symbol in case it contains special characters
	decoded, err := url.QueryUnescape(systemSymbol)
	if err != nil {
		return "", fmt.Errorf("invalid system symbol encoding: %w", err)
	}

	return decoded, nil
}

// getWaypointTypeCounts returns a count of waypoints by type
func (r *WaypointsResource) getWaypointTypeCounts(waypoints []spacetraders.SystemWaypoint) map[string]int {
	counts := make(map[string]int)
	for _, waypoint := range waypoints {
		counts[waypoint.Type]++
	}
	return counts
}

// getShipyardWaypoints returns waypoints that have shipyards
func (r *WaypointsResource) getShipyardWaypoints(waypoints []spacetraders.SystemWaypoint) []string {
	var shipyards []string
	for _, waypoint := range waypoints {
		for _, trait := range waypoint.Traits {
			if trait.Symbol == "SHIPYARD" {
				shipyards = append(shipyards, waypoint.Symbol)
				break
			}
		}
	}
	return shipyards
}

// getMarketWaypoints returns waypoints that have markets
func (r *WaypointsResource) getMarketWaypoints(waypoints []spacetraders.SystemWaypoint) []string {
	var markets []string
	for _, waypoint := range waypoints {
		for _, trait := range waypoint.Traits {
			if trait.Symbol == "MARKETPLACE" {
				markets = append(markets, waypoint.Symbol)
				break
			}
		}
	}
	return markets
}
