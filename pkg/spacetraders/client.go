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

// PurchaseShipRequest represents a request to purchase a ship
type PurchaseShipRequest struct {
	ShipType       string `json:"shipType"`
	WaypointSymbol string `json:"waypointSymbol"`
}

// PurchaseShipResponse represents the response from purchasing a ship
type PurchaseShipResponse struct {
	Data struct {
		Agent       Agent       `json:"agent"`
		Ship        Ship        `json:"ship"`
		Transaction Transaction `json:"transaction"`
	} `json:"data"`
}

// Transaction represents a ship purchase transaction
type Transaction struct {
	WaypointSymbol string `json:"waypointSymbol"`
	ShipSymbol     string `json:"shipSymbol"`
	ShipType       string `json:"shipType"`
	Price          int    `json:"price"`
	AgentSymbol    string `json:"agentSymbol"`
	Timestamp      string `json:"timestamp"`
}

// NavigateRequest represents the request body for navigating a ship
type NavigateRequest struct {
	WaypointSymbol string `json:"waypointSymbol"`
}

// NavigateResponse represents the response from navigating a ship
type NavigateResponse struct {
	Data NavigateData `json:"data"`
}

type NavigateData struct {
	Fuel  Fuel       `json:"fuel"`
	Nav   Navigation `json:"nav"`
	Event *Event     `json:"event,omitempty"`
}

// OrbitResponse represents the response from orbiting a ship
type OrbitResponse struct {
	Data OrbitData `json:"data"`
}

type OrbitData struct {
	Nav Navigation `json:"nav"`
}

// DockResponse represents the response from docking a ship
type DockResponse struct {
	Data DockData `json:"data"`
}

type DockData struct {
	Nav Navigation `json:"nav"`
}

// PatchNavRequest represents the request body for patching ship navigation
type PatchNavRequest struct {
	FlightMode string `json:"flightMode"`
}

// PatchNavResponse represents the response from patching ship navigation
type PatchNavResponse struct {
	Data Navigation `json:"data"`
}

// WarpResponse represents the response from warping a ship
type WarpResponse struct {
	Data WarpData `json:"data"`
}

type WarpData struct {
	Fuel  Fuel       `json:"fuel"`
	Nav   Navigation `json:"nav"`
	Event *Event     `json:"event,omitempty"`
}

// JumpRequest represents the request body for jumping a ship
type JumpRequest struct {
	SystemSymbol string `json:"systemSymbol"`
}

// RefuelRequest represents a ship refuel request
type RefuelRequest struct {
	Units     int  `json:"units,omitempty"`     // Optional: specific units to refuel (if not specified, refuels to capacity)
	FromCargo bool `json:"fromCargo,omitempty"` // Optional: refuel from cargo instead of marketplace
}

// ExtractRequest represents a ship extraction request
type ExtractRequest struct {
	Survey *Survey `json:"survey,omitempty"` // Optional: survey data to improve extraction
}

// Survey represents survey data for mining operations
type Survey struct {
	Signature  string          `json:"signature"`
	Symbol     string          `json:"symbol"`
	Deposits   []SurveyDeposit `json:"deposits"`
	Expiration string          `json:"expiration"`
	Size       string          `json:"size"`
}

// SurveyDeposit represents a mineral deposit in a survey
type SurveyDeposit struct {
	Symbol string `json:"symbol"`
}

// JettisonRequest represents a cargo jettison request
type JettisonRequest struct {
	Symbol string `json:"symbol"` // The cargo symbol to jettison
	Units  int    `json:"units"`  // Number of units to jettison
}

type SellCargoRequest struct {
	Symbol string `json:"symbol"` // The cargo symbol to sell
	Units  int    `json:"units"`  // Number of units to sell
}

type BuyCargoRequest struct {
	Symbol string `json:"symbol"` // The cargo symbol to buy
	Units  int    `json:"units"`  // Number of units to buy
}

