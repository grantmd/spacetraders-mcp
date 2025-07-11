package resources

import (
	"context"
	"spacetraders-mcp/pkg/client"
	"spacetraders-mcp/pkg/logging"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ResourceHandler defines the interface for all resource handlers
type ResourceHandler interface {
	Resource() mcp.Resource
	Handler() func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error)
}

// Registry manages all MCP resources
type Registry struct {
	client   *client.Client
	logger   *logging.Logger
	handlers []ResourceHandler
}

// NewRegistry creates a new resource registry
func NewRegistry(client *client.Client, logger *logging.Logger) *Registry {
	registry := &Registry{
		client:   client,
		logger:   logger,
		handlers: make([]ResourceHandler, 0),
	}

	// Register all available resources
	registry.registerResources()

	return registry
}

// registerResources registers all available resource handlers
func (r *Registry) registerResources() {
	// Agent information resource
	r.handlers = append(r.handlers, NewAgentResource(r.client, r.logger))

	// Ships list resource
	r.handlers = append(r.handlers, NewShipsResource(r.client, r.logger))

	// Contracts list resource
	r.handlers = append(r.handlers, NewContractsResource(r.client, r.logger))

	// System waypoints resource
	r.handlers = append(r.handlers, NewWaypointsResource(r.client, r.logger))

	// Shipyard resource
	r.handlers = append(r.handlers, NewShipyardResource(r.client, r.logger))

	// Market resource
	r.handlers = append(r.handlers, NewMarketResource(r.client, r.logger))

	// Systems resource
	r.handlers = append(r.handlers, NewSystemsResource(r.client, r.logger))

	// Factions resource
	r.handlers = append(r.handlers, NewFactionsResource(r.client, r.logger))

	// Individual ship resource
	r.handlers = append(r.handlers, NewShipResource(r.client, r.logger))

	// Ship cooldown resource
	r.handlers = append(r.handlers, NewShipCooldownResource(r.client, r.logger))
}

// RegisterWithServer registers all resources with the MCP server
func (r *Registry) RegisterWithServer(s *server.MCPServer) {
	for _, handler := range r.handlers {
		s.AddResource(handler.Resource(), handler.Handler())
	}
}

// GetResources returns all registered resources (useful for testing/debugging)
func (r *Registry) GetResources() []mcp.Resource {
	resources := make([]mcp.Resource, len(r.handlers))
	for i, handler := range r.handlers {
		resources[i] = handler.Resource()
	}
	return resources
}
