package navigation

import (
	"context"
	"fmt"
	"time"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"

	"github.com/mark3labs/mcp-go/mcp"
)

// JumpShipTool handles jumping ships to different systems
type JumpShipTool struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewJumpShipTool creates a new jump ship tool
func NewJumpShipTool(client *spacetraders.Client, logger *logging.Logger) *JumpShipTool {
	return &JumpShipTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *JumpShipTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "jump_ship",
		Description: "Jump a ship to a different system using a jump gate. Ship must have a jump drive and be in orbit to jump. Creates a cooldown period.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"ship_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the ship to jump (e.g., 'SHIP_1234')",
				},
				"system_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the destination system (e.g., 'X1-AB12')",
				},
			},
			Required: []string{"ship_symbol", "system_symbol"},
		},
	}
}

// Handler returns the tool handler function
func (t *JumpShipTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contextLogger := t.logger.WithContext(ctx, "jump-ship-tool")

		// Extract ship symbol and system symbol
		var shipSymbol string
		var systemSymbol string
		if request.Params.Arguments != nil {
			if argsMap, ok := request.Params.Arguments.(map[string]interface{}); ok {
				if val, exists := argsMap["ship_symbol"]; exists {
					if s, ok := val.(string); ok && s != "" {
						shipSymbol = s
					}
				}
				if val, exists := argsMap["system_symbol"]; exists {
					if s, ok := val.(string); ok && s != "" {
						systemSymbol = s
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

		if systemSymbol == "" {
			contextLogger.Error("Missing or invalid system_symbol parameter")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("Error: system_symbol parameter is required and must be a non-empty string"),
				},
				IsError: true,
			}, nil
		}

		contextLogger.Info(fmt.Sprintf("Attempting to jump ship %s to system %s", shipSymbol, systemSymbol))

		// Jump the ship
		nav, cooldown, event, err := t.client.JumpShip(shipSymbol, systemSymbol)
		if err != nil {
			contextLogger.Error(fmt.Sprintf("Failed to jump ship %s to %s: %v", shipSymbol, systemSymbol, err))
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("Failed to jump ship %s to system %s: %v", shipSymbol, systemSymbol, err)),
				},
				IsError: true,
			}, nil
		}

		contextLogger.ToolCall("jump_ship", true)
		contextLogger.Info(fmt.Sprintf("Successfully jumped ship %s to system %s", shipSymbol, systemSymbol))

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
			"cooldown": map[string]interface{}{
				"ship_symbol":       cooldown.ShipSymbol,
				"total_seconds":     cooldown.TotalSeconds,
				"remaining_seconds": cooldown.RemainingSeconds,
				"expiration":        cooldown.Expiration,
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
		textSummary := fmt.Sprintf("## Ship Jump Completed\n\n")
		textSummary += fmt.Sprintf("**Ship:** %s\n", shipSymbol)
		textSummary += fmt.Sprintf("**Status:** %s\n", nav.Status)
		textSummary += fmt.Sprintf("**New Location:** %s (%s)\n", nav.WaypointSymbol, nav.SystemSymbol)
		textSummary += fmt.Sprintf("**Flight Mode:** %s\n", nav.FlightMode)

		if nav.Route.Destination.Symbol != "" {
			textSummary += fmt.Sprintf("\n**Jump Details:**\n")
			textSummary += fmt.Sprintf("- **From:** %s (%s) at coordinates (%d, %d)\n",
				nav.Route.Origin.Symbol, nav.Route.Origin.Type, nav.Route.Origin.X, nav.Route.Origin.Y)
			textSummary += fmt.Sprintf("- **To:** %s (%s) at coordinates (%d, %d)\n",
				nav.Route.Destination.Symbol, nav.Route.Destination.Type, nav.Route.Destination.X, nav.Route.Destination.Y)
			textSummary += fmt.Sprintf("- **Departure:** %s\n", nav.Route.DepartureTime)
			textSummary += fmt.Sprintf("- **Arrival:** %s\n", nav.Route.Arrival)
		}

		// Add cooldown information
		textSummary += fmt.Sprintf("\n**Jump Drive Cooldown:**\n")
		textSummary += fmt.Sprintf("- **Total Cooldown:** %d seconds\n", cooldown.TotalSeconds)
		textSummary += fmt.Sprintf("- **Remaining:** %d seconds\n", cooldown.RemainingSeconds)
		textSummary += fmt.Sprintf("- **Ready At:** %s\n", cooldown.Expiration)

		// Calculate cooldown duration
		if cooldown.Expiration != "" {
			if expirationTime, err := time.Parse(time.RFC3339, cooldown.Expiration); err == nil {
				now := time.Now()
				if expirationTime.After(now) {
					duration := expirationTime.Sub(now)
					textSummary += fmt.Sprintf("- **Time Until Ready:** %s\n", duration.String())
				} else {
					textSummary += fmt.Sprintf("- **Status:** Jump drive is ready for use\n")
				}
			}
		}

		if event != nil {
			textSummary += fmt.Sprintf("\n**Jump Event:**\n")
			textSummary += fmt.Sprintf("- **Event:** %s\n", event.Name)
			textSummary += fmt.Sprintf("- **Description:** %s\n", event.Description)
			if event.Component != "" {
				textSummary += fmt.Sprintf("- **Component:** %s\n", event.Component)
			}
		}

		textSummary += fmt.Sprintf("\n**Jump Drive Information:**\n")
		textSummary += fmt.Sprintf("- Jump drives enable instant travel between systems via jump gates\n")
		textSummary += fmt.Sprintf("- Each jump creates a cooldown period before the next jump\n")
		textSummary += fmt.Sprintf("- Ships must be in orbit and have a functional jump drive\n")
		textSummary += fmt.Sprintf("- Jump gates connect specific systems - not all systems are connected\n")

		if cooldown.RemainingSeconds > 0 {
			textSummary += fmt.Sprintf("\n**Current Status:** The ship's jump drive is cooling down and cannot be used until the cooldown expires.\n")
			textSummary += fmt.Sprintf("During cooldown, the ship can still use regular navigation and warp drives.\n")
		} else {
			textSummary += fmt.Sprintf("\n**Current Status:** The ship's jump drive is ready for immediate use.\n")
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%+v\n```", result)),
			},
		}, nil
	}
}
