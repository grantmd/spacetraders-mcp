package tools

import (
	"context"
	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"
	"spacetraders-mcp/pkg/tools/contract"
	"spacetraders-mcp/pkg/tools/info"
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

	// TODO: Add more tool handlers here as we implement them:
	// - NavigateShip tool
	// - FulfillContract tool
	// - PurchaseShip tool
	// - SellCargo tool
	// - BuyCargo tool
	// - RefuelShip tool
	// - RepairShip tool
	// - ExtractResources tool
	// - JumpShip tool
	// - ScanSystems tool
	// - ScanWaypoints tool
	// - ScanShips tool
	// - OrbitShip tool
	// - DockShip tool
	// etc.
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
