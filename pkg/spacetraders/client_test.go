package spacetraders

import (
	"encoding/json"
	"io"
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

func TestClient_GetAllShips(t *testing.T) {
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

	// Test GetAllShips
	ships, err := client.GetAllShips()
	if err != nil {
		t.Fatalf("GetAllShips returned error: %v", err)
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

func TestClient_GetAllContracts(t *testing.T) {
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

	// Test GetAllContracts
	contracts, err := client.GetAllContracts()
	if err != nil {
		t.Fatalf("GetAllContracts returned error: %v", err)
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

func TestClient_AcceptContract(t *testing.T) {
	// Mock successful accept contract response
	mockContract := Contract{
		ID:            "test-contract-123",
		FactionSymbol: "COSMIC",
		Type:          "PROCUREMENT",
		Terms: ContractTerms{
			Deadline: "2025-12-30T23:59:59.000Z",
			Payment: ContractPayment{
				OnAccepted:  10000,
				OnFulfilled: 50000,
			},
			Deliver: []ContractDeliverGood{
				{
					TradeSymbol:       "IRON_ORE",
					DestinationSymbol: "X1-TEST-STATION",
					UnitsRequired:     100,
					UnitsFulfilled:    0,
				},
			},
		},
		Accepted:         true,
		Fulfilled:        false,
		Expiration:       "2025-12-31T23:59:59.000Z",
		DeadlineToAccept: "2025-12-25T23:59:59.000Z",
	}

	mockAgent := Agent{
		AccountID:       "test-account-id",
		Symbol:          "TEST_AGENT",
		Headquarters:    "X1-TEST-HQ",
		Credits:         110000, // Increased after accepting contract
		StartingFaction: "COSMIC",
		ShipCount:       1,
	}

	mockResponse := AcceptContractResponse{
		Data: struct {
			Contract Contract `json:"contract"`
			Agent    Agent    `json:"agent"`
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
		// Check request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		expectedPath := "/my/contracts/test-contract-123/accept"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
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

	// Test AcceptContract
	contract, agent, err := client.AcceptContract("test-contract-123")
	if err != nil {
		t.Fatalf("AcceptContract returned error: %v", err)
	}

	// Verify contract response
	if contract.ID != mockContract.ID {
		t.Errorf("Expected contract ID %s, got %s", mockContract.ID, contract.ID)
	}
	if contract.FactionSymbol != mockContract.FactionSymbol {
		t.Errorf("Expected faction symbol %s, got %s", mockContract.FactionSymbol, contract.FactionSymbol)
	}
	if contract.Type != mockContract.Type {
		t.Errorf("Expected contract type %s, got %s", mockContract.Type, contract.Type)
	}
	if !contract.Accepted {
		t.Error("Expected contract to be accepted")
	}
	if contract.Terms.Payment.OnAccepted != mockContract.Terms.Payment.OnAccepted {
		t.Errorf("Expected on accepted payment %d, got %d",
			mockContract.Terms.Payment.OnAccepted, contract.Terms.Payment.OnAccepted)
	}

	// Verify agent response
	if agent.Symbol != mockAgent.Symbol {
		t.Errorf("Expected agent symbol %s, got %s", mockAgent.Symbol, agent.Symbol)
	}
	if agent.Credits != mockAgent.Credits {
		t.Errorf("Expected agent credits %d, got %d", mockAgent.Credits, agent.Credits)
	}
	if agent.StartingFaction != mockAgent.StartingFaction {
		t.Errorf("Expected starting faction %s, got %s", mockAgent.StartingFaction, agent.StartingFaction)
	}
}

func TestClient_AcceptContract_NotFound(t *testing.T) {
	// Test server that returns 404 for contract not found
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(`{"error": {"message": "Contract not found"}}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	contract, agent, err := client.AcceptContract("nonexistent-contract")
	if err == nil {
		t.Fatal("Expected error for nonexistent contract, got nil")
	}
	if contract != nil {
		t.Error("Expected nil contract on error, got non-nil")
	}
	if agent != nil {
		t.Error("Expected nil agent on error, got non-nil")
	}

	// Check that the error message contains status code
	if !contains(err.Error(), "404") {
		t.Errorf("Expected error to contain '404', got: %s", err.Error())
	}
}

func TestClient_AcceptContract_AlreadyAccepted(t *testing.T) {
	// Test server that returns 409 for already accepted contract
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		if _, err := w.Write([]byte(`{"error": {"message": "Contract already accepted"}}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	contract, agent, err := client.AcceptContract("already-accepted-contract")
	if err == nil {
		t.Fatal("Expected error for already accepted contract, got nil")
	}
	if contract != nil {
		t.Error("Expected nil contract on error, got non-nil")
	}
	if agent != nil {
		t.Error("Expected nil agent on error, got non-nil")
	}

	// Check that the error message contains status code
	if !contains(err.Error(), "409") {
		t.Errorf("Expected error to contain '409', got: %s", err.Error())
	}
}

func TestClient_AcceptContract_InvalidJSON(t *testing.T) {
	// Test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{invalid json}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	contract, agent, err := client.AcceptContract("test-contract")
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
	if contract != nil {
		t.Error("Expected nil contract on error, got non-nil")
	}
	if agent != nil {
		t.Error("Expected nil agent on error, got non-nil")
	}

	// Check that the error message indicates parsing failure
	if !contains(err.Error(), "failed to parse response") {
		t.Errorf("Expected error to contain 'failed to parse response', got: %s", err.Error())
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

func TestClient_GetAllSystemWaypoints(t *testing.T) {
	// Mock waypoints response
	mockWaypoints := []SystemWaypoint{
		{
			Symbol: "X1-TEST-A1",
			Type:   "PLANET",
			X:      10,
			Y:      20,
			Orbitals: []WaypointOrbital{
				{Symbol: "X1-TEST-A1-M1"},
			},
			Traits: []WaypointTrait{
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
			Traits: []WaypointTrait{
				{
					Symbol:      "SHIPYARD",
					Name:        "Shipyard",
					Description: "Shipyard for purchasing ships",
				},
			},
		},
	}

	mockResponse := SystemWaypointsResponse{
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
		expectedPath := "/systems/X1-TEST/waypoints"
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

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	// Test GetAllSystemWaypoints
	waypoints, err := client.GetAllSystemWaypoints("X1-TEST")
	if err != nil {
		t.Fatalf("GetAllSystemWaypoints returned error: %v", err)
	}

	// Verify response
	if len(waypoints) != 2 {
		t.Errorf("Expected 2 waypoints, got %d", len(waypoints))
	}

	if waypoints[0].Symbol != "X1-TEST-A1" {
		t.Errorf("Expected first waypoint symbol X1-TEST-A1, got %s", waypoints[0].Symbol)
	}

	if waypoints[0].Type != "PLANET" {
		t.Errorf("Expected first waypoint type PLANET, got %s", waypoints[0].Type)
	}

	if len(waypoints[0].Traits) != 1 {
		t.Errorf("Expected 1 trait for first waypoint, got %d", len(waypoints[0].Traits))
	}

	if waypoints[0].Traits[0].Symbol != "MARKETPLACE" {
		t.Errorf("Expected first waypoint trait MARKETPLACE, got %s", waypoints[0].Traits[0].Symbol)
	}

	if waypoints[1].Symbol != "X1-TEST-B2" {
		t.Errorf("Expected second waypoint symbol X1-TEST-B2, got %s", waypoints[1].Symbol)
	}

	if waypoints[1].Traits[0].Symbol != "SHIPYARD" {
		t.Errorf("Expected second waypoint trait SHIPYARD, got %s", waypoints[1].Traits[0].Symbol)
	}
}

func TestClient_GetAllSystemWaypoints_Error(t *testing.T) {
	// Test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(`{"error": "System not found"}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	waypoints, err := client.GetAllSystemWaypoints("INVALID-SYSTEM")
	if err == nil {
		t.Fatal("Expected error for invalid system, got nil")
	}
	if waypoints != nil {
		t.Error("Expected nil waypoints on error, got non-nil")
	}
}

func TestClient_GetShipyard(t *testing.T) {
	// Mock shipyard response
	mockShipyard := Shipyard{
		Symbol: "X1-TEST-SHIPYARD",
		ShipTypes: []ShipyardShipType{
			{Type: "SHIP_PROBE"},
			{Type: "SHIP_MINING_DRONE"},
		},
		Ships: []ShipyardShip{
			{
				Type:          "SHIP_PROBE",
				Name:          "Probe",
				Description:   "A small exploration vessel",
				Supply:        "ABUNDANT",
				PurchasePrice: 50000,
				Frame: ShipyardShipFrame{
					Symbol:         "FRAME_PROBE",
					Name:           "Probe Frame",
					Description:    "Small frame for probe ships",
					ModuleSlots:    2,
					MountingPoints: 1,
					FuelCapacity:   400,
					Condition:      100,
					Integrity:      100,
				},
				Reactor: ShipyardShipReactor{
					Symbol:      "REACTOR_FISSION_I",
					Name:        "Fission Reactor I",
					Description: "Basic fission reactor",
					Condition:   100,
					Integrity:   100,
					PowerOutput: 31,
				},
				Engine: ShipyardShipEngine{
					Symbol:      "ENGINE_IMPULSE_DRIVE_I",
					Name:        "Impulse Drive I",
					Description: "Basic impulse drive",
					Condition:   100,
					Integrity:   100,
					Speed:       30,
				},
				Modules: []ShipyardShipModule{
					{
						Symbol:      "MODULE_CARGO_HOLD_I",
						Name:        "Cargo Hold I",
						Description: "Basic cargo storage",
						Capacity:    30,
					},
				},
				Mounts: []ShipyardShipMount{
					{
						Symbol:      "MOUNT_SENSOR_ARRAY_I",
						Name:        "Sensor Array I",
						Description: "Basic sensor array",
						Strength:    1,
					},
				},
				Crew: ShipyardShipCrew{
					Required: 1,
					Capacity: 3,
				},
			},
		},
		ModificationsFee: 1000,
	}

	mockResponse := ShipyardResponse{
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

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	// Test GetShipyard
	shipyard, err := client.GetShipyard("X1-TEST", "X1-TEST-SHIPYARD")
	if err != nil {
		t.Fatalf("GetShipyard returned error: %v", err)
	}

	// Verify response
	if shipyard.Symbol != "X1-TEST-SHIPYARD" {
		t.Errorf("Expected shipyard symbol X1-TEST-SHIPYARD, got %s", shipyard.Symbol)
	}

	if len(shipyard.ShipTypes) != 2 {
		t.Errorf("Expected 2 ship types, got %d", len(shipyard.ShipTypes))
	}

	if shipyard.ShipTypes[0].Type != "SHIP_PROBE" {
		t.Errorf("Expected first ship type SHIP_PROBE, got %s", shipyard.ShipTypes[0].Type)
	}

	if len(shipyard.Ships) != 1 {
		t.Errorf("Expected 1 ship available, got %d", len(shipyard.Ships))
	}

	ship := shipyard.Ships[0]
	if ship.Type != "SHIP_PROBE" {
		t.Errorf("Expected ship type SHIP_PROBE, got %s", ship.Type)
	}

	if ship.PurchasePrice != 50000 {
		t.Errorf("Expected purchase price 50000, got %d", ship.PurchasePrice)
	}

	if ship.Supply != "ABUNDANT" {
		t.Errorf("Expected supply ABUNDANT, got %s", ship.Supply)
	}

	if shipyard.ModificationsFee != 1000 {
		t.Errorf("Expected modifications fee 1000, got %d", shipyard.ModificationsFee)
	}
}

func TestClient_GetShipyard_Error(t *testing.T) {
	// Test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(`{"error": "Shipyard not found"}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	shipyard, err := client.GetShipyard("INVALID-SYSTEM", "INVALID-WAYPOINT")
	if err == nil {
		t.Fatal("Expected error for invalid shipyard, got nil")
	}
	if shipyard != nil {
		t.Error("Expected nil shipyard on error, got non-nil")
	}
}

func TestClient_PurchaseShip(t *testing.T) {
	// Mock successful ship purchase response
	mockShip := Ship{
		Symbol: "TEST_SHIP_NEW",
		Registration: Registration{
			Name:          "Test Mining Drone",
			FactionSymbol: "COSMIC",
			Role:          "EXCAVATOR",
		},
		Nav: Navigation{
			SystemSymbol:   "X1-TEST",
			WaypointSymbol: "X1-TEST-SHIPYARD",
			Status:         "DOCKED",
			FlightMode:     "CRUISE",
		},
		Cargo: Cargo{
			Capacity: 30,
			Units:    0,
		},
		Fuel: Fuel{
			Current:  400,
			Capacity: 400,
		},
		Crew: Crew{
			Current:  1,
			Capacity: 3,
		},
	}

	mockAgent := Agent{
		AccountID:       "test-account",
		Symbol:          "TEST_AGENT",
		Headquarters:    "X1-TEST-HQ",
		Credits:         125000, // After spending 75000 on ship
		StartingFaction: "COSMIC",
		ShipCount:       2, // Increased after purchase
	}

	mockTransaction := Transaction{
		WaypointSymbol: "X1-TEST-SHIPYARD",
		ShipSymbol:     "TEST_SHIP_NEW",
		ShipType:       "SHIP_MINING_DRONE",
		Price:          75000,
		AgentSymbol:    "TEST_AGENT",
		Timestamp:      "2025-01-01T12:00:00.000Z",
	}

	mockResponse := PurchaseShipResponse{
		Data: struct {
			Agent       Agent       `json:"agent"`
			Ship        Ship        `json:"ship"`
			Transaction Transaction `json:"transaction"`
		}{
			Agent:       mockAgent,
			Ship:        mockShip,
			Transaction: mockTransaction,
		},
	}

	responseJSON, err := json.Marshal(mockResponse)
	if err != nil {
		t.Fatalf("Failed to marshal mock response: %v", err)
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		expectedPath := "/my/ships"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
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

		// Check request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
		}

		var purchaseReq PurchaseShipRequest
		if err := json.Unmarshal(body, &purchaseReq); err != nil {
			t.Errorf("Failed to parse request body: %v", err)
		}

		if purchaseReq.ShipType != "SHIP_MINING_DRONE" {
			t.Errorf("Expected ship type SHIP_MINING_DRONE, got %s", purchaseReq.ShipType)
		}

		if purchaseReq.WaypointSymbol != "X1-TEST-SHIPYARD" {
			t.Errorf("Expected waypoint X1-TEST-SHIPYARD, got %s", purchaseReq.WaypointSymbol)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
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

	// Test PurchaseShip
	ship, agent, transaction, err := client.PurchaseShip("SHIP_MINING_DRONE", "X1-TEST-SHIPYARD")
	if err != nil {
		t.Fatalf("PurchaseShip returned error: %v", err)
	}

	// Verify ship response
	if ship.Symbol != mockShip.Symbol {
		t.Errorf("Expected ship symbol %s, got %s", mockShip.Symbol, ship.Symbol)
	}
	if ship.Registration.Role != mockShip.Registration.Role {
		t.Errorf("Expected ship role %s, got %s", mockShip.Registration.Role, ship.Registration.Role)
	}
	if ship.Nav.WaypointSymbol != mockShip.Nav.WaypointSymbol {
		t.Errorf("Expected ship waypoint %s, got %s", mockShip.Nav.WaypointSymbol, ship.Nav.WaypointSymbol)
	}

	// Verify agent response
	if agent.Credits != mockAgent.Credits {
		t.Errorf("Expected agent credits %d, got %d", mockAgent.Credits, agent.Credits)
	}
	if agent.ShipCount != mockAgent.ShipCount {
		t.Errorf("Expected ship count %d, got %d", mockAgent.ShipCount, agent.ShipCount)
	}

	// Verify transaction response
	if transaction.Price != mockTransaction.Price {
		t.Errorf("Expected transaction price %d, got %d", mockTransaction.Price, transaction.Price)
	}
	if transaction.ShipType != mockTransaction.ShipType {
		t.Errorf("Expected transaction ship type %s, got %s", mockTransaction.ShipType, transaction.ShipType)
	}
}

func TestClient_PurchaseShip_InsufficientFunds(t *testing.T) {
	// Test server that returns insufficient funds error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(`{"error": {"message": "Insufficient funds"}}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	ship, agent, transaction, err := client.PurchaseShip("SHIP_MINING_DRONE", "X1-TEST-SHIPYARD")
	if err == nil {
		t.Fatal("Expected error for insufficient funds, got nil")
	}
	if ship != nil {
		t.Error("Expected nil ship on error, got non-nil")
	}
	if agent != nil {
		t.Error("Expected nil agent on error, got non-nil")
	}
	if transaction != nil {
		t.Error("Expected nil transaction on error, got non-nil")
	}

	// Check that the error message contains status code
	if !contains(err.Error(), "400") {
		t.Errorf("Expected error to contain '400', got: %s", err.Error())
	}
}

func TestClient_PurchaseShip_ShipNotAvailable(t *testing.T) {
	// Test server that returns ship not available error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		if _, err := w.Write([]byte(`{"error": {"message": "Ship type not available at this shipyard"}}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	ship, agent, transaction, err := client.PurchaseShip("SHIP_MINING_DRONE", "X1-TEST-SHIPYARD")
	if err == nil {
		t.Fatal("Expected error for unavailable ship, got nil")
	}
	if ship != nil {
		t.Error("Expected nil ship on error, got non-nil")
	}
	if agent != nil {
		t.Error("Expected nil agent on error, got non-nil")
	}
	if transaction != nil {
		t.Error("Expected nil transaction on error, got non-nil")
	}

	// Check that the error message contains status code
	if !contains(err.Error(), "409") {
		t.Errorf("Expected error to contain '409', got: %s", err.Error())
	}
}

func TestClient_GetAllSystems(t *testing.T) {
	// Mock systems response
	mockSystems := []System{
		{
			Symbol:       "X1-TEST",
			SectorSymbol: "X1",
			Type:         "STAR_SYSTEM",
			X:            10,
			Y:            20,
			Waypoints: []Waypoint{
				{Symbol: "X1-TEST-A1", Type: "PLANET", X: 10, Y: 20},
			},
			Factions: []struct {
				Symbol string `json:"symbol"`
			}{
				{Symbol: "COSMIC"},
			},
		},
	}

	// Test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/systems" {
			t.Errorf("Expected path '/systems', got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected method 'GET', got %s", r.Method)
		}

		response := SystemsResponse{
			Data: mockSystems,
			Meta: struct {
				Total int `json:"total"`
				Page  int `json:"page"`
				Limit int `json:"limit"`
			}{
				Total: 1,
				Page:  1,
				Limit: 10,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	// Test GetAllSystems
	systems, err := client.GetAllSystems()
	if err != nil {
		t.Fatalf("GetAllSystems failed: %v", err)
	}

	if len(systems) != 1 {
		t.Errorf("Expected 1 system, got %d", len(systems))
	}

	if systems[0].Symbol != "X1-TEST" {
		t.Errorf("Expected system symbol 'X1-TEST', got %s", systems[0].Symbol)
	}
}

func TestClient_GetSystem(t *testing.T) {
	// Mock system response
	mockSystem := System{
		Symbol:       "X1-TEST",
		SectorSymbol: "X1",
		Type:         "STAR_SYSTEM",
		X:            10,
		Y:            20,
		Waypoints: []Waypoint{
			{Symbol: "X1-TEST-A1", Type: "PLANET", X: 10, Y: 20},
		},
		Factions: []struct {
			Symbol string `json:"symbol"`
		}{
			{Symbol: "COSMIC"},
		},
	}

	// Test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/systems/X1-TEST" {
			t.Errorf("Expected path '/systems/X1-TEST', got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected method 'GET', got %s", r.Method)
		}

		response := SystemResponse{
			Data: mockSystem,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	// Test GetSystem
	system, err := client.GetSystem("X1-TEST")
	if err != nil {
		t.Fatalf("GetSystem failed: %v", err)
	}

	if system.Symbol != "X1-TEST" {
		t.Errorf("Expected system symbol 'X1-TEST', got %s", system.Symbol)
	}

	if system.Type != "STAR_SYSTEM" {
		t.Errorf("Expected system type 'STAR_SYSTEM', got %s", system.Type)
	}
}

func TestClient_GetAllFactions(t *testing.T) {
	// Mock factions response
	mockFactions := []Faction{
		{
			Symbol:       "COSMIC",
			Name:         "Cosmic Syndicate",
			Description:  "A peaceful faction focused on exploration and trade",
			Headquarters: "X1-TEST-A1",
			Traits: []struct {
				Symbol      string `json:"symbol"`
				Name        string `json:"name"`
				Description string `json:"description"`
			}{
				{
					Symbol:      "PEACEFUL",
					Name:        "Peaceful",
					Description: "Avoids conflict when possible",
				},
			},
			IsRecruiting: true,
		},
	}

	// Test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/factions" {
			t.Errorf("Expected path '/factions', got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected method 'GET', got %s", r.Method)
		}

		response := FactionsResponse{
			Data: mockFactions,
			Meta: struct {
				Total int `json:"total"`
				Page  int `json:"page"`
				Limit int `json:"limit"`
			}{
				Total: 1,
				Page:  1,
				Limit: 10,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	// Test GetAllFactions
	factions, err := client.GetAllFactions()
	if err != nil {
		t.Fatalf("GetAllFactions failed: %v", err)
	}

	if len(factions) != 1 {
		t.Errorf("Expected 1 faction, got %d", len(factions))
	}

	if factions[0].Symbol != "COSMIC" {
		t.Errorf("Expected faction symbol 'COSMIC', got %s", factions[0].Symbol)
	}

	if factions[0].Name != "Cosmic Syndicate" {
		t.Errorf("Expected faction name 'Cosmic Syndicate', got %s", factions[0].Name)
	}
}

func TestClient_GetFaction(t *testing.T) {
	// Mock faction response
	mockFaction := Faction{
		Symbol:       "COSMIC",
		Name:         "Cosmic Syndicate",
		Description:  "A peaceful faction focused on exploration and trade",
		Headquarters: "X1-TEST-A1",
		Traits: []struct {
			Symbol      string `json:"symbol"`
			Name        string `json:"name"`
			Description string `json:"description"`
		}{
			{
				Symbol:      "PEACEFUL",
				Name:        "Peaceful",
				Description: "Avoids conflict when possible",
			},
		},
		IsRecruiting: true,
	}

	// Test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/factions/COSMIC" {
			t.Errorf("Expected path '/factions/COSMIC', got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected method 'GET', got %s", r.Method)
		}

		response := FactionResponse{
			Data: mockFaction,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	// Test GetFaction
	faction, err := client.GetFaction("COSMIC")
	if err != nil {
		t.Fatalf("GetFaction failed: %v", err)
	}

	if faction.Symbol != "COSMIC" {
		t.Errorf("Expected faction symbol 'COSMIC', got %s", faction.Symbol)
	}

	if faction.Name != "Cosmic Syndicate" {
		t.Errorf("Expected faction name 'Cosmic Syndicate', got %s", faction.Name)
	}

	if !faction.IsRecruiting {
		t.Error("Expected faction to be recruiting")
	}
}

func TestClient_GetAllSystems_Error(t *testing.T) {
	// Test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(`{"error": "Internal server error"}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	// Test GetAllSystems with error
	systems, err := client.GetAllSystems()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if systems != nil {
		t.Error("Expected nil systems on error, got non-nil")
	}
}

func TestClient_GetAllFactions_Error(t *testing.T) {
	// Test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(`{"error": "Not found"}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	// Test GetAllFactions with error
	factions, err := client.GetAllFactions()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if factions != nil {
		t.Error("Expected nil factions on error, got non-nil")
	}
}

func TestClient_SellCargo(t *testing.T) {
	// Mock sell cargo response
	mockAgent := Agent{
		AccountID:       "test-account",
		Symbol:          "TEST_AGENT",
		Headquarters:    "X1-TEST-A1",
		Credits:         150000,
		StartingFaction: "COSMIC",
		ShipCount:       1,
	}

	mockCargo := Cargo{
		Capacity: 100,
		Units:    50,
		Inventory: []CargoItem{
			{
				Symbol:      "FUEL",
				Name:        "Fuel",
				Description: "Ship fuel",
				Units:       50,
			},
		},
	}

	mockTransaction := MarketTransaction{
		WaypointSymbol: "X1-TEST-MARKET",
		ShipSymbol:     "TEST_SHIP",
		TradeSymbol:    "IRON_ORE",
		Type:           "SELL",
		Units:          10,
		PricePerUnit:   100,
		TotalPrice:     1000,
		Timestamp:      "2023-01-01T00:00:00.000Z",
	}

	// Test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/my/ships/TEST_SHIP/sell" {
			t.Errorf("Expected path '/my/ships/TEST_SHIP/sell', got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected method 'POST', got %s", r.Method)
		}

		response := SellCargoResponse{
			Data: SellCargoData{
				Agent:       mockAgent,
				Cargo:       mockCargo,
				Transaction: mockTransaction,
			},
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	// Test SellCargo
	agent, cargo, transaction, err := client.SellCargo("TEST_SHIP", "IRON_ORE", 10)
	if err != nil {
		t.Fatalf("SellCargo failed: %v", err)
	}

	if agent.Credits != 150000 {
		t.Errorf("Expected agent credits 150000, got %d", agent.Credits)
	}

	if cargo.Units != 50 {
		t.Errorf("Expected cargo units 50, got %d", cargo.Units)
	}

	if transaction.TotalPrice != 1000 {
		t.Errorf("Expected transaction total price 1000, got %d", transaction.TotalPrice)
	}
}

func TestClient_BuyCargo(t *testing.T) {
	// Mock buy cargo response
	mockAgent := Agent{
		AccountID:       "test-account",
		Symbol:          "TEST_AGENT",
		Headquarters:    "X1-TEST-A1",
		Credits:         50000,
		StartingFaction: "COSMIC",
		ShipCount:       1,
	}

	mockCargo := Cargo{
		Capacity: 100,
		Units:    75,
		Inventory: []CargoItem{
			{
				Symbol:      "FUEL",
				Name:        "Fuel",
				Description: "Ship fuel",
				Units:       50,
			},
			{
				Symbol:      "FOOD",
				Name:        "Food",
				Description: "Ship provisions",
				Units:       25,
			},
		},
	}

	mockTransaction := MarketTransaction{
		WaypointSymbol: "X1-TEST-MARKET",
		ShipSymbol:     "TEST_SHIP",
		TradeSymbol:    "FOOD",
		Type:           "PURCHASE",
		Units:          25,
		PricePerUnit:   80,
		TotalPrice:     2000,
		Timestamp:      "2023-01-01T00:00:00.000Z",
	}

	// Test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/my/ships/TEST_SHIP/purchase" {
			t.Errorf("Expected path '/my/ships/TEST_SHIP/purchase', got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected method 'POST', got %s", r.Method)
		}

		response := BuyCargoResponse{
			Data: BuyCargoData{
				Agent:       mockAgent,
				Cargo:       mockCargo,
				Transaction: mockTransaction,
			},
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	// Test BuyCargo
	agent, cargo, transaction, err := client.BuyCargo("TEST_SHIP", "FOOD", 25)
	if err != nil {
		t.Fatalf("BuyCargo failed: %v", err)
	}

	if agent.Credits != 50000 {
		t.Errorf("Expected agent credits 50000, got %d", agent.Credits)
	}

	if cargo.Units != 75 {
		t.Errorf("Expected cargo units 75, got %d", cargo.Units)
	}

	if transaction.TotalPrice != 2000 {
		t.Errorf("Expected transaction total price 2000, got %d", transaction.TotalPrice)
	}
}

func TestClient_FulfillContract(t *testing.T) {
	// Mock fulfill contract response
	mockAgent := Agent{
		AccountID:       "test-account",
		Symbol:          "TEST_AGENT",
		Headquarters:    "X1-TEST-A1",
		Credits:         200000,
		StartingFaction: "COSMIC",
		ShipCount:       1,
	}

	mockContract := Contract{
		ID:               "CONTRACT_123",
		FactionSymbol:    "COSMIC",
		Type:             "PROCUREMENT",
		Accepted:         true,
		Fulfilled:        true,
		Expiration:       "2023-02-01T00:00:00.000Z",
		DeadlineToAccept: "2023-01-15T00:00:00.000Z",
		Terms: ContractTerms{
			Deadline: "2023-01-31T00:00:00.000Z",
			Payment: ContractPayment{
				OnAccepted:  10000,
				OnFulfilled: 50000,
			},
			Deliver: []ContractDeliverGood{
				{
					TradeSymbol:       "IRON_ORE",
					DestinationSymbol: "X1-TEST-STATION",
					UnitsRequired:     100,
					UnitsFulfilled:    100,
				},
			},
		},
	}

	// Test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/my/contracts/CONTRACT_123/fulfill" {
			t.Errorf("Expected path '/my/contracts/CONTRACT_123/fulfill', got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected method 'POST', got %s", r.Method)
		}

		response := FulfillContractResponse{
			Data: FulfillContractData{
				Agent:    mockAgent,
				Contract: mockContract,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	// Test FulfillContract
	agent, contract, err := client.FulfillContract("CONTRACT_123")
	if err != nil {
		t.Fatalf("FulfillContract failed: %v", err)
	}

	if agent.Credits != 200000 {
		t.Errorf("Expected agent credits 200000, got %d", agent.Credits)
	}

	if !contract.Fulfilled {
		t.Error("Expected contract to be fulfilled")
	}

	if contract.ID != "CONTRACT_123" {
		t.Errorf("Expected contract ID 'CONTRACT_123', got %s", contract.ID)
	}
}

func TestClient_SellCargo_Error(t *testing.T) {
	// Test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(`{"error": "Ship not docked at marketplace"}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	// Test SellCargo with error
	agent, cargo, transaction, err := client.SellCargo("TEST_SHIP", "IRON_ORE", 10)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if agent != nil {
		t.Error("Expected nil agent on error, got non-nil")
	}
	if cargo != nil {
		t.Error("Expected nil cargo on error, got non-nil")
	}
	if transaction != nil {
		t.Error("Expected nil transaction on error, got non-nil")
	}
}

func TestClient_FulfillContract_Error(t *testing.T) {
	// Test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if _, err := w.Write([]byte(`{"error": "Contract requirements not met"}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		APIToken: "test-token",
		BaseURL:  server.URL,
	}

	// Test FulfillContract with error
	agent, contract, err := client.FulfillContract("CONTRACT_123")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if agent != nil {
		t.Error("Expected nil agent on error, got non-nil")
	}
	if contract != nil {
		t.Error("Expected nil contract on error, got non-nil")
	}
}