// JumpResponse represents the response from jumping a ship
type JumpResponse struct {
	Data JumpData `json:"data"`
}

// RefuelResponse represents the response from refueling a ship
type RefuelResponse struct {
	Data RefuelData `json:"data"`
}

// RefuelData contains the refuel operation results
type RefuelData struct {
	Agent       Agent       `json:"agent"`
	Fuel        Fuel        `json:"fuel"`
	Transaction Transaction `json:"transaction"`
}

// ExtractResponse represents the response from extracting resources
type ExtractResponse struct {
	Data ExtractData `json:"data"`
}

// ExtractData contains the extraction operation results
type ExtractData struct {
	Cooldown   Cooldown   `json:"cooldown"`
	Extraction Extraction `json:"extraction"`
	Cargo      Cargo      `json:"cargo"`
	Events     []Event    `json:"events"`
}

// Extraction represents an extraction operation result
type Extraction struct {
	ShipSymbol string `json:"shipSymbol"`
	Yield      struct {
		Symbol string `json:"symbol"`
		Units  int    `json:"units"`
	} `json:"yield"`
}

// JettisonResponse represents the response from jettisoning cargo
type JettisonResponse struct {
	Data JettisonData `json:"data"`
}

// JettisonData contains the jettison operation results
type JettisonData struct {
	Cargo Cargo `json:"cargo"`
}

type SellCargoResponse struct {
	Data SellCargoData `json:"data"`
}

type SellCargoData struct {
	Agent       Agent             `json:"agent"`
	Cargo       Cargo             `json:"cargo"`
	Transaction MarketTransaction `json:"transaction"`
}

type BuyCargoResponse struct {
	Data BuyCargoData `json:"data"`
}

type BuyCargoData struct {
	Agent       Agent             `json:"agent"`
	Cargo       Cargo             `json:"cargo"`
	Transaction MarketTransaction `json:"transaction"`
}

type FulfillContractResponse struct {
	Data FulfillContractData `json:"data"`
}

type FulfillContractData struct {
	Agent    Agent    `json:"agent"`
	Contract Contract `json:"contract"`
}

type JumpData struct {
	Cooldown Cooldown   `json:"cooldown"`
	Nav      Navigation `json:"nav"`
	Event    *Event     `json:"event,omitempty"`
}

