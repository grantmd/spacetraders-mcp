package exploration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"spacetraders-mcp/pkg/client"
	"spacetraders-mcp/pkg/logging"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestFindWaypointsTool_Tool(t *testing.T) {
	client := client.NewClient("test-token")
	logger := logging.NewLogger(nil)
	tool := NewFindWaypointsTool(client, logger)

	toolDef := tool.Tool()

	if toolDef.Name != "find_waypoints" {
		t.Errorf("Expected tool name 'find_waypoints', got %s", toolDef.Name)
	}

	if len(toolDef.InputSchema.Required) != 2 {
		t.Errorf("Expected 2 required parameters, got %d", len(toolDef.InputSchema.Required))
	}

	expectedRequired := []string{"system_symbol", "trait"}
	for i, req := range expectedRequired {
		if toolDef.InputSchema.Required[i] != req {
			t.Errorf("Expected required param %s, got %s", req, toolDef.InputSchema.Required[i])
		}
	}
}

func TestFindWaypointsTool_Handler_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/systems/X1-TEST/waypoints") {
			t.Errorf("Expected waypoints endpoint, got %s", r.URL.Path)
		}

		mockWaypoints := []client.SystemWaypoint{
			{
				Symbol: "X1-TEST-SHIPYARD",
				Type:   "PLANET",
				X:      10,
				Y:      20,
				Traits: []client.WaypointTrait{
					{
						Symbol:      "SHIPYARD",
						Name:        "Shipyard",
						Description: "A facility for building ships",
					},
					{
						Symbol:      "MARKETPLACE",
						Name:        "Marketplace",
						Description: "A trading facility",
					},
				},
			},
			{
				Symbol: "X1-TEST-MARKET",
				Type:   "MOON",
				X:      30,
				Y:      40,
				Traits: []client.WaypointTrait{
					{
						Symbol:      "MARKETPLACE",
						Name:        "Marketplace",
						Description: "A trading facility",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			Data []client.SystemWaypoint `json:"data"`
			Meta struct {
				Total int `json:"total"`
				Page  int `json:"page"`
				Limit int `json:"limit"`
			} `json:"meta"`
		}{
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
		})
	}))
	defer server.Close()

	client := client.NewClientWithBaseURL("test-token", server.URL)
	logger := logging.NewLogger(nil)
	tool := NewFindWaypointsTool(client, logger)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "find_waypoints",
			Arguments: map[string]interface{}{
				"system_symbol": "X1-TEST",
				"trait":         "SHIPYARD",
			},
		},
	}

	handler := tool.Handler()
	result, err := handler(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.IsError {
		t.Fatalf("Expected success, got error: %v", result.Content)
	}

	if len(result.Content) != 2 {
		t.Fatalf("Expected 2 content items, got %d", len(result.Content))
	}

	// Check that the text content mentions the found shipyard
	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("Expected TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(textContent.Text, "X1-TEST-SHIPYARD") {
		t.Errorf("Expected text to contain shipyard waypoint, got: %s", textContent.Text)
	}

	if !strings.Contains(textContent.Text, "Found 1 waypoint") {
		t.Errorf("Expected text to mention 1 waypoint found, got: %s", textContent.Text)
	}
}

func TestFindWaypointsTool_Handler_NoResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse := struct {
			Data []client.SystemWaypoint `json:"data"`
			Meta struct {
				Total int `json:"total"`
				Page  int `json:"page"`
				Limit int `json:"limit"`
			} `json:"meta"`
		}{
			Data: []client.SystemWaypoint{
				{
					Symbol: "X1-TEST-ASTEROID",
					Type:   "ASTEROID",
					X:      10,
					Y:      20,
					Traits: []client.WaypointTrait{
						{
							Symbol:      "MARKETPLACE",
							Name:        "Marketplace",
							Description: "A trading facility",
						},
					},
				},
			},
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

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := client.NewClientWithBaseURL("test-token", server.URL)
	logger := logging.NewLogger(nil)
	tool := NewFindWaypointsTool(client, logger)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "find_waypoints",
			Arguments: map[string]interface{}{
				"system_symbol": "X1-TEST",
				"trait":         "SHIPYARD",
			},
		},
	}

	handler := tool.Handler()
	result, err := handler(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.IsError {
		t.Fatalf("Expected success, got error: %v", result.Content)
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("Expected TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(textContent.Text, "No waypoints found") {
		t.Errorf("Expected 'No waypoints found' message, got: %s", textContent.Text)
	}
}

func TestFindWaypointsTool_Handler_MissingParameters(t *testing.T) {
	client := client.NewClient("test-token")
	logger := logging.NewLogger(nil)
	tool := NewFindWaypointsTool(client, logger)

	// Test missing system_symbol
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "find_waypoints",
			Arguments: map[string]interface{}{
				"trait": "SHIPYARD",
			},
		},
	}

	handler := tool.Handler()
	result, err := handler(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.IsError {
		t.Fatalf("Expected error for missing system_symbol")
	}

	// Test missing trait
	request = mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "find_waypoints",
			Arguments: map[string]interface{}{
				"system_symbol": "X1-TEST",
			},
		},
	}

	result, err = handler(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.IsError {
		t.Fatalf("Expected error for missing trait")
	}
}

