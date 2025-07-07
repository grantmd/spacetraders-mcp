package client

import (
	"fmt"
	"time"

	spacetraders "github.com/grantmd/spacetraders-mcp/spacetraders"
)

// convertRegistration converts generated ShipRegistration to wrapper Registration
func convertRegistration(gen spacetraders.ShipRegistration) Registration {
	return Registration{
		Name:          gen.Name,
		FactionSymbol: gen.FactionSymbol,
		Role:          string(gen.Role),
	}
}

// convertNavigation converts generated ShipNav to wrapper Navigation
func convertNavigation(gen spacetraders.ShipNav) Navigation {
	return Navigation{
		SystemSymbol:   gen.SystemSymbol,
		WaypointSymbol: gen.WaypointSymbol,
		Route:          convertRoute(gen.Route),
		Status:         string(gen.Status),
		FlightMode:     string(gen.FlightMode),
	}
}

// convertRoute converts generated ShipNavRoute to wrapper Route
func convertRoute(gen spacetraders.ShipNavRoute) Route {
	return Route{
		Destination:   convertWaypointFromNavRoute(gen.Destination),
		Origin:        convertWaypointFromNavRoute(gen.Origin),
		DepartureTime: gen.DepartureTime.Format("2006-01-02T15:04:05.000Z"),
		Arrival:       gen.Arrival.Format("2006-01-02T15:04:05.000Z"),
	}
}

// convertWaypointFromNavRoute converts nav route waypoint to wrapper Waypoint
func convertWaypointFromNavRoute(gen spacetraders.ShipNavRouteWaypoint) Waypoint {
	return Waypoint{
		Symbol: gen.Symbol,
		Type:   string(gen.Type),
		X:      int(gen.X),
		Y:      int(gen.Y),
	}
}

// convertCrew converts generated ShipCrew to wrapper Crew
func convertCrew(gen spacetraders.ShipCrew) Crew {
	return Crew{
		Current:  int(gen.Current),
		Required: int(gen.Required),
		Capacity: int(gen.Capacity),
		Rotation: string(gen.Rotation),
		Morale:   int(gen.Morale),
		Wages:    int(gen.Wages),
	}
}

// convertFrame converts generated ShipFrame to wrapper Frame with FIXED types
func convertFrame(gen spacetraders.ShipFrame) Frame {
	return Frame{
		Symbol:         gen.Symbol,
		Name:           gen.Name,
		Description:    gen.Description,
		Condition:      gen.Condition, // ✅ FIXED: now float64
		Integrity:      gen.Integrity, // ✅ FIXED: now float64
		ModuleSlots:    int(gen.ModuleSlots),
		MountingPoints: int(gen.MountingPoints),
		FuelCapacity:   int(gen.FuelCapacity),
		Requirements:   convertShipRequirements(gen.Requirements),
		Quality:        gen.Quality, // ✅ NEW: was missing
	}
}

// convertReactor converts generated ShipReactor to wrapper Reactor with FIXED types
func convertReactor(gen spacetraders.ShipReactor) Reactor {
	return Reactor{
		Symbol:       gen.Symbol,
		Name:         gen.Name,
		Description:  gen.Description,
		Condition:    gen.Condition, // ✅ FIXED: now float64
		Integrity:    gen.Integrity, // ✅ FIXED: now float64
		PowerOutput:  int(gen.PowerOutput),
		Requirements: convertShipRequirements(gen.Requirements),
		Quality:      gen.Quality, // ✅ NEW: was missing
	}
}

// convertEngine converts generated ShipEngine to wrapper Engine with FIXED types
func convertEngine(gen spacetraders.ShipEngine) Engine {
	return Engine{
		Symbol:       gen.Symbol,
		Name:         gen.Name,
		Description:  gen.Description,
		Condition:    gen.Condition, // ✅ FIXED: now float64
		Integrity:    gen.Integrity, // ✅ FIXED: now float64
		Speed:        int(gen.Speed),
		Requirements: convertShipRequirements(gen.Requirements),
		Quality:      gen.Quality, // ✅ NEW: was missing
	}
}

