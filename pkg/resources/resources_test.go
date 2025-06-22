package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func createMockLogger() *logging.Logger {
	// Create a mock logger that doesn't require an MCP server
	return logging.NewLogger(nil)
}

func TestAgentResource_Resource(t *testing.T) {
	client := spacetraders.NewClient("test-token")
	logger := createMockLogger()
	resource := NewAgentResource(client, logger)

	mcpResource := resource.Resource()

	expectedURI := "spacetraders://agent/info"
	if mcpResource.URI != expectedURI {
		t.Errorf("Expected URI %s, got %s", expectedURI, mcpResource.URI)
	}

	expectedName := "Agent Information"
	if mcpResource.Name != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, mcpResource.Name)
	}

	expectedMIMEType := "application/json"
	if mcpResource.MIMEType != expectedMIMEType {
		t.Errorf("Expected MIME type %s, got %s", expectedMIMEType, mcpResource.MIMEType)
	}
}

func TestAgentResource_Handler_Success(t *testing.T) {
	// Mock successful agent response
	mockAgent := spacetraders.Agent{
		AccountID:       "test-account",
		Symbol:          "TEST_AGENT",
		Headquarters:    "X1-TEST-HQ",
		Credits:         50000,
		StartingFaction: "TEST_FACTION",
		ShipCount:       2,
	}

	mockResponse := spacetraders.AgentResponse{
		Data: mockAgent,
	}

	responseJSON, err := json.Marshal(mockResponse)
	if err != nil {
		t.Fatalf("Failed to marshal mock response: %v", err)
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(responseJSON); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with test server
	client := &spacetraders.Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}
	logger := createMockLogger()
	resource := NewAgentResource(client, logger)

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "spacetraders://agent/info",
		},
	}

	// Test handler
	handler := resource.Handler()
	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify result
	if len(result) != 1 {
		t.Fatalf("Expected 1 resource content, got %d", len(result))
	}

	content, ok := result[0].(*mcp.TextResourceContents)
	if !ok {
		t.Fatalf("Expected TextResourceContents, got %T", result[0])
	}

	if content.URI != "spacetraders://agent/info" {
		t.Errorf("Expected URI spacetraders://agent/info, got %s", content.URI)
	}

	if content.MIMEType != "application/json" {
		t.Errorf("Expected MIME type application/json, got %s", content.MIMEType)
	}

	// Parse and verify JSON content
	var jsonResult map[string]interface{}
	err = json.Unmarshal([]byte(content.Text), &jsonResult)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	agent, ok := jsonResult["agent"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected agent object in response")
	}

	if agent["symbol"] != "TEST_AGENT" {
		t.Errorf("Expected agent symbol TEST_AGENT, got %v", agent["symbol"])
	}

	if agent["credits"] != float64(50000) {
		t.Errorf("Expected credits 50000, got %v", agent["credits"])
	}
}

func TestAgentResource_Handler_InvalidURI(t *testing.T) {
	client := spacetraders.NewClient("test-token")
	logger := createMockLogger()
	resource := NewAgentResource(client, logger)

	// Create test request with invalid URI
	request := mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "spacetraders://invalid/uri",
		},
	}

	// Test handler
	handler := resource.Handler()
	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify error response
	if len(result) != 1 {
		t.Fatalf("Expected 1 resource content, got %d", len(result))
	}

	content, ok := result[0].(*mcp.TextResourceContents)
	if !ok {
		t.Fatalf("Expected TextResourceContents, got %T", result[0])
	}

	if content.MIMEType != "text/plain" {
		t.Errorf("Expected MIME type text/plain for error, got %s", content.MIMEType)
	}

	if content.Text != "Invalid resource URI" {
		t.Errorf("Expected error message 'Invalid resource URI', got %s", content.Text)
	}
}

func TestShipsResource_Resource(t *testing.T) {
	client := spacetraders.NewClient("test-token")
	logger := createMockLogger()
	resource := NewShipsResource(client, logger)

	mcpResource := resource.Resource()

	expectedURI := "spacetraders://ships/list"
	if mcpResource.URI != expectedURI {
		t.Errorf("Expected URI %s, got %s", expectedURI, mcpResource.URI)
	}

	expectedName := "Ships List"
	if mcpResource.Name != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, mcpResource.Name)
	}
}

