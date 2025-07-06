package contract

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"spacetraders-mcp/pkg/client"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestAcceptContractTool_Tool(t *testing.T) {
	client := client.NewClient("test-token")
	tool := NewAcceptContractTool(client)

	mcpTool := tool.Tool()

	// Test tool definition
	if mcpTool.Name != "accept_contract" {
		t.Errorf("Expected tool name 'accept_contract', got '%s'", mcpTool.Name)
	}

	if mcpTool.Description == "" {
		t.Error("Expected tool description to be non-empty")
	}

	// Test schema structure
	if mcpTool.InputSchema.Type != "object" {
		t.Errorf("Expected input schema type 'object', got '%s'", mcpTool.InputSchema.Type)
	}

	properties, ok := mcpTool.InputSchema.Properties["contract_id"]
	if !ok {
		t.Error("Expected 'contract_id' property in input schema")
	}

	propertyMap, ok := properties.(map[string]interface{})
	if !ok {
		t.Error("Expected contract_id property to be a map")
	}

	if propertyMap["type"] != "string" {
		t.Errorf("Expected contract_id type to be 'string', got '%v'", propertyMap["type"])
	}

	// Test required fields
	found := false
	for _, required := range mcpTool.InputSchema.Required {
		if required == "contract_id" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'contract_id' to be required")
	}
}

func TestAcceptContractTool_Handler_Success(t *testing.T) {
	// Mock successful contract acceptance
	mockContract := client.Contract{
		ID:            "test-contract-123",
		FactionSymbol: "COSMIC",
		Type:          "PROCUREMENT",
		Accepted:      true,
		Fulfilled:     false,
		Expiration:    "2024-12-31T23:59:59Z",
		Terms: client.ContractTerms{
			Deadline: "2024-12-30T23:59:59Z",
			Payment: client.ContractPayment{
				OnAccepted:  10000,
				OnFulfilled: 50000,
			},
			Deliver: []client.ContractDeliverGood{
				{
					TradeSymbol:       "IRON_ORE",
					DestinationSymbol: "X1-TEST-STATION",
					UnitsRequired:     100,
					UnitsFulfilled:    0,
				},
			},
		},
	}

	mockAgent := client.Agent{
		Symbol:          "TEST_AGENT",
		Credits:         110000,
		ShipCount:       1,
		StartingFaction: "COSMIC",
	}

	mockResponse := client.AcceptContractResponse{
		Data: struct {
			Contract client.Contract `json:"contract"`
			Agent    client.Agent    `json:"agent"`
		}{
			Contract: mockContract,
			Agent:    mockAgent,
		},
	}

	responseJSON, err := json.Marshal(mockResponse)
	if err != nil {
		t.Fatalf("Failed to marshal mock response: %v", err)
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		expectedPath := "/my/contracts/test-contract-123/accept"
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
	client := &client.Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	tool := NewAcceptContractTool(client)
	handler := tool.Handler()

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "accept_contract",
			Arguments: map[string]interface{}{
				"contract_id": "test-contract-123",
			},
		},
	}

	result, err := handler(context.Background(), request)

	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if result.IsError {
		t.Fatalf("Handler returned error result: %v", result.Content)
	}

	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(result.Content))
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("Expected TextContent, got %T", result.Content[0])
	}

	// Parse the JSON response
	var response map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Verify response structure
	if response["success"] != true {
		t.Error("Expected success to be true")
	}

	if response["message"] != "Successfully accepted contract test-contract-123" {
		t.Errorf("Unexpected message: %v", response["message"])
	}

	// Verify contract data
	contractData, ok := response["contract"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected contract data to be a map")
	}

	if contractData["id"] != "test-contract-123" {
		t.Errorf("Expected contract ID 'test-contract-123', got '%v'", contractData["id"])
	}

	if contractData["accepted"] != true {
		t.Error("Expected contract to be accepted")
	}

	// Verify agent data
	agentData, ok := response["agent"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected agent data to be a map")
	}

	if agentData["symbol"] != "TEST_AGENT" {
		t.Errorf("Expected agent symbol 'TEST_AGENT', got '%v'", agentData["symbol"])
	}

	if agentData["credits"] != float64(110000) {
		t.Errorf("Expected agent credits 110000, got %v", agentData["credits"])
	}
}

func TestAcceptContractTool_Handler_APIError(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(`{"error": {"message": "Contract not found"}}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &client.Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	tool := NewAcceptContractTool(client)
	handler := tool.Handler()

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "accept_contract",
			Arguments: map[string]interface{}{
				"contract_id": "nonexistent-contract",
			},
		},
	}

	result, err := handler(context.Background(), request)

	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if !result.IsError {
		t.Fatal("Expected error result")
	}

	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(result.Content))
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("Expected TextContent, got %T", result.Content[0])
	}

	if !contains(textContent.Text, "Failed to accept contract") {
		t.Errorf("Expected error message to contain 'Failed to accept contract', got '%s'", textContent.Text)
	}
}

func TestAcceptContractTool_Handler_MissingContractID(t *testing.T) {
	client := client.NewClient("test-token")
	tool := NewAcceptContractTool(client)
	handler := tool.Handler()

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "accept_contract",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := handler(context.Background(), request)

	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if !result.IsError {
		t.Fatal("Expected error result")
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("Expected TextContent, got %T", result.Content[0])
	}

	if !contains(textContent.Text, "contract_id must be a valid string") {
		t.Errorf("Expected error message to contain 'contract_id must be a valid string', got '%s'", textContent.Text)
	}
}

func TestAcceptContractTool_Handler_EmptyContractID(t *testing.T) {
	client := client.NewClient("test-token")
	tool := NewAcceptContractTool(client)
	handler := tool.Handler()

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "accept_contract",
			Arguments: map[string]interface{}{
				"contract_id": "",
			},
		},
	}

	result, err := handler(context.Background(), request)

	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if !result.IsError {
		t.Fatal("Expected error result")
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("Expected TextContent, got %T", result.Content[0])
	}

	expectedError := "contract_id cannot be empty"
	if textContent.Text != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, textContent.Text)
	}
}

func TestAcceptContractTool_Handler_InvalidContractIDType(t *testing.T) {
	client := client.NewClient("test-token")
	tool := NewAcceptContractTool(client)
	handler := tool.Handler()

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "accept_contract",
			Arguments: map[string]interface{}{
				"contract_id": 123, // Invalid type - should be string
			},
		},
	}

	result, err := handler(context.Background(), request)

	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if !result.IsError {
		t.Fatal("Expected error result")
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("Expected TextContent, got %T", result.Content[0])
	}

	if !contains(textContent.Text, "contract_id must be a valid string") {
		t.Errorf("Expected error message to contain 'contract_id must be a valid string', got '%s'", textContent.Text)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[0:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			indexOfSubstring(s, substr) >= 0)))
}

// Simple substring search helper
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
