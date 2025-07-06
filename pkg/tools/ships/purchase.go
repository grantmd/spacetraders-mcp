package ships

import (
	"context"
	"fmt"
	"strings"
	"time"

	"spacetraders-mcp/pkg/client"
	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/tools/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

// PurchaseShipTool handles ship purchasing at shipyards
type PurchaseShipTool struct {
	client *client.Client
	logger *logging.Logger
}

// NewPurchaseShipTool creates a new ship purchase tool
func NewPurchaseShipTool(client *client.Client, logger *logging.Logger) *PurchaseShipTool {
	return &PurchaseShipTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *PurchaseShipTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "purchase_ship",
		Description: "Purchase a ship at a shipyard. Requires being docked at the shipyard and having sufficient credits.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"ship_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of ship to purchase (e.g., SHIP_MINING_DRONE, SHIP_PROBE, SHIP_LIGHT_HAULER)",
				},
				"waypoint_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Waypoint symbol of the shipyard where you want to purchase the ship (e.g., X1-FM66-B2)",
				},
			},
			Required: []string{"ship_type", "waypoint_symbol"},
		},
	}
}

// Handler returns the tool handler function
func (t *PurchaseShipTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Set up context logger
		ctxLogger := t.logger.WithContext(ctx, "purchase-ship-tool")
		ctxLogger.Debug("Processing ship purchase request")

		// Parse arguments
		shipType := ""
		waypointSymbol := ""

		if request.Params.Arguments == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("❌ Missing required arguments: ship_type and waypoint_symbol"),
				},
				IsError: true,
			}, nil
		}

		if argsMap, ok := request.Params.Arguments.(map[string]interface{}); ok {
			if st, exists := argsMap["ship_type"]; exists {
				if stStr, ok := st.(string); ok {
					shipType = strings.TrimSpace(stStr)
				}
			}
			if ws, exists := argsMap["waypoint_symbol"]; exists {
				if wsStr, ok := ws.(string); ok {
					waypointSymbol = strings.TrimSpace(wsStr)
				}
			}
		}

		if shipType == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("❌ ship_type is required"),
				},
				IsError: true,
			}, nil
		}

		if waypointSymbol == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("❌ waypoint_symbol is required"),
				},
				IsError: true,
			}, nil
		}

		ctxLogger.Info("Attempting to purchase %s at %s", shipType, waypointSymbol)

		// Purchase the ship
		start := time.Now()
		req := client.PurchaseShipRequest{
			ShipType:       shipType,
			WaypointSymbol: waypointSymbol,
		}
		resp, err := t.client.PurchaseShip(req)
		duration := time.Since(start)

		if err != nil {
			ctxLogger.Error("Failed to purchase ship: %v", err)
			ctxLogger.APICall("/my/ships", 0, duration.String())
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("❌ Failed to purchase ship: %s", err.Error())),
				},
				IsError: true,
			}, nil
		}

		ctxLogger.APICall("/my/ships", 201, duration.String())
		ctxLogger.Info("Successfully purchased ship %s for %d credits", resp.Data.Ship.Symbol, resp.Data.Transaction.Price)

		// Format the response
		result := map[string]interface{}{
			"success": true,
			"ship": map[string]interface{}{
				"symbol": resp.Data.Ship.Symbol,
				"type":   shipType,
				"name":   resp.Data.Ship.Registration.Name,
				"role":   resp.Data.Ship.Registration.Role,
				"location": map[string]interface{}{
					"system":   resp.Data.Ship.Nav.SystemSymbol,
					"waypoint": resp.Data.Ship.Nav.WaypointSymbol,
					"status":   resp.Data.Ship.Nav.Status,
				},
				"specs": map[string]interface{}{
					"cargo_capacity": resp.Data.Ship.Cargo.Capacity,
					"fuel_capacity":  resp.Data.Ship.Fuel.Capacity,
					"crew": map[string]interface{}{
						"current":  resp.Data.Ship.Crew.Current,
						"capacity": resp.Data.Ship.Crew.Capacity,
					},
				},
			},
			"transaction": map[string]interface{}{
				"price":     resp.Data.Transaction.Price,
				"timestamp": resp.Data.Transaction.Timestamp,
			},
			"agent": map[string]interface{}{
				"credits":   resp.Data.Agent.Credits,
				"shipCount": resp.Data.Agent.ShipCount,
			},
		}

		jsonData := utils.FormatJSON(result)

		// Create formatted text summary
		textSummary := "🚢 **Ship Purchase Successful!**\n\n"
		textSummary += fmt.Sprintf("**New Ship:** %s (%s)\n", resp.Data.Ship.Symbol, resp.Data.Ship.Registration.Name)
		textSummary += fmt.Sprintf("**Type:** %s\n", shipType)
		textSummary += fmt.Sprintf("**Role:** %s\n", resp.Data.Ship.Registration.Role)
		textSummary += fmt.Sprintf("**Location:** %s (Status: %s)\n", resp.Data.Ship.Nav.WaypointSymbol, resp.Data.Ship.Nav.Status)
		textSummary += fmt.Sprintf("**Cost:** %d credits\n", resp.Data.Transaction.Price)
		textSummary += fmt.Sprintf("**Remaining Credits:** %d\n", resp.Data.Agent.Credits)
		textSummary += fmt.Sprintf("**Total Ships:** %d\n\n", resp.Data.Agent.ShipCount)

		textSummary += "**Ship Specifications:**\n"
		textSummary += fmt.Sprintf("• Cargo Capacity: %d units\n", resp.Data.Ship.Cargo.Capacity)
		textSummary += fmt.Sprintf("• Fuel Capacity: %d units\n", resp.Data.Ship.Fuel.Capacity)
		textSummary += fmt.Sprintf("• Crew Capacity: %d/%d\n\n", resp.Data.Ship.Crew.Current, resp.Data.Ship.Crew.Capacity)

		textSummary += "💡 **Next Steps:**\n"
		textSummary += "• Use `get_status_summary` to see your updated fleet\n"
		textSummary += "• Your new ship is ready for missions!\n"

		ctxLogger.ToolCall("purchase_ship", true)
		ctxLogger.Debug("Purchase ship response size: %d bytes", len(jsonData))

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%s\n```", jsonData)),
			},
		}, nil
	}
}
