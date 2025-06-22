package spacetraders

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	token := "test-token"
	client := NewClient(token)

	if client.APIToken != token {
		t.Errorf("Expected APIToken %s, got %s", token, client.APIToken)
	}

	expectedBaseURL := "https://api.spacetraders.io/v2"
	if client.BaseURL != expectedBaseURL {
		t.Errorf("Expected BaseURL %s, got %s", expectedBaseURL, client.BaseURL)
	}
}

func TestClient_GetAgent(t *testing.T) {
	// Mock successful response
	mockAgent := Agent{
		AccountID:       "test-account-id",
		Symbol:          "TEST_AGENT",
		Headquarters:    "X1-TEST-HQ",
		Credits:         100000,
		StartingFaction: "TEST_FACTION",
		ShipCount:       3,
	}

	mockResponse := AgentResponse{
		Data: mockAgent,
	}

	responseJSON, err := json.Marshal(mockResponse)
	if err != nil {
		t.Fatalf("Failed to marshal mock response: %v", err)
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/my/agent" {
			t.Errorf("Expected path /my/agent, got %s", r.URL.Path)
		}

		// Check authorization header
		authHeader := r.Header.Get("Authorization")
		expectedAuth := "Bearer test-token"
		if authHeader != expectedAuth {
			t.Errorf("Expected Authorization header %s, got %s", expectedAuth, authHeader)
		}

		// Check content type
		contentType := r.Header.Get("Content-Type")
		expectedContentType := "application/json"
		if contentType != expectedContentType {
			t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(responseJSON); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with test server URL
	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	// Test GetAgent
	agent, err := client.GetAgent()
	if err != nil {
		t.Fatalf("GetAgent returned error: %v", err)
	}

	// Verify response
	if agent.AccountID != mockAgent.AccountID {
		t.Errorf("Expected AccountID %s, got %s", mockAgent.AccountID, agent.AccountID)
	}
	if agent.Symbol != mockAgent.Symbol {
		t.Errorf("Expected Symbol %s, got %s", mockAgent.Symbol, agent.Symbol)
	}
	if agent.Credits != mockAgent.Credits {
		t.Errorf("Expected Credits %d, got %d", mockAgent.Credits, agent.Credits)
	}
	if agent.ShipCount != mockAgent.ShipCount {
		t.Errorf("Expected ShipCount %d, got %d", mockAgent.ShipCount, agent.ShipCount)
	}
}

func TestClient_GetAgent_Error(t *testing.T) {
	// Test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		if _, err := w.Write([]byte(`{"error": "Unauthorized"}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "invalid-token",
		BaseURL:  server.URL,
	}

	agent, err := client.GetAgent()
	if err == nil {
		t.Fatal("Expected error for unauthorized request, got nil")
	}
	if agent != nil {
		t.Error("Expected nil agent on error, got non-nil")
	}
}

func TestClient_GetShips(t *testing.T) {
	// Mock ships response
	mockShips := []Ship{
		{
			Symbol: "TEST_SHIP_1",
			Registration: Registration{
				Name:          "Test Ship 1",
				FactionSymbol: "TEST_FACTION",
				Role:          "COMMAND",
			},
			Nav: Navigation{
				SystemSymbol:   "X1-TEST",
				WaypointSymbol: "X1-TEST-A1",
				Status:         "DOCKED",
				FlightMode:     "CRUISE",
			},
		},
		{
			Symbol: "TEST_SHIP_2",
			Registration: Registration{
				Name:          "Test Ship 2",
				FactionSymbol: "TEST_FACTION",
				Role:          "HAULER",
			},
			Nav: Navigation{
				SystemSymbol:   "X1-TEST",
				WaypointSymbol: "X1-TEST-B2",
				Status:         "IN_ORBIT",
				FlightMode:     "CRUISE",
			},
		},
	}

	mockResponse := ShipsResponse{
		Data: mockShips,
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
		if r.URL.Path != "/my/ships" {
			t.Errorf("Expected path /my/ships, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(responseJSON); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	// Test GetShips
	ships, err := client.GetShips()
	if err != nil {
		t.Fatalf("GetShips returned error: %v", err)
	}

	// Verify response
	if len(ships) != 2 {
		t.Errorf("Expected 2 ships, got %d", len(ships))
	}

	if ships[0].Symbol != "TEST_SHIP_1" {
		t.Errorf("Expected first ship symbol TEST_SHIP_1, got %s", ships[0].Symbol)
	}

	if ships[1].Symbol != "TEST_SHIP_2" {
		t.Errorf("Expected second ship symbol TEST_SHIP_2, got %s", ships[1].Symbol)
	}
}

func TestClient_GetContracts(t *testing.T) {
	// Mock contracts response
	mockContracts := []Contract{
		{
			ID:            "test-contract-1",
			FactionSymbol: "TEST_FACTION",
			Type:          "PROCUREMENT",
			Terms: ContractTerms{
				Deadline: "2025-12-31T23:59:59.000Z",
				Payment: ContractPayment{
					OnAccepted:  1000,
					OnFulfilled: 5000,
				},
				Deliver: []ContractDeliverGood{
					{
						TradeSymbol:       "IRON_ORE",
						DestinationSymbol: "X1-TEST-DEST",
						UnitsRequired:     100,
						UnitsFulfilled:    0,
					},
				},
			},
			Accepted:         false,
			Fulfilled:        false,
			Expiration:       "2025-12-31T23:59:59.000Z",
			DeadlineToAccept: "2025-12-25T23:59:59.000Z",
		},
	}

	mockResponse := ContractsResponse{
		Data: mockContracts,
		Meta: struct {
			Total int `json:"total"`
			Page  int `json:"page"`
			Limit int `json:"limit"`
		}{
			Total: 1,
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
		if r.URL.Path != "/my/contracts" {
			t.Errorf("Expected path /my/contracts, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(responseJSON); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	// Test GetContracts
	contracts, err := client.GetContracts()
	if err != nil {
		t.Fatalf("GetContracts returned error: %v", err)
	}

	// Verify response
	if len(contracts) != 1 {
		t.Errorf("Expected 1 contract, got %d", len(contracts))
	}

	contract := contracts[0]
	if contract.ID != "test-contract-1" {
		t.Errorf("Expected contract ID test-contract-1, got %s", contract.ID)
	}

	if contract.Type != "PROCUREMENT" {
		t.Errorf("Expected contract type PROCUREMENT, got %s", contract.Type)
	}

	if len(contract.Terms.Deliver) != 1 {
		t.Errorf("Expected 1 delivery requirement, got %d", len(contract.Terms.Deliver))
	}

	delivery := contract.Terms.Deliver[0]
	if delivery.TradeSymbol != "IRON_ORE" {
		t.Errorf("Expected trade symbol IRON_ORE, got %s", delivery.TradeSymbol)
	}

	if delivery.UnitsRequired != 100 {
		t.Errorf("Expected 100 units required, got %d", delivery.UnitsRequired)
	}
}

func TestClient_makeRequest_InvalidURL(t *testing.T) {
	client := &Client{
		APIToken: "test-token",
		BaseURL:  "://invalid-url",
	}

	_, err := client.makeRequest("GET", "/test", nil)
	if err == nil {
		t.Fatal("Expected error for invalid URL, got nil")
	}
}
