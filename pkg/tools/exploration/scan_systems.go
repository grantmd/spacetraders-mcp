package exploration

import (
	"context"
	"fmt"
	"strings"

	"spacetraders-mcp/pkg/client"
	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/tools/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

// ScanSystemsTool allows scanning for systems around a ship
type ScanSystemsTool struct {
	client *client.Client
	logger *logging.Logger
}

// NewScanSystemsTool creates a new scan systems tool
func NewScanSystemsTool(client *client.Client, logger *logging.Logger) *ScanSystemsTool {
	return &ScanSystemsTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *ScanSystemsTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "scan_systems",
		Description: "Scan for systems around a ship using its sensors. Requires appropriate scanning equipment.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"ship_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the ship to scan with (e.g., 'MYSHIP-1')",
				},
			},
			Required: []string{"ship_symbol"},
		},
	}
}

// Handler returns the tool handler function
func (t *ScanSystemsTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contextLogger := t.logger.WithContext(ctx, "scan-systems-tool")

		// Extract parameters
		var shipSymbol string
		if request.Params.Arguments != nil {
			if argsMap, ok := request.Params.Arguments.(map[string]interface{}); ok {
				if val, exists := argsMap["ship_symbol"]; exists {
					if s, ok := val.(string); ok {
						shipSymbol = strings.ToUpper(s)
					}
				}
			}
		}

		if shipSymbol == "" {
			contextLogger.Error("Missing ship_symbol parameter")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("Error: ship_symbol parameter is required"),
				},
				IsError: true,
			}, nil
		}

		contextLogger.Info(fmt.Sprintf("Scanning for systems using ship %s", shipSymbol))

		// Perform the scan
		scanData, err := t.client.ScanSystems(shipSymbol)
		if err != nil {
			contextLogger.Error(fmt.Sprintf("Failed to scan systems with ship %s: %v", shipSymbol, err))
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("Failed to scan systems with ship %s: %v", shipSymbol, err)),
				},
				IsError: true,
			}, nil
		}

		contextLogger.ToolCall("scan_systems", true)
		contextLogger.Info(fmt.Sprintf("Successfully scanned %d systems with ship %s", len(scanData.Data.Systems), shipSymbol))

		// Create structured response
		result := map[string]interface{}{
			"ship_symbol":   shipSymbol,
			"systems_found": len(scanData.Data.Systems),
			"cooldown":      scanData.Data.Cooldown,
			"systems":       []map[string]interface{}{},
		}

		// Build systems data
		for _, system := range scanData.Data.Systems {
			systemData := map[string]interface{}{
				"symbol":        system.Symbol,
				"sector_symbol": system.SectorSymbol,
				"type":          system.Type,
				"x":             system.X,
				"y":             system.Y,
				"distance":      system.Distance,
			}

			result["systems"] = append(result["systems"].([]map[string]interface{}), systemData)
		}

		// Create text summary
		textSummary := fmt.Sprintf("## 🔍 System Scan Results for %s\n\n", shipSymbol)

		if len(scanData.Data.Systems) == 0 {
			textSummary += "❌ **No systems detected** in scanning range.\n\n"
			textSummary += "This could mean:\n"
			textSummary += "- Your ship doesn't have appropriate scanning equipment\n"
			textSummary += "- No systems are within scanning range\n"
			textSummary += "- The ship is on cooldown from previous scans\n\n"
		} else {
			textSummary += fmt.Sprintf("✅ **Detected %d system(s)** in scanning range:\n\n", len(scanData.Data.Systems))

			for i, system := range scanData.Data.Systems {
				textSummary += fmt.Sprintf("### %d. %s (%s)\n", i+1, system.Symbol, system.Type)
				textSummary += fmt.Sprintf("**Sector:** %s\n", system.SectorSymbol)
				textSummary += fmt.Sprintf("**Location:** (%d, %d)\n", system.X, system.Y)
				textSummary += fmt.Sprintf("**Distance:** %d units\n", system.Distance)

				textSummary += "\n"
			}
		}

		// Add cooldown information
		if scanData.Data.Cooldown.RemainingSeconds > 0 {
			textSummary += "## ⏳ Cooldown Information\n\n"
			textSummary += fmt.Sprintf("**Total Cooldown:** %d seconds\n", scanData.Data.Cooldown.TotalSeconds)
			textSummary += fmt.Sprintf("**Cooldown:** %d seconds remaining\n", scanData.Data.Cooldown.RemainingSeconds)
			if scanData.Data.Cooldown.Expiration != "" {
				textSummary += fmt.Sprintf("**Ready At:** %s\n", scanData.Data.Cooldown.Expiration)
			}
			textSummary += "\n"
		}

		// Add next steps
		textSummary += "## 🚀 Next Steps\n\n"
		if len(scanData.Data.Systems) > 0 {
			textSummary += "To explore discovered systems:\n"
			for _, system := range scanData.Data.Systems {
				textSummary += fmt.Sprintf("- Get detailed info about %s: `system_overview` tool\n", system.Symbol)
				textSummary += fmt.Sprintf("- Jump to %s: `jump_ship` tool (requires jump drive)\n", system.Symbol)
			}
		} else {
			textSummary += "- Try scanning again after cooldown expires\n"
			textSummary += "- Move to a different location and scan again\n"
			textSummary += "- Ensure your ship has appropriate scanning equipment\n"
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%s\n```", utils.FormatJSON(result))),
			},
		}, nil
	}
}