// Event represents an event that occurred during navigation
type Event struct {
	Symbol      string `json:"symbol"`
	Component   string `json:"component"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Scan-related structures
type ScanSystemsResponse struct {
	Data ScanSystemsData `json:"data"`
}

type ScanSystemsData struct {
	Cooldown Cooldown `json:"cooldown"`
	Systems  []System `json:"systems"`
}

type ScanWaypointsResponse struct {
	Data ScanWaypointsData `json:"data"`
}

type ScanWaypointsData struct {
	Cooldown  Cooldown         `json:"cooldown"`
	Waypoints []SystemWaypoint `json:"waypoints"`
}

type ScanShipsResponse struct {
	Data ScanShipsData `json:"data"`
}

type ScanShipsData struct {
	Cooldown Cooldown      `json:"cooldown"`
	Ships    []ScannedShip `json:"ships"`
}

type ScannedShip struct {
	Symbol       string       `json:"symbol"`
	Registration Registration `json:"registration"`
	Nav          Navigation   `json:"nav"`
	Frame        Frame        `json:"frame"`
	Reactor      Reactor      `json:"reactor"`
	Engine       Engine       `json:"engine"`
	Mounts       []Mount      `json:"mounts"`
}

// Repair-related structures
type RepairShipResponse struct {
	Data RepairShipData `json:"data"`
}

type RepairShipData struct {
	Agent       Agent       `json:"agent"`
	Ship        Ship        `json:"ship"`
	Transaction Transaction `json:"transaction"`
}

// Market data structures
type Market struct {
	Symbol       string              `json:"symbol"`
	Exports      []TradeGood         `json:"exports"`
	Imports      []TradeGood         `json:"imports"`
	Exchange     []TradeGood         `json:"exchange"`
	Transactions []MarketTransaction `json:"transactions"`
	TradeGoods   []MarketTradeGood   `json:"tradeGoods"`
}

type TradeGood struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type MarketTradeGood struct {
	Symbol        string `json:"symbol"`
	Type          string `json:"type"`
	TradeVolume   int    `json:"tradeVolume"`
	Supply        string `json:"supply"`
	Activity      string `json:"activity"`
	PurchasePrice int    `json:"purchasePrice"`
	SellPrice     int    `json:"sellPrice"`
}

type MarketTransaction struct {
	WaypointSymbol string `json:"waypointSymbol"`
	ShipSymbol     string `json:"shipSymbol"`
	TradeSymbol    string `json:"tradeSymbol"`
	Type           string `json:"type"`
	Units          int    `json:"units"`
	PricePerUnit   int    `json:"pricePerUnit"`
	TotalPrice     int    `json:"totalPrice"`
	Timestamp      string `json:"timestamp"`
}

type MarketResponse struct {
	Data Market `json:"data"`
}

// Faction represents a faction in the SpaceTraders universe
type Faction struct {
	Symbol       string `json:"symbol"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Headquarters string `json:"headquarters"`
	Traits       []struct {
		Symbol      string `json:"symbol"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"traits"`
	IsRecruiting bool `json:"isRecruiting"`
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

type SystemsResponse struct {
	Data []System `json:"data"`
	Meta struct {
		Total int `json:"total"`
		Page  int `json:"page"`
		Limit int `json:"limit"`
	} `json:"meta"`
}

type SystemResponse struct {
	Data System `json:"data"`
}

type FactionsResponse struct {
	Data []Faction `json:"data"`
	Meta struct {
		Total int `json:"total"`
		Page  int `json:"page"`
		Limit int `json:"limit"`
	} `json:"meta"`
}

type FactionResponse struct {
	Data Faction `json:"data"`
}

// NewClient creates a new SpaceTraders client
func NewClient(apiToken string) *Client {
	return &Client{
		APIToken: apiToken,
		BaseURL:  "https://api.spacetraders.io/v2",
	}
}

// NewClientWithBaseURL creates a new SpaceTraders client with a custom base URL (for testing)
func NewClientWithBaseURL(apiToken, baseURL string) *Client {
	return &Client{
		APIToken: apiToken,
		BaseURL:  baseURL,
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

// GetAllShips fetches all ships for the agent with pagination
func (c *Client) GetAllShips() ([]Ship, error) {
	var allShips []Ship
	page := 1
	limit := 20 // SpaceTraders default page size

	for {
		endpoint := fmt.Sprintf("/my/ships?page=%d&limit=%d", page, limit)

		resp, err := c.makeRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()

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

		allShips = append(allShips, shipsResp.Data...)

		// Check if we have all ships
		if len(allShips) >= shipsResp.Meta.Total {
			break
		}

		page++
	}

	return allShips, nil
}

// GetAllContracts fetches all contracts for the agent with pagination
func (c *Client) GetAllContracts() ([]Contract, error) {
	var allContracts []Contract
	page := 1
	limit := 20 // SpaceTraders default page size

	for {
		endpoint := fmt.Sprintf("/my/contracts?page=%d&limit=%d", page, limit)

		resp, err := c.makeRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()

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

		allContracts = append(allContracts, contractsResp.Data...)

		// Check if we have all contracts
		if len(allContracts) >= contractsResp.Meta.Total {
			break
		}

		page++
	}

	return allContracts, nil
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

// GetAllSystemWaypoints fetches all waypoints in a system with pagination
func (c *Client) GetAllSystemWaypoints(systemSymbol string) ([]SystemWaypoint, error) {
	var allWaypoints []SystemWaypoint
	page := 1
	limit := 20 // SpaceTraders default page size

	for {
		endpoint := fmt.Sprintf("/systems/%s/waypoints?page=%d&limit=%d", systemSymbol, page, limit)

		resp, err := c.makeRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()

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

		allWaypoints = append(allWaypoints, waypointsResp.Data...)

		// Check if we have all waypoints
		if len(allWaypoints) >= waypointsResp.Meta.Total {
			break
		}

		page++
	}

	return allWaypoints, nil
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

// GetMarket fetches market information for a waypoint
func (c *Client) GetMarket(systemSymbol, waypointSymbol string) (*Market, error) {
	endpoint := fmt.Sprintf("/systems/%s/waypoints/%s/market", systemSymbol, waypointSymbol)

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

	var marketResp MarketResponse
	if err := json.Unmarshal(body, &marketResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &marketResp.Data, nil
}

// PurchaseShip purchases a ship at a shipyard
func (c *Client) PurchaseShip(shipType, waypointSymbol string) (*Ship, *Agent, *Transaction, error) {
	endpoint := "/my/ships"

	// Create the purchase request
	purchaseReq := PurchaseShipRequest{
		ShipType:       shipType,
		WaypointSymbol: waypointSymbol,
	}

	// Marshal the request body
	requestBody, err := json.Marshal(purchaseReq)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to marshal purchase request: %w", err)
	}

	resp, err := c.makeRequest("POST", endpoint, strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, nil, nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var purchaseResp PurchaseShipResponse
	if err := json.Unmarshal(body, &purchaseResp); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &purchaseResp.Data.Ship, &purchaseResp.Data.Agent, &purchaseResp.Data.Transaction, nil
}

// OrbitShip puts a ship into orbit around a waypoint
func (c *Client) OrbitShip(shipSymbol string) (*Navigation, error) {
	endpoint := fmt.Sprintf("/my/ships/%s/orbit", shipSymbol)

	// SpaceTraders API requires an empty JSON object for POST requests
	emptyBody := strings.NewReader("{}")
	resp, err := c.makeRequest("POST", endpoint, emptyBody)
	if err != nil {
		return nil, fmt.Errorf("failed to orbit ship: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var orbitResp OrbitResponse
	if err := json.NewDecoder(resp.Body).Decode(&orbitResp); err != nil {
		return nil, fmt.Errorf("failed to decode orbit response: %w", err)
	}

	return &orbitResp.Data.Nav, nil
}

// DockShip docks a ship at a waypoint
func (c *Client) DockShip(shipSymbol string) (*Navigation, error) {
	endpoint := fmt.Sprintf("/my/ships/%s/dock", shipSymbol)

	// SpaceTraders API requires an empty JSON object for POST requests
	emptyBody := strings.NewReader("{}")
	resp, err := c.makeRequest("POST", endpoint, emptyBody)
	if err != nil {
		return nil, fmt.Errorf("failed to dock ship: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var dockResp DockResponse
	if err := json.NewDecoder(resp.Body).Decode(&dockResp); err != nil {
		return nil, fmt.Errorf("failed to decode dock response: %w", err)
	}

	return &dockResp.Data.Nav, nil
}

// NavigateShip navigates a ship to a waypoint
func (c *Client) NavigateShip(shipSymbol, waypointSymbol string) (*Navigation, *Fuel, *Event, error) {
	endpoint := fmt.Sprintf("/my/ships/%s/navigate", shipSymbol)

	reqBody := NavigateRequest{
		WaypointSymbol: waypointSymbol,
	}

	// Marshal the request body
	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to marshal navigate request: %w", err)
	}

	resp, err := c.makeRequest("POST", endpoint, strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to navigate ship: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var navResp NavigateResponse
	if err := json.NewDecoder(resp.Body).Decode(&navResp); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode navigate response: %w", err)
	}

	return &navResp.Data.Nav, &navResp.Data.Fuel, navResp.Data.Event, nil
}

// PatchShipNav updates a ship's navigation settings (e.g., flight mode)
func (c *Client) PatchShipNav(shipSymbol, flightMode string) (*Navigation, error) {
	endpoint := fmt.Sprintf("/my/ships/%s/nav", shipSymbol)

	reqBody := PatchNavRequest{
		FlightMode: flightMode,
	}

	// Marshal the request body
	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal patch nav request: %w", err)
	}

	resp, err := c.makeRequest("PATCH", endpoint, strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to patch ship nav: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var patchResp PatchNavResponse
	if err := json.NewDecoder(resp.Body).Decode(&patchResp); err != nil {
		return nil, fmt.Errorf("failed to decode patch nav response: %w", err)
	}

	return &patchResp.Data, nil
}

// WarpShip warps a ship to a waypoint (requires warp drive)
func (c *Client) WarpShip(shipSymbol, waypointSymbol string) (*Navigation, *Fuel, *Event, error) {
	endpoint := fmt.Sprintf("/my/ships/%s/warp", shipSymbol)

	reqBody := NavigateRequest{
		WaypointSymbol: waypointSymbol,
	}

	// Marshal the request body
	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to marshal warp request: %w", err)
	}

	resp, err := c.makeRequest("POST", endpoint, strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to warp ship: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var warpResp WarpResponse
	if err := json.NewDecoder(resp.Body).Decode(&warpResp); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode warp response: %w", err)
	}

	return &warpResp.Data.Nav, &warpResp.Data.Fuel, warpResp.Data.Event, nil
}

// JumpShip jumps a ship to a different system (requires jump drive)
func (c *Client) JumpShip(shipSymbol, systemSymbol string) (*Navigation, *Cooldown, *Event, error) {
	endpoint := fmt.Sprintf("/my/ships/%s/jump", shipSymbol)

	reqBody := JumpRequest{
		SystemSymbol: systemSymbol,
	}

	// Marshal the request body
	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to marshal jump request: %w", err)
	}

	resp, err := c.makeRequest("POST", endpoint, strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to jump ship: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var jumpResp JumpResponse
	if err := json.NewDecoder(resp.Body).Decode(&jumpResp); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode jump response: %w", err)
	}

	return &jumpResp.Data.Nav, &jumpResp.Data.Cooldown, jumpResp.Data.Event, nil
}

// RefuelShip refuels a ship at the current waypoint
func (c *Client) RefuelShip(shipSymbol string, units int, fromCargo bool) (*Agent, *Fuel, *Transaction, error) {
	endpoint := fmt.Sprintf("/my/ships/%s/refuel", shipSymbol)

	reqBody := RefuelRequest{}
	if units > 0 {
		reqBody.Units = units
	}
	if fromCargo {
		reqBody.FromCargo = fromCargo
	}

	// Marshal the request body
	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to marshal refuel request: %w", err)
	}

	resp, err := c.makeRequest("POST", endpoint, strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to refuel ship: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var refuelResp RefuelResponse
	if err := json.NewDecoder(resp.Body).Decode(&refuelResp); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode refuel response: %w", err)
	}

	return &refuelResp.Data.Agent, &refuelResp.Data.Fuel, &refuelResp.Data.Transaction, nil
}

// ExtractResources extracts resources from the current waypoint (asteroid, etc.)
func (c *Client) ExtractResources(shipSymbol string, survey *Survey) (*Cooldown, *Extraction, *Cargo, []Event, error) {
	endpoint := fmt.Sprintf("/my/ships/%s/extract", shipSymbol)

	reqBody := ExtractRequest{}
	if survey != nil {
		reqBody.Survey = survey
	}

	// Marshal the request body
	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to marshal extract request: %w", err)
	}

	resp, err := c.makeRequest("POST", endpoint, strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to extract resources: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, nil, nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var extractResp ExtractResponse
	if err := json.NewDecoder(resp.Body).Decode(&extractResp); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to decode extract response: %w", err)
	}

	return &extractResp.Data.Cooldown, &extractResp.Data.Extraction, &extractResp.Data.Cargo, extractResp.Data.Events, nil
}

// JettisonCargo jettisons cargo from a ship
func (c *Client) JettisonCargo(shipSymbol, cargoSymbol string, units int) (*Cargo, error) {
	endpoint := fmt.Sprintf("/my/ships/%s/jettison", shipSymbol)

	reqBody := JettisonRequest{
		Symbol: cargoSymbol,
		Units:  units,
	}

	// Marshal the request body
	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal jettison request: %w", err)
	}

	resp, err := c.makeRequest("POST", endpoint, strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to jettison cargo: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var jettisonResp JettisonResponse
	if err := json.NewDecoder(resp.Body).Decode(&jettisonResp); err != nil {
		return nil, fmt.Errorf("failed to decode jettison response: %w", err)
	}

	return &jettisonResp.Data.Cargo, nil
}

// GetAllSystems gets a list of all systems in the universe with pagination
func (c *Client) GetAllSystems() ([]System, error) {
	var allSystems []System
	page := 1
	limit := 20 // SpaceTraders default page size

	for {
		endpoint := fmt.Sprintf("/systems?page=%d&limit=%d", page, limit)

		resp, err := c.makeRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}

		var systemsResp SystemsResponse
		if err := json.NewDecoder(resp.Body).Decode(&systemsResp); err != nil {
			return nil, fmt.Errorf("failed to decode systems response: %w", err)
		}

		allSystems = append(allSystems, systemsResp.Data...)

		// Check if we have all systems
		if len(allSystems) >= systemsResp.Meta.Total {
			break
		}

		page++
	}

	return allSystems, nil
}

// GetSystem gets detailed information about a specific system
func (c *Client) GetSystem(systemSymbol string) (*System, error) {
	endpoint := fmt.Sprintf("/systems/%s", systemSymbol)

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

	var systemResp SystemResponse
	if err := json.NewDecoder(resp.Body).Decode(&systemResp); err != nil {
		return nil, fmt.Errorf("failed to decode system response: %w", err)
	}

	return &systemResp.Data, nil
}

// GetAllFactions gets a list of all factions in the universe with pagination
func (c *Client) GetAllFactions() ([]Faction, error) {
	var allFactions []Faction
	page := 1
	limit := 20 // SpaceTraders default page size

	for {
		endpoint := fmt.Sprintf("/factions?page=%d&limit=%d", page, limit)

		resp, err := c.makeRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}

		var factionsResp FactionsResponse
		if err := json.NewDecoder(resp.Body).Decode(&factionsResp); err != nil {
			return nil, fmt.Errorf("failed to decode factions response: %w", err)
		}

		allFactions = append(allFactions, factionsResp.Data...)

		// Check if we have all factions
		if len(allFactions) >= factionsResp.Meta.Total {
			break
		}

		page++
	}

	return allFactions, nil
}

// GetFaction gets detailed information about a specific faction
func (c *Client) GetFaction(factionSymbol string) (*Faction, error) {
	endpoint := fmt.Sprintf("/factions/%s", factionSymbol)

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

	var factionResp FactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&factionResp); err != nil {
		return nil, fmt.Errorf("failed to decode faction response: %w", err)
	}

	return &factionResp.Data, nil
}

// SellCargo sells cargo from a ship at a market
func (c *Client) SellCargo(shipSymbol, cargoSymbol string, units int) (*Agent, *Cargo, *MarketTransaction, error) {
	endpoint := fmt.Sprintf("/my/ships/%s/sell", shipSymbol)

	reqBody := SellCargoRequest{
		Symbol: cargoSymbol,
		Units:  units,
	}

	// Marshal the request body
	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to marshal sell cargo request: %w", err)
	}

	resp, err := c.makeRequest("POST", endpoint, strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, nil, nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var sellResp SellCargoResponse
	if err := json.NewDecoder(resp.Body).Decode(&sellResp); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode sell cargo response: %w", err)
	}

	return &sellResp.Data.Agent, &sellResp.Data.Cargo, &sellResp.Data.Transaction, nil
}

// BuyCargo purchases cargo for a ship at a market
func (c *Client) BuyCargo(shipSymbol, cargoSymbol string, units int) (*Agent, *Cargo, *MarketTransaction, error) {
	endpoint := fmt.Sprintf("/my/ships/%s/purchase", shipSymbol)

	reqBody := BuyCargoRequest{
		Symbol: cargoSymbol,
		Units:  units,
	}

	// Marshal the request body
	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to marshal buy cargo request: %w", err)
	}

	resp, err := c.makeRequest("POST", endpoint, strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, nil, nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var buyResp BuyCargoResponse
	if err := json.NewDecoder(resp.Body).Decode(&buyResp); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode buy cargo response: %w", err)
	}

	return &buyResp.Data.Agent, &buyResp.Data.Cargo, &buyResp.Data.Transaction, nil
}

// FulfillContract fulfills a contract by delivering the required cargo
func (c *Client) FulfillContract(contractID string) (*Agent, *Contract, error) {
	endpoint := fmt.Sprintf("/my/contracts/%s/fulfill", contractID)

	resp, err := c.makeRequest("POST", endpoint, nil)
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

	var fulfillResp FulfillContractResponse
	if err := json.NewDecoder(resp.Body).Decode(&fulfillResp); err != nil {
		return nil, nil, fmt.Errorf("failed to decode fulfill contract response: %w", err)
	}

	return &fulfillResp.Data.Agent, &fulfillResp.Data.Contract, nil
}

// ScanSystems scans for systems around the ship
func (c *Client) ScanSystems(shipSymbol string) (*ScanSystemsData, error) {
	url := fmt.Sprintf("/my/ships/%s/scan/systems", shipSymbol)
	resp, err := c.makeRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to scan systems: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var scanResp ScanSystemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&scanResp); err != nil {
		return nil, fmt.Errorf("failed to decode scan systems response: %w", err)
	}

	return &scanResp.Data, nil
}

