package info

import (
	"context"
	"fmt"
	"strings"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/spacetraders"

	"github.com/mark3labs/mcp-go/mcp"
)

// FleetAnalysisTool analyzes fleet capabilities vs contract requirements
type FleetAnalysisTool struct {
	client *spacetraders.Client
	logger *logging.Logger
}

// NewFleetAnalysisTool creates a new fleet analysis tool
func NewFleetAnalysisTool(client *spacetraders.Client, logger *logging.Logger) *FleetAnalysisTool {
	return &FleetAnalysisTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *FleetAnalysisTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "analyze_fleet_capabilities",
		Description: "Analyze your current fleet's capabilities against contract requirements and suggest needed ships",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"include_recommendations": map[string]interface{}{
					"type":        "boolean",
					"description": "Include ship purchase recommendations",
					"default":     true,
				},
			},
		},
	}
}

// Handler returns the tool handler function
func (t *FleetAnalysisTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Set up context logger
		ctxLogger := t.logger.WithContext(ctx, "fleet-analysis-tool")
		ctxLogger.Debug("Analyzing fleet capabilities")

		// Parse arguments
		includeRecommendations := true
		if request.Params.Arguments != nil {
			if argsMap, ok := request.Params.Arguments.(map[string]interface{}); ok {
				if rec, exists := argsMap["include_recommendations"]; exists {
					if recBool, ok := rec.(bool); ok {
						includeRecommendations = recBool
					}
				}
			}
		}

		// Get current fleet
		ships, err := t.client.GetShips()
		if err != nil {
			ctxLogger.Error("Failed to fetch ships: %v", err)
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("❌ Error fetching ships: %s", err.Error())),
				},
				IsError: true,
			}, nil
		}

		// Get current contracts
		contracts, err := t.client.GetContracts()
		if err != nil {
			ctxLogger.Error("Failed to fetch contracts: %v", err)
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("❌ Error fetching contracts: %s", err.Error())),
				},
				IsError: true,
			}, nil
		}

		// Analyze fleet composition
		fleetAnalysis := t.analyzeFleet(ships)

		// Analyze contract requirements
		contractRequirements := t.analyzeContractRequirements(contracts)

		// Build response
		var response strings.Builder
		response.WriteString("🚢 **Fleet Capability Analysis**\n\n")

		// Current fleet summary
		response.WriteString("**Current Fleet:**\n")
		response.WriteString(fmt.Sprintf("• Total Ships: %d\n", len(ships)))
		response.WriteString(fmt.Sprintf("• Total Cargo Capacity: %d units\n", fleetAnalysis.TotalCargo))
		response.WriteString(fmt.Sprintf("• Mining Capable Ships: %d\n", fleetAnalysis.MiningShips))
		response.WriteString(fmt.Sprintf("• Hauling Capable Ships: %d\n", fleetAnalysis.HaulingShips))
		response.WriteString(fmt.Sprintf("• Combat Capable Ships: %d\n", fleetAnalysis.CombatShips))

		if len(fleetAnalysis.ShipsByType) > 0 {
			response.WriteString("\n**Fleet Composition:**\n")
			for shipType, count := range fleetAnalysis.ShipsByType {
				capability := t.getShipCapabilityDescription(shipType)
				response.WriteString(fmt.Sprintf("• %s: %d ship(s) - %s\n", shipType, count, capability))
			}
		}

		// Contract requirements analysis
		if len(contractRequirements.ActiveContracts) > 0 {
			response.WriteString("\n**Contract Requirements Analysis:**\n")

			for _, req := range contractRequirements.ActiveContracts {
				response.WriteString(fmt.Sprintf("\n📋 **Contract %s:**\n", req.ContractID))
				response.WriteString(fmt.Sprintf("• Status: %s\n", req.Status))

				if len(req.RequiredMaterials) > 0 {
					response.WriteString("• Required Materials:\n")
					for _, material := range req.RequiredMaterials {
						response.WriteString(fmt.Sprintf("  - %s: %d units\n", material.Symbol, material.UnitsNeeded))
						if material.RequiresMining {
							response.WriteString("    *Requires mining*\n")
						}
					}
				}

				response.WriteString(fmt.Sprintf("• Required Cargo Space: %d units\n", req.TotalCargoNeeded))
			}

			// Gap analysis
			response.WriteString(t.performGapAnalysis(fleetAnalysis, contractRequirements))
		}

		// Recommendations
		if includeRecommendations {
			response.WriteString(t.generateRecommendations(fleetAnalysis, contractRequirements))
		}

		ctxLogger.ToolCall("analyze_fleet_capabilities", true)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(response.String()),
			},
		}, nil
	}
}

