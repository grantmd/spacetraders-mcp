package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/your-username/spacetraders-mcp/generated/spacetraders"
)

// Client wraps the generated OpenAPI client to provide a compatible interface
// with the existing manual client while fixing type issues like reactor integrity.
type Client struct {
	apiClient *spacetraders.APIClient
	ctx       context.Context
}

// NewClient creates a new SpaceTraders client using the generated OpenAPI client
func NewClient(apiToken string) *Client {
	return NewClientWithBaseURL(apiToken, "https://api.spacetraders.io/v2")
}

// NewClientWithBaseURL creates a new SpaceTraders client with a custom base URL (for testing)
func NewClientWithBaseURL(apiToken, baseURL string) *Client {
	cfg := spacetraders.NewConfiguration()
	cfg.AddDefaultHeader("Authorization", "Bearer "+apiToken)
	cfg.Servers = []spacetraders.ServerConfiguration{
		{URL: baseURL},
	}
	cfg.HTTPClient = &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Client{
		apiClient: spacetraders.NewAPIClient(cfg),
		ctx:       context.Background(),
	}
}

// GetAgent returns the agent information
func (c *Client) GetAgent() (*Agent, error) {
	resp, _, err := c.apiClient.AgentsAPI.GetMyAgent(c.ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	return &Agent{
		AccountID:       resp.Data.AccountId,
		Symbol:          resp.Data.Symbol,
		Headquarters:    resp.Data.Headquarters,
		Credits:         resp.Data.Credits,
		StartingFaction: resp.Data.StartingFaction,
		ShipCount:       int(resp.Data.ShipCount),
	}, nil
}

// GetAllShips returns all ships for the agent
func (c *Client) GetAllShips() ([]Ship, error) {
	var allShips []Ship
	page := int32(1)
	limit := int32(20)

	for {
		resp, _, err := c.apiClient.FleetAPI.GetMyShips(c.ctx).Page(page).Limit(limit).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to get ships: %w", err)
		}

		for _, ship := range resp.Data {
			convertedShip := Ship{
				Symbol:       ship.Symbol,
				Registration: convertRegistration(ship.Registration),
				Nav:          convertNavigation(ship.Nav),
				Crew:         convertCrew(ship.Crew),
				Frame:        convertFrame(ship.Frame),
				Reactor:      convertReactor(ship.Reactor),
				Engine:       convertEngine(ship.Engine),
				Cooldown:     convertCooldown(ship.Cooldown),
				Modules:      convertModules(ship.Modules),
				Mounts:       convertMounts(ship.Mounts),
				Cargo:        convertCargo(ship.Cargo),
				Fuel:         convertFuel(ship.Fuel),
			}
			allShips = append(allShips, convertedShip)
		}

		// Check if we have more pages
		if len(resp.Data) < int(limit) || int32(len(allShips)) >= resp.Meta.Total {
			break
		}
		page++
	}

	return allShips, nil
}

// GetAllContracts returns all contracts for the agent
func (c *Client) GetAllContracts() ([]Contract, error) {
	var allContracts []Contract
	page := int32(1)
	limit := int32(20)

	for {
		resp, _, err := c.apiClient.ContractsAPI.GetContracts(c.ctx).Page(page).Limit(limit).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to get contracts: %w", err)
		}

		for _, contract := range resp.Data {
			var expiration, deadlineToAccept string
			expiration = contract.Expiration.Format("2006-01-02T15:04:05.000Z")
			if contract.DeadlineToAccept != nil {
				deadlineToAccept = contract.DeadlineToAccept.Format("2006-01-02T15:04:05.000Z")
			}

			convertedContract := Contract{
				ID:               contract.Id,
				FactionSymbol:    contract.FactionSymbol,
				Type:             contract.Type,
				Terms:            convertContractTerms(contract.Terms),
				Accepted:         contract.Accepted,
				Fulfilled:        contract.Fulfilled,
				Expiration:       expiration,
				DeadlineToAccept: deadlineToAccept,
			}
			allContracts = append(allContracts, convertedContract)
		}

		// Check if we have more pages
		if len(resp.Data) < int(limit) || int32(len(allContracts)) >= resp.Meta.Total {
			break
		}
		page++
	}

	return allContracts, nil
}