// convertShipRequirements converts generated ShipRequirements to wrapper ShipRequirements
func convertShipRequirements(gen spacetraders.ShipRequirements) ShipRequirements {
	power := 0
	crew := 0
	slots := 0

	if gen.Power != nil {
		power = int(*gen.Power)
	}
	if gen.Crew != nil {
		crew = int(*gen.Crew)
	}
	if gen.Slots != nil {
		slots = int(*gen.Slots)
	}

	return ShipRequirements{
		Power: power,
		Crew:  crew,
		Slots: slots,
	}
}

// convertCooldown converts generated Cooldown to wrapper Cooldown
func convertCooldown(gen spacetraders.Cooldown) Cooldown {
	expiration := ""
	if gen.Expiration != nil {
		expiration = gen.Expiration.Format("2006-01-02T15:04:05.000Z")
	}

	return Cooldown{
		ShipSymbol:       gen.ShipSymbol,
		TotalSeconds:     int(gen.TotalSeconds),
		RemainingSeconds: int(gen.RemainingSeconds),
		Expiration:       expiration,
	}
}

// convertModules converts generated ShipModule slice to wrapper Module slice
func convertModules(gen []spacetraders.ShipModule) []Module {
	modules := make([]Module, len(gen))
	for i, m := range gen {
		capacity := 0
		if m.Capacity != nil {
			capacity = int(*m.Capacity)
		}

		rangeVal := 0
		if m.Range != nil {
			rangeVal = int(*m.Range)
		}

		modules[i] = Module{
			Symbol:       m.Symbol,
			Capacity:     capacity,
			Range:        rangeVal,
			Name:         m.Name,
			Description:  m.Description,
			Requirements: convertShipRequirements(m.Requirements),
		}
	}
	return modules
}

// convertMounts converts generated ShipMount slice to wrapper Mount slice
func convertMounts(gen []spacetraders.ShipMount) []Mount {
	mounts := make([]Mount, len(gen))
	for i, m := range gen {
		strength := 0
		if m.Strength != nil {
			strength = int(*m.Strength)
		}

		var deposits []string
		if m.Deposits != nil {
			deposits = make([]string, len(m.Deposits))
			for j, d := range m.Deposits {
				deposits[j] = string(d)
			}
		}

		description := ""
		if m.Description != nil {
			description = *m.Description
		}

		mounts[i] = Mount{
			Symbol:       m.Symbol,
			Name:         m.Name,
			Description:  description,
			Strength:     strength,
			Deposits:     deposits,
			Requirements: convertShipRequirements(m.Requirements),
		}
	}
	return mounts
}

// convertCargo converts generated ShipCargo to wrapper Cargo
func convertCargo(gen spacetraders.ShipCargo) Cargo {
	inventory := make([]CargoItem, len(gen.Inventory))
	for i, item := range gen.Inventory {
		inventory[i] = CargoItem{
			Symbol:      string(item.Symbol),
			Name:        item.Name,
			Description: item.Description,
			Units:       int(item.Units),
		}
	}

	return Cargo{
		Capacity:  int(gen.Capacity),
		Units:     int(gen.Units),
		Inventory: inventory,
	}
}

// convertFuel converts generated ShipFuel to wrapper Fuel
func convertFuel(gen spacetraders.ShipFuel) Fuel {
	fuel := Fuel{
		Current:  int(gen.Current),
		Capacity: int(gen.Capacity),
	}

	if gen.Consumed != nil {
		fuel.Consumed = &FuelConsumed{
			Amount:    int(gen.Consumed.Amount),
			Timestamp: gen.Consumed.Timestamp.Format("2006-01-02T15:04:05.000Z"),
		}
	}

	return fuel
}

