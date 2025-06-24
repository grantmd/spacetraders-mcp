package spacetraders

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client handles SpaceTraders API interactions
type Client struct {
	APIToken string
	BaseURL  string
}

// Agent represents the SpaceTraders agent data structure
type Agent struct {
	AccountID       string `json:"accountId"`
	Symbol          string `json:"symbol"`
	Headquarters    string `json:"headquarters"`
	Credits         int64  `json:"credits"`
	StartingFaction string `json:"startingFaction"`
	ShipCount       int    `json:"shipCount"`
}

// Ship represents a SpaceTraders ship
type Ship struct {
	Symbol       string       `json:"symbol"`
	Registration Registration `json:"registration"`
	Nav          Navigation   `json:"nav"`
	Crew         Crew         `json:"crew"`
	Frame        Frame        `json:"frame"`
	Reactor      Reactor      `json:"reactor"`
	Engine       Engine       `json:"engine"`
	Cooldown     Cooldown     `json:"cooldown"`
	Modules      []Module     `json:"modules"`
	Mounts       []Mount      `json:"mounts"`
	Cargo        Cargo        `json:"cargo"`
	Fuel         Fuel         `json:"fuel"`
}

// Registration represents ship registration info
type Registration struct {
	Name          string `json:"name"`
	FactionSymbol string `json:"factionSymbol"`
	Role          string `json:"role"`
}

// Navigation represents ship navigation info
type Navigation struct {
	SystemSymbol   string `json:"systemSymbol"`
	WaypointSymbol string `json:"waypointSymbol"`
	Route          Route  `json:"route"`
	Status         string `json:"status"`
	FlightMode     string `json:"flightMode"`
}

// Route represents a navigation route
type Route struct {
	Destination   Waypoint `json:"destination"`
	Origin        Waypoint `json:"origin"`
	DepartureTime string   `json:"departureTime"`
	Arrival       string   `json:"arrival"`
}

// Waypoint represents a waypoint in space
type Waypoint struct {
	Symbol string `json:"symbol"`
	Type   string `json:"type"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
}

// Crew represents ship crew information
type Crew struct {
	Current  int    `json:"current"`
	Required int    `json:"required"`
	Capacity int    `json:"capacity"`
	Rotation string `json:"rotation"`
	Morale   int    `json:"morale"`
	Wages    int    `json:"wages"`
}

// Frame represents ship frame information
type Frame struct {
	Symbol         string `json:"symbol"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Condition      int    `json:"condition"`
	Integrity      int    `json:"integrity"`
	ModuleSlots    int    `json:"moduleSlots"`
	MountingPoints int    `json:"mountingPoints"`
	FuelCapacity   int    `json:"fuelCapacity"`
	Requirements   struct {
		Power int `json:"power"`
		Crew  int `json:"crew"`
		Slots int `json:"slots"`
	} `json:"requirements"`
}

// Reactor represents ship reactor information
type Reactor struct {
	Symbol       string `json:"symbol"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Condition    int    `json:"condition"`
	Integrity    int    `json:"integrity"`
	PowerOutput  int    `json:"powerOutput"`
	Requirements struct {
		Power int `json:"power"`
		Crew  int `json:"crew"`
		Slots int `json:"slots"`
	} `json:"requirements"`
}

// Engine represents ship engine information
type Engine struct {
	Symbol       string `json:"symbol"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Condition    int    `json:"condition"`
	Integrity    int    `json:"integrity"`
	Speed        int    `json:"speed"`
	Requirements struct {
		Power int `json:"power"`
		Crew  int `json:"crew"`
		Slots int `json:"slots"`
	} `json:"requirements"`
}

// Cooldown represents ship cooldown information
type Cooldown struct {
	ShipSymbol       string `json:"shipSymbol"`
	TotalSeconds     int    `json:"totalSeconds"`
	RemainingSeconds int    `json:"remainingSeconds"`
	Expiration       string `json:"expiration"`
}