// AcceptContract accepts a contract by ID
func (c *Client) AcceptContract(contractID string) (*AcceptContractResponse, error) {
	resp, _, err := c.apiClient.ContractsAPI.AcceptContract(c.ctx, contractID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to accept contract: %w", err)
	}

	var expiration, deadlineToAccept string
	expiration = resp.Data.Contract.Expiration.Format("2006-01-02T15:04:05.000Z")
	if resp.Data.Contract.DeadlineToAccept != nil {
		deadlineToAccept = resp.Data.Contract.DeadlineToAccept.Format("2006-01-02T15:04:05.000Z")
	}

	return &AcceptContractResponse{
		Data: AcceptContractData{
			Contract: Contract{
				ID:               resp.Data.Contract.Id,
				FactionSymbol:    resp.Data.Contract.FactionSymbol,
				Type:             resp.Data.Contract.Type,
				Terms:            convertContractTerms(resp.Data.Contract.Terms),
				Accepted:         resp.Data.Contract.Accepted,
				Fulfilled:        resp.Data.Contract.Fulfilled,
				Expiration:       expiration,
				DeadlineToAccept: deadlineToAccept,
			},
			Agent: Agent{
				AccountID:       resp.Data.Agent.AccountId,
				Symbol:          resp.Data.Agent.Symbol,
				Headquarters:    resp.Data.Agent.Headquarters,
				Credits:         resp.Data.Agent.Credits,
				StartingFaction: resp.Data.Agent.StartingFaction,
				ShipCount:       int(resp.Data.Agent.ShipCount),
			},
		},
	}, nil
}

// GetAllSystemWaypoints returns all waypoints in a system
func (c *Client) GetAllSystemWaypoints(systemSymbol string) ([]SystemWaypoint, error) {
	var allWaypoints []SystemWaypoint
	page := int32(1)
	limit := int32(20)

	for {
		resp, _, err := c.apiClient.SystemsAPI.GetSystemWaypoints(c.ctx, systemSymbol).Page(page).Limit(limit).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to get system waypoints: %w", err)
		}

		for _, waypoint := range resp.Data {
			convertedWaypoint := SystemWaypoint{
				Symbol:    waypoint.Symbol,
				Type:      string(waypoint.Type),
				X:         int(waypoint.X),
				Y:         int(waypoint.Y),
				Orbitals:  convertOrbitals(waypoint.Orbitals),
				Traits:    convertWaypointTraits(waypoint.Traits),
				Modifiers: convertWaypointModifiers(waypoint.Modifiers),
				Chart:     convertChart(waypoint.Chart),
				Faction:   convertWaypointFaction(waypoint.Faction),
			}
			allWaypoints = append(allWaypoints, convertedWaypoint)
		}

		// Check if we have more pages
		if len(resp.Data) < int(limit) || int32(len(allWaypoints)) >= resp.Meta.Total {
			break
		}
		page++
	}

	return allWaypoints, nil
}

// GetShipyard returns shipyard information for a waypoint
func (c *Client) GetShipyard(systemSymbol, waypointSymbol string) (*Shipyard, error) {
	resp, _, err := c.apiClient.SystemsAPI.GetShipyard(c.ctx, systemSymbol, waypointSymbol).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get shipyard: %w", err)
	}

	return &Shipyard{
		Symbol:           resp.Data.Symbol,
		ShipTypes:        convertShipyardShipTypes(resp.Data.ShipTypes),
		Transactions:     convertShipyardTransactions(resp.Data.Transactions),
		Ships:            convertShipyardShips(resp.Data.Ships),
		ModificationsFee: int(resp.Data.ModificationsFee),
	}, nil
}

// GetMarket returns market information for a waypoint
func (c *Client) GetMarket(systemSymbol, waypointSymbol string) (*Market, error) {
	resp, _, err := c.apiClient.SystemsAPI.GetMarket(c.ctx, systemSymbol, waypointSymbol).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get market: %w", err)
	}

	return &Market{
		Symbol:       resp.Data.Symbol,
		Exports:      convertTradeGoods(resp.Data.Exports),
		Imports:      convertTradeGoods(resp.Data.Imports),
		Exchange:     convertTradeGoods(resp.Data.Exchange),
		Transactions: convertMarketTransactions(resp.Data.Transactions),
		TradeGoods:   convertMarketTradeGoods(resp.Data.TradeGoods),
	}, nil
}

