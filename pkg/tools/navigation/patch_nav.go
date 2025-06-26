package navigation

import (
	"context"
	"fmt"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"

	"github.com/mark3labs/mcp-go/mcp"
)

// PatchNavTool handles changing ship navigation settings
type PatchNavTool struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewPatchNavTool creates a new patch nav tool
func NewPatchNavTool(client *spacetraders.Client, logger *logging.Logger) *PatchNavTool {
	return &PatchNavTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *PatchNavTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "patch_ship_nav",
		Description: "Change a ship's navigation settings, particularly the flight mode. Available flight modes: DRIFT, STEALTH, CRUISE, BURN.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"ship_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the ship to modify (e.g., 'SHIP_1234')",
				},
				"flight_mode": map[string]interface{}{
					"type":        "string",
					"description": "Flight mode to set. Options: DRIFT (slowest, most fuel efficient), STEALTH (slow, hard to detect), CRUISE (balanced), BURN (fastest, most fuel consumption)",
					"enum":        []string{"DRIFT", "STEALTH", "CRUISE", "BURN"},
				},
			},
			Required: []string{"ship_symbol", "flight_mode"},
		},
	}
}

// Handler returns the tool handler function
func (t *PatchNavTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contextLogger := t.logger.WithContext(ctx, "patch-nav-tool")

		// Extract ship symbol and flight mode
		var shipSymbol string
		var flightMode string
		if request.Params.Arguments != nil {
			if argsMap, ok := request.Params.Arguments.(map[string]interface{}); ok {
				if val, exists := argsMap["ship_symbol"]; exists {
					if s, ok := val.(string); ok && s != "" {
						shipSymbol = s
					}
				}
				if val, exists := argsMap["flight_mode"]; exists {
					if s, ok := val.(string); ok && s != "" {
						flightMode = s
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

		if flightMode == "" {
			contextLogger.Error("Missing or invalid flight_mode parameter")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("Error: flight_mode parameter is required and must be one of: DRIFT, STEALTH, CRUISE, BURN"),
				},
				IsError: true,
			}, nil
		}

		// Validate flight mode
		validModes := map[string]bool{
			"DRIFT":   true,
			"STEALTH": true,
			"CRUISE":  true,
			"BURN":    true,
		}
		if !validModes[flightMode] {
			contextLogger.Error(fmt.Sprintf("Invalid flight mode: %s", flightMode))
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("Error: Invalid flight mode '%s'. Must be one of: DRIFT, STEALTH, CRUISE, BURN", flightMode)),
				},
				IsError: true,
			}, nil
		}

		contextLogger.Info(fmt.Sprintf("Attempting to change flight mode for ship %s to %s", shipSymbol, flightMode))

		// Patch the ship's navigation
		nav, err := t.client.PatchShipNav(shipSymbol, flightMode)
		if err != nil {
			contextLogger.Error(fmt.Sprintf("Failed to patch nav for ship %s: %v", shipSymbol, err))
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("Failed to change flight mode for ship %s: %v", shipSymbol, err)),
				},
				IsError: true,
			}, nil
		}

		contextLogger.ToolCall("patch_ship_nav", true)
		contextLogger.Info(fmt.Sprintf("Successfully changed flight mode for ship %s to %s", shipSymbol, flightMode))

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

		// Create text summary with flight mode descriptions
		textSummary := fmt.Sprintf("## Flight Mode Updated\n\n")
		textSummary += fmt.Sprintf("**Ship:** %s\n", shipSymbol)
		textSummary += fmt.Sprintf("**Flight Mode:** %s\n", nav.FlightMode)
		textSummary += fmt.Sprintf("**Status:** %s\n", nav.Status)
		textSummary += fmt.Sprintf("**Location:** %s (%s)\n", nav.WaypointSymbol, nav.SystemSymbol)

		// Add flight mode description
		modeDescriptions := map[string]string{
			"DRIFT":   "Slowest speed, most fuel efficient, longest travel times",
			"STEALTH": "Slow speed, harder to detect, moderate fuel consumption",
			"CRUISE":  "Balanced speed and fuel consumption",
			"BURN":    "Fastest speed, highest fuel consumption, shortest travel times",
		}
		if desc, exists := modeDescriptions[nav.FlightMode]; exists {
			textSummary += fmt.Sprintf("**Mode Description:** %s\n", desc)
		}

		if nav.Route.Destination.Symbol != "" {
			textSummary += fmt.Sprintf("\n**Current Route:**\n")
			textSummary += fmt.Sprintf("- From: %s (%s)\n", nav.Route.Origin.Symbol, nav.Route.Origin.Type)
			textSummary += fmt.Sprintf("- To: %s (%s)\n", nav.Route.Destination.Symbol, nav.Route.Destination.Type)
			textSummary += fmt.Sprintf("- Departure: %s\n", nav.Route.DepartureTime)
			textSummary += fmt.Sprintf("- Arrival: %s\n", nav.Route.Arrival)
			textSummary += fmt.Sprintf("\n**Note:** The arrival time may have changed due to the flight mode change.\n")
		}

		textSummary += fmt.Sprintf("\n**Flight Mode Effects:**\n")
		textSummary += fmt.Sprintf("- **DRIFT:** 25%% speed, 1x fuel consumption\n")
		textSummary += fmt.Sprintf("- **STEALTH:** 30%% speed, 1x fuel consumption, stealth bonus\n")
		textSummary += fmt.Sprintf("- **CRUISE:** 100%% speed, 1x fuel consumption (default)\n")
		textSummary += fmt.Sprintf("- **BURN:** 200%% speed, 2x fuel consumption\n")

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%+v\n```", result)),
			},
		}, nil
	}
}
