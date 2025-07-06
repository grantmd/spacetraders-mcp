package client

// Agent represents agent information with correct types
type Agent struct {
	AccountID       *string `json:"accountId,omitempty"`
	Symbol          string  `json:"symbol"`
	Headquarters    string  `json:"headquarters"`
	Credits         int64   `json:"credits"`
	StartingFaction string  `json:"startingFaction"`
	ShipCount       int     `json:"shipCount"`
}

// Ship represents a ship with FIXED reactor integrity types
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

// Registration represents ship registration
type Registration struct {
	Name          string `json:"name"`
	FactionSymbol string `json:"factionSymbol"`
	Role          string `json:"role"`
}

// Navigation represents ship navigation
type Navigation struct {
	SystemSymbol   string `json:"systemSymbol"`
	WaypointSymbol string `json:"waypointSymbol"`
	Route          Route  `json:"route"`
	Status         string `json:"status"`
	FlightMode     string `json:"flightMode"`
}

// Route represents navigation route
type Route struct {
	Destination   Waypoint `json:"destination"`
	Origin        Waypoint `json:"origin"`
	DepartureTime string   `json:"departureTime"`
	Arrival       string   `json:"arrival"`
}

// Waypoint represents a waypoint location
type Waypoint struct {
	Symbol string `json:"symbol"`
	Type   string `json:"type"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
}

// Crew represents ship crew
type Crew struct {
	Current  int    `json:"current"`
	Required int    `json:"required"`
	Capacity int    `json:"capacity"`
	Rotation string `json:"rotation"`
	Morale   int    `json:"morale"`
	Wages    int    `json:"wages"`
}

// Frame represents ship frame with FIXED integrity type
type Frame struct {
	Symbol         string           `json:"symbol"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	Condition      float64          `json:"condition"` // ✅ FIXED: was int, now float64
	Integrity      float64          `json:"integrity"` // ✅ FIXED: was int, now float64
	ModuleSlots    int              `json:"moduleSlots"`
	MountingPoints int              `json:"mountingPoints"`
	FuelCapacity   int              `json:"fuelCapacity"`
	Requirements   ShipRequirements `json:"requirements"`
	Quality        float32          `json:"quality"` // ✅ NEW: was missing
}

// Reactor represents ship reactor with FIXED integrity type
type Reactor struct {
	Symbol       string           `json:"symbol"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Condition    float64          `json:"condition"` // ✅ FIXED: was int, now float64
	Integrity    float64          `json:"integrity"` // ✅ FIXED: was int, now float64
	PowerOutput  int              `json:"powerOutput"`
	Requirements ShipRequirements `json:"requirements"`
	Quality      float32          `json:"quality"` // ✅ NEW: was missing
}

// Engine represents ship engine with FIXED integrity type
type Engine struct {
	Symbol       string           `json:"symbol"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Condition    float64          `json:"condition"` // ✅ FIXED: was int, now float64
	Integrity    float64          `json:"integrity"` // ✅ FIXED: was int, now float64
	Speed        int              `json:"speed"`
	Requirements ShipRequirements `json:"requirements"`
	Quality      float32          `json:"quality"` // ✅ NEW: was missing
}

// ShipRequirements represents component requirements
type ShipRequirements struct {
	Power int `json:"power"`
	Crew  int `json:"crew"`
	Slots int `json:"slots"`
}

// Cooldown represents ship cooldown
type Cooldown struct {
	ShipSymbol       string `json:"shipSymbol"`
	TotalSeconds     int    `json:"totalSeconds"`
	RemainingSeconds int    `json:"remainingSeconds"`
	Expiration       string `json:"expiration,omitempty"`
}