// convertContractTerms converts generated ContractTerms to wrapper ContractTerms
func convertContractTerms(gen spacetraders.ContractTerms) ContractTerms {
	terms := ContractTerms{
		Deadline: gen.Deadline.Format("2006-01-02T15:04:05.000Z"),
		Payment: ContractPayment{
			OnAccepted:  int(gen.Payment.OnAccepted),
			OnFulfilled: int(gen.Payment.OnFulfilled),
		},
	}

	if gen.Deliver != nil {
		terms.Deliver = make([]ContractDeliverGood, len(gen.Deliver))
		for i, d := range gen.Deliver {
			terms.Deliver[i] = ContractDeliverGood{
				TradeSymbol:       d.TradeSymbol,
				DestinationSymbol: d.DestinationSymbol,
				UnitsRequired:     int(d.UnitsRequired),
				UnitsFulfilled:    int(d.UnitsFulfilled),
			}
		}
	}

	return terms
}

// convertOrbitals converts generated WaypointOrbital slice to wrapper WaypointOrbital slice
func convertOrbitals(gen []spacetraders.WaypointOrbital) []WaypointOrbital {
	orbitals := make([]WaypointOrbital, len(gen))
	for i, o := range gen {
		orbitals[i] = WaypointOrbital{
			Symbol: o.Symbol,
		}
	}
	return orbitals
}

// convertWaypointTraits converts generated WaypointTrait slice to wrapper WaypointTrait slice
func convertWaypointTraits(gen []spacetraders.WaypointTrait) []WaypointTrait {
	traits := make([]WaypointTrait, len(gen))
	for i, t := range gen {
		traits[i] = WaypointTrait{
			Symbol:      string(t.Symbol),
			Name:        t.Name,
			Description: t.Description,
		}
	}
	return traits
}

// convertWaypointModifiers converts generated WaypointModifier slice to wrapper WaypointModifier slice
func convertWaypointModifiers(gen []spacetraders.WaypointModifier) []WaypointModifier {
	modifiers := make([]WaypointModifier, len(gen))
	for i, m := range gen {
		modifiers[i] = WaypointModifier{
			Symbol:      string(m.Symbol),
			Name:        m.Name,
			Description: m.Description,
		}
	}
	return modifiers
}

// convertChart converts generated Chart to wrapper WaypointChart
func convertChart(gen *spacetraders.Chart) *WaypointChart {
	if gen == nil {
		return nil
	}

	var waypointSymbol, submittedBy, submittedOn string
	if gen.WaypointSymbol != nil {
		waypointSymbol = *gen.WaypointSymbol
	}
	if gen.SubmittedBy != nil {
		submittedBy = *gen.SubmittedBy
	}
	if gen.SubmittedOn != nil {
		submittedOn = gen.SubmittedOn.Format("2006-01-02T15:04:05.000Z")
	}

	return &WaypointChart{
		WaypointSymbol: waypointSymbol,
		SubmittedBy:    submittedBy,
		SubmittedOn:    submittedOn,
	}
}

// convertWaypointFaction converts generated WaypointFaction to wrapper WaypointFaction
func convertWaypointFaction(gen *spacetraders.WaypointFaction) *WaypointFaction {
	if gen == nil {
		return nil
	}
	return &WaypointFaction{
		Symbol: string(gen.Symbol),
	}
}

// convertSystemWaypoints converts generated SystemWaypoint slice to wrapper SystemWaypoint slice
func convertSystemWaypoints(gen []spacetraders.SystemWaypoint) []SystemWaypoint {
	waypoints := make([]SystemWaypoint, len(gen))
	for i, w := range gen {
		waypoints[i] = SystemWaypoint{
			Symbol:   w.Symbol,
			Type:     string(w.Type),
			X:        int(w.X),
			Y:        int(w.Y),
			Orbitals: convertOrbitals(w.Orbitals),
		}
	}
	return waypoints
}

