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

// ShipyardResource handles the shipyard information resource
type ShipyardResource struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewShipyardResource creates a new shipyard resource handler
func NewShipyardResource(client *spacetraders.Client, logger *logging.Logger) *ShipyardResource {
	return &ShipyardResource{
		client: client,
		logger: logger,
	}
}

// Resource returns the MCP resource definition
func (r *ShipyardResource) Resource() mcp.Resource {
	return mcp.Resource{
		URI:         "spacetraders://systems/{systemSymbol}/waypoints/{waypointSymbol}/shipyard",
		Name:        "Shipyard Information",
		Description: "Information about available ships, prices, and transactions at a shipyard",
		MIMEType:    "application/json",
	}
}

// Handler returns the resource handler function
func (r *ShipyardResource) Handler() func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// Parse the system and waypoint symbols from the URI
		systemSymbol, waypointSymbol, err := r.parseShipyardURI(request.Params.URI)
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
		ctxLogger := r.logger.WithContext(ctx, "shipyard-resource")
		ctxLogger.Debug("Fetching shipyard info for %s at %s from API", waypointSymbol, systemSymbol)

		// Get shipyard information from the API
		start := time.Now()
		shipyard, err := r.client.GetShipyard(systemSymbol, waypointSymbol)
		duration := time.Since(start)

		if err != nil {
			ctxLogger.Error("Failed to fetch shipyard info for %s at %s: %v", waypointSymbol, systemSymbol, err)
			ctxLogger.APICall(fmt.Sprintf("/systems/%s/waypoints/%s/shipyard", systemSymbol, waypointSymbol), 0, duration.String())
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     fmt.Sprintf("Error fetching shipyard info for %s at %s: %s", waypointSymbol, systemSymbol, err.Error()),
				},
			}, nil
		}

		ctxLogger.APICall(fmt.Sprintf("/systems/%s/waypoints/%s/shipyard", systemSymbol, waypointSymbol), 200, duration.String())
		ctxLogger.Info("Successfully retrieved shipyard info for %s at %s", waypointSymbol, systemSymbol)

		// Format the response as structured JSON with additional analysis
		result := map[string]interface{}{
			"system":   systemSymbol,
			"waypoint": waypointSymbol,
			"shipyard": shipyard,
			"summary": map[string]interface{}{
				"availableShipTypes":  r.getAvailableShipTypes(shipyard),
				"totalShipsAvailable": len(shipyard.Ships),
				"priceRange":          r.getPriceRange(shipyard.Ships),
				"shipsByType":         r.getShipsByType(shipyard.Ships),
				"shipsBySupply":       r.getShipsBySupply(shipyard.Ships),
				"modificationsFee":    shipyard.ModificationsFee,
				"recentTransactions":  len(shipyard.Transactions),
			},
		}

		// Convert to JSON for response
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			ctxLogger.Error("Failed to marshal shipyard data to JSON: %v", err)
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Error formatting shipyard information",
				},
			}, nil
		}

		ctxLogger.ResourceRead(request.Params.URI, true)
		ctxLogger.Debug("Shipyard resource response size: %d bytes", len(jsonData))

		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		}, nil
	}
}

// parseShipyardURI extracts the system and waypoint symbols from the resource URI
func (r *ShipyardResource) parseShipyardURI(uri string) (string, string, error) {
	// Expected format: spacetraders://systems/{systemSymbol}/waypoints/{waypointSymbol}/shipyard
	if !strings.HasPrefix(uri, "spacetraders://systems/") {
		return "", "", fmt.Errorf("invalid URI format")
	}

	// Remove the protocol prefix
	path := strings.TrimPrefix(uri, "spacetraders://systems/")

	// Split by '/' and expect 4 parts: systemSymbol, 'waypoints', waypointSymbol, 'shipyard'
	parts := strings.Split(path, "/")
	if len(parts) != 4 || parts[1] != "waypoints" || parts[3] != "shipyard" {
		return "", "", fmt.Errorf("invalid URI format, expected spacetraders://systems/{systemSymbol}/waypoints/{waypointSymbol}/shipyard")
	}

	systemSymbol := parts[0]
	waypointSymbol := parts[2]

	if systemSymbol == "" {
		return "", "", fmt.Errorf("system symbol cannot be empty")
	}
	if waypointSymbol == "" {
		return "", "", fmt.Errorf("waypoint symbol cannot be empty")
	}

	// URL decode the symbols in case they contain special characters
	decodedSystem, err := url.QueryUnescape(systemSymbol)
	if err != nil {
		return "", "", fmt.Errorf("invalid system symbol encoding: %w", err)
	}

	decodedWaypoint, err := url.QueryUnescape(waypointSymbol)
	if err != nil {
		return "", "", fmt.Errorf("invalid waypoint symbol encoding: %w", err)
	}

	return decodedSystem, decodedWaypoint, nil
}

// getAvailableShipTypes returns a list of unique ship types available
func (r *ShipyardResource) getAvailableShipTypes(shipyard *spacetraders.Shipyard) []string {
	typeSet := make(map[string]bool)
	var types []string

	// From ship types
	for _, shipType := range shipyard.ShipTypes {
		if !typeSet[shipType.Type] {
			typeSet[shipType.Type] = true
			types = append(types, shipType.Type)
		}
	}

	// From available ships
	for _, ship := range shipyard.Ships {
		if !typeSet[ship.Type] {
			typeSet[ship.Type] = true
			types = append(types, ship.Type)
		}
	}

	return types
}

// getPriceRange returns the min and max prices of available ships
func (r *ShipyardResource) getPriceRange(ships []spacetraders.ShipyardShip) map[string]interface{} {
	if len(ships) == 0 {
		return map[string]interface{}{
			"min": nil,
			"max": nil,
		}
	}

	min := ships[0].PurchasePrice
	max := ships[0].PurchasePrice

	for _, ship := range ships {
		if ship.PurchasePrice < min {
			min = ship.PurchasePrice
		}
		if ship.PurchasePrice > max {
			max = ship.PurchasePrice
		}
	}

	return map[string]interface{}{
		"min": min,
		"max": max,
	}
}

// getShipsByType groups ships by their type
func (r *ShipyardResource) getShipsByType(ships []spacetraders.ShipyardShip) map[string]int {
	counts := make(map[string]int)
	for _, ship := range ships {
		counts[ship.Type]++
	}
	return counts
}

// getShipsBySupply groups ships by their supply level
func (r *ShipyardResource) getShipsBySupply(ships []spacetraders.ShipyardShip) map[string]int {
	counts := make(map[string]int)
	for _, ship := range ships {
		counts[ship.Supply]++
	}
	return counts
}