// Module represents a ship module
type Module struct {
	Symbol       string           `json:"symbol"`
	Capacity     int              `json:"capacity,omitempty"`
	Range        int              `json:"range,omitempty"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Requirements ShipRequirements `json:"requirements"`
}

// Mount represents a ship mount
type Mount struct {
	Symbol       string           `json:"symbol"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Strength     int              `json:"strength,omitempty"`
	Deposits     []string         `json:"deposits,omitempty"`
	Requirements ShipRequirements `json:"requirements"`
}

// Cargo represents ship cargo
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

// Fuel represents ship fuel
type Fuel struct {
	Current  int           `json:"current"`
	Capacity int           `json:"capacity"`
	Consumed *FuelConsumed `json:"consumed,omitempty"`
}

// FuelConsumed represents consumed fuel
type FuelConsumed struct {
	Amount    int    `json:"amount"`
	Timestamp string `json:"timestamp"`
}

// Contract represents a contract
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
	Deliver  []ContractDeliverGood `json:"deliver,omitempty"`
}

// ContractPayment represents contract payment
type ContractPayment struct {
	OnAccepted  int `json:"onAccepted"`
	OnFulfilled int `json:"onFulfilled"`
}

// ContractDeliverGood represents a delivery requirement
type ContractDeliverGood struct {
	TradeSymbol       string `json:"tradeSymbol"`
	DestinationSymbol string `json:"destinationSymbol"`
	UnitsRequired     int    `json:"unitsRequired"`
	UnitsFulfilled    int    `json:"unitsFulfilled"`
}

// AcceptContractResponse represents the response from accepting a contract
type AcceptContractResponse struct {
	Data AcceptContractData `json:"data"`
}

// AcceptContractData represents the data from accepting a contract
type AcceptContractData struct {
	Contract Contract `json:"contract"`
	Agent    Agent    `json:"agent"`
}

// System represents a star system
type System struct {
	Symbol       string           `json:"symbol"`
	SectorSymbol string           `json:"sectorSymbol"`
	Type         string           `json:"type"`
	X            int              `json:"x"`
	Y            int              `json:"y"`
	Waypoints    []SystemWaypoint `json:"waypoints"`
	Factions     []SystemFaction  `json:"factions"`
}

// SystemWaypoint represents a waypoint in a system
type SystemWaypoint struct {
	Symbol    string             `json:"symbol"`
	Type      string             `json:"type"`
	X         int                `json:"x"`
	Y         int                `json:"y"`
	Orbitals  []WaypointOrbital  `json:"orbitals"`
	Traits    []WaypointTrait    `json:"traits"`
	Modifiers []WaypointModifier `json:"modifiers"`
	Chart     *WaypointChart     `json:"chart,omitempty"`
	Faction   *WaypointFaction   `json:"faction,omitempty"`
}

// WaypointOrbital represents an orbital waypoint
type WaypointOrbital struct {
	Symbol string `json:"symbol"`
}

// WaypointTrait represents a waypoint trait
type WaypointTrait struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// WaypointModifier represents a waypoint modifier
type WaypointModifier struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// WaypointChart represents waypoint chart information
type WaypointChart struct {
	WaypointSymbol string `json:"waypointSymbol"`
	SubmittedBy    string `json:"submittedBy"`
	SubmittedOn    string `json:"submittedOn"`
}

// WaypointFaction represents the faction controlling a waypoint
type WaypointFaction struct {
	Symbol string `json:"symbol"`
}

// SystemFaction represents a faction in a system
type SystemFaction struct {
	Symbol string `json:"symbol"`
}

// Shipyard represents a shipyard
type Shipyard struct {
	Symbol           string                `json:"symbol"`
	ShipTypes        []ShipyardShipType    `json:"shipTypes"`
	Transactions     []ShipyardTransaction `json:"transactions"`
	Ships            []ShipyardShip        `json:"ships"`
	ModificationsFee int                   `json:"modificationsFee"`
}

// ShipyardShipType represents available ship type
type ShipyardShipType struct {
	Type string `json:"type"`
}