// PurchaseShip purchases a new ship
func (c *Client) PurchaseShip(request PurchaseShipRequest) (*PurchaseShipResponse, error) {
	req := spacetraders.PurchaseShipRequest{
		ShipType:       spacetraders.ShipType(request.ShipType),
		WaypointSymbol: request.WaypointSymbol,
	}

	resp, _, err := c.apiClient.FleetAPI.PurchaseShip(c.ctx).PurchaseShipRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to purchase ship: %w", err)
	}

	return &PurchaseShipResponse{
		Data: PurchaseShipData{
			Agent: Agent{
				AccountID:       resp.Data.Agent.AccountId,
				Symbol:          resp.Data.Agent.Symbol,
				Headquarters:    resp.Data.Agent.Headquarters,
				Credits:         resp.Data.Agent.Credits,
				StartingFaction: resp.Data.Agent.StartingFaction,
				ShipCount:       int(resp.Data.Agent.ShipCount),
			},
			Ship:        convertShipFromGenerated(resp.Data.Ship),
			Transaction: convertTransactionFromGenerated(resp.Data.Transaction),
		},
	}, nil
}

// OrbitShip moves a ship to orbit
func (c *Client) OrbitShip(shipSymbol string) (*OrbitResponse, error) {
	resp, _, err := c.apiClient.FleetAPI.OrbitShip(c.ctx, shipSymbol).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to orbit ship: %w", err)
	}

	return &OrbitResponse{
		Data: OrbitData{
			Nav: convertNavigation(resp.Data.Nav),
		},
	}, nil
}

// DockShip docks a ship
func (c *Client) DockShip(shipSymbol string) (*DockResponse, error) {
	resp, _, err := c.apiClient.FleetAPI.DockShip(c.ctx, shipSymbol).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to dock ship: %w", err)
	}

	return &DockResponse{
		Data: DockData{
			Nav: convertNavigation(resp.Data.Nav),
		},
	}, nil
}

// NavigateShip navigates a ship to a waypoint
func (c *Client) NavigateShip(shipSymbol, waypointSymbol string) (*NavigateResponse, error) {
	req := spacetraders.NavigateShipRequest{
		WaypointSymbol: waypointSymbol,
	}

	resp, _, err := c.apiClient.FleetAPI.NavigateShip(c.ctx, shipSymbol).NavigateShipRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to navigate ship: %w", err)
	}

	return &NavigateResponse{
		Data: NavigateData{
			Fuel:  convertFuel(resp.Data.Fuel),
			Nav:   convertNavigation(resp.Data.Nav),
			Event: convertEvent(resp.Data.Events),
		},
	}, nil
}

// GetAllSystems returns all systems
func (c *Client) GetAllSystems() ([]System, error) {
	var allSystems []System
	page := int32(1)
	limit := int32(20)

	for {
		resp, _, err := c.apiClient.SystemsAPI.GetSystems(c.ctx).Page(page).Limit(limit).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to get systems: %w", err)
		}

		for _, system := range resp.Data {
			convertedSystem := System{
				Symbol:       system.Symbol,
				SectorSymbol: system.SectorSymbol,
				Type:         string(system.Type),
				X:            int(system.X),
				Y:            int(system.Y),
				Waypoints:    convertSystemWaypoints(system.Waypoints),
				Factions:     convertSystemFactions(system.Factions),
			}
			allSystems = append(allSystems, convertedSystem)
		}

		// Check if we have more pages
		if len(resp.Data) < int(limit) || int32(len(allSystems)) >= resp.Meta.Total {
			break
		}
		page++
	}

	return allSystems, nil
}

// GetSystem returns a specific system
func (c *Client) GetSystem(systemSymbol string) (*System, error) {
	resp, _, err := c.apiClient.SystemsAPI.GetSystem(c.ctx, systemSymbol).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get system: %w", err)
	}

	return &System{
		Symbol:       resp.Data.Symbol,
		SectorSymbol: resp.Data.SectorSymbol,
		Type:         string(resp.Data.Type),
		X:            int(resp.Data.X),
		Y:            int(resp.Data.Y),
		Waypoints:    convertSystemWaypoints(resp.Data.Waypoints),
		Factions:     convertSystemFactions(resp.Data.Factions),
	}, nil
}

