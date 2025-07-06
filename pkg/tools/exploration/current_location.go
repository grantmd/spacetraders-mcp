package exploration

import (
	"context"
	"fmt"
	"strings"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"
	"spacetraders-mcp/pkg/tools/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

// CurrentLocationTool analyzes where the player's ships are currently located
type CurrentLocationTool struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewCurrentLocationTool creates a new current location analysis tool
func NewCurrentLocationTool(client *spacetraders.Client, logger *logging.Logger) *CurrentLocationTool {
	return &CurrentLocationTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *CurrentLocationTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "current_location",
		Description: "Analyze where your ships are currently located and what facilities are nearby",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"include_nearby": map[string]interface{}{
					"type":        "boolean",
					"description": "Include nearby waypoints and facilities in the same system (default: true)",
					"default":     true,
				},
				"ship_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Optional: Analyze specific ship only (otherwise analyzes all ships)",
				},
			},
		},
	}
}

// Handler returns the tool handler function
func (t *CurrentLocationTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contextLogger := t.logger.WithContext(ctx, "current-location-tool")

		// Extract parameters
		includeNearby := true
		var specificShip string
		if request.Params.Arguments != nil {
			if argsMap, ok := request.Params.Arguments.(map[string]interface{}); ok {
				if val, exists := argsMap["include_nearby"]; exists {
					if b, ok := val.(bool); ok {
						includeNearby = b
					}
				}
				if val, exists := argsMap["ship_symbol"]; exists {
					if s, ok := val.(string); ok && s != "" {
						specificShip = strings.ToUpper(s)
					}
				}
			}
		}

		contextLogger.Info("Analyzing current ship locations")

		// Get all ships
		ships, err := t.client.GetAllShips()
		if err != nil {
			contextLogger.Error(fmt.Sprintf("Failed to get ships: %v", err))
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("Failed to retrieve ships: %v", err)),
				},
				IsError: true,
			}, nil
		}

		// Filter ships if specific ship requested
		var shipsToAnalyze []spacetraders.Ship
		if specificShip != "" {
			for _, ship := range ships {
				if ship.Symbol == specificShip {
					shipsToAnalyze = append(shipsToAnalyze, ship)
					break
				}
			}
			if len(shipsToAnalyze) == 0 {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						mcp.NewTextContent(fmt.Sprintf("Ship '%s' not found", specificShip)),
					},
					IsError: true,
				}, nil
			}
		} else {
			shipsToAnalyze = ships
		}

		// Analyze locations
		locationAnalysis := t.analyzeShipLocations(shipsToAnalyze, includeNearby)

		contextLogger.ToolCall("current_location", true)
		contextLogger.Info(fmt.Sprintf("Analyzed %d ships across %d systems", len(shipsToAnalyze), len(locationAnalysis.SystemSummary)))

		// Create structured response
		result := map[string]interface{}{
			"total_ships":      len(shipsToAnalyze),
			"systems_occupied": len(locationAnalysis.SystemSummary),
			"ship_locations":   locationAnalysis.ShipLocations,
			"system_summary":   locationAnalysis.SystemSummary,
			"status_summary":   locationAnalysis.StatusSummary,
			"recommendations":  locationAnalysis.Recommendations,
		}

		if includeNearby {
			result["nearby_facilities"] = locationAnalysis.NearbyFacilities
		}

		// Create text summary
		textSummary := t.generateLocationSummary(locationAnalysis, includeNearby, specificShip)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%s\n```", utils.FormatJSON(result))),
			},
		}, nil
	}
}

// LocationAnalysis holds the analysis results
type LocationAnalysis struct {
	ShipLocations    []map[string]interface{}
	SystemSummary    map[string]map[string]interface{}
	StatusSummary    map[string]int
	NearbyFacilities map[string][]map[string]interface{}
	Recommendations  []string
}

// analyzeShipLocations performs comprehensive analysis of ship locations
func (t *CurrentLocationTool) analyzeShipLocations(ships []spacetraders.Ship, includeNearby bool) *LocationAnalysis {
	analysis := &LocationAnalysis{
		ShipLocations:    []map[string]interface{}{},
		SystemSummary:    make(map[string]map[string]interface{}),
		StatusSummary:    make(map[string]int),
		NearbyFacilities: make(map[string][]map[string]interface{}),
		Recommendations:  []string{},
	}

	systemsToCheck := make(map[string]bool)

	// Analyze each ship
	for _, ship := range ships {
		shipLocation := map[string]interface{}{
			"symbol":         ship.Symbol,
			"role":           ship.Registration.Role,
			"system":         ship.Nav.SystemSymbol,
			"waypoint":       ship.Nav.WaypointSymbol,
			"status":         ship.Nav.Status,
			"flight_mode":    ship.Nav.FlightMode,
			"fuel_current":   ship.Fuel.Current,
			"fuel_capacity":  ship.Fuel.Capacity,
			"cargo_units":    ship.Cargo.Units,
			"cargo_capacity": ship.Cargo.Capacity,
		}

		// Add route information if in transit
		if ship.Nav.Status == "IN_TRANSIT" && ship.Nav.Route.Destination.Symbol != "" {
			shipLocation["destination"] = ship.Nav.Route.Destination.Symbol
			shipLocation["arrival_time"] = ship.Nav.Route.Arrival
		}

		analysis.ShipLocations = append(analysis.ShipLocations, shipLocation)

		// Update system summary
		system := ship.Nav.SystemSymbol
		if analysis.SystemSummary[system] == nil {
			analysis.SystemSummary[system] = map[string]interface{}{
				"ship_count": 0,
				"ships":      []string{},
				"statuses":   make(map[string]int),
			}
		}

		analysis.SystemSummary[system]["ship_count"] = analysis.SystemSummary[system]["ship_count"].(int) + 1
		analysis.SystemSummary[system]["ships"] = append(analysis.SystemSummary[system]["ships"].([]string), ship.Symbol)

		systemStatuses := analysis.SystemSummary[system]["statuses"].(map[string]int)
		systemStatuses[ship.Nav.Status]++

		// Update overall status summary
		analysis.StatusSummary[ship.Nav.Status]++

		// Mark system for nearby facility checking
		if includeNearby {
			systemsToCheck[system] = true
		}
	}

	// Get nearby facilities for each system
	if includeNearby {
		for system := range systemsToCheck {
			facilities := t.getNearbyFacilities(system)
			if len(facilities) > 0 {
				analysis.NearbyFacilities[system] = facilities
			}
		}
	}

	// Generate recommendations
	analysis.Recommendations = t.generateRecommendations(ships, analysis.SystemSummary, analysis.NearbyFacilities)

	return analysis
}

// getNearbyFacilities gets key facilities in a system
func (t *CurrentLocationTool) getNearbyFacilities(systemSymbol string) []map[string]interface{} {
	waypoints, err := t.client.GetAllSystemWaypoints(systemSymbol)
	if err != nil {
		return []map[string]interface{}{}
	}

	var facilities []map[string]interface{}
	importantTraits := map[string]string{
		"SHIPYARD":       "🚀 Shipyard",
		"MARKETPLACE":    "🏪 Marketplace",
		"JUMP_GATE":      "🚪 Jump Gate",
		"FUEL_STATION":   "⛽ Fuel Station",
		"ASTEROID_FIELD": "⛏️ Mining Site",
	}

	for _, waypoint := range waypoints {
		for _, trait := range waypoint.Traits {
			if name, isImportant := importantTraits[trait.Symbol]; isImportant {
				facilities = append(facilities, map[string]interface{}{
					"waypoint":    waypoint.Symbol,
					"type":        waypoint.Type,
					"trait":       trait.Symbol,
					"trait_name":  name,
					"description": trait.Description,
					"coordinates": fmt.Sprintf("(%d, %d)", waypoint.X, waypoint.Y),
				})
			}
		}
	}

	return facilities
}

// generateRecommendations creates actionable recommendations
func (t *CurrentLocationTool) generateRecommendations(ships []spacetraders.Ship, systemSummary map[string]map[string]interface{}, nearbyFacilities map[string][]map[string]interface{}) []string {
	var recommendations []string

	// Check for ships needing fuel
	for _, ship := range ships {
		fuelPercent := float64(ship.Fuel.Current) / float64(ship.Fuel.Capacity) * 100
		if fuelPercent < 25 {
			recommendations = append(recommendations, fmt.Sprintf("⛽ Ship %s is low on fuel (%.0f%%) - find a fuel station", ship.Symbol, fuelPercent))
		}
	}

	// Check for ships at capacity
	for _, ship := range ships {
		if ship.Cargo.Units == ship.Cargo.Capacity && ship.Cargo.Capacity > 0 {
			recommendations = append(recommendations, fmt.Sprintf("📦 Ship %s cargo is full - consider selling goods at a marketplace", ship.Symbol))
		}
	}

	// Check for docked ships that could be exploring
	dockedCount := 0
	for _, ship := range ships {
		if ship.Nav.Status == "DOCKED" {
			dockedCount++
		}
	}
	if dockedCount > 1 {
		recommendations = append(recommendations, fmt.Sprintf("🚀 You have %d docked ships - consider putting some in orbit for exploration", dockedCount))
	}

	// System-specific recommendations
	for system, facilities := range nearbyFacilities {
		// Look for shipyards
		for _, facility := range facilities {
			if facility["trait"] == "SHIPYARD" {
				recommendations = append(recommendations, fmt.Sprintf("🛒 System %s has a shipyard at %s - consider expanding your fleet", system, facility["waypoint"]))
				break
			}
		}

		// Look for trading opportunities
		marketCount := 0
		for _, facility := range facilities {
			if facility["trait"] == "MARKETPLACE" {
				marketCount++
			}
		}
		if marketCount > 1 {
			recommendations = append(recommendations, fmt.Sprintf("💰 System %s has %d marketplaces - good for trade routes", system, marketCount))
		}
	}

	return recommendations
}

// generateLocationSummary creates a human-readable summary
func (t *CurrentLocationTool) generateLocationSummary(analysis *LocationAnalysis, includeNearby bool, specificShip string) string {
	var summary string

	if specificShip != "" {
		summary = fmt.Sprintf("# 📍 Location Analysis: %s\n\n", specificShip)
	} else {
		summary = "# 📍 Fleet Location Analysis\n\n"
		summary += fmt.Sprintf("**Total Ships:** %d\n", len(analysis.ShipLocations))
		summary += fmt.Sprintf("**Systems Occupied:** %d\n\n", len(analysis.SystemSummary))
	}

	// Ship status summary
	if len(analysis.StatusSummary) > 0 {
		summary += "## 🚢 Ship Status Overview\n"
		for status, count := range analysis.StatusSummary {
			icon := "•"
			switch status {
			case "DOCKED":
				icon = "⚓"
			case "IN_ORBIT":
				icon = "🌌"
			case "IN_TRANSIT":
				icon = "🚀"
			}
			summary += fmt.Sprintf("%s **%s:** %d ships\n", icon, status, count)
		}
		summary += "\n"
	}

	// System breakdown
	summary += "## 🌌 System Breakdown\n\n"
	for system, systemInfo := range analysis.SystemSummary {
		shipCount := systemInfo["ship_count"].(int)
		ships := systemInfo["ships"].([]string)

		summary += fmt.Sprintf("### %s (%d ships)\n", system, shipCount)
		summary += "**Ships:** " + strings.Join(ships, ", ") + "\n"

		systemStatuses := systemInfo["statuses"].(map[string]int)
		summary += "**Status:** "
		statusParts := []string{}
		for status, count := range systemStatuses {
			statusParts = append(statusParts, fmt.Sprintf("%s: %d", status, count))
		}
		summary += strings.Join(statusParts, ", ") + "\n\n"
	}

	// Individual ship details
	summary += "## 🚀 Individual Ship Details\n\n"
	for _, shipLoc := range analysis.ShipLocations {
		summary += fmt.Sprintf("### %s (%s)\n", shipLoc["symbol"], shipLoc["role"])
		summary += fmt.Sprintf("**Location:** %s → %s\n", shipLoc["system"], shipLoc["waypoint"])
		summary += fmt.Sprintf("**Status:** %s (%s mode)\n", shipLoc["status"], shipLoc["flight_mode"])

		if shipLoc["destination"] != nil {
			summary += fmt.Sprintf("**Destination:** %s (ETA: %s)\n", shipLoc["destination"], shipLoc["arrival_time"])
		}

		fuelPercent := float64(shipLoc["fuel_current"].(int)) / float64(shipLoc["fuel_capacity"].(int)) * 100
		summary += fmt.Sprintf("**Fuel:** %d/%d (%.1f%%)\n",
			shipLoc["fuel_current"], shipLoc["fuel_capacity"], fuelPercent)

		if shipLoc["cargo_capacity"].(int) > 0 {
			cargoPercent := float64(shipLoc["cargo_units"].(int)) / float64(shipLoc["cargo_capacity"].(int)) * 100
			summary += fmt.Sprintf("**Cargo:** %d/%d (%.1f%%)\n",
				shipLoc["cargo_units"], shipLoc["cargo_capacity"], cargoPercent)
		}
		summary += "\n"
	}

	// Nearby facilities
	if includeNearby && len(analysis.NearbyFacilities) > 0 {
		summary += "## 🏪 Nearby Facilities\n\n"
		for system, facilities := range analysis.NearbyFacilities {
			if len(facilities) > 0 {
				summary += fmt.Sprintf("### %s\n", system)
				for _, facility := range facilities {
					summary += fmt.Sprintf("- %s at **%s** %s\n",
						facility["trait_name"], facility["waypoint"], facility["coordinates"])
				}
				summary += "\n"
			}
		}
	}

	// Recommendations
	if len(analysis.Recommendations) > 0 {
		summary += "## 💡 Recommendations\n\n"
		for _, rec := range analysis.Recommendations {
			summary += fmt.Sprintf("- %s\n", rec)
		}
		summary += "\n"
	}

	// Next steps
	summary += "## 🎯 Next Steps\n\n"
	summary += "**Navigation Commands:**\n"
	summary += "- `orbit_ship` - Put a docked ship into orbit\n"
	summary += "- `navigate_ship` - Move ship to another waypoint\n"
	summary += "- `dock_ship` - Dock an orbiting ship\n\n"

	summary += "**Exploration Commands:**\n"
	summary += "- `find_waypoints` - Search for specific facilities\n"
	summary += "- `system_overview` - Get full system analysis\n"

	return summary
}
