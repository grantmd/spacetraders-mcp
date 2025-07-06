package navigation

import (
	"context"
	"fmt"

	"spacetraders-mcp/pkg/client"
	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/tools/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

// OrbitShipTool handles putting ships into orbit
type OrbitShipTool struct {
	client *client.Client
	logger *logging.Logger
}

// NewOrbitShipTool creates a new orbit ship tool
func NewOrbitShipTool(client *client.Client, logger *logging.Logger) *OrbitShipTool {
	return &OrbitShipTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *OrbitShipTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "orbit_ship",
		Description: "Put a ship into orbit around its current waypoint. Ship must be docked to use this command.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"ship_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the ship to put into orbit (e.g., 'SHIP_1234')",
				},
			},
			Required: []string{"ship_symbol"},
		},
	}
}

// Handler returns the tool handler function
func (t *OrbitShipTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contextLogger := t.logger.WithContext(ctx, "orbit-ship-tool")

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

		contextLogger.Info(fmt.Sprintf("Attempting to orbit ship: %s", shipSymbol))

		// Orbit the ship
		nav, err := t.client.OrbitShip(shipSymbol)
		if err != nil {
			contextLogger.Error(fmt.Sprintf("Failed to orbit ship %s: %v", shipSymbol, err))
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("Failed to orbit ship %s: %v", shipSymbol, err)),
				},
				IsError: true,
			}, nil
		}

		contextLogger.ToolCall("orbit_ship", true)
		contextLogger.Info(fmt.Sprintf("Successfully put ship %s into orbit at %s", shipSymbol, nav.Data.Nav.WaypointSymbol))

		// Create structured response
		result := map[string]interface{}{
			"success":     true,
			"ship_symbol": shipSymbol,
			"navigation": map[string]interface{}{
				"system_symbol":   nav.Data.Nav.SystemSymbol,
				"waypoint_symbol": nav.Data.Nav.WaypointSymbol,
				"status":          nav.Data.Nav.Status,
				"flight_mode":     nav.Data.Nav.FlightMode,
			},
		}

		// Add route information if available
		if nav.Data.Nav.Route.Destination.Symbol != "" {
			result["route"] = map[string]interface{}{
				"destination": map[string]interface{}{
					"symbol": nav.Data.Nav.Route.Destination.Symbol,
					"type":   nav.Data.Nav.Route.Destination.Type,
					"x":      nav.Data.Nav.Route.Destination.X,
					"y":      nav.Data.Nav.Route.Destination.Y,
				},
				"origin": map[string]interface{}{
					"symbol": nav.Data.Nav.Route.Origin.Symbol,
					"type":   nav.Data.Nav.Route.Origin.Type,
					"x":      nav.Data.Nav.Route.Origin.X,
					"y":      nav.Data.Nav.Route.Origin.Y,
				},
				"departure_time": nav.Data.Nav.Route.DepartureTime,
				"arrival":        nav.Data.Nav.Route.Arrival,
			}
		}

		// Create text summary
		textSummary := "## Ship Orbit Successful\n\n"
		textSummary += fmt.Sprintf("**Ship:** %s\n", shipSymbol)
		textSummary += fmt.Sprintf("**Status:** %s\n", nav.Data.Nav.Status)
		textSummary += fmt.Sprintf("**Location:** %s (%s)\n", nav.Data.Nav.WaypointSymbol, nav.Data.Nav.SystemSymbol)
		textSummary += fmt.Sprintf("**Flight Mode:** %s\n", nav.Data.Nav.FlightMode)

		if nav.Data.Nav.Route.Destination.Symbol != "" {
			textSummary += "\n**Current Route:**\n"
			textSummary += fmt.Sprintf("- From: %s (%s)\n", nav.Data.Nav.Route.Origin.Symbol, nav.Data.Nav.Route.Origin.Type)
			textSummary += fmt.Sprintf("- To: %s (%s)\n", nav.Data.Nav.Route.Destination.Symbol, nav.Data.Nav.Route.Destination.Type)
			textSummary += fmt.Sprintf("- Departure: %s\n", nav.Data.Nav.Route.DepartureTime)
			textSummary += fmt.Sprintf("- Arrival: %s\n", nav.Data.Nav.Route.Arrival)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%s\n```", utils.FormatJSON(result))),
			},
		}, nil
	}
}