// ShipyardTransaction represents a shipyard transaction
type ShipyardTransaction struct {
	WaypointSymbol string `json:"waypointSymbol"`
	ShipSymbol     string `json:"shipSymbol"`
	ShipType       string `json:"shipType"`
	Price          int    `json:"price"`
	AgentSymbol    string `json:"agentSymbol"`
	Timestamp      string `json:"timestamp"`
}

// ShipyardShip represents a ship available at shipyard
type ShipyardShip struct {
	Type          string               `json:"type"`
	Name          string               `json:"name"`
	Description   string               `json:"description"`
	Supply        string               `json:"supply"`
	Activity      string               `json:"activity"`
	PurchasePrice int                  `json:"purchasePrice"`
	Frame         ShipyardShipFrame    `json:"frame"`
	Reactor       ShipyardShipReactor  `json:"reactor"`
	Engine        ShipyardShipEngine   `json:"engine"`
	Modules       []ShipyardShipModule `json:"modules"`
	Mounts        []ShipyardShipMount  `json:"mounts"`
	Crew          ShipyardShipCrew     `json:"crew"`
}

// ShipyardShipFrame represents frame available at shipyard
type ShipyardShipFrame struct {
	Symbol         string           `json:"symbol"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	ModuleSlots    int              `json:"moduleSlots"`
	MountingPoints int              `json:"mountingPoints"`
	FuelCapacity   int              `json:"fuelCapacity"`
	Condition      float64          `json:"condition"`
	Integrity      float64          `json:"integrity"`
	Requirements   ShipRequirements `json:"requirements"`
}

// ShipyardShipReactor represents reactor available at shipyard
type ShipyardShipReactor struct {
	Symbol       string           `json:"symbol"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Condition    float64          `json:"condition"`
	Integrity    float64          `json:"integrity"`
	PowerOutput  int              `json:"powerOutput"`
	Requirements ShipRequirements `json:"requirements"`
}

// ShipyardShipEngine represents engine available at shipyard
type ShipyardShipEngine struct {
	Symbol       string           `json:"symbol"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Condition    float64          `json:"condition"`
	Integrity    float64          `json:"integrity"`
	Speed        int              `json:"speed"`
	Requirements ShipRequirements `json:"requirements"`
}

// ShipyardShipModule represents module available at shipyard
type ShipyardShipModule struct {
	Symbol       string           `json:"symbol"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Capacity     int              `json:"capacity,omitempty"`
	Range        int              `json:"range,omitempty"`
	Requirements ShipRequirements `json:"requirements"`
}

// ShipyardShipMount represents mount available at shipyard
type ShipyardShipMount struct {
	Symbol       string           `json:"symbol"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Strength     int              `json:"strength,omitempty"`
	Deposits     []string         `json:"deposits,omitempty"`
	Requirements ShipRequirements `json:"requirements"`
}

// ShipyardShipCrew represents crew requirements
type ShipyardShipCrew struct {
	Required int `json:"required"`
	Capacity int `json:"capacity"`
}

// Market represents a marketplace
type Market struct {
	Symbol       string              `json:"symbol"`
	Exports      []TradeGood         `json:"exports"`
	Imports      []TradeGood         `json:"imports"`
	Exchange     []TradeGood         `json:"exchange"`
	Transactions []MarketTransaction `json:"transactions"`
	TradeGoods   []MarketTradeGood   `json:"tradeGoods"`
}

// TradeGood represents a tradeable good
type TradeGood struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// MarketTradeGood represents trade good with market data
type MarketTradeGood struct {
	Symbol        string `json:"symbol"`
	Type          string `json:"type"`
	TradeVolume   int    `json:"tradeVolume"`
	Supply        string `json:"supply"`
	Activity      string `json:"activity"`
	PurchasePrice int    `json:"purchasePrice"`
	SellPrice     int    `json:"sellPrice"`
}

// MarketTransaction represents a market transaction
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

// Faction represents a faction
type Faction struct {
	Symbol       string         `json:"symbol"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Headquarters string         `json:"headquarters"`
	Traits       []FactionTrait `json:"traits"`
	IsRecruiting bool           `json:"isRecruiting"`
}