// GetAllFactions returns all factions
func (c *Client) GetAllFactions() ([]Faction, error) {
	var allFactions []Faction
	page := int32(1)
	limit := int32(20)

	for {
		resp, _, err := c.apiClient.FactionsAPI.GetFactions(c.ctx).Page(page).Limit(limit).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to get factions: %w", err)
		}

		for _, faction := range resp.Data {
			var headquarters string
			if faction.Headquarters != nil {
				headquarters = *faction.Headquarters
			}

			convertedFaction := Faction{
				Symbol:       string(faction.Symbol),
				Name:         faction.Name,
				Description:  faction.Description,
				Headquarters: headquarters,
				Traits:       convertFactionTraits(faction.Traits),
				IsRecruiting: faction.IsRecruiting,
			}
			allFactions = append(allFactions, convertedFaction)
		}

		// Check if we have more pages
		if len(resp.Data) < int(limit) || int32(len(allFactions)) >= resp.Meta.Total {
			break
		}
		page++
	}

	return allFactions, nil
}

// GetFaction returns a specific faction
func (c *Client) GetFaction(factionSymbol string) (*Faction, error) {
	resp, _, err := c.apiClient.FactionsAPI.GetFaction(c.ctx, factionSymbol).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get faction: %w", err)
	}

	var headquarters string
	if resp.Data.Headquarters != nil {
		headquarters = *resp.Data.Headquarters
	}

	return &Faction{
		Symbol:       string(resp.Data.Symbol),
		Name:         resp.Data.Name,
		Description:  resp.Data.Description,
		Headquarters: headquarters,
		Traits:       convertFactionTraits(resp.Data.Traits),
		IsRecruiting: resp.Data.IsRecruiting,
	}, nil
}

// SellCargo sells cargo from a ship
func (c *Client) SellCargo(shipSymbol, symbol string, units int) (*SellCargoResponse, error) {
	req := spacetraders.SellCargoRequest{
		Symbol: spacetraders.TradeSymbol(symbol),
		Units:  int32(units),
	}

	resp, _, err := c.apiClient.FleetAPI.SellCargo(c.ctx, shipSymbol).SellCargoRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to sell cargo: %w", err)
	}

	return &SellCargoResponse{
		Data: SellCargoData{
			Agent:       convertAgentFromGenerated(resp.Data.Agent),
			Cargo:       convertCargo(resp.Data.Cargo),
			Transaction: convertMarketTransactionFromGenerated(resp.Data.Transaction),
		},
	}, nil
}

// BuyCargo buys cargo for a ship
func (c *Client) BuyCargo(shipSymbol, symbol string, units int) (*BuyCargoResponse, error) {
	req := spacetraders.PurchaseCargoRequest{
		Symbol: spacetraders.TradeSymbol(symbol),
		Units:  int32(units),
	}

	resp, _, err := c.apiClient.FleetAPI.PurchaseCargo(c.ctx, shipSymbol).PurchaseCargoRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to buy cargo: %w", err)
	}

	return &BuyCargoResponse{
		Data: BuyCargoData{
			Agent:       convertAgentFromGenerated(resp.Data.Agent),
			Cargo:       convertCargo(resp.Data.Cargo),
			Transaction: convertMarketTransactionFromGenerated(resp.Data.Transaction),
		},
	}, nil
}

// DeliverContract delivers goods to a contract
func (c *Client) DeliverContract(contractID, shipSymbol, tradeSymbol string, units int) (*DeliverContractResponse, error) {
	req := spacetraders.DeliverContractRequest{
		ShipSymbol:  shipSymbol,
		TradeSymbol: tradeSymbol,
		Units:       int32(units),
	}

	resp, _, err := c.apiClient.ContractsAPI.DeliverContract(c.ctx, contractID).DeliverContractRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to deliver contract goods: %w", err)
	}

	var expiration, deadlineToAccept string
	expiration = resp.Data.Contract.Expiration.Format("2006-01-02T15:04:05.000Z")
	if resp.Data.Contract.DeadlineToAccept != nil {
		deadlineToAccept = resp.Data.Contract.DeadlineToAccept.Format("2006-01-02T15:04:05.000Z")
	}

	return &DeliverContractResponse{
		Data: DeliverContractData{
			Contract: Contract{
				ID:               resp.Data.Contract.Id,
				FactionSymbol:    resp.Data.Contract.FactionSymbol,
				Type:             resp.Data.Contract.Type,
				Terms:            convertContractTerms(resp.Data.Contract.Terms),
				Accepted:         resp.Data.Contract.Accepted,
				Fulfilled:        resp.Data.Contract.Fulfilled,
				Expiration:       expiration,
				DeadlineToAccept: deadlineToAccept,
			},
			Cargo: convertCargo(resp.Data.Cargo),
		},
	}, nil
}