// ScanWaypoints scans for waypoints around the ship
func (c *Client) ScanWaypoints(shipSymbol string) (*ScanWaypointsData, error) {
	url := fmt.Sprintf("/my/ships/%s/scan/waypoints", shipSymbol)
	resp, err := c.makeRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to scan waypoints: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var scanResp ScanWaypointsResponse
	if err := json.NewDecoder(resp.Body).Decode(&scanResp); err != nil {
		return nil, fmt.Errorf("failed to decode scan waypoints response: %w", err)
	}

	return &scanResp.Data, nil
}

// ScanShips scans for ships around the ship
func (c *Client) ScanShips(shipSymbol string) (*ScanShipsData, error) {
	url := fmt.Sprintf("/my/ships/%s/scan/ships", shipSymbol)
	resp, err := c.makeRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to scan ships: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var scanResp ScanShipsResponse
	if err := json.NewDecoder(resp.Body).Decode(&scanResp); err != nil {
		return nil, fmt.Errorf("failed to decode scan ships response: %w", err)
	}

	return &scanResp.Data, nil
}

// RepairShip repairs a ship at a shipyard
func (c *Client) RepairShip(shipSymbol string) (*Agent, *Ship, *Transaction, error) {
	url := fmt.Sprintf("/my/ships/%s/repair", shipSymbol)
	resp, err := c.makeRequest("POST", url, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to repair ship: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var repairResp RepairShipResponse
	if err := json.NewDecoder(resp.Body).Decode(&repairResp); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode repair ship response: %w", err)
	}

	return &repairResp.Data.Agent, &repairResp.Data.Ship, &repairResp.Data.Transaction, nil
}
