package navigation

import (
	"context"
	"fmt"
	"time"

	"spacetraders-mcp/pkg/client"
	"spacetraders-mcp/pkg/logging"

	"github.com/mark3labs/mcp-go/mcp"
)

// WarpShipTool handles warping ships to waypoints
type WarpShipTool struct {
	client *client.Client
	logger *logging.Logger
}

// NewWarpShipTool creates a new warp ship tool
func NewWarpShipTool(client *client.Client, logger *logging.Logger) *WarpShipTool {
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
		resp, err := t.client.WarpShip(shipSymbol, waypointSymbol)
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
				"system_symbol":   resp.Data.Nav.SystemSymbol,
				"waypoint_symbol": resp.Data.Nav.WaypointSymbol,
				"status":          resp.Data.Nav.Status,
				"flight_mode":     resp.Data.Nav.FlightMode,
			},
			"fuel": map[string]interface{}{
				"current":  resp.Data.Fuel.Current,
				"capacity": resp.Data.Fuel.Capacity,
			},
		}

		// Add route information
		if resp.Data.Nav.Route.Destination.Symbol != "" {
			result["route"] = map[string]interface{}{
				"destination": map[string]interface{}{
					"symbol": resp.Data.Nav.Route.Destination.Symbol,
					"type":   resp.Data.Nav.Route.Destination.Type,
					"x":      resp.Data.Nav.Route.Destination.X,
					"y":      resp.Data.Nav.Route.Destination.Y,
				},
				"origin": map[string]interface{}{
					"symbol": resp.Data.Nav.Route.Origin.Symbol,
					"type":   resp.Data.Nav.Route.Origin.Type,
					"x":      resp.Data.Nav.Route.Origin.X,
					"y":      resp.Data.Nav.Route.Origin.Y,
				},
				"departure_time": resp.Data.Nav.Route.DepartureTime,
				"arrival":        resp.Data.Nav.Route.Arrival,
			}
		}

		// Add fuel consumption information if available
		if resp.Data.Fuel.Consumed.Amount > 0 {
			result["fuel_consumed"] = map[string]interface{}{
				"amount":    resp.Data.Fuel.Consumed.Amount,
				"timestamp": resp.Data.Fuel.Consumed.Timestamp,
			}
		}

		// Add event information if available
		if resp.Data.Event.Symbol != "" {
			result["event"] = map[string]interface{}{
				"symbol":      resp.Data.Event.Symbol,
				"component":   resp.Data.Event.Component,
				"name":        resp.Data.Event.Name,
				"description": resp.Data.Event.Description,
			}
		}

		// Create text summary
		textSummary := "## Ship Warp Initiated\n\n"
		textSummary += fmt.Sprintf("**Ship:** %s\n", shipSymbol)
		textSummary += fmt.Sprintf("**Status:** %s\n", resp.Data.Nav.Status)
		textSummary += fmt.Sprintf("**Current Location:** %s (%s)\n", resp.Data.Nav.WaypointSymbol, resp.Data.Nav.SystemSymbol)
		textSummary += fmt.Sprintf("**Flight Mode:** %s\n", resp.Data.Nav.FlightMode)
		textSummary += fmt.Sprintf("**Fuel:** %d/%d units\n", resp.Data.Fuel.Current, resp.Data.Fuel.Capacity)

		if resp.Data.Nav.Route.Destination.Symbol != "" {
			textSummary += "\n**Warp Route Details:**\n"
			textSummary += fmt.Sprintf("- **From:** %s (%s) at coordinates (%d, %d)\n",
				resp.Data.Nav.Route.Origin.Symbol, resp.Data.Nav.Route.Origin.Type, resp.Data.Nav.Route.Origin.X, resp.Data.Nav.Route.Origin.Y)
			textSummary += fmt.Sprintf("- **To:** %s (%s) at coordinates (%d, %d)\n",
				resp.Data.Nav.Route.Destination.Symbol, resp.Data.Nav.Route.Destination.Type, resp.Data.Nav.Route.Destination.X, resp.Data.Nav.Route.Destination.Y)
			textSummary += fmt.Sprintf("- **Departure:** %s\n", resp.Data.Nav.Route.DepartureTime)
			textSummary += fmt.Sprintf("- **Arrival:** %s\n", resp.Data.Nav.Route.Arrival)

			// Calculate travel time if possible
			if resp.Data.Nav.Route.DepartureTime != "" && resp.Data.Nav.Route.Arrival != "" {
				if departureTime, err := time.Parse(time.RFC3339, resp.Data.Nav.Route.DepartureTime); err == nil {
					if arrivalTime, err := time.Parse(time.RFC3339, resp.Data.Nav.Route.Arrival); err == nil {
						duration := arrivalTime.Sub(departureTime)
						textSummary += fmt.Sprintf("- **Warp Time:** %s\n", duration.String())
					}
				}
			}
		}

		if resp.Data.Fuel.Consumed.Amount > 0 {
			textSummary += "\n**Fuel Consumption:**\n"
			textSummary += fmt.Sprintf("- **Amount Used:** %d units\n", resp.Data.Fuel.Consumed.Amount)
			textSummary += fmt.Sprintf("- **Remaining:** %d units\n", resp.Data.Fuel.Current)
			textSummary += "- **Efficiency:** Warp drives consume significant fuel for inter-system travel\n"
		}

		if resp.Data.Event.Symbol != "" {
			textSummary += "\n**Warp Event:**\n"
			textSummary += fmt.Sprintf("- **Event:** %s\n", resp.Data.Event.Name)
			textSummary += fmt.Sprintf("- **Description:** %s\n", resp.Data.Event.Description)
			if resp.Data.Event.Component != "" {
				textSummary += fmt.Sprintf("- **Component:** %s\n", resp.Data.Event.Component)
			}
		}

		if resp.Data.Nav.Status == "IN_TRANSIT" {
			textSummary += "\n**Status:** The ship is currently warping through space. It will automatically arrive at the destination system at the scheduled time.\n"
			textSummary += "Warp travel allows ships to move between different star systems much faster than conventional navigation.\n"
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