// convertSystemFactions converts generated SystemFaction slice to wrapper SystemFaction slice
func convertSystemFactions(gen []spacetraders.SystemFaction) []SystemFaction {
	factions := make([]SystemFaction, len(gen))
	for i, f := range gen {
		factions[i] = SystemFaction{
			Symbol: string(f.Symbol),
		}
	}
	return factions
}

// convertShipyardShipTypes converts generated shipyard ship types
func convertShipyardShipTypes(gen []spacetraders.ShipyardShipTypesInner) []ShipyardShipType {
	types := make([]ShipyardShipType, len(gen))
	for i, t := range gen {
		types[i] = ShipyardShipType{
			Type: string(t.Type),
		}
	}
	return types
}

// convertShipyardTransactions converts generated shipyard transactions
func convertShipyardTransactions(gen []spacetraders.ShipyardTransaction) []ShipyardTransaction {
	transactions := make([]ShipyardTransaction, len(gen))
	for i, t := range gen {
		transactions[i] = ShipyardTransaction{
			WaypointSymbol: t.WaypointSymbol,
			ShipSymbol:     t.ShipSymbol,
			ShipType:       t.ShipType,
			Price:          int(t.Price),
			AgentSymbol:    t.AgentSymbol,
			Timestamp:      t.Timestamp.Format("2006-01-02T15:04:05.000Z"),
		}
	}
	return transactions
}

// convertShipyardShips converts generated shipyard ships
func convertShipyardShips(gen []spacetraders.ShipyardShip) []ShipyardShip {
	var ships []ShipyardShip
	for _, s := range gen {
		var activity string
		if s.Activity != nil {
			activity = string(*s.Activity)
		}
		ships = append(ships, ShipyardShip{
			Type:          string(s.Type),
			Name:          s.Name,
			Description:   s.Description,
			Supply:        string(s.Supply),
			Activity:      activity,
			PurchasePrice: int(s.PurchasePrice),
			Frame:         convertShipyardShipFrame(s.Frame),
			Reactor:       convertShipyardShipReactor(s.Reactor),
			Engine:        convertShipyardShipEngine(s.Engine),
			Modules:       convertShipyardShipModules(s.Modules),
			Mounts:        convertShipyardShipMounts(s.Mounts),
			Crew:          convertShipyardShipCrew(s.Crew),
		})
	}
	return ships
}

// convertShipyardShipFrame converts generated shipyard ship frame
func convertShipyardShipFrame(gen spacetraders.ShipFrame) ShipyardShipFrame {
	return ShipyardShipFrame{
		Symbol:         gen.Symbol,
		Name:           gen.Name,
		Description:    gen.Description,
		ModuleSlots:    int(gen.ModuleSlots),
		MountingPoints: int(gen.MountingPoints),
		FuelCapacity:   int(gen.FuelCapacity),
		Condition:      gen.Condition,
		Integrity:      gen.Integrity,
		Requirements:   convertShipRequirements(gen.Requirements),
	}
}

// convertShipyardShipReactor converts generated shipyard ship reactor
func convertShipyardShipReactor(gen spacetraders.ShipReactor) ShipyardShipReactor {
	return ShipyardShipReactor{
		Symbol:       gen.Symbol,
		Name:         gen.Name,
		Description:  gen.Description,
		Condition:    gen.Condition,
		Integrity:    gen.Integrity,
		PowerOutput:  int(gen.PowerOutput),
		Requirements: convertShipRequirements(gen.Requirements),
	}
}

// convertShipyardShipEngine converts generated shipyard ship engine
func convertShipyardShipEngine(gen spacetraders.ShipEngine) ShipyardShipEngine {
	return ShipyardShipEngine{
		Symbol:       gen.Symbol,
		Name:         gen.Name,
		Description:  gen.Description,
		Condition:    gen.Condition,
		Integrity:    gen.Integrity,
		Speed:        int(gen.Speed),
		Requirements: convertShipRequirements(gen.Requirements),
	}
}