func TestShipsResource_Handler_Success(t *testing.T) {
	// Mock ships response
	mockShips := []spacetraders.Ship{
		{
			Symbol: "TEST_SHIP_1",
			Registration: spacetraders.Registration{
				Name:          "Test Ship 1",
				FactionSymbol: "TEST_FACTION",
				Role:          "COMMAND",
			},
		},
		{
			Symbol: "TEST_SHIP_2",
			Registration: spacetraders.Registration{
				Name:          "Test Ship 2",
				FactionSymbol: "TEST_FACTION",
				Role:          "HAULER",
			},
		},
	}

	mockResponse := spacetraders.ShipsResponse{
		Data: mockShips,
	}

	responseJSON, err := json.Marshal(mockResponse)
	if err != nil {
		t.Fatalf("Failed to marshal mock response: %v", err)
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(responseJSON); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with test server
	client := &spacetraders.Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}
	logger := createMockLogger()
	resource := NewShipsResource(client, logger)

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "spacetraders://ships/list",
		},
	}

	// Test handler
	handler := resource.Handler()
	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify result
	if len(result) != 1 {
		t.Fatalf("Expected 1 resource content, got %d", len(result))
	}

	content, ok := result[0].(*mcp.TextResourceContents)
	if !ok {
		t.Fatalf("Expected TextResourceContents, got %T", result[0])
	}

	// Parse and verify JSON content
	var jsonResult map[string]interface{}
	err = json.Unmarshal([]byte(content.Text), &jsonResult)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	ships, ok := jsonResult["ships"].([]interface{})
	if !ok {
		t.Fatal("Expected ships array in response")
	}

	if len(ships) != 2 {
		t.Errorf("Expected 2 ships, got %d", len(ships))
	}

	meta, ok := jsonResult["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected meta object in response")
	}

	if meta["count"] != float64(2) {
		t.Errorf("Expected count 2, got %v", meta["count"])
	}
}

func TestContractsResource_Resource(t *testing.T) {
	client := spacetraders.NewClient("test-token")
	logger := createMockLogger()
	resource := NewContractsResource(client, logger)

	mcpResource := resource.Resource()

	expectedURI := "spacetraders://contracts/list"
	if mcpResource.URI != expectedURI {
		t.Errorf("Expected URI %s, got %s", expectedURI, mcpResource.URI)
	}

	expectedName := "Contracts List"
	if mcpResource.Name != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, mcpResource.Name)
	}
}

func TestRegistry_NewRegistry(t *testing.T) {
	client := spacetraders.NewClient("test-token")
	logger := createMockLogger()
	registry := NewRegistry(client, logger)

	if registry == nil {
		t.Fatal("Expected non-nil registry")
	}

	if registry.client != client {
		t.Error("Registry client not set correctly")
	}

	if registry.logger != logger {
		t.Error("Registry logger not set correctly")
	}

	// Verify resources are registered
	resources := registry.GetResources()
	if len(resources) < 3 {
		t.Errorf("Expected at least 3 resources, got %d", len(resources))
	}

	// Check for expected resource URIs
	expectedURIs := map[string]bool{
		"spacetraders://agent/info":     false,
		"spacetraders://ships/list":     false,
		"spacetraders://contracts/list": false,
	}

	for _, resource := range resources {
		if _, exists := expectedURIs[resource.URI]; exists {
			expectedURIs[resource.URI] = true
		}
	}

	for uri, found := range expectedURIs {
		if !found {
			t.Errorf("Expected resource URI %s not found", uri)
		}
	}
}

func TestRegistry_RegisterWithServer(t *testing.T) {
	client := spacetraders.NewClient("test-token")
	logger := createMockLogger()
	registry := NewRegistry(client, logger)

	// Create a test MCP server
	s := server.NewMCPServer(
		"Test Server",
		"1.0.0",
		server.WithResourceCapabilities(false, false),
	)

	// This should not panic
	registry.RegisterWithServer(s)

	// Basic verification that registration completed without error
	resources := registry.GetResources()
	if len(resources) == 0 {
		t.Error("No resources found after registration")
	}
}

func TestResourceHandler_Interface(t *testing.T) {
	client := spacetraders.NewClient("test-token")
	logger := createMockLogger()

	// Verify all resource types implement ResourceHandler interface
	var _ ResourceHandler = NewAgentResource(client, logger)
	var _ ResourceHandler = NewShipsResource(client, logger)
	var _ ResourceHandler = NewContractsResource(client, logger)
}