// FulfillContract fulfills a contract
func (c *Client) FulfillContract(contractID string) (*FulfillContractResponse, error) {
	resp, _, err := c.apiClient.ContractsAPI.FulfillContract(c.ctx, contractID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to fulfill contract: %w", err)
	}

	var expiration, deadlineToAccept string
	expiration = resp.Data.Contract.Expiration.Format("2006-01-02T15:04:05.000Z")
	if resp.Data.Contract.DeadlineToAccept != nil {
		deadlineToAccept = resp.Data.Contract.DeadlineToAccept.Format("2006-01-02T15:04:05.000Z")
	}

	return &FulfillContractResponse{
		Data: FulfillContractData{
			Agent: convertAgentFromGenerated(resp.Data.Agent),
			Contract: Contract{
				ID:               resp.Data.Contract.Id,
				FactionSymbol:    resp.Data.Contract.FactionSymbol,
				Type:             resp.Data.Contract.Type,
				Terms:            convertContractTerms(resp.Data.Contract.Terms),
				Accepted:         resp.Data.Contract.Accepted,
				Fulfilled:        resp.Data.Contract.Fulfilled,
				Expiration:       expiration,
				DeadlineToAccept: deadlineToAccept,
			},
		},
	}, nil
}

// ExtractResources extracts resources from a waypoint
func (c *Client) ExtractResources(shipSymbol string, survey *Survey) (*ExtractResponse, error) {
	var req spacetraders.ExtractResourcesRequest
	if survey != nil {
		req.Survey = &spacetraders.Survey{
			Signature:  survey.Signature,
			Symbol:     survey.Symbol,
			Deposits:   convertSurveyDepositsToGenerated(survey.Deposits),
			Expiration: parseTime(survey.Expiration),
			Size:       survey.Size,
		}
	}

	resp, _, err := c.apiClient.FleetAPI.ExtractResources(c.ctx, shipSymbol).ExtractResourcesRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to extract resources: %w", err)
	}

	return &ExtractResponse{
		Data: ExtractData{
			Cooldown:   convertCooldown(resp.Data.Cooldown),
			Extraction: convertExtraction(resp.Data.Extraction),
			Cargo:      convertCargo(resp.Data.Cargo),
			Events:     convertEvents(resp.Data.Events),
		},
	}, nil
}

// JettisonCargo jettisons cargo from a ship
func (c *Client) JettisonCargo(shipSymbol, symbol string, units int) (*JettisonResponse, error) {
	req := spacetraders.JettisonRequest{
		Symbol: spacetraders.TradeSymbol(symbol),
		Units:  int32(units),
	}

	resp, _, err := c.apiClient.FleetAPI.Jettison(c.ctx, shipSymbol).JettisonRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to jettison cargo: %w", err)
	}

	return &JettisonResponse{
		Data: JettisonData{
			Cargo: convertCargo(resp.Data.Cargo),
		},
	}, nil
}

// RefuelShip refuels a ship
func (c *Client) RefuelShip(shipSymbol string, units *int, fromCargo bool) (*RefuelResponse, error) {
	req := spacetraders.RefuelShipRequest{
		FromCargo: &fromCargo,
	}
	if units != nil {
		units32 := int32(*units)
		req.Units = &units32
	}

	resp, _, err := c.apiClient.FleetAPI.RefuelShip(c.ctx, shipSymbol).RefuelShipRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to refuel ship: %w", err)
	}

	return &RefuelResponse{
		Data: RefuelData{
			Agent:       convertAgentFromGenerated(resp.Data.Agent),
			Fuel:        convertFuel(resp.Data.Fuel),
			Transaction: convertMarketTransactionFromGenerated(resp.Data.Transaction),
		},
	}, nil
}