// convertShipyardShipModules converts generated shipyard ship modules
func convertShipyardShipModules(gen []spacetraders.ShipModule) []ShipyardShipModule {
	modules := make([]ShipyardShipModule, len(gen))
	for i, m := range gen {
		capacity := 0
		if m.Capacity != nil {
			capacity = int(*m.Capacity)
		}

		rangeVal := 0
		if m.Range != nil {
			rangeVal = int(*m.Range)
		}

		modules[i] = ShipyardShipModule{
			Symbol:       m.Symbol,
			Name:         m.Name,
			Description:  m.Description,
			Capacity:     capacity,
			Range:        rangeVal,
			Requirements: convertShipRequirements(m.Requirements),
		}
	}
	return modules
}

// convertShipyardShipMounts converts generated shipyard ship mounts
func convertShipyardShipMounts(gen []spacetraders.ShipMount) []ShipyardShipMount {
	mounts := make([]ShipyardShipMount, len(gen))
	for i, m := range gen {
		strength := 0
		if m.Strength != nil {
			strength = int(*m.Strength)
		}

		var deposits []string
		if m.Deposits != nil {
			deposits = make([]string, len(m.Deposits))
			for j, d := range m.Deposits {
				deposits[j] = string(d)
			}
		}

		description := ""
		if m.Description != nil {
			description = *m.Description
		}

		mounts[i] = ShipyardShipMount{
			Symbol:       m.Symbol,
			Name:         m.Name,
			Description:  description,
			Strength:     strength,
			Deposits:     deposits,
			Requirements: convertShipRequirements(m.Requirements),
		}
	}
	return mounts
}

// convertShipyardShipCrew converts generated shipyard ship crew
func convertShipyardShipCrew(gen spacetraders.ShipyardShipCrew) ShipyardShipCrew {
	return ShipyardShipCrew{
		Required: int(gen.Required),
		Capacity: int(gen.Capacity),
	}
}

// convertTradeGoods converts generated trade goods
func convertTradeGoods(gen []spacetraders.TradeGood) []TradeGood {
	goods := make([]TradeGood, len(gen))
	for i, g := range gen {
		goods[i] = TradeGood{
			Symbol:      string(g.Symbol),
			Name:        g.Name,
			Description: g.Description,
		}
	}
	return goods
}

// convertMarketTransactions converts generated market transactions
func convertMarketTransactions(gen []spacetraders.MarketTransaction) []MarketTransaction {
	transactions := make([]MarketTransaction, len(gen))
	for i, t := range gen {
		transactions[i] = MarketTransaction{
			WaypointSymbol: t.WaypointSymbol,
			ShipSymbol:     t.ShipSymbol,
			TradeSymbol:    t.TradeSymbol,
			Type:           t.Type,
			Units:          int(t.Units),
			PricePerUnit:   int(t.PricePerUnit),
			TotalPrice:     int(t.TotalPrice),
			Timestamp:      t.Timestamp.Format("2006-01-02T15:04:05.000Z"),
		}
	}
	return transactions
}

// convertMarketTradeGoods converts generated market trade goods
func convertMarketTradeGoods(gen []spacetraders.MarketTradeGood) []MarketTradeGood {
	goods := make([]MarketTradeGood, len(gen))
	for i, g := range gen {
		activity := ""
		if g.Activity != nil {
			activity = string(*g.Activity)
		}

		goods[i] = MarketTradeGood{
			Symbol:        string(g.Symbol),
			Type:          g.Type,
			TradeVolume:   int(g.TradeVolume),
			Supply:        string(g.Supply),
			Activity:      activity,
			PurchasePrice: int(g.PurchasePrice),
			SellPrice:     int(g.SellPrice),
		}
	}
	return goods
}

