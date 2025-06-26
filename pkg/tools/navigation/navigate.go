package navigation

import (
	"context"
	"fmt"
	"time"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"

	"github.com/mark3labs/mcp-go/mcp"
)

// NavigateShipTool handles navigating ships to waypoints
type NavigateShipTool struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewNavigateShipTool creates a new navigate ship tool
func NewNavigateShipTool(client *spacetraders.Client, logger *logging.Logger) *NavigateShipTool {
	return &NavigateShipTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *NavigateShipTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "navigate_ship",
		Description: "Navigate a ship to a waypoint within the same system. Ship must be in orbit to navigate.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"ship_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the ship to navigate (e.g., 'SHIP_1234')",
				},
				"waypoint_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the destination waypoint (e.g., 'X1-DF55-20250Z')",
				},
			},
			Required: []string{"ship_symbol", "waypoint_symbol"},
		},
	}
}

// Handler returns the tool handler function
func (t *NavigateShipTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contextLogger := t.logger.WithContext(ctx, "navigate-ship-tool")

		// Extract ship symbol
		var shipSymbol string
		var waypointSymbol string
		if request.Params.Arguments != nil {
			if argsMap, ok := request.Params.Arguments.(map[string]interface{}); ok {
				if val, exists := argsMap["ship_symbol"]; exists {
					if s, ok := val.(string); ok && s != "" {
						shipSymbol = s
					}
				}
				if val, exists := argsMap["waypoint_symbol"]; exists {
					if s, ok := val.(string); ok && s != "" {
						waypointSymbol = s
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

		if waypointSymbol == "" {
			contextLogger.Error("Missing or invalid waypoint_symbol parameter")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("Error: waypoint_symbol parameter is required and must be a non-empty string"),
				},
				IsError: true,
			}, nil
		}

		contextLogger.Info(fmt.Sprintf("Attempting to navigate ship %s to %s", shipSymbol, waypointSymbol))

		// Navigate the ship
		nav, fuel, event, err := t.client.NavigateShip(shipSymbol, waypointSymbol)
		if err != nil {
			contextLogger.Error(fmt.Sprintf("Failed to navigate ship %s to %s: %v", shipSymbol, waypointSymbol, err))
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("Failed to navigate ship %s to %s: %v", shipSymbol, waypointSymbol, err)),
				},
				IsError: true,
			}, nil
		}

		contextLogger.ToolCall("navigate_ship", true)
		contextLogger.Info(fmt.Sprintf("Successfully started navigation for ship %s to %s", shipSymbol, waypointSymbol))

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
			"fuel": map[string]interface{}{
				"current":  fuel.Current,
				"capacity": fuel.Capacity,
			},
		}

		// Add route information
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

		// Add fuel consumption information if available
		if fuel.Consumed.Amount > 0 {
			result["fuel_consumed"] = map[string]interface{}{
				"amount":    fuel.Consumed.Amount,
				"timestamp": fuel.Consumed.Timestamp,
			}
		}

		// Add event information if available
		if event != nil {
			result["event"] = map[string]interface{}{
				"symbol":      event.Symbol,
				"component":   event.Component,
				"name":        event.Name,
				"description": event.Description,
			}
		}

		// Create text summary
		textSummary := "## Ship Navigation Started\n\n"
		textSummary += fmt.Sprintf("**Ship:** %s\n", shipSymbol)
		textSummary += fmt.Sprintf("**Status:** %s\n", nav.Status)
		textSummary += fmt.Sprintf("**Current Location:** %s (%s)\n", nav.WaypointSymbol, nav.SystemSymbol)
		textSummary += fmt.Sprintf("**Flight Mode:** %s\n", nav.FlightMode)
		textSummary += fmt.Sprintf("**Fuel:** %d/%d units\n", fuel.Current, fuel.Capacity)

		if nav.Route.Destination.Symbol != "" {
			textSummary += "\n**Route Details:**\n"
			textSummary += fmt.Sprintf("- **From:** %s (%s) at coordinates (%d, %d)\n",
				nav.Route.Origin.Symbol, nav.Route.Origin.Type, nav.Route.Origin.X, nav.Route.Origin.Y)
			textSummary += fmt.Sprintf("- **To:** %s (%s) at coordinates (%d, %d)\n",
				nav.Route.Destination.Symbol, nav.Route.Destination.Type, nav.Route.Destination.X, nav.Route.Destination.Y)
			textSummary += fmt.Sprintf("- **Departure:** %s\n", nav.Route.DepartureTime)
			textSummary += fmt.Sprintf("- **Arrival:** %s\n", nav.Route.Arrival)

			// Calculate travel time if possible
			if nav.Route.DepartureTime != "" && nav.Route.Arrival != "" {
				if departureTime, err := time.Parse(time.RFC3339, nav.Route.DepartureTime); err == nil {
					if arrivalTime, err := time.Parse(time.RFC3339, nav.Route.Arrival); err == nil {
						duration := arrivalTime.Sub(departureTime)
						textSummary += fmt.Sprintf("- **Travel Time:** %s\n", duration.String())
					}
				}
			}
		}

		if fuel.Consumed.Amount > 0 {
			textSummary += "\n**Fuel Consumption:**\n"
			textSummary += fmt.Sprintf("- **Amount Used:** %d units\n", fuel.Consumed.Amount)
			textSummary += fmt.Sprintf("- **Remaining:** %d units\n", fuel.Current)
		}

		if event != nil {
			textSummary += "\n**Navigation Event:**\n"
			textSummary += fmt.Sprintf("- **Event:** %s\n", event.Name)
			textSummary += fmt.Sprintf("- **Description:** %s\n", event.Description)
			if event.Component != "" {
				textSummary += fmt.Sprintf("- **Component:** %s\n", event.Component)
			}
		}

		if nav.Status == "IN_TRANSIT" {
			textSummary += "\n**Status:** The ship is currently in transit. It will automatically arrive at the destination at the scheduled time.\n"
			textSummary += "Use the `get_status_summary` tool to check the current status of all your ships.\n"
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%+v\n```", result)),
			},
		}, nil
	}
}
