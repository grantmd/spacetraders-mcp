package main

import (
	"context"
	"fmt"
	"os"

	"spacetraders-mcp/pkg/config"
	"spacetraders-mcp/pkg/resources"
	"spacetraders-mcp/pkg/spacetraders"
	"spacetraders-mcp/pkg/tools"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Create SpaceTraders client
	client := spacetraders.NewClient(cfg.SpaceTradersAPIToken)

	// Create MCP server with resource capabilities
	s := server.NewMCPServer(
		"SpaceTraders MCP Server",
		"1.0.0",
		server.WithResourceCapabilities(false, false), // subscribe=false, listChanged=false
	)

	// Register all resources
	resourceRegistry := resources.NewRegistry(client)
	resourceRegistry.RegisterWithServer(s)

	// Register all tools (when we have them)
	toolRegistry := tools.NewRegistry(client)
	toolRegistry.RegisterWithServer(s)

	// Start the stdio server (ServeStdio already handles signals gracefully)
	if err := server.ServeStdio(s); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
	}
}