// FactionTrait represents a faction trait
type FactionTrait struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// PurchaseShipRequest represents a ship purchase request
type PurchaseShipRequest struct {
	ShipType       string `json:"shipType"`
	WaypointSymbol string `json:"waypointSymbol"`
}

// PurchaseShipResponse represents ship purchase response
type PurchaseShipResponse struct {
	Data PurchaseShipData `json:"data"`
}

// PurchaseShipData represents ship purchase data
type PurchaseShipData struct {
	Agent       Agent       `json:"agent"`
	Ship        Ship        `json:"ship"`
	Transaction Transaction `json:"transaction"`
}

// Transaction represents a transaction
type Transaction struct {
	WaypointSymbol string `json:"waypointSymbol"`
	ShipSymbol     string `json:"shipSymbol"`
	ShipType       string `json:"shipType"`
	Price          int    `json:"price"`
	AgentSymbol    string `json:"agentSymbol"`
	Timestamp      string `json:"timestamp"`
}

// NavigateResponse represents navigation response
type NavigateResponse struct {
	Data NavigateData `json:"data"`
}

// NavigateData represents navigation data
type NavigateData struct {
	Fuel  Fuel       `json:"fuel"`
	Nav   Navigation `json:"nav"`
	Event Event      `json:"event"`
}

// OrbitResponse represents orbit response
type OrbitResponse struct {
	Data OrbitData `json:"data"`
}

// OrbitData represents orbit data
type OrbitData struct {
	Nav Navigation `json:"nav"`
}

// DockResponse represents dock response
type DockResponse struct {
	Data DockData `json:"data"`
}

// DockData represents dock data
type DockData struct {
	Nav Navigation `json:"nav"`
}

