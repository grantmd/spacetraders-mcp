package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"spacetraders-mcp/pkg/client"
	"spacetraders-mcp/pkg/logging"

	"github.com/mark3labs/mcp-go/mcp"
)

// ShipResource handles individual ship information resources
type ShipResource struct {
	client *client.Client
	logger *logging.Logger
}

// NewShipResource creates a new ship resource handler
func NewShipResource(client *client.Client, logger *logging.Logger) *ShipResource {
	return &ShipResource{
		client: client,
		logger: logger,
	}
}

// Resource returns the MCP resource definition
func (r *ShipResource) Resource() mcp.Resource {
	return mcp.Resource{
		URI:         "spacetraders://ships/{shipSymbol}",
		Name:        "Individual Ship Details",
		Description: "Detailed information about a specific ship including status, location, cargo, cooldown, and all components",
		MIMEType:    "application/json",
	}
}

// Handler returns the resource handler function
func (r *ShipResource) Handler() func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// Extract ship symbol from URI
		shipSymbol := r.extractShipSymbol(request.Params.URI)
		if shipSymbol == "" {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Invalid ship resource URI. Expected format: spacetraders://ships/{shipSymbol}",
				},
			}, nil
		}

		// Set up context logger
		ctxLogger := r.logger.WithContext(ctx, "ship-resource")
		ctxLogger.Debug("Fetching ship details for %s", shipSymbol)

		// Get ship information from the API
		start := time.Now()
		ship, err := r.client.GetShip(shipSymbol)
		duration := time.Since(start)

		if err != nil {
			ctxLogger.Error("Failed to fetch ship %s: %v", shipSymbol, err)
			ctxLogger.APICall(fmt.Sprintf("/my/ships/%s", shipSymbol), 0, duration.String())
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     fmt.Sprintf("Error fetching ship %s: %s", shipSymbol, err.Error()),
				},
			}, nil
		}

		ctxLogger.APICall(fmt.Sprintf("/my/ships/%s", shipSymbol), 200, duration.String())
		ctxLogger.Info("Successfully retrieved ship %s", shipSymbol)

		// Get detailed cooldown information
		cooldown, cooldownErr := r.client.GetShipCooldown(shipSymbol)
		if cooldownErr != nil {
			ctxLogger.Debug("Could not get detailed cooldown for %s: %v", shipSymbol, cooldownErr)
			// Don't fail the entire request, just use the cooldown from ship data
		}

		// Create enhanced ship data with additional analysis
		result := r.createEnhancedShipData(ship, cooldown)

		// Convert to JSON for response
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			ctxLogger.Error("Failed to marshal ship data to JSON: %v", err)
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Error formatting ship information",
				},
			}, nil
		}

		ctxLogger.ResourceRead(request.Params.URI, true)
		ctxLogger.Debug("Ship resource response size: %d bytes", len(jsonData))

		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		}, nil
	}
}

// extractShipSymbol extracts the ship symbol from the URI
func (r *ShipResource) extractShipSymbol(uri string) string {
	// Match pattern: spacetraders://ships/{shipSymbol}
	re := regexp.MustCompile(`^spacetraders://ships/([A-Za-z0-9_-]+)$`)
	matches := re.FindStringSubmatch(uri)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}

// createEnhancedShipData creates enhanced ship data with additional analysis
func (r *ShipResource) createEnhancedShipData(ship *client.Ship, detailedCooldown *client.Cooldown) map[string]interface{} {
	// Use detailed cooldown if available, otherwise use ship's cooldown
	activeCooldown := &ship.Cooldown
	if detailedCooldown != nil {
		activeCooldown = detailedCooldown
	}

	// Calculate cooldown status
	cooldownStatus := r.analyzeCooldown(activeCooldown)

	// Calculate cargo utilization
	cargoUtilization := r.analyzeCargoUtilization(ship.Cargo)

	// Analyze ship status
	shipStatus := r.analyzeShipStatus(ship, cooldownStatus)

	// Analyze location and travel capabilities
	locationAnalysis := r.analyzeLocation(ship.Nav)

	// Create the enhanced result
	result := map[string]interface{}{
		"ship": map[string]interface{}{
			"symbol":       ship.Symbol,
			"registration": ship.Registration,
			"nav":          ship.Nav,
			"crew":         ship.Crew,
			"frame":        ship.Frame,
			"reactor":      ship.Reactor,
			"engine":       ship.Engine,
			"cooldown":     activeCooldown,
			"modules":      ship.Modules,
			"mounts":       ship.Mounts,
			"cargo":        ship.Cargo,
			"fuel":         ship.Fuel,
		},
		"analysis": map[string]interface{}{
			"status":             shipStatus,
			"cooldown_status":    cooldownStatus,
			"cargo_utilization":  cargoUtilization,
			"location_analysis":  locationAnalysis,
			"operational_status": r.determineOperationalStatus(ship, cooldownStatus),
		},
		"capabilities":    r.analyzeCapabilities(ship),
		"recommendations": r.generateRecommendations(ship, cooldownStatus, cargoUtilization),
		"meta": map[string]interface{}{
			"last_updated": time.Now().Format(time.RFC3339),
			"ship_symbol":  ship.Symbol,
		},
	}

	return result
}