// FleetAnalysis holds fleet capability data
type FleetAnalysis struct {
	TotalCargo   int
	MiningShips  int
	HaulingShips int
	CombatShips  int
	ShipsByType  map[string]int
}

// ContractRequirements holds contract requirement data
type ContractRequirements struct {
	ActiveContracts  []ContractRequirement
	TotalCargoNeeded int
	RequiresMining   bool
}

// ContractRequirement holds individual contract requirements
type ContractRequirement struct {
	ContractID        string
	Status            string
	RequiredMaterials []MaterialRequirement
	TotalCargoNeeded  int
}

// MaterialRequirement holds material-specific requirements
type MaterialRequirement struct {
	Symbol         string
	UnitsNeeded    int
	RequiresMining bool
}

// analyzeFleet analyzes current fleet capabilities
func (t *FleetAnalysisTool) analyzeFleet(ships []spacetraders.Ship) FleetAnalysis {
	analysis := FleetAnalysis{
		ShipsByType: make(map[string]int),
	}

	for _, ship := range ships {
		// Count by type
		shipType := ship.Registration.Role
		analysis.ShipsByType[shipType]++

		// Add cargo capacity
		analysis.TotalCargo += ship.Cargo.Capacity

		// Categorize by capability
		switch shipType {
		case "EXCAVATOR", "MINING":
			analysis.MiningShips++
		case "HAULER", "TRANSPORT":
			analysis.HaulingShips++
		case "INTERCEPTOR", "FIGHTER":
			analysis.CombatShips++
		default:
			// Check if it has mining mounts
			for _, mount := range ship.Mounts {
				if strings.Contains(mount.Symbol, "MINING") || strings.Contains(mount.Symbol, "EXCAVATOR") {
					analysis.MiningShips++
					break
				}
			}
		}
	}

	return analysis
}

// analyzeContractRequirements analyzes what contracts need
func (t *FleetAnalysisTool) analyzeContractRequirements(contracts []spacetraders.Contract) ContractRequirements {
	requirements := ContractRequirements{
		ActiveContracts: make([]ContractRequirement, 0),
	}

	for _, contract := range contracts {
		if contract.Fulfilled {
			continue
		}

		contractReq := ContractRequirement{
			ContractID:        contract.ID,
			RequiredMaterials: make([]MaterialRequirement, 0),
		}

		if contract.Accepted {
			contractReq.Status = "In Progress"
		} else {
			contractReq.Status = "Available"
		}

		for _, delivery := range contract.Terms.Deliver {
			unitsNeeded := delivery.UnitsRequired - delivery.UnitsFulfilled
			if unitsNeeded > 0 {
				materialReq := MaterialRequirement{
					Symbol:         delivery.TradeSymbol,
					UnitsNeeded:    unitsNeeded,
					RequiresMining: t.isMiningMaterial(delivery.TradeSymbol),
				}
				contractReq.RequiredMaterials = append(contractReq.RequiredMaterials, materialReq)
				contractReq.TotalCargoNeeded += unitsNeeded

				if materialReq.RequiresMining {
					requirements.RequiresMining = true
				}
			}
		}

		if len(contractReq.RequiredMaterials) > 0 {
			requirements.ActiveContracts = append(requirements.ActiveContracts, contractReq)
			requirements.TotalCargoNeeded += contractReq.TotalCargoNeeded
		}
	}

	return requirements
}

// performGapAnalysis identifies fleet capability gaps
func (t *FleetAnalysisTool) performGapAnalysis(fleet FleetAnalysis, requirements ContractRequirements) string {
	var analysis strings.Builder
	analysis.WriteString("\n🔍 **Gap Analysis:**\n")

	// Check mining capability
	if requirements.RequiresMining && fleet.MiningShips == 0 {
		analysis.WriteString("❌ **CRITICAL GAP:** No mining ships available but contracts require mining\n")
		analysis.WriteString("   *You need a SHIP_MINING_DRONE or similar mining vessel*\n")
	} else if requirements.RequiresMining && fleet.MiningShips > 0 {
		analysis.WriteString("✅ Mining capability available\n")
	}

	// Check cargo capacity
	if requirements.TotalCargoNeeded > fleet.TotalCargo {
		shortage := requirements.TotalCargoNeeded - fleet.TotalCargo
		analysis.WriteString(fmt.Sprintf("⚠️ **CARGO SHORTAGE:** Need %d more cargo capacity\n", shortage))
		analysis.WriteString("   *Consider buying hauler ships or ships with larger cargo holds*\n")
	} else if requirements.TotalCargoNeeded > 0 {
		analysis.WriteString("✅ Sufficient cargo capacity available\n")
	}

	// Check if we have any contracts but no ships
	if len(requirements.ActiveContracts) > 0 && len(fleet.ShipsByType) == 0 {
		analysis.WriteString("❌ **CRITICAL GAP:** You have contracts but no ships!\n")
	}

	return analysis.String()
}

