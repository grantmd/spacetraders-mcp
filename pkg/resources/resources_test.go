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

func TestWaypointsResource_Resource(t *testing.T) {
	client := spacetraders.NewClient("test-token")
	logger := createMockLogger()
	resource := NewWaypointsResource(client, logger)

	mcpResource := resource.Resource()

	expectedURI := "spacetraders://systems/{systemSymbol}/waypoints"
	if mcpResource.URI != expectedURI {
		t.Errorf("Expected URI %s, got %s", expectedURI, mcpResource.URI)
	}

	expectedName := "System Waypoints"
	if mcpResource.Name != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, mcpResource.Name)
	}

	expectedMIMEType := "application/json"
	if mcpResource.MIMEType != expectedMIMEType {
		t.Errorf("Expected MIME type %s, got %s", expectedMIMEType, mcpResource.MIMEType)
	}
}

func TestWaypointsResource_Handler_Success(t *testing.T) {
	// Mock successful waypoints response
	mockWaypoints := []spacetraders.SystemWaypoint{
		{
			Symbol: "X1-TEST-A1",
			Type:   "PLANET",
			X:      10,
			Y:      20,
			Traits: []spacetraders.WaypointTrait{
				{
					Symbol:      "MARKETPLACE",
					Name:        "Marketplace",
					Description: "A thriving marketplace",
				},
			},
		},
		{
			Symbol: "X1-TEST-B2",
			Type:   "MOON",
			X:      15,
			Y:      25,
			Traits: []spacetraders.WaypointTrait{
				{
					Symbol:      "SHIPYARD",
					Name:        "Shipyard",
					Description: "Shipyard for purchasing ships",
				},
			},
		},
	}

	mockResponse := spacetraders.SystemWaypointsResponse{
		Data: mockWaypoints,
		Meta: struct {
			Total int `json:"total"`
			Page  int `json:"page"`
			Limit int `json:"limit"`
		}{
			Total: 2,
			Page:  1,
			Limit: 20,
		},
	}

	responseJSON, err := json.Marshal(mockResponse)
	if err != nil {
		t.Fatalf("Failed to marshal mock response: %v", err)
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/systems/X1-TEST/waypoints" {
			t.Errorf("Expected path /systems/X1-TEST/waypoints, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(responseJSON); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with test server URL
	client := &spacetraders.Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	logger := createMockLogger()
	resource := NewWaypointsResource(client, logger)
	handler := resource.Handler()

	// Test the handler
	request := mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "spacetraders://systems/X1-TEST/waypoints",
		},
	}

	contents, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if len(contents) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(contents))
	}

	textContent, ok := contents[0].(*mcp.TextResourceContents)
	if !ok {
		t.Fatalf("Expected TextResourceContents, got %T", contents[0])
	}

	if textContent.MIMEType != "application/json" {
		t.Errorf("Expected MIME type application/json, got %s", textContent.MIMEType)
	}

	// Parse the JSON response to verify structure
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &result); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Verify response structure
	if result["system"] != "X1-TEST" {
		t.Errorf("Expected system X1-TEST, got %v", result["system"])
	}

	waypoints, ok := result["waypoints"].([]interface{})
	if !ok {
		t.Fatal("Expected waypoints to be an array")
	}

	if len(waypoints) != 2 {
		t.Errorf("Expected 2 waypoints, got %d", len(waypoints))
	}
}

func TestWaypointsResource_Handler_InvalidURI(t *testing.T) {
	client := spacetraders.NewClient("test-token")
	logger := createMockLogger()
	resource := NewWaypointsResource(client, logger)
	handler := resource.Handler()

	request := mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "invalid://uri",
		},
	}

	contents, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if len(contents) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(contents))
	}

	textContent, ok := contents[0].(*mcp.TextResourceContents)
	if !ok {
		t.Fatalf("Expected TextResourceContents, got %T", contents[0])
	}

	if textContent.MIMEType != "text/plain" {
		t.Errorf("Expected MIME type text/plain, got %s", textContent.MIMEType)
	}

	if !contains(textContent.Text, "Invalid resource URI") {
		t.Errorf("Expected error message about invalid URI, got: %s", textContent.Text)
	}
}