// analyzeCooldown analyzes the cooldown status
func (r *ShipResource) analyzeCooldown(cooldown *client.Cooldown) map[string]interface{} {
	if cooldown == nil || cooldown.RemainingSeconds <= 0 {
		return map[string]interface{}{
			"active":            false,
			"remaining_seconds": 0,
			"status":            "ready",
			"message":           "Ship is ready for actions",
		}
	}

	// Calculate time remaining
	remainingMinutes := cooldown.RemainingSeconds / 60
	remainingSecondsDisplay := cooldown.RemainingSeconds % 60

	var timeDisplay string
	if remainingMinutes > 0 {
		timeDisplay = fmt.Sprintf("%dm %ds", remainingMinutes, remainingSecondsDisplay)
	} else {
		timeDisplay = fmt.Sprintf("%ds", remainingSecondsDisplay)
	}

	var statusMessage string
	if cooldown.RemainingSeconds > 300 { // 5 minutes
		statusMessage = "Long cooldown active - consider using other ships"
	} else if cooldown.RemainingSeconds > 60 { // 1 minute
		statusMessage = "Moderate cooldown active"
	} else {
		statusMessage = "Short cooldown active - almost ready"
	}

	return map[string]interface{}{
		"active":            true,
		"remaining_seconds": cooldown.RemainingSeconds,
		"total_seconds":     cooldown.TotalSeconds,
		"expiration":        cooldown.Expiration,
		"time_display":      timeDisplay,
		"status":            "cooling_down",
		"message":           statusMessage,
	}
}

// analyzeCargoUtilization analyzes cargo capacity and utilization
func (r *ShipResource) analyzeCargoUtilization(cargo client.Cargo) map[string]interface{} {
	utilizationPercent := float64(cargo.Units) / float64(cargo.Capacity) * 100

	var status string
	var message string

	if utilizationPercent == 0 {
		status = "empty"
		message = "Cargo hold is empty"
	} else if utilizationPercent < 25 {
		status = "low"
		message = "Plenty of cargo space available"
	} else if utilizationPercent < 75 {
		status = "moderate"
		message = "Good cargo capacity remaining"
	} else if utilizationPercent < 100 {
		status = "high"
		message = "Cargo hold is nearly full"
	} else {
		status = "full"
		message = "Cargo hold is at maximum capacity"
	}

	// Analyze cargo composition
	cargoComposition := make(map[string]interface{})
	for _, item := range cargo.Inventory {
		cargoComposition[item.Symbol] = map[string]interface{}{
			"units":       item.Units,
			"name":        item.Name,
			"description": item.Description,
		}
	}

	return map[string]interface{}{
		"capacity":            cargo.Capacity,
		"units":               cargo.Units,
		"available_space":     cargo.Capacity - cargo.Units,
		"utilization_percent": utilizationPercent,
		"status":              status,
		"message":             message,
		"composition":         cargoComposition,
		"item_count":          len(cargo.Inventory),
	}
}

// analyzeShipStatus determines overall ship status
func (r *ShipResource) analyzeShipStatus(ship *client.Ship, cooldownStatus map[string]interface{}) map[string]interface{} {
	status := "operational"
	message := "Ship is operational"

	// Check various conditions
	conditions := []string{}

	if cooldownStatus["active"].(bool) {
		conditions = append(conditions, "cooling_down")
	}

	if ship.Nav.Status == "DOCKED" {
		conditions = append(conditions, "docked")
	} else if ship.Nav.Status == "IN_ORBIT" {
		conditions = append(conditions, "in_orbit")
	} else if ship.Nav.Status == "IN_TRANSIT" {
		conditions = append(conditions, "in_transit")
		status = "traveling"
		message = "Ship is currently traveling"
	}

	// Check fuel levels
	if ship.Fuel.Current < ship.Fuel.Capacity/4 {
		conditions = append(conditions, "low_fuel")
	}

	// Check if cargo is full
	if ship.Cargo.Units >= ship.Cargo.Capacity {
		conditions = append(conditions, "cargo_full")
	}

	return map[string]interface{}{
		"status":     status,
		"message":    message,
		"conditions": conditions,
		"nav_status": ship.Nav.Status,
		"location":   ship.Nav.WaypointSymbol,
		"system":     ship.Nav.SystemSymbol,
	}
}

// analyzeLocation analyzes ship location and travel capabilities
func (r *ShipResource) analyzeLocation(nav client.Navigation) map[string]interface{} {
	location := map[string]interface{}{
		"system":   nav.SystemSymbol,
		"waypoint": nav.WaypointSymbol,
		"status":   nav.Status,
	}

	if nav.Route.Origin.Symbol != "" {
		location["route"] = map[string]interface{}{
			"origin":      nav.Route.Origin,
			"destination": nav.Route.Destination,
			"departure":   nav.Route.DepartureTime,
			"arrival":     nav.Route.Arrival,
		}
	}

	return location
}