// Module represents a ship module
type Module struct {
	Symbol       string `json:"symbol"`
	Capacity     int    `json:"capacity,omitempty"`
	Range        int    `json:"range,omitempty"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Requirements struct {
		Power int `json:"power"`
		Crew  int `json:"crew"`
		Slots int `json:"slots"`
	} `json:"requirements"`
}

// Mount represents a ship mount
type Mount struct {
	Symbol       string   `json:"symbol"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Strength     int      `json:"strength,omitempty"`
	Deposits     []string `json:"deposits,omitempty"`
	Requirements struct {
		Power int `json:"power"`
		Crew  int `json:"crew"`
		Slots int `json:"slots"`
	} `json:"requirements"`
}

// Cargo represents ship cargo information
type Cargo struct {
	Capacity  int         `json:"capacity"`
	Units     int         `json:"units"`
	Inventory []CargoItem `json:"inventory"`
}

// CargoItem represents an item in cargo
type CargoItem struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Units       int    `json:"units"`
}

// Fuel represents ship fuel information
type Fuel struct {
	Current  int `json:"current"`
	Capacity int `json:"capacity"`
	Consumed struct {
		Amount    int    `json:"amount"`
		Timestamp string `json:"timestamp"`
	} `json:"consumed"`
}

// Contract represents a SpaceTraders contract
type Contract struct {
	ID               string        `json:"id"`
	FactionSymbol    string        `json:"factionSymbol"`
	Type             string        `json:"type"`
	Terms            ContractTerms `json:"terms"`
	Accepted         bool          `json:"accepted"`
	Fulfilled        bool          `json:"fulfilled"`
	Expiration       string        `json:"expiration"`
	DeadlineToAccept string        `json:"deadlineToAccept"`
}

// ContractTerms represents contract terms
type ContractTerms struct {
	Deadline string                `json:"deadline"`
	Payment  ContractPayment       `json:"payment"`
	Deliver  []ContractDeliverGood `json:"deliver"`
}

// ContractPayment represents contract payment info
type ContractPayment struct {
	OnAccepted  int `json:"onAccepted"`
	OnFulfilled int `json:"onFulfilled"`
}

// ContractDeliverGood represents a contract delivery requirement
type ContractDeliverGood struct {
	TradeSymbol       string `json:"tradeSymbol"`
	DestinationSymbol string `json:"destinationSymbol"`
	UnitsRequired     int    `json:"unitsRequired"`
	UnitsFulfilled    int    `json:"unitsFulfilled"`
}

// AcceptContractResponse represents the response from accepting a contract
type AcceptContractResponse struct {
	Data struct {
		Contract Contract `json:"contract"`
		Agent    Agent    `json:"agent"`
	} `json:"data"`
}

// System represents a SpaceTraders system
type System struct {
	Symbol       string     `json:"symbol"`
	SectorSymbol string     `json:"sectorSymbol"`
	Type         string     `json:"type"`
	X            int        `json:"x"`
	Y            int        `json:"y"`
	Waypoints    []Waypoint `json:"waypoints"`
	Factions     []struct {
		Symbol string `json:"symbol"`
	} `json:"factions"`
}

// SystemWaypoint represents a waypoint in a system (different from navigation waypoint)
type SystemWaypoint struct {
	Symbol    string             `json:"symbol"`
	Type      string             `json:"type"`
	X         int                `json:"x"`
	Y         int                `json:"y"`
	Orbitals  []WaypointOrbital  `json:"orbitals"`
	Traits    []WaypointTrait    `json:"traits"`
	Modifiers []WaypointModifier `json:"modifiers,omitempty"`
	Chart     *WaypointChart     `json:"chart,omitempty"`
	Faction   *WaypointFaction   `json:"faction,omitempty"`
}

// WaypointOrbital represents an orbital around a waypoint
type WaypointOrbital struct {
	Symbol string `json:"symbol"`
}