func TestSystemOverviewTool_Tool(t *testing.T) {
	client := client.NewClient("test-token")
	logger := logging.NewLogger(nil)
	tool := NewSystemOverviewTool(client, logger)

	toolDef := tool.Tool()

	if toolDef.Name != "system_overview" {
		t.Errorf("Expected tool name 'system_overview', got %s", toolDef.Name)
	}

	if len(toolDef.InputSchema.Required) != 1 {
		t.Errorf("Expected 1 required parameter, got %d", len(toolDef.InputSchema.Required))
	}

	if toolDef.InputSchema.Required[0] != "system_symbol" {
		t.Errorf("Expected required param 'system_symbol', got %s", toolDef.InputSchema.Required[0])
	}
}

func TestCurrentLocationTool_Tool(t *testing.T) {
	client := client.NewClient("test-token")
	logger := logging.NewLogger(nil)
	tool := NewCurrentLocationTool(client, logger)

	toolDef := tool.Tool()

	if toolDef.Name != "current_location" {
		t.Errorf("Expected tool name 'current_location', got %s", toolDef.Name)
	}

	// current_location has no required parameters
	if len(toolDef.InputSchema.Required) != 0 {
		t.Errorf("Expected 0 required parameters, got %d", len(toolDef.InputSchema.Required))
	}
}

func TestCurrentLocationTool_Handler_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/systems/X1-TEST/waypoints") {
			mockWaypoints := struct {
				Data []client.SystemWaypoint `json:"data"`
				Meta struct {
					Total int `json:"total"`
					Page  int `json:"page"`
					Limit int `json:"limit"`
				} `json:"meta"`
			}{
				Data: []client.SystemWaypoint{
					{
						Symbol: "X1-TEST-A1",
						Type:   "PLANET",
						X:      10,
						Y:      20,
						Traits: []client.WaypointTrait{
							{
								Symbol:      "MARKETPLACE",
								Name:        "Marketplace",
								Description: "A trading facility",
							},
						},
					},
				},
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
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockWaypoints)
		} else if strings.Contains(r.URL.Path, "/my/ships") {
			mockResponse := struct {
				Data []client.Ship `json:"data"`
				Meta struct {
					Total int `json:"total"`
					Page  int `json:"page"`
					Limit int `json:"limit"`
				} `json:"meta"`
			}{
				Data: []client.Ship{
					{
						Symbol: "SHIP_1234",
						Registration: client.Registration{
							Name:          "Explorer",
							FactionSymbol: "COSMIC",
							Role:          "COMMAND",
						},
						Nav: client.Navigation{
							SystemSymbol:   "X1-TEST",
							WaypointSymbol: "X1-TEST-A1",
							Route: client.Route{
								Destination: client.Waypoint{
									Symbol: "X1-TEST-A1",
									Type:   "PLANET",
									X:      10,
									Y:      20,
								},
								Origin: client.Waypoint{
									Symbol: "X1-TEST-A1",
									Type:   "PLANET",
									X:      10,
									Y:      20,
								},
								DepartureTime: "2024-01-01T00:00:00.000Z",
								Arrival:       "2024-01-01T00:00:00.000Z",
							},
							Status:     "DOCKED",
							FlightMode: "CRUISE",
						},
						Crew: client.Crew{
							Current:  3,
							Required: 3,
							Capacity: 5,
							Rotation: "STRICT",
							Morale:   100,
							Wages:    0,
						},
						Frame: client.Frame{
							Symbol:         "FRAME_PROBE",
							Name:           "Probe Frame",
							Description:    "Small frame for probe ships",
							ModuleSlots:    2,
							MountingPoints: 1,
							FuelCapacity:   400,
							Condition:      1.0,
							Integrity:      1.0,
							Requirements: client.ShipRequirements{
								Power: 1,
								Crew:  1,
								Slots: 1,
							},
						},
						Reactor: client.Reactor{
							Symbol:      "REACTOR_SOLAR_I",
							Name:        "Solar Reactor I",
							Description: "Basic solar reactor",
							Condition:   1.0,
							Integrity:   1.0,
							PowerOutput: 40,
							Requirements: client.ShipRequirements{
								Power: 0,
								Crew:  0,
								Slots: 1,
							},
						},
						Engine: client.Engine{
							Symbol:      "ENGINE_IMPULSE_DRIVE_I",
							Name:        "Impulse Drive I",
							Description: "Basic impulse drive",
							Condition:   1.0,
							Integrity:   1.0,
							Speed:       30,
							Requirements: client.ShipRequirements{
								Power: 1,
								Crew:  0,
								Slots: 1,
							},
						},
						Cooldown: client.Cooldown{
							ShipSymbol:       "SHIP_1234",
							TotalSeconds:     0,
							RemainingSeconds: 0,
						},
						Modules: []client.Module{},
						Mounts:  []client.Mount{},
						Fuel: client.Fuel{
							Current:  80,
							Capacity: 100,
						},
						Cargo: client.Cargo{
							Capacity:  40,
							Units:     10,
							Inventory: []client.CargoItem{},
						},
					},
				},
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
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockResponse)
		}
	}))
	defer server.Close()

	client := client.NewClientWithBaseURL("test-token", server.URL)
	logger := logging.NewLogger(nil)
	tool := NewCurrentLocationTool(client, logger)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "current_location",
			Arguments: map[string]interface{}{},
		},
	}

	handler := tool.Handler()
	result, err := handler(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.IsError {
		t.Fatalf("Expected success, got error: %v", result.Content)
	}

	if len(result.Content) != 2 {
		t.Fatalf("Expected 2 content items, got %d", len(result.Content))
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("Expected TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(textContent.Text, "SHIP_1234") {
		t.Errorf("Expected text to contain ship symbol, got: %s", textContent.Text)
	}

	if !strings.Contains(textContent.Text, "X1-TEST") {
		t.Errorf("Expected text to contain system symbol, got: %s", textContent.Text)
	}
}
