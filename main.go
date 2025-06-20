package main

import (
    "context"
    "fmt"
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
)

func main() {
    // Create a new MCP server
    s := server.NewMCPServer(
        "SpaceTraders MCP Server",
        "1.0.0",
        server.WithToolCapabilities(false),
    )

    // Add tools for SpaceTraders API interactions
    // (we'll implement these)

    // Start the stdio server
    if err := server.ServeStdio(s); err != nil {
        fmt.Printf("Server error: %v\n", err)
    }
}