// WaypointTrait represents a trait of a waypoint
type WaypointTrait struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// WaypointModifier represents a modifier affecting a waypoint
type WaypointModifier struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// WaypointChart represents chart information for a waypoint
type WaypointChart struct {
	WaypointSymbol string `json:"waypointSymbol,omitempty"`
	SubmittedBy    string `json:"submittedBy,omitempty"`
	SubmittedOn    string `json:"submittedOn,omitempty"`
}

// WaypointFaction represents faction control of a waypoint
type WaypointFaction struct {
	Symbol string `json:"symbol"`
}

// Shipyard represents a shipyard at a waypoint
type Shipyard struct {
	Symbol           string                `json:"symbol"`
	ShipTypes        []ShipyardShipType    `json:"shipTypes"`
	Transactions     []ShipyardTransaction `json:"transactions,omitempty"`
	Ships            []ShipyardShip        `json:"ships,omitempty"`
	ModificationsFee int                   `json:"modificationsFee"`
}

// ShipyardShipType represents a type of ship available at a shipyard
type ShipyardShipType struct {
	Type string `json:"type"`
}

// ShipyardTransaction represents a transaction at a shipyard
type ShipyardTransaction struct {
	WaypointSymbol string `json:"waypointSymbol"`
	ShipSymbol     string `json:"shipSymbol"`
	ShipType       string `json:"shipType"`
	Price          int    `json:"price"`
	AgentSymbol    string `json:"agentSymbol"`
	Timestamp      string `json:"timestamp"`
}

// ShipyardShip represents a ship available for purchase at a shipyard
type ShipyardShip struct {
	Type          string               `json:"type"`
	Name          string               `json:"name"`
	Description   string               `json:"description"`
	Supply        string               `json:"supply"`
	Activity      string               `json:"activity,omitempty"`
	PurchasePrice int                  `json:"purchasePrice"`
	Frame         ShipyardShipFrame    `json:"frame"`
	Reactor       ShipyardShipReactor  `json:"reactor"`
	Engine        ShipyardShipEngine   `json:"engine"`
	Modules       []ShipyardShipModule `json:"modules"`
	Mounts        []ShipyardShipMount  `json:"mounts"`
	Crew          ShipyardShipCrew     `json:"crew"`
}

// ShipyardShipFrame represents frame information for a ship at shipyard
type ShipyardShipFrame struct {
	Symbol         string `json:"symbol"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ModuleSlots    int    `json:"moduleSlots"`
	MountingPoints int    `json:"mountingPoints"`
	FuelCapacity   int    `json:"fuelCapacity"`
	Condition      int    `json:"condition"`
	Integrity      int    `json:"integrity"`
	Requirements   struct {
		Power int `json:"power"`
		Crew  int `json:"crew"`
		Slots int `json:"slots"`
	} `json:"requirements"`
}

// ShipyardShipReactor represents reactor information for a ship at shipyard
type ShipyardShipReactor struct {
	Symbol       string `json:"symbol"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Condition    int    `json:"condition"`
	Integrity    int    `json:"integrity"`
	PowerOutput  int    `json:"powerOutput"`
	Requirements struct {
		Power int `json:"power"`
		Crew  int `json:"crew"`
		Slots int `json:"slots"`
	} `json:"requirements"`
}

// ShipyardShipEngine represents engine information for a ship at shipyard
type ShipyardShipEngine struct {
	Symbol       string `json:"symbol"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Condition    int    `json:"condition"`
	Integrity    int    `json:"integrity"`
	Speed        int    `json:"speed"`
	Requirements struct {
		Power int `json:"power"`
		Crew  int `json:"crew"`
		Slots int `json:"slots"`
	} `json:"requirements"`
}

// ShipyardShipModule represents a module for a ship at shipyard
type ShipyardShipModule struct {
	Symbol       string `json:"symbol"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Capacity     int    `json:"capacity,omitempty"`
	Range        int    `json:"range,omitempty"`
	Requirements struct {
		Power int `json:"power"`
		Crew  int `json:"crew"`
		Slots int `json:"slots"`
	} `json:"requirements"`
}

