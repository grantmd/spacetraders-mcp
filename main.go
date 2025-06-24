package main

import (
	"context"
	"log"
	"os"

	"spacetraders-mcp/pkg/config"
	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/resources"
	"spacetraders-mcp/pkg/spacetraders"
	"spacetraders-mcp/pkg/tools"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Set up error logging
	errorLogger := log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		errorLogger.Printf("Configuration error: %v", err)
		os.Exit(1)
	}

	// Create SpaceTraders client
	client := spacetraders.NewClient(cfg.SpaceTradersAPIToken)

	// Create MCP server with resource and logging capabilities
	s := server.NewMCPServer(
		"SpaceTraders MCP Server",
		"1.0.0",
		server.WithResourceCapabilities(false, false), // subscribe=false, listChanged=false
		server.WithLogging(),                          // Enable MCP logging support
	)

	// Create application logger
	appLogger := logging.NewLogger(s)

	// Add logging support - send log messages to MCP client
	s.AddNotificationHandler("logging/setLevel", func(ctx context.Context, notification mcp.JSONRPCNotification) {
		errorLogger.Printf("Client requested logging level change: %+v", notification)
	})

	// Note: MCP framework handles resources/list and tools/list automatically
	// To see these calls, you would need to monitor the stdio communication directly
	appLogger.Debug("MCP server configured - resources/list and tools/list calls will be handled automatically")

	appLogger.Info("Starting SpaceTraders MCP Server")

	// Register all resources
	resourceRegistry := resources.NewRegistry(client, appLogger)
	resourceRegistry.RegisterWithServer(s)

	// Register all tools (when we have them)
	toolRegistry := tools.NewRegistry(client, appLogger)
	toolRegistry.RegisterWithServer(s)

	appLogger.Info("Server initialization complete")

	// Start the stdio server with error logging (ServeStdio already handles signals gracefully)
	if err := server.ServeStdio(s, server.WithErrorLogger(errorLogger)); err != nil && err != context.Canceled {
		errorLogger.Printf("Server error: %v", err)
	}
}
