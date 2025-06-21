package resources

import (
	"context"
	"spacetraders-mcp/pkg/spacetraders"

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
	client   *spacetraders.Client
	handlers []ResourceHandler
}

// NewRegistry creates a new resource registry
func NewRegistry(client *spacetraders.Client) *Registry {
	registry := &Registry{
		client:   client,
		handlers: make([]ResourceHandler, 0),
	}

	// Register all available resources
	registry.registerResources()

	return registry
}

// registerResources registers all available resource handlers
func (r *Registry) registerResources() {
	// Agent information resource
	r.handlers = append(r.handlers, NewAgentResource(r.client))

	// Ships list resource
	r.handlers = append(r.handlers, NewShipsResource(r.client))

	// Contracts list resource
	r.handlers = append(r.handlers, NewContractsResource(r.client))

	// TODO: Add more resources here as we implement them:
	// - Systems resource
	// - Waypoints resource
	// - Markets resource
	// - Faction resource
	// etc.
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