// ShipyardShipMount represents a mount for a ship at shipyard
type ShipyardShipMount struct {
	Symbol       string   `json:"symbol"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Strength     int      `json:"strength,omitempty"`
	Deposits     []string `json:"deposits,omitempty"`
	Requirements struct {
		Power int `json:"power"`
		Crew  int `json:"crew"`
		Slots int `json:"slots"`
	} `json:"requirements"`
}

// ShipyardShipCrew represents crew requirements for a ship at shipyard
type ShipyardShipCrew struct {
	Required int `json:"required"`
	Capacity int `json:"capacity"`
}

// API Response wrappers
type AgentResponse struct {
	Data Agent `json:"data"`
}

type ShipsResponse struct {
	Data []Ship `json:"data"`
	Meta struct {
		Total int `json:"total"`
		Page  int `json:"page"`
		Limit int `json:"limit"`
	} `json:"meta"`
}

type ContractsResponse struct {
	Data []Contract `json:"data"`
	Meta struct {
		Total int `json:"total"`
		Page  int `json:"page"`
		Limit int `json:"limit"`
	} `json:"meta"`
}

type ContractResponse struct {
	Data Contract `json:"data"`
}

type SystemWaypointsResponse struct {
	Data []SystemWaypoint `json:"data"`
	Meta struct {
		Total int `json:"total"`
		Page  int `json:"page"`
		Limit int `json:"limit"`
	} `json:"meta"`
}

type ShipyardResponse struct {
	Data Shipyard `json:"data"`
}

// NewClient creates a new SpaceTraders client
func NewClient(apiToken string) *Client {
	return &Client{
		APIToken: apiToken,
		BaseURL:  "https://api.spacetraders.io/v2",
	}
}

// makeRequest makes an HTTP request to the SpaceTraders API
func (c *Client) makeRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)

	req, err := http.NewRequest(method, url, body)
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

	return resp, nil
}

// GetAgent fetches agent information from the SpaceTraders API
func (c *Client) GetAgent() (*Agent, error) {
	resp, err := c.makeRequest("GET", "/my/agent", nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

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

// GetShips fetches all ships for the agent
func (c *Client) GetShips() ([]Ship, error) {
	resp, err := c.makeRequest("GET", "/my/ships", nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var shipsResp ShipsResponse
	if err := json.Unmarshal(body, &shipsResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return shipsResp.Data, nil
}

// GetContracts fetches all contracts for the agent
func (c *Client) GetContracts() ([]Contract, error) {
	resp, err := c.makeRequest("GET", "/my/contracts", nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var contractsResp ContractsResponse
	if err := json.Unmarshal(body, &contractsResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return contractsResp.Data, nil
}

// AcceptContract accepts a contract by its ID
func (c *Client) AcceptContract(contractID string) (*Contract, *Agent, error) {
	endpoint := fmt.Sprintf("/my/contracts/%s/accept", contractID)

	// SpaceTraders API requires an empty JSON object for POST requests
	emptyBody := strings.NewReader("{}")
	resp, err := c.makeRequest("POST", endpoint, emptyBody)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var acceptResp AcceptContractResponse
	if err := json.Unmarshal(body, &acceptResp); err != nil {
		return nil, nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &acceptResp.Data.Contract, &acceptResp.Data.Agent, nil
}

// GetSystemWaypoints fetches all waypoints in a system
func (c *Client) GetSystemWaypoints(systemSymbol string) ([]SystemWaypoint, error) {
	endpoint := fmt.Sprintf("/systems/%s/waypoints", systemSymbol)

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var waypointsResp SystemWaypointsResponse
	if err := json.Unmarshal(body, &waypointsResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return waypointsResp.Data, nil
}

// GetShipyard fetches shipyard information for a waypoint
func (c *Client) GetShipyard(systemSymbol, waypointSymbol string) (*Shipyard, error) {
	endpoint := fmt.Sprintf("/systems/%s/waypoints/%s/shipyard", systemSymbol, waypointSymbol)

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var shipyardResp ShipyardResponse
	if err := json.Unmarshal(body, &shipyardResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &shipyardResp.Data, nil
}