// determineOperationalStatus determines if ship can perform various actions
func (r *ShipResource) determineOperationalStatus(ship *client.Ship, cooldownStatus map[string]interface{}) map[string]interface{} {
	canAct := !cooldownStatus["active"].(bool)
	canDock := ship.Nav.Status == "IN_ORBIT"
	canOrbit := ship.Nav.Status == "DOCKED"
	canNavigate := ship.Nav.Status == "IN_ORBIT" && canAct
	canExtract := ship.Nav.Status == "IN_ORBIT" && canAct && r.hasExtractionCapability(ship)
	canTrade := ship.Nav.Status == "DOCKED"
	canRefuel := ship.Nav.Status == "DOCKED" && ship.Fuel.Current < ship.Fuel.Capacity

	return map[string]interface{}{
		"can_act":      canAct,
		"can_dock":     canDock,
		"can_orbit":    canOrbit,
		"can_navigate": canNavigate,
		"can_extract":  canExtract,
		"can_trade":    canTrade,
		"can_refuel":   canRefuel,
	}
}

// hasExtractionCapability checks if ship has mining/extraction mounts
func (r *ShipResource) hasExtractionCapability(ship *client.Ship) bool {
	for _, mount := range ship.Mounts {
		if strings.Contains(strings.ToUpper(mount.Symbol), "MINING") ||
			strings.Contains(strings.ToUpper(mount.Symbol), "LASER") ||
			strings.Contains(strings.ToUpper(mount.Symbol), "SIPHON") {
			return true
		}
	}
	return false
}

// analyzeCapabilities analyzes ship capabilities based on mounts and modules
func (r *ShipResource) analyzeCapabilities(ship *client.Ship) map[string]interface{} {
	capabilities := map[string]interface{}{
		"mining":    false,
		"scanning":  false,
		"trading":   true, // All ships can trade
		"combat":    false,
		"surveying": false,
	}

	mountCapabilities := []string{}
	for _, mount := range ship.Mounts {
		mountType := strings.ToUpper(mount.Symbol)
		if strings.Contains(mountType, "MINING") || strings.Contains(mountType, "LASER") {
			capabilities["mining"] = true
			mountCapabilities = append(mountCapabilities, "mining")
		}
		if strings.Contains(mountType, "SENSOR") {
			capabilities["scanning"] = true
			mountCapabilities = append(mountCapabilities, "scanning")
		}
		if strings.Contains(mountType, "SURVEYOR") {
			capabilities["surveying"] = true
			mountCapabilities = append(mountCapabilities, "surveying")
		}
		if strings.Contains(mountType, "WEAPON") {
			capabilities["combat"] = true
			mountCapabilities = append(mountCapabilities, "combat")
		}
	}

	return map[string]interface{}{
		"primary_capabilities": capabilities,
		"mount_capabilities":   mountCapabilities,
		"cargo_capacity":       ship.Cargo.Capacity,
		"fuel_capacity":        ship.Fuel.Capacity,
		"crew_capacity":        ship.Crew.Capacity,
	}
}

// generateRecommendations generates actionable recommendations
func (r *ShipResource) generateRecommendations(ship *client.Ship, cooldownStatus, cargoUtilization map[string]interface{}) []string {
	recommendations := []string{}

	// Cooldown recommendations
	if cooldownStatus["active"].(bool) {
		remaining := cooldownStatus["remaining_seconds"].(int)
		if remaining > 300 {
			recommendations = append(recommendations, "Ship has long cooldown - consider using other ships for immediate actions")
		} else if remaining > 60 {
			recommendations = append(recommendations, "Ship cooling down - plan next action in advance")
		}
	}

	// Cargo recommendations
	utilizationPercent := cargoUtilization["utilization_percent"].(float64)
	if utilizationPercent > 90 {
		recommendations = append(recommendations, "Cargo hold nearly full - consider selling goods or transferring to another ship")
	} else if utilizationPercent == 0 && ship.Nav.Status == "DOCKED" {
		recommendations = append(recommendations, "Cargo hold empty - good opportunity to purchase goods")
	}

	// Fuel recommendations
	fuelPercent := float64(ship.Fuel.Current) / float64(ship.Fuel.Capacity) * 100
	if fuelPercent < 25 {
		recommendations = append(recommendations, "Fuel level low - refuel before long-distance travel")
	}

	// Location-based recommendations
	if ship.Nav.Status == "DOCKED" {
		recommendations = append(recommendations, "Ship is docked - can access market, shipyard, and refueling")
	} else if ship.Nav.Status == "IN_ORBIT" {
		if r.hasExtractionCapability(ship) && !cooldownStatus["active"].(bool) {
			recommendations = append(recommendations, "Ship in orbit with extraction capability - ready for mining operations")
		}
	}

	// No recommendations case
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Ship is in good operational condition")
	}

	return recommendations
}