// convertFactionTraits converts generated faction traits
func convertFactionTraits(gen []spacetraders.FactionTrait) []FactionTrait {
	traits := make([]FactionTrait, len(gen))
	for i, t := range gen {
		traits[i] = FactionTrait{
			Symbol:      string(t.Symbol),
			Name:        t.Name,
			Description: t.Description,
		}
	}
	return traits
}

// Helper conversion functions for responses

// convertShipFromGenerated converts a generated Ship to wrapper Ship
func convertShipFromGenerated(gen spacetraders.Ship) Ship {
	return Ship{
		Symbol:       gen.Symbol,
		Registration: convertRegistration(gen.Registration),
		Nav:          convertNavigation(gen.Nav),
		Crew:         convertCrew(gen.Crew),
		Frame:        convertFrame(gen.Frame),
		Reactor:      convertReactor(gen.Reactor),
		Engine:       convertEngine(gen.Engine),
		Cooldown:     convertCooldown(gen.Cooldown),
		Modules:      convertModules(gen.Modules),
		Mounts:       convertMounts(gen.Mounts),
		Cargo:        convertCargo(gen.Cargo),
		Fuel:         convertFuel(gen.Fuel),
	}
}

// convertTransactionFromGenerated converts a generated transaction
func convertTransactionFromGenerated(gen spacetraders.ShipyardTransaction) Transaction {
	return Transaction{
		WaypointSymbol: gen.WaypointSymbol,
		ShipSymbol:     gen.ShipSymbol,
		ShipType:       string(gen.ShipType),
		Price:          int(gen.Price),
		AgentSymbol:    gen.AgentSymbol,
		Timestamp:      gen.Timestamp.Format("2006-01-02T15:04:05.000Z"),
	}
}

// convertAgentFromGenerated converts a generated Agent to wrapper Agent
func convertAgentFromGenerated(gen spacetraders.Agent) Agent {
	return Agent{
		AccountID:       gen.AccountId,
		Symbol:          gen.Symbol,
		Headquarters:    gen.Headquarters,
		Credits:         gen.Credits,
		StartingFaction: gen.StartingFaction,
		ShipCount:       int(gen.ShipCount),
	}
}

// Helper functions for new methods

// parseTime parses a time string in RFC3339 format
func parseTime(timeStr string) time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05.000Z", timeStr)
	return t
}

// convertSurveyDepositsToGenerated converts survey deposits to generated format
func convertSurveyDepositsToGenerated(deposits []SurveyDeposit) []spacetraders.SurveyDeposit {
	result := make([]spacetraders.SurveyDeposit, len(deposits))
	for i, d := range deposits {
		result[i] = spacetraders.SurveyDeposit{
			Symbol: d.Symbol,
		}
	}
	return result
}

// convertExtraction converts generated Extraction to wrapper Extraction
func convertExtraction(gen spacetraders.Extraction) Extraction {
	return Extraction{
		ShipSymbol: gen.ShipSymbol,
		Yield: ExtractionYield{
			Symbol: string(gen.Yield.Symbol),
			Units:  int(gen.Yield.Units),
		},
	}
}

// convertEvents converts generated events to wrapper events
func convertEvents(gen []spacetraders.ShipConditionEvent) []Event {
	events := make([]Event, len(gen))
	for i, e := range gen {
		events[i] = Event{
			Symbol:      string(e.Symbol),
			Component:   string(e.Component),
			Name:        e.Name,
			Description: e.Description,
		}
	}
	return events
}

// convertScannedSystems converts generated scanned systems
func convertScannedSystems(gen []spacetraders.ScannedSystem) []ScannedSystem {
	systems := make([]ScannedSystem, len(gen))
	for i, s := range gen {
		systems[i] = ScannedSystem{
			Symbol:       s.Symbol,
			SectorSymbol: s.SectorSymbol,
			Type:         string(s.Type),
			X:            int(s.X),
			Y:            int(s.Y),
			Distance:     int(s.Distance),
		}
	}
	return systems
}