// generateRecommendations provides ship purchase recommendations
func (t *FleetAnalysisTool) generateRecommendations(fleet FleetAnalysis, requirements ContractRequirements) string {
	var recommendations strings.Builder
	recommendations.WriteString("\n💡 **Recommendations:**\n")

	// Mining recommendations
	if requirements.RequiresMining && fleet.MiningShips == 0 {
		recommendations.WriteString("🔥 **URGENT:** Purchase a SHIP_MINING_DRONE\n")
		recommendations.WriteString("   • Required for mining contracts\n")
		recommendations.WriteString("   • Use `purchase_ship` with ship_type=SHIP_MINING_DRONE\n")
		recommendations.WriteString("   • Must be at a shipyard that sells mining drones\n\n")
	}

	// Cargo recommendations
	if requirements.TotalCargoNeeded > fleet.TotalCargo {
		recommendations.WriteString("📦 **CARGO:** Consider purchasing hauler ships\n")
		recommendations.WriteString("   • SHIP_LIGHT_HAULER for medium cargo needs\n")
		recommendations.WriteString("   • SHIP_ORE_HOUND for mining + hauling combo\n\n")
	}

	// General fleet recommendations
	if len(fleet.ShipsByType) == 1 {
		recommendations.WriteString("🚢 **FLEET DIVERSITY:** Consider diversifying your fleet\n")
		recommendations.WriteString("   • Different ship types have different capabilities\n")
		recommendations.WriteString("   • Specialized ships are more efficient at their tasks\n\n")
	}

	// Next steps
	recommendations.WriteString("**Next Steps:**\n")
	recommendations.WriteString("• Use `get_status_summary` to see your current credits\n")
	recommendations.WriteString("• Use waypoint resources to find shipyards in your systems\n")
	recommendations.WriteString("• Check shipyard inventories before traveling\n")

	return recommendations.String()
}

// getShipCapabilityDescription returns a description of what a ship type can do
func (t *FleetAnalysisTool) getShipCapabilityDescription(shipType string) string {
	descriptions := map[string]string{
		"COMMAND":     "Multi-purpose command vessel",
		"EXCAVATOR":   "Mining and excavation specialist",
		"MINING":      "Resource extraction specialist",
		"HAULER":      "Large cargo capacity transport",
		"TRANSPORT":   "Medium cargo transport",
		"INTERCEPTOR": "Fast combat vessel",
		"FIGHTER":     "Combat specialist",
		"DRONE":       "Automated utility vessel",
		"PROBE":       "Exploration and reconnaissance",
		"SURVEYOR":    "System mapping and analysis",
	}

	if desc, exists := descriptions[shipType]; exists {
		return desc
	}
	return "General purpose vessel"
}

// isMiningMaterial checks if a material requires mining (reused from contract_info.go)
func (t *FleetAnalysisTool) isMiningMaterial(tradeSymbol string) bool {
	miningMaterials := map[string]bool{
		"IRON_ORE":          true,
		"COPPER_ORE":        true,
		"ALUMINUM_ORE":      true,
		"SILVER_ORE":        true,
		"GOLD_ORE":          true,
		"PLATINUM_ORE":      true,
		"URANITE_ORE":       true,
		"MERITIUM_ORE":      true,
		"HYDROCARBON":       true,
		"QUARTZ_SAND":       true,
		"SILICON_CRYSTALS":  true,
		"AMMONIA_ICE":       true,
		"LIQUID_HYDROGEN":   true,
		"LIQUID_NITROGEN":   true,
		"ICE_WATER":         true,
		"EXOTIC_MATTER":     true,
		"GRAVITON_EMITTERS": true,
		"IRON":              true,
		"COPPER":            true,
		"ALUMINUM":          true,
		"SILVER":            true,
		"GOLD":              true,
		"PLATINUM":          true,
		"URANITE":           true,
		"MERITIUM":          true,
	}
	return miningMaterials[tradeSymbol]
}
