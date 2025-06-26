package navigation

import (
	"context"
	"fmt"
	"time"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"

	"github.com/mark3labs/mcp-go/mcp"
)

// WarpShipTool handles warping ships to waypoints
type WarpShipTool struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewWarpShipTool creates a new warp ship tool
func NewWarpShipTool(client *spacetraders.Client, logger *logging.Logger) *WarpShipTool {
	return &WarpShipTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *WarpShipTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "warp_ship",
		Description: "Warp a ship to a waypoint in a different system. Ship must have a warp drive and be in orbit to warp.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"ship_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the ship to warp (e.g., 'SHIP_1234')",
				},
				"waypoint_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the destination waypoint in another system (e.g., 'X1-AB12-34567Z')",
				},
			},
			Required: []string{"ship_symbol", "waypoint_symbol"},
		},
	}
}

// Handler returns the tool handler function
func (t *WarpShipTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contextLogger := t.logger.WithContext(ctx, "warp-ship-tool")

		// Extract ship symbol and waypoint symbol
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

		contextLogger.Info(fmt.Sprintf("Attempting to warp ship %s to %s", shipSymbol, waypointSymbol))

		// Warp the ship
		nav, fuel, event, err := t.client.WarpShip(shipSymbol, waypointSymbol)
		if err != nil {
			contextLogger.Error(fmt.Sprintf("Failed to warp ship %s to %s: %v", shipSymbol, waypointSymbol, err))
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("Failed to warp ship %s to %s: %v", shipSymbol, waypointSymbol, err)),
				},
				IsError: true,
			}, nil
		}

		contextLogger.ToolCall("warp_ship", true)
		contextLogger.Info(fmt.Sprintf("Successfully initiated warp for ship %s to %s", shipSymbol, waypointSymbol))

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
		textSummary := "## Ship Warp Initiated\n\n"
		textSummary += fmt.Sprintf("**Ship:** %s\n", shipSymbol)
		textSummary += fmt.Sprintf("**Status:** %s\n", nav.Status)
		textSummary += fmt.Sprintf("**Current Location:** %s (%s)\n", nav.WaypointSymbol, nav.SystemSymbol)
		textSummary += fmt.Sprintf("**Flight Mode:** %s\n", nav.FlightMode)
		textSummary += fmt.Sprintf("**Fuel:** %d/%d units\n", fuel.Current, fuel.Capacity)

		if nav.Route.Destination.Symbol != "" {
			textSummary += "\n**Warp Route Details:**\n"
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
						textSummary += fmt.Sprintf("- **Warp Time:** %s\n", duration.String())
					}
				}
			}
		}

		if fuel.Consumed.Amount > 0 {
			textSummary += "\n**Fuel Consumption:**\n"
			textSummary += fmt.Sprintf("- **Amount Used:** %d units\n", fuel.Consumed.Amount)
			textSummary += fmt.Sprintf("- **Remaining:** %d units\n", fuel.Current)
			textSummary += "- **Efficiency:** Warp drives consume significant fuel for inter-system travel\n"
		}

		if event != nil {
			textSummary += "\n**Warp Event:**\n"
			textSummary += fmt.Sprintf("- **Event:** %s\n", event.Name)
			textSummary += fmt.Sprintf("- **Description:** %s\n", event.Description)
			if event.Component != "" {
				textSummary += fmt.Sprintf("- **Component:** %s\n", event.Component)
			}
		}

		if nav.Status == "IN_TRANSIT" {
			textSummary += "\n**Status:** The ship is currently warping through space. It will automatically arrive at the destination system at the scheduled time.\n"
			textSummary += "Warp travel allows ships to move between different star systems much faster than conventional navigation.\n"
			textSummary += "Use the `get_status_summary` tool to check the current status of all your ships.\n"
		}

		textSummary += "\n**Important Notes:**\n"
		textSummary += "- Warp drives enable inter-system travel\n"
		textSummary += "- Warp travel is faster than regular navigation but consumes more fuel\n"
		textSummary += "- Ships must have a functional warp drive installed\n"
		textSummary += "- Ships must be in orbit before initiating warp\n"

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%+v\n```", result)),
			},
		}, nil
	}
}
