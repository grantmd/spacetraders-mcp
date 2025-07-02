package tools

import (
	"context"
	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"
	"spacetraders-mcp/pkg/tools/contract"
	"spacetraders-mcp/pkg/tools/exploration"
	"spacetraders-mcp/pkg/tools/info"
	"spacetraders-mcp/pkg/tools/navigation"
	"spacetraders-mcp/pkg/tools/ships"
	"spacetraders-mcp/pkg/tools/status"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ToolHandler defines the interface for all tool handlers
type ToolHandler interface {
	Tool() mcp.Tool
	Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

// Registry manages all MCP tools
type Registry struct {
	client   *spacetraders.Client
	logger   *logging.Logger
	handlers []ToolHandler
}

// NewRegistry creates a new tool registry
func NewRegistry(client *spacetraders.Client, logger *logging.Logger) *Registry {
	registry := &Registry{
		client:   client,
		logger:   logger,
		handlers: make([]ToolHandler, 0),
	}

	// Register all available tools
	registry.registerTools()

	return registry
}

// registerTools registers all available tool handlers
func (r *Registry) registerTools() {
	// Register AcceptContract tool
	r.handlers = append(r.handlers, contract.NewAcceptContractTool(r.client))

	// Register Status Summary tool
	r.handlers = append(r.handlers, status.NewStatusTool(r.client, r.logger))

	// Register Contract Info tool
	r.handlers = append(r.handlers, info.NewContractInfoTool(r.client, r.logger))

	// Register Fleet Analysis tool
	r.handlers = append(r.handlers, info.NewFleetAnalysisTool(r.client, r.logger))

	// Register Ship Purchase tool
	r.handlers = append(r.handlers, ships.NewPurchaseShipTool(r.client, r.logger))

	// Register Refuel Ship tool
	r.handlers = append(r.handlers, ships.NewRefuelShipTool(r.client, r.logger))

	// Register Extract Resources tool
	r.handlers = append(r.handlers, ships.NewExtractResourcesTool(r.client, r.logger))

	// Register Jettison Cargo tool
	r.handlers = append(r.handlers, ships.NewJettisonCargoTool(r.client, r.logger))

	// Register Navigation tools
	r.handlers = append(r.handlers, navigation.NewOrbitShipTool(r.client, r.logger))
	r.handlers = append(r.handlers, navigation.NewDockShipTool(r.client, r.logger))
	r.handlers = append(r.handlers, navigation.NewNavigateShipTool(r.client, r.logger))
	r.handlers = append(r.handlers, navigation.NewPatchNavTool(r.client, r.logger))
	r.handlers = append(r.handlers, navigation.NewWarpShipTool(r.client, r.logger))
	r.handlers = append(r.handlers, navigation.NewJumpShipTool(r.client, r.logger))

	// Register Exploration tools
	r.handlers = append(r.handlers, exploration.NewFindWaypointsTool(r.client, r.logger))
	r.handlers = append(r.handlers, exploration.NewSystemOverviewTool(r.client, r.logger))
	r.handlers = append(r.handlers, exploration.NewCurrentLocationTool(r.client, r.logger))

	// TODO: Add more tool handlers here as we implement them:
	// - FulfillContract tool
	// - SellCargo tool
	// - BuyCargo tool
	// - RepairShip tool
	// - ScanSystems tool
	// - ScanWaypoints tool
	// - ScanShips tool
	// etc.
	//
	// IMPLEMENTED:
	// - RefuelShip tool ✅
	// - ExtractResources tool ✅
	// - JettisonCargo tool ✅
}

// RegisterWithServer registers all tools with the MCP server
func (r *Registry) RegisterWithServer(s *server.MCPServer) {
	for _, handler := range r.handlers {
		s.AddTool(handler.Tool(), handler.Handler())
	}
}

// GetTools returns all registered tools (useful for testing/debugging)
func (r *Registry) GetTools() []mcp.Tool {
	tools := make([]mcp.Tool, len(r.handlers))
	for i, handler := range r.handlers {
		tools[i] = handler.Tool()
	}
	return tools
}