// convertScannedWaypoints converts generated scanned waypoints
func convertScannedWaypoints(gen []spacetraders.ScannedWaypoint) []ScannedWaypoint {
	waypoints := make([]ScannedWaypoint, len(gen))
	for i, w := range gen {
		waypoints[i] = ScannedWaypoint{
			Symbol:       w.Symbol,
			Type:         string(w.Type),
			SystemSymbol: w.SystemSymbol,
			X:            int(w.X),
			Y:            int(w.Y),
			Orbitals:     convertOrbitals(w.Orbitals),
			Faction:      convertWaypointFaction(w.Faction),
			Traits:       convertWaypointTraits(w.Traits),
			Chart:        convertChart(w.Chart),
		}
	}
	return waypoints
}

// convertScannedShips converts generated scanned ships
func convertScannedShips(gen []spacetraders.ScannedShip) []ScannedShip {
	ships := make([]ScannedShip, len(gen))
	for i, s := range gen {
		var frame *ScannedShipFrame
		if s.Frame != nil {
			frame = &ScannedShipFrame{Symbol: s.Frame.Symbol}
		}

		var reactor *ScannedShipReactor
		if s.Reactor != nil {
			reactor = &ScannedShipReactor{Symbol: s.Reactor.Symbol}
		}

		engine := &ScannedShipEngine{Symbol: s.Engine.Symbol}

		mounts := make([]ScannedShipMount, len(s.Mounts))
		for j, m := range s.Mounts {
			mounts[j] = ScannedShipMount{Symbol: m.Symbol}
		}

		ships[i] = ScannedShip{
			Symbol:       s.Symbol,
			Registration: convertRegistration(s.Registration),
			Nav:          convertNavigation(s.Nav),
			Frame:        frame,
			Reactor:      reactor,
			Engine:       engine,
			Mounts:       mounts,
		}
	}
	return ships
}

// convertRepairTransactionFromGenerated converts repair transaction
func convertRepairTransactionFromGenerated(gen spacetraders.RepairTransaction) RepairTransaction {
	return RepairTransaction{
		WaypointSymbol: gen.WaypointSymbol,
		ShipSymbol:     gen.ShipSymbol,
		TotalPrice:     int(gen.TotalPrice),
		Timestamp:      gen.Timestamp.Format("2006-01-02T15:04:05.000Z"),
	}
}

// convertEventFromTransaction converts a transaction to an event
func convertEventFromTransaction(gen spacetraders.MarketTransaction) Event {
	return Event{
		Symbol:      string(gen.TradeSymbol),
		Component:   "transaction",
		Name:        string(gen.Type),
		Description: fmt.Sprintf("Transaction at %s", gen.WaypointSymbol),
	}
}

// convertMarketTransactionFromGenerated converts a generated market transaction
func convertMarketTransactionFromGenerated(gen spacetraders.MarketTransaction) MarketTransaction {
	return MarketTransaction{
		WaypointSymbol: gen.WaypointSymbol,
		ShipSymbol:     gen.ShipSymbol,
		TradeSymbol:    string(gen.TradeSymbol),
		Type:           string(gen.Type),
		Units:          int(gen.Units),
		PricePerUnit:   int(gen.PricePerUnit),
		TotalPrice:     int(gen.TotalPrice),
		Timestamp:      gen.Timestamp.Format("2006-01-02T15:04:05.000Z"),
	}
}

// convertEvent converts events (handling slice to single for compatibility)
func convertEvent(events []spacetraders.ShipConditionEvent) Event {
	if len(events) == 0 {
		return Event{}
	}

	// Take the first event for compatibility
	e := events[0]
	return Event{
		Symbol:      string(e.Symbol),
		Component:   string(e.Component),
		Name:        e.Name,
		Description: e.Description,
	}
}