func TestShipyardResource_Resource(t *testing.T) {
	client := spacetraders.NewClient("test-token")
	logger := createMockLogger()
	resource := NewShipyardResource(client, logger)

	mcpResource := resource.Resource()

	expectedURI := "spacetraders://systems/{systemSymbol}/waypoints/{waypointSymbol}/shipyard"
	if mcpResource.URI != expectedURI {
		t.Errorf("Expected URI %s, got %s", expectedURI, mcpResource.URI)
	}

	expectedName := "Shipyard Information"
	if mcpResource.Name != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, mcpResource.Name)
	}

	expectedMIMEType := "application/json"
	if mcpResource.MIMEType != expectedMIMEType {
		t.Errorf("Expected MIME type %s, got %s", expectedMIMEType, mcpResource.MIMEType)
	}
}

func TestShipyardResource_Handler_Success(t *testing.T) {
	// Mock successful shipyard response
	mockShipyard := spacetraders.Shipyard{
		Symbol: "X1-TEST-SHIPYARD",
		ShipTypes: []spacetraders.ShipyardShipType{
			{Type: "SHIP_PROBE"},
		},
		Ships: []spacetraders.ShipyardShip{
			{
				Type:          "SHIP_PROBE",
				Name:          "Probe",
				Description:   "A small exploration vessel",
				Supply:        "ABUNDANT",
				PurchasePrice: 50000,
				Frame: spacetraders.ShipyardShipFrame{
					Symbol:         "FRAME_PROBE",
					Name:           "Probe Frame",
					Description:    "Small frame for probe ships",
					ModuleSlots:    2,
					MountingPoints: 1,
					FuelCapacity:   400,
				},
			},
		},
		ModificationsFee: 1000,
	}

	mockResponse := spacetraders.ShipyardResponse{
		Data: mockShipyard,
	}

	responseJSON, err := json.Marshal(mockResponse)
	if err != nil {
		t.Fatalf("Failed to marshal mock response: %v", err)
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/systems/X1-TEST/waypoints/X1-TEST-SHIPYARD/shipyard"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(responseJSON); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with test server URL
	client := &spacetraders.Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	logger := createMockLogger()
	resource := NewShipyardResource(client, logger)
	handler := resource.Handler()

	// Test the handler
	request := mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "spacetraders://systems/X1-TEST/waypoints/X1-TEST-SHIPYARD/shipyard",
		},
	}

	contents, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if len(contents) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(contents))
	}

	textContent, ok := contents[0].(*mcp.TextResourceContents)
	if !ok {
		t.Fatalf("Expected TextResourceContents, got %T", contents[0])
	}

	if textContent.MIMEType != "application/json" {
		t.Errorf("Expected MIME type application/json, got %s", textContent.MIMEType)
	}

	// Parse the JSON response to verify structure
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &result); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Verify response structure
	if result["system"] != "X1-TEST" {
		t.Errorf("Expected system X1-TEST, got %v", result["system"])
	}

	if result["waypoint"] != "X1-TEST-SHIPYARD" {
		t.Errorf("Expected waypoint X1-TEST-SHIPYARD, got %v", result["waypoint"])
	}

	shipyard, ok := result["shipyard"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected shipyard to be an object")
	}

	if shipyard["symbol"] != "X1-TEST-SHIPYARD" {
		t.Errorf("Expected shipyard symbol X1-TEST-SHIPYARD, got %v", shipyard["symbol"])
	}
}

func TestShipyardResource_Handler_InvalidURI(t *testing.T) {
	client := spacetraders.NewClient("test-token")
	logger := createMockLogger()
	resource := NewShipyardResource(client, logger)
	handler := resource.Handler()

	request := mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "invalid://uri",
		},
	}

	contents, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if len(contents) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(contents))
	}

	textContent, ok := contents[0].(*mcp.TextResourceContents)
	if !ok {
		t.Fatalf("Expected TextResourceContents, got %T", contents[0])
	}

	if textContent.MIMEType != "text/plain" {
		t.Errorf("Expected MIME type text/plain, got %s", textContent.MIMEType)
	}

	if !contains(textContent.Text, "Invalid resource URI") {
		t.Errorf("Expected error message about invalid URI, got: %s", textContent.Text)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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
