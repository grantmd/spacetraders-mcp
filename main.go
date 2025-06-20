package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/viper"
)

// Agent represents the SpaceTraders agent data structure
type Agent struct {
	AccountID       string `json:"accountId"`
	Symbol          string `json:"symbol"`
	Headquarters    string `json:"headquarters"`
	Credits         int64  `json:"credits"`
	StartingFaction string `json:"startingFaction"`
	ShipCount       int    `json:"shipCount"`
}

// AgentResponse represents the API response structure
type AgentResponse struct {
	Data Agent `json:"data"`
}

// SpaceTradersClient handles API interactions
type SpaceTradersClient struct {
	APIToken string
	BaseURL  string
}

// NewSpaceTradersClient creates a new client instance
func NewSpaceTradersClient() *SpaceTradersClient {
	token := viper.GetString("SPACETRADERS_API_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "Warning: SPACETRADERS_API_TOKEN not found in configuration")
	}

	return &SpaceTradersClient{
		APIToken: token,
		BaseURL:  "https://api.spacetraders.io/v2",
	}
}

// GetAgent fetches agent information from the SpaceTraders API
func (c *SpaceTradersClient) GetAgent() (*Agent, error) {
	url := fmt.Sprintf("%s/my/agent", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var agentResp AgentResponse
	if err := json.Unmarshal(body, &agentResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &agentResp.Data, nil
}

// initConfig initializes Viper configuration
func initConfig() {
	// Set config name and type
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	// Add current directory to search path
	viper.AddConfigPath(".")

	// Enable automatic environment variable binding
	viper.AutomaticEnv()

	// Try to read the config file (silently)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error since we can use env vars
			// Silent - no logging needed for normal operation
		} else {
			// Config file was found but another error was produced
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		}
	}
	// Silent success - no logging needed for normal operation
}

func main() {
	// Initialize configuration with Viper
	initConfig()

	// Create SpaceTraders client
	client := NewSpaceTradersClient()

	// Create a new MCP server with resource capabilities
	s := server.NewMCPServer(
		"SpaceTraders MCP Server",
		"1.0.0",
		server.WithResourceCapabilities(false, false),
	)

	// Add the agent info resource
	agentResource := mcp.Resource{
		URI:         "spacetraders://agent/info",
		Name:        "Agent Information",
		Description: "Current agent information including credits, headquarters, faction, and ship count",
		MIMEType:    "application/json",
	}

	s.AddResource(agentResource, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// Validate the resource URI
		if request.Params.URI != "spacetraders://agent/info" {
			return nil, fmt.Errorf("unknown resource URI: %s", request.Params.URI)
		}

		// Get agent information from the API
		agent, err := client.GetAgent()
		if err != nil {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     fmt.Sprintf("Error fetching agent info: %v", err),
				},
			}, nil
		}

		// Format the response as structured JSON
		result := map[string]interface{}{
			"agent": map[string]interface{}{
				"accountId":       agent.AccountID,
				"symbol":          agent.Symbol,
				"headquarters":    agent.Headquarters,
				"credits":         agent.Credits,
				"startingFaction": agent.StartingFaction,
				"shipCount":       agent.ShipCount,
			},
		}

		// Convert to JSON for response
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Error formatting agent information",
				},
			}, nil
		}

		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		}, nil
	})

	// Start the stdio server (ServeStdio already handles signals gracefully)
	if err := server.ServeStdio(s); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
	}
}
