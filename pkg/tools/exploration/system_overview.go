package exploration

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/client"
	"spacetraders-mcp/pkg/tools/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

// SystemOverviewTool provides a comprehensive overview of a system
type SystemOverviewTool struct {
	client *client.Client
	logger *logging.Logger
}

// NewSystemOverviewTool creates a new system overview tool
func NewSystemOverviewTool(client *client.Client, logger *logging.Logger) *SystemOverviewTool {
	return &SystemOverviewTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *SystemOverviewTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "system_overview",
		Description: "Get a comprehensive overview of a system including all facilities, waypoints, and strategic opportunities",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"system_symbol": map[string]interface{}{
					"type":        "string",
					"description": "System symbol to analyze (e.g., 'X1-FM66')",
				},
				"include_shipyards": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether to include detailed shipyard information (default: true)",
					"default":     true,
				},
			},
			Required: []string{"system_symbol"},
		},
	}
}

// Handler returns the tool handler function
func (t *SystemOverviewTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contextLogger := t.logger.WithContext(ctx, "system-overview-tool")

		// Extract parameters
		var systemSymbol string
		includeShipyards := true
		if request.Params.Arguments != nil {
			if argsMap, ok := request.Params.Arguments.(map[string]interface{}); ok {
				if val, exists := argsMap["system_symbol"]; exists {
					if s, ok := val.(string); ok {
						systemSymbol = strings.ToUpper(s)
					}
				}
				if val, exists := argsMap["include_shipyards"]; exists {
					if b, ok := val.(bool); ok {
						includeShipyards = b
					}
				}
			}
		}

		if systemSymbol == "" {
			contextLogger.Error("Missing system_symbol parameter")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("Error: system_symbol parameter is required"),
				},
				IsError: true,
			}, nil
		}

		contextLogger.Info(fmt.Sprintf("Generating overview for system %s", systemSymbol))

		// Get waypoints from the system
		waypoints, err := t.client.GetAllSystemWaypoints(systemSymbol)
		if err != nil {
			contextLogger.Error(fmt.Sprintf("Failed to get waypoints for system %s: %v", systemSymbol, err))
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("Failed to retrieve waypoints for system %s: %v", systemSymbol, err)),
				},
				IsError: true,
			}, nil
		}

		// Analyze the system
		analysis := t.analyzeSystem(systemSymbol, waypoints)

		// Get shipyard details if requested
		var shipyardDetails []map[string]interface{}
		if includeShipyards && len(analysis.Shipyards) > 0 {
			for _, shipyardSymbol := range analysis.Shipyards {
				shipyard, err := t.client.GetShipyard(systemSymbol, shipyardSymbol)
				if err != nil {
					contextLogger.Error(fmt.Sprintf("Failed to get shipyard details for %s: %v", shipyardSymbol, err))
					continue
				}

				shipyardInfo := map[string]interface{}{
					"symbol":            shipyard.Symbol,
					"ship_types":        []string{},
					"total_ships":       len(shipyard.Ships),
					"modifications_fee": shipyard.ModificationsFee,
				}

				for _, shipType := range shipyard.ShipTypes {
					shipyardInfo["ship_types"] = append(shipyardInfo["ship_types"].([]string), shipType.Type)
				}

				shipyardDetails = append(shipyardDetails, shipyardInfo)
			}
		}

		contextLogger.ToolCall("system_overview", true)
		contextLogger.Info(fmt.Sprintf("Generated overview for system %s with %d waypoints", systemSymbol, len(waypoints)))

		// Create structured response
		result := map[string]interface{}{
			"system_symbol":     systemSymbol,
			"total_waypoints":   len(waypoints),
			"waypoint_types":    analysis.WaypointTypes,
			"key_facilities":    analysis.KeyFacilities,
			"shipyards":         analysis.Shipyards,
			"marketplaces":      analysis.Marketplaces,
			"mining_sites":      analysis.MiningSites,
			"jump_gates":        analysis.JumpGates,
			"fuel_stations":     analysis.FuelStations,
			"strategic_summary": analysis.StrategicSummary,
		}

		if includeShipyards && len(shipyardDetails) > 0 {
			result["shipyard_details"] = shipyardDetails
		}

		// Create text summary
		textSummary := t.generateTextSummary(analysis, shipyardDetails, includeShipyards)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%s\n```", utils.FormatJSON(result))),
			},
		}, nil
	}
}

// SystemAnalysis holds the analysis results
type SystemAnalysis struct {
	SystemSymbol     string
	WaypointTypes    map[string]int
	KeyFacilities    map[string][]string
	Shipyards        []string
	Marketplaces     []string
	MiningSites      []string
	JumpGates        []string
	FuelStations     []string
	StrategicSummary map[string]interface{}
}

// analyzeSystem performs comprehensive analysis of a system
func (t *SystemOverviewTool) analyzeSystem(systemSymbol string, waypoints []client.SystemWaypoint) *SystemAnalysis {
	analysis := &SystemAnalysis{
		SystemSymbol:     systemSymbol,
		WaypointTypes:    make(map[string]int),
		KeyFacilities:    make(map[string][]string),
		Shipyards:        []string{},
		Marketplaces:     []string{},
		MiningSites:      []string{},
		JumpGates:        []string{},
		FuelStations:     []string{},
		StrategicSummary: make(map[string]interface{}),
	}

	traitCounts := make(map[string]int)

	for _, waypoint := range waypoints {
		// Count waypoint types
		analysis.WaypointTypes[waypoint.Type]++

		// Analyze traits
		for _, trait := range waypoint.Traits {
			traitCounts[trait.Symbol]++

			// Categorize key facilities
			switch trait.Symbol {
			case "SHIPYARD":
				analysis.Shipyards = append(analysis.Shipyards, waypoint.Symbol)
			case "MARKETPLACE":
				analysis.Marketplaces = append(analysis.Marketplaces, waypoint.Symbol)
			case "ASTEROID_FIELD", "MINERAL_DEPOSITS", "RARE_METAL_DEPOSITS", "ICE_CRYSTALS":
				analysis.MiningSites = append(analysis.MiningSites, waypoint.Symbol)
			case "JUMP_GATE":
				analysis.JumpGates = append(analysis.JumpGates, waypoint.Symbol)
			case "FUEL_STATION":
				analysis.FuelStations = append(analysis.FuelStations, waypoint.Symbol)
			}

			// Group facilities by trait
			if analysis.KeyFacilities[trait.Symbol] == nil {
				analysis.KeyFacilities[trait.Symbol] = []string{}
			}
			analysis.KeyFacilities[trait.Symbol] = append(analysis.KeyFacilities[trait.Symbol], waypoint.Symbol)
		}
	}

	// Generate strategic summary
	analysis.StrategicSummary["total_waypoints"] = len(waypoints)
	analysis.StrategicSummary["trade_opportunities"] = len(analysis.Marketplaces)
	analysis.StrategicSummary["ship_acquisition"] = len(analysis.Shipyards)
	analysis.StrategicSummary["mining_potential"] = len(analysis.MiningSites)
	analysis.StrategicSummary["system_connectivity"] = len(analysis.JumpGates)

	// Determine system classification
	classification := "Unknown"
	if len(analysis.Shipyards) > 0 && len(analysis.Marketplaces) > 0 {
		classification = "Commercial Hub"
	} else if len(analysis.MiningSites) > 3 {
		classification = "Mining System"
	} else if len(analysis.Marketplaces) > 2 {
		classification = "Trading System"
	} else if len(analysis.Shipyards) > 0 {
		classification = "Industrial System"
	} else if len(analysis.JumpGates) > 0 {
		classification = "Transit System"
	}
	analysis.StrategicSummary["classification"] = classification

	return analysis
}

// generateTextSummary creates a human-readable summary
func (t *SystemOverviewTool) generateTextSummary(analysis *SystemAnalysis, shipyardDetails []map[string]interface{}, includeShipyards bool) string {
	summary := fmt.Sprintf("# 🌌 System Overview: %s\n\n", analysis.SystemSymbol)

	// Classification and basic stats
	classification := analysis.StrategicSummary["classification"].(string)
	summary += fmt.Sprintf("**Classification:** %s\n", classification)
	summary += fmt.Sprintf("**Total Waypoints:** %d\n\n", len(analysis.WaypointTypes))

	// Waypoint types breakdown
	summary += "## 🏗️ Waypoint Types\n"
	for waypointType, count := range analysis.WaypointTypes {
		icon := "•"
		switch waypointType {
		case "PLANET":
			icon = "🪐"
		case "MOON":
			icon = "🌙"
		case "ASTEROID":
			icon = "🪨"
		case "DEBRIS_FIELD":
			icon = "💫"
		case "JUMP_GATE":
			icon = "🚪"
		}
		summary += fmt.Sprintf("%s **%s:** %d\n", icon, waypointType, count)
	}
	summary += "\n"

	// Key facilities
	summary += "## 🏪 Key Facilities\n\n"

	if len(analysis.Shipyards) > 0 {
		summary += fmt.Sprintf("### 🚀 Shipyards (%d)\n", len(analysis.Shipyards))
		for _, shipyard := range analysis.Shipyards {
			summary += fmt.Sprintf("- **%s**\n", shipyard)
		}
		summary += "\n"
	}

	if len(analysis.Marketplaces) > 0 {
		summary += fmt.Sprintf("### 🏪 Marketplaces (%d)\n", len(analysis.Marketplaces))
		for _, marketplace := range analysis.Marketplaces {
			summary += fmt.Sprintf("- **%s**\n", marketplace)
		}
		summary += "\n"
	}

	if len(analysis.MiningSites) > 0 {
		summary += fmt.Sprintf("### ⛏️ Mining Sites (%d)\n", len(analysis.MiningSites))
		for _, miningSite := range analysis.MiningSites {
			summary += fmt.Sprintf("- **%s**\n", miningSite)
		}
		summary += "\n"
	}

	if len(analysis.JumpGates) > 0 {
		summary += fmt.Sprintf("### 🚪 Jump Gates (%d)\n", len(analysis.JumpGates))
		for _, jumpGate := range analysis.JumpGates {
			summary += fmt.Sprintf("- **%s**\n", jumpGate)
		}
		summary += "\n"
	}

	// Shipyard details
	if includeShipyards && len(shipyardDetails) > 0 {
		summary += "## 🛠️ Shipyard Details\n\n"
		for _, shipyard := range shipyardDetails {
			summary += fmt.Sprintf("### %s\n", shipyard["symbol"])
			summary += fmt.Sprintf("**Available Ship Types:** %d\n", len(shipyard["ship_types"].([]string)))

			shipTypes := shipyard["ship_types"].([]string)
			sort.Strings(shipTypes)
			for _, shipType := range shipTypes {
				summary += fmt.Sprintf("- %s\n", shipType)
			}

			summary += fmt.Sprintf("**Total Ships Available:** %d\n", shipyard["total_ships"])
			summary += fmt.Sprintf("**Modification Fee:** %d credits\n\n", shipyard["modifications_fee"])
		}
	}

	// Strategic recommendations
	summary += "## 📋 Strategic Recommendations\n\n"

	if len(analysis.Shipyards) > 0 {
		summary += "✅ **Ship Acquisition:** This system has shipyards for expanding your fleet\n"
	}

	if len(analysis.Marketplaces) > 1 {
		summary += "✅ **Trading Opportunities:** Multiple marketplaces enable profitable trade routes\n"
	}

	if len(analysis.MiningSites) > 0 {
		summary += "✅ **Resource Extraction:** Mining opportunities available for resource gathering\n"
	}

	if len(analysis.JumpGates) > 0 {
		summary += "✅ **System Connectivity:** Jump gates provide access to other systems\n"
	}

	// Next steps
	summary += "\n## 🚀 Next Steps\n\n"
	summary += "**To navigate to facilities:**\n"
	summary += "- Use `navigate_ship` tool with ship symbol and waypoint symbol\n"
	summary += "- Ensure ship is in orbit before navigating\n\n"

	if len(analysis.Shipyards) > 0 {
		summary += "**To check available ships:**\n"
		for _, shipyard := range analysis.Shipyards {
			summary += fmt.Sprintf("- `spacetraders://systems/%s/waypoints/%s/shipyard`\n", analysis.SystemSymbol, shipyard)
		}
		summary += "\n"
	}

	if len(analysis.Marketplaces) > 0 {
		summary += "**To check market prices:**\n"
		for _, marketplace := range analysis.Marketplaces {
			summary += fmt.Sprintf("- `spacetraders://systems/%s/waypoints/%s/market`\n", analysis.SystemSymbol, marketplace)
		}
	}

	return summary
}