// Event represents a game event
type Event struct {
	Symbol      string `json:"symbol"`
	Component   string `json:"component"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SellCargoResponse represents sell cargo response
type SellCargoResponse struct {
	Data SellCargoData `json:"data"`
}

// SellCargoData represents sell cargo data
type SellCargoData struct {
	Agent       Agent             `json:"agent"`
	Cargo       Cargo             `json:"cargo"`
	Transaction MarketTransaction `json:"transaction"`
}

// BuyCargoResponse represents buy cargo response
type BuyCargoResponse struct {
	Data BuyCargoData `json:"data"`
}

// BuyCargoData represents buy cargo data
type BuyCargoData struct {
	Agent       Agent             `json:"agent"`
	Cargo       Cargo             `json:"cargo"`
	Transaction MarketTransaction `json:"transaction"`
}

// FulfillContractResponse represents fulfill contract response
type FulfillContractResponse struct {
	Data FulfillContractData `json:"data"`
}

// FulfillContractData represents fulfill contract data
type FulfillContractData struct {
	Agent    Agent    `json:"agent"`
	Contract Contract `json:"contract"`
}

type ExtractResponse struct {
	Data ExtractData `json:"data"`
}

type ExtractData struct {
	Cooldown   Cooldown   `json:"cooldown"`
	Extraction Extraction `json:"extraction"`
	Cargo      Cargo      `json:"cargo"`
	Events     []Event    `json:"events"`
}

type Extraction struct {
	ShipSymbol string          `json:"shipSymbol"`
	Yield      ExtractionYield `json:"yield"`
}

type ExtractionYield struct {
	Symbol string `json:"symbol"`
	Units  int    `json:"units"`
}

type JettisonResponse struct {
	Data JettisonData `json:"data"`
}

type JettisonData struct {
	Cargo Cargo `json:"cargo"`
}

type RefuelResponse struct {
	Data RefuelData `json:"data"`
}

type RefuelData struct {
	Agent       Agent             `json:"agent"`
	Fuel        Fuel              `json:"fuel"`
	Transaction MarketTransaction `json:"transaction"`
}

type ScanSystemsResponse struct {
	Data ScanSystemsData `json:"data"`
}

type ScanSystemsData struct {
	Cooldown Cooldown        `json:"cooldown"`
	Systems  []ScannedSystem `json:"systems"`
}

type ScannedSystem struct {
	Symbol       string `json:"symbol"`
	SectorSymbol string `json:"sectorSymbol"`
	Type         string `json:"type"`
	X            int    `json:"x"`
	Y            int    `json:"y"`
	Distance     int    `json:"distance"`
}

type ScanWaypointsResponse struct {
	Data ScanWaypointsData `json:"data"`
}

type ScanWaypointsData struct {
	Cooldown  Cooldown          `json:"cooldown"`
	Waypoints []ScannedWaypoint `json:"waypoints"`
}

type ScannedWaypoint struct {
	Symbol       string            `json:"symbol"`
	Type         string            `json:"type"`
	SystemSymbol string            `json:"systemSymbol"`
	X            int               `json:"x"`
	Y            int               `json:"y"`
	Orbitals     []WaypointOrbital `json:"orbitals"`
	Faction      *WaypointFaction  `json:"faction"`
	Traits       []WaypointTrait   `json:"traits"`
	Chart        *WaypointChart    `json:"chart"`
}

type ScanShipsResponse struct {
	Data ScanShipsData `json:"data"`
}

type ScanShipsData struct {
	Cooldown Cooldown      `json:"cooldown"`
	Ships    []ScannedShip `json:"ships"`
}

type ScannedShip struct {
	Symbol       string              `json:"symbol"`
	Registration Registration        `json:"registration"`
	Nav          Navigation          `json:"nav"`
	Frame        *ScannedShipFrame   `json:"frame"`
	Reactor      *ScannedShipReactor `json:"reactor"`
	Engine       *ScannedShipEngine  `json:"engine"`
	Mounts       []ScannedShipMount  `json:"mounts"`
}

type ScannedShipFrame struct {
	Symbol string `json:"symbol"`
}

type ScannedShipReactor struct {
	Symbol string `json:"symbol"`
}

type ScannedShipEngine struct {
	Symbol string `json:"symbol"`
}

type ScannedShipMount struct {
	Symbol string `json:"symbol"`
}

type RepairShipResponse struct {
	Data RepairShipData `json:"data"`
}

type RepairShipData struct {
	Agent       Agent             `json:"agent"`
	Ship        Ship              `json:"ship"`
	Transaction RepairTransaction `json:"transaction"`
}

type RepairTransaction struct {
	WaypointSymbol string `json:"waypointSymbol"`
	ShipSymbol     string `json:"shipSymbol"`
	TotalPrice     int    `json:"totalPrice"`
	Timestamp      string `json:"timestamp"`
}

type JumpResponse struct {
	Data JumpData `json:"data"`
}

type JumpData struct {
	Cooldown Cooldown   `json:"cooldown"`
	Nav      Navigation `json:"nav"`
	Event    Event      `json:"event"`
}

type WarpResponse struct {
	Data WarpData `json:"data"`
}

type WarpData struct {
	Fuel  Fuel       `json:"fuel"`
	Nav   Navigation `json:"nav"`
	Event Event      `json:"event"`
}

type PatchNavRequest struct {
	FlightMode string `json:"flightMode"`
}

type PatchNavResponse struct {
	Data Navigation `json:"data"`
}

type Survey struct {
	Signature  string          `json:"signature"`
	Symbol     string          `json:"symbol"`
	Deposits   []SurveyDeposit `json:"deposits"`
	Expiration string          `json:"expiration"`
	Size       string          `json:"size"`
}

type SurveyDeposit struct {
	Symbol string `json:"symbol"`
}
