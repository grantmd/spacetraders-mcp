package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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
	token := os.Getenv("SPACETRADERS_API_TOKEN")
	if token == "" {
		fmt.Println("Warning: SPACETRADERS_API_TOKEN not found in environment")
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

// loadEnvFile loads environment variables from .env file
func loadEnvFile() {
	if _, err := os.Stat(".env"); err != nil {
		return // .env file doesn't exist
	}

	data, err := os.ReadFile(".env")
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		// Find the first = character
		eqIndex := strings.Index(line, "=")
		if eqIndex == -1 {
			continue
		}

		key := strings.TrimSpace(line[:eqIndex])
		value := strings.TrimSpace(line[eqIndex+1:])

		// Remove surrounding quotes if present
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		os.Setenv(key, value)
	}
}

func main() {
	// Load environment variables from .env file if it exists
	loadEnvFile()

	// Create SpaceTraders client
	client := NewSpaceTradersClient()

	// Create a new MCP server
	s := server.NewMCPServer(
		"SpaceTraders MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Add the get_agent_info tool
	getAgentInfoTool := mcp.Tool{
		Name:        "get_agent_info",
		Description: "Retrieves the current agent's information from SpaceTraders API, including credits, headquarters, faction, and ship count.",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}

	s.AddTool(getAgentInfoTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get agent information from the API
		agent, err := client.GetAgent()
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to get agent info", err), nil
		}

		// Format the response as structured data
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

		// Convert to JSON for pretty formatting
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("Failed to format response"), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