// ScanSystems scans for systems around the ship
func (c *Client) ScanSystems(shipSymbol string) (*ScanSystemsResponse, error) {
	resp, _, err := c.apiClient.FleetAPI.CreateShipSystemScan(c.ctx, shipSymbol).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to scan systems: %w", err)
	}

	return &ScanSystemsResponse{
		Data: ScanSystemsData{
			Cooldown: convertCooldown(resp.Data.Cooldown),
			Systems:  convertScannedSystems(resp.Data.Systems),
		},
	}, nil
}

// ScanWaypoints scans for waypoints around the ship
func (c *Client) ScanWaypoints(shipSymbol string) (*ScanWaypointsResponse, error) {
	resp, _, err := c.apiClient.FleetAPI.CreateShipWaypointScan(c.ctx, shipSymbol).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to scan waypoints: %w", err)
	}

	return &ScanWaypointsResponse{
		Data: ScanWaypointsData{
			Cooldown:  convertCooldown(resp.Data.Cooldown),
			Waypoints: convertScannedWaypoints(resp.Data.Waypoints),
		},
	}, nil
}

// ScanShips scans for ships around the ship
func (c *Client) ScanShips(shipSymbol string) (*ScanShipsResponse, error) {
	resp, _, err := c.apiClient.FleetAPI.CreateShipShipScan(c.ctx, shipSymbol).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to scan ships: %w", err)
	}

	return &ScanShipsResponse{
		Data: ScanShipsData{
			Cooldown: convertCooldown(resp.Data.Cooldown),
			Ships:    convertScannedShips(resp.Data.Ships),
		},
	}, nil
}

// RepairShip repairs a ship
func (c *Client) RepairShip(shipSymbol string) (*RepairShipResponse, error) {
	resp, _, err := c.apiClient.FleetAPI.RepairShip(c.ctx, shipSymbol).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to repair ship: %w", err)
	}

	return &RepairShipResponse{
		Data: RepairShipData{
			Agent:       convertAgentFromGenerated(resp.Data.Agent),
			Ship:        convertShipFromGenerated(resp.Data.Ship),
			Transaction: convertRepairTransactionFromGenerated(resp.Data.Transaction),
		},
	}, nil
}

// JumpShip jumps a ship to a system
func (c *Client) JumpShip(shipSymbol, systemSymbol string) (*JumpResponse, error) {
	req := spacetraders.JumpShipRequest{
		WaypointSymbol: systemSymbol,
	}

	resp, _, err := c.apiClient.FleetAPI.JumpShip(c.ctx, shipSymbol).JumpShipRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to jump ship: %w", err)
	}

	return &JumpResponse{
		Data: JumpData{
			Cooldown: convertCooldown(resp.Data.Cooldown),
			Nav:      convertNavigation(resp.Data.Nav),
			Event:    convertEventFromTransaction(resp.Data.Transaction),
		},
	}, nil
}

// WarpShip warps a ship to a waypoint
func (c *Client) WarpShip(shipSymbol, waypointSymbol string) (*WarpResponse, error) {
	req := spacetraders.NavigateShipRequest{
		WaypointSymbol: waypointSymbol,
	}

	resp, _, err := c.apiClient.FleetAPI.WarpShip(c.ctx, shipSymbol).NavigateShipRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to warp ship: %w", err)
	}

	return &WarpResponse{
		Data: WarpData{
			Fuel: convertFuel(resp.Data.Fuel),
			Nav:  convertNavigation(resp.Data.Nav),
		},
	}, nil
}

// PatchShipNav updates ship navigation configuration
func (c *Client) PatchShipNav(shipSymbol, flightMode string) (*PatchNavResponse, error) {
	req := spacetraders.PatchShipNavRequest{
		FlightMode: (*spacetraders.ShipNavFlightMode)(&flightMode),
	}

	resp, _, err := c.apiClient.FleetAPI.PatchShipNav(c.ctx, shipSymbol).PatchShipNavRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to patch ship nav: %w", err)
	}

	return &PatchNavResponse{
		Data: convertNavigation(resp.Data.Nav),
	}, nil
}
