package navigation

import (
	"context"
	"fmt"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"

	"github.com/mark3labs/mcp-go/mcp"
)

// DockShipTool handles docking ships at waypoints
type DockShipTool struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewDockShipTool creates a new dock ship tool
func NewDockShipTool(client *spacetraders.Client, logger *logging.Logger) *DockShipTool {
	return &DockShipTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *DockShipTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "dock_ship",
		Description: "Dock a ship at its current waypoint. Ship must be in orbit to use this command.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"ship_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the ship to dock (e.g., 'SHIP_1234')",
				},
			},
			Required: []string{"ship_symbol"},
		},
	}
}

// Handler returns the tool handler function
func (t *DockShipTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contextLogger := t.logger.WithContext(ctx, "dock-ship-tool")

		// Extract ship symbol
		var shipSymbol string
		if request.Params.Arguments != nil {
			if argsMap, ok := request.Params.Arguments.(map[string]interface{}); ok {
				if val, exists := argsMap["ship_symbol"]; exists {
					if s, ok := val.(string); ok && s != "" {
						shipSymbol = s
					}
				}
			}
		}

		if shipSymbol == "" {
			contextLogger.Error("Missing or invalid ship_symbol parameter")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("Error: ship_symbol parameter is required and must be a non-empty string"),
				},
				IsError: true,
			}, nil
		}

		contextLogger.Info(fmt.Sprintf("Attempting to dock ship: %s", shipSymbol))

		// Dock the ship
		nav, err := t.client.DockShip(shipSymbol)
		if err != nil {
			contextLogger.Error(fmt.Sprintf("Failed to dock ship %s: %v", shipSymbol, err))
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("Failed to dock ship %s: %v", shipSymbol, err)),
				},
				IsError: true,
			}, nil
		}

		contextLogger.ToolCall("dock_ship", true)
		contextLogger.Info(fmt.Sprintf("Successfully docked ship %s at %s", shipSymbol, nav.WaypointSymbol))

		// Create structured response
		result := map[string]interface{}{
			"success":     true,
			"ship_symbol": shipSymbol,
			"navigation": map[string]interface{}{
				"system_symbol":   nav.SystemSymbol,
				"waypoint_symbol": nav.WaypointSymbol,
				"status":          nav.Status,
				"flight_mode":     nav.FlightMode,
			},
		}

		// Add route information if available
		if nav.Route.Destination.Symbol != "" {
			result["route"] = map[string]interface{}{
				"destination": map[string]interface{}{
					"symbol": nav.Route.Destination.Symbol,
					"type":   nav.Route.Destination.Type,
					"x":      nav.Route.Destination.X,
					"y":      nav.Route.Destination.Y,
				},
				"origin": map[string]interface{}{
					"symbol": nav.Route.Origin.Symbol,
					"type":   nav.Route.Origin.Type,
					"x":      nav.Route.Origin.X,
					"y":      nav.Route.Origin.Y,
				},
				"departure_time": nav.Route.DepartureTime,
				"arrival":        nav.Route.Arrival,
			}
		}

		// Create text summary
		textSummary := "## Ship Dock Successful\n\n"
		textSummary += fmt.Sprintf("**Ship:** %s\n", shipSymbol)
		textSummary += fmt.Sprintf("**Status:** %s\n", nav.Status)
		textSummary += fmt.Sprintf("**Location:** %s (%s)\n", nav.WaypointSymbol, nav.SystemSymbol)
		textSummary += fmt.Sprintf("**Flight Mode:** %s\n", nav.FlightMode)

		if nav.Route.Destination.Symbol != "" {
			textSummary += "\n**Current Route:**\n"
			textSummary += fmt.Sprintf("- From: %s (%s)\n", nav.Route.Origin.Symbol, nav.Route.Origin.Type)
			textSummary += fmt.Sprintf("- To: %s (%s)\n", nav.Route.Destination.Symbol, nav.Route.Destination.Type)
			textSummary += fmt.Sprintf("- Departure: %s\n", nav.Route.DepartureTime)
			textSummary += fmt.Sprintf("- Arrival: %s\n", nav.Route.Arrival)
		}

		textSummary += "\n**Available Actions:**\n"
		textSummary += "- Now that the ship is docked, you can:\n"
		textSummary += "  - Trade goods at the marketplace\n"
		textSummary += "  - Purchase ships at the shipyard\n"
		textSummary += "  - Refuel the ship\n"
		textSummary += "  - Repair the ship\n"

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%+v\n```", result)),
			},
		}, nil
	}
}
