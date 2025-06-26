package spacetraders

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_OrbitShip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/my/ships/TEST_SHIP_1/orbit") {
			t.Errorf("Expected orbit endpoint, got %s", r.URL.Path)
		}

		mockResponse := OrbitResponse{
			Data: OrbitData{
				Nav: Navigation{
					SystemSymbol:   "X1-TEST",
					WaypointSymbol: "X1-TEST-A1",
					Status:         "IN_ORBIT",
					FlightMode:     "CRUISE",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL)
	nav, err := client.OrbitShip("TEST_SHIP_1")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if nav.Status != "IN_ORBIT" {
		t.Errorf("Expected status 'IN_ORBIT', got %s", nav.Status)
	}
	if nav.SystemSymbol != "X1-TEST" {
		t.Errorf("Expected system 'X1-TEST', got %s", nav.SystemSymbol)
	}
}

func TestClient_DockShip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/my/ships/TEST_SHIP_1/dock") {
			t.Errorf("Expected dock endpoint, got %s", r.URL.Path)
		}

		mockResponse := DockResponse{
			Data: DockData{
				Nav: Navigation{
					SystemSymbol:   "X1-TEST",
					WaypointSymbol: "X1-TEST-A1",
					Status:         "DOCKED",
					FlightMode:     "CRUISE",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL)
	nav, err := client.DockShip("TEST_SHIP_1")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if nav.Status != "DOCKED" {
		t.Errorf("Expected status 'DOCKED', got %s", nav.Status)
	}
	if nav.SystemSymbol != "X1-TEST" {
		t.Errorf("Expected system 'X1-TEST', got %s", nav.SystemSymbol)
	}
}

func TestClient_NavigateShip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/my/ships/TEST_SHIP_1/navigate") {
			t.Errorf("Expected navigate endpoint, got %s", r.URL.Path)
		}

		// Check request body
		var req NavigateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if req.WaypointSymbol != "X1-TEST-B2" {
			t.Errorf("Expected waypoint 'X1-TEST-B2', got %s", req.WaypointSymbol)
		}

		mockResponse := NavigateResponse{
			Data: NavigateData{
				Nav: Navigation{
					SystemSymbol:   "X1-TEST",
					WaypointSymbol: "X1-TEST-A1",
					Status:         "IN_TRANSIT",
					FlightMode:     "CRUISE",
					Route: Route{
						Origin: Waypoint{
							Symbol: "X1-TEST-A1",
							Type:   "PLANET",
							X:      10,
							Y:      20,
						},
						Destination: Waypoint{
							Symbol: "X1-TEST-B2",
							Type:   "MOON",
							X:      30,
							Y:      40,
						},
						DepartureTime: "2023-01-01T10:00:00Z",
						Arrival:       "2023-01-01T11:00:00Z",
					},
				},
				Fuel: Fuel{
					Current:  80,
					Capacity: 100,
					Consumed: struct {
						Amount    int    `json:"amount"`
						Timestamp string `json:"timestamp"`
					}{
						Amount:    20,
						Timestamp: "2023-01-01T10:00:00Z",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL)
	nav, fuel, event, err := client.NavigateShip("TEST_SHIP_1", "X1-TEST-B2")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if nav.Status != "IN_TRANSIT" {
		t.Errorf("Expected status 'IN_TRANSIT', got %s", nav.Status)
	}
	if fuel.Current != 80 {
		t.Errorf("Expected fuel 80, got %d", fuel.Current)
	}
	if fuel.Consumed.Amount != 20 {
		t.Errorf("Expected consumed fuel 20, got %d", fuel.Consumed.Amount)
	}
	if event != nil {
		t.Errorf("Expected no event, got %v", event)
	}
}

func TestClient_PatchShipNav(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/my/ships/TEST_SHIP_1/nav") {
			t.Errorf("Expected nav endpoint, got %s", r.URL.Path)
		}

		// Check request body
		var req PatchNavRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if req.FlightMode != "BURN" {
			t.Errorf("Expected flight mode 'BURN', got %s", req.FlightMode)
		}

		mockResponse := PatchNavResponse{
			Data: Navigation{
				SystemSymbol:   "X1-TEST",
				WaypointSymbol: "X1-TEST-A1",
				Status:         "DOCKED",
				FlightMode:     "BURN",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL)
	nav, err := client.PatchShipNav("TEST_SHIP_1", "BURN")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if nav.FlightMode != "BURN" {
		t.Errorf("Expected flight mode 'BURN', got %s", nav.FlightMode)
	}
}

func TestClient_WarpShip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/my/ships/TEST_SHIP_1/warp") {
			t.Errorf("Expected warp endpoint, got %s", r.URL.Path)
		}

		mockResponse := WarpResponse{
			Data: WarpData{
				Nav: Navigation{
					SystemSymbol:   "X1-OTHER",
					WaypointSymbol: "X1-OTHER-A1",
					Status:         "IN_TRANSIT",
					FlightMode:     "CRUISE",
				},
				Fuel: Fuel{
					Current:  50,
					Capacity: 100,
					Consumed: struct {
						Amount    int    `json:"amount"`
						Timestamp string `json:"timestamp"`
					}{
						Amount:    50,
						Timestamp: "2023-01-01T10:00:00Z",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL)
	nav, fuel, event, err := client.WarpShip("TEST_SHIP_1", "X1-OTHER-A1")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if nav.SystemSymbol != "X1-OTHER" {
		t.Errorf("Expected system 'X1-OTHER', got %s", nav.SystemSymbol)
	}
	if fuel.Consumed.Amount != 50 {
		t.Errorf("Expected consumed fuel 50, got %d", fuel.Consumed.Amount)
	}
	if event != nil {
		t.Errorf("Expected no event, got %v", event)
	}
}

func TestClient_JumpShip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/my/ships/TEST_SHIP_1/jump") {
			t.Errorf("Expected jump endpoint, got %s", r.URL.Path)
		}

		mockResponse := JumpResponse{
			Data: JumpData{
				Nav: Navigation{
					SystemSymbol:   "X1-JUMP",
					WaypointSymbol: "X1-JUMP-GATE",
					Status:         "IN_ORBIT",
					FlightMode:     "CRUISE",
				},
				Cooldown: Cooldown{
					ShipSymbol:       "TEST_SHIP_1",
					TotalSeconds:     300,
					RemainingSeconds: 300,
					Expiration:       "2023-01-01T10:05:00Z",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL)
	nav, cooldown, event, err := client.JumpShip("TEST_SHIP_1", "X1-JUMP")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if nav.SystemSymbol != "X1-JUMP" {
		t.Errorf("Expected system 'X1-JUMP', got %s", nav.SystemSymbol)
	}
	if cooldown.TotalSeconds != 300 {
		t.Errorf("Expected cooldown 300 seconds, got %d", cooldown.TotalSeconds)
	}
	if event != nil {
		t.Errorf("Expected no event, got %v", event)
	}
}

func TestClient_OrbitShip_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": {"message": "Ship is already in orbit"}}`))
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL)
	_, err := client.OrbitShip("TEST_SHIP_1")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to orbit ship") {
		t.Errorf("Expected orbit error message, got %s", err.Error())
	}
}

func TestClient_NavigateShip_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": {"message": "Ship must be in orbit to navigate"}}`))
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL)
	_, _, _, err := client.NavigateShip("TEST_SHIP_1", "X1-TEST-B2")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to navigate ship") {
		t.Errorf("Expected navigate error message, got %s", err.Error())
	}
}
