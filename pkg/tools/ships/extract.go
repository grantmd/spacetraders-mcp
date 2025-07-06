package ships

import (
	"context"
	"fmt"
	"strings"
	"time"

	"spacetraders-mcp/pkg/client"
	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/tools/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

// ExtractResourcesTool handles extracting resources from asteroids and mining sites
type ExtractResourcesTool struct {
	client *client.Client
	logger *logging.Logger
}

// NewExtractResourcesTool creates a new extract resources tool
func NewExtractResourcesTool(client *client.Client, logger *logging.Logger) *ExtractResourcesTool {
	return &ExtractResourcesTool{
		client: client,
		logger: logger,
	}
}

// Tool returns the MCP tool definition
func (t *ExtractResourcesTool) Tool() mcp.Tool {
	return mcp.Tool{
		Name:        "extract_resources",
		Description: "Extract resources from the current waypoint (asteroid fields, mining sites). Ship must be in orbit and have mining capabilities.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"ship_symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol of the ship to perform extraction (e.g., 'SHIP_1234')",
				},
				"survey": map[string]interface{}{
					"type":        "object",
					"description": "Optional: Survey data to improve extraction efficiency and target specific resources",
					"properties": map[string]interface{}{
						"signature": map[string]interface{}{
							"type":        "string",
							"description": "Survey signature identifier",
						},
						"symbol": map[string]interface{}{
							"type":        "string",
							"description": "Waypoint symbol where survey was conducted",
						},
						"deposits": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"symbol": map[string]interface{}{
										"type":        "string",
										"description": "Resource symbol that can be extracted",
									},
								},
								"required": []string{"symbol"},
							},
							"description": "List of available resource deposits",
						},
						"expiration": map[string]interface{}{
							"type":        "string",
							"description": "When the survey expires (ISO 8601 format)",
						},
						"size": map[string]interface{}{
							"type":        "string",
							"description": "Survey size (SMALL, MODERATE, LARGE)",
						},
					},
					"required": []string{"signature", "symbol", "deposits", "expiration", "size"},
				},
			},
			Required: []string{"ship_symbol"},
		},
	}
}

// Handler returns the tool handler function
func (t *ExtractResourcesTool) Handler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Set up context logger
		ctxLogger := t.logger.WithContext(ctx, "extract-resources-tool")
		ctxLogger.Debug("Processing resource extraction request")

		// Parse arguments
		shipSymbol := ""
		var survey *client.Survey

		if request.Params.Arguments == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("❌ Missing required argument: ship_symbol"),
				},
				IsError: true,
			}, nil
		}

		if argsMap, ok := request.Params.Arguments.(map[string]interface{}); ok {
			if ss, exists := argsMap["ship_symbol"]; exists {
				if ssStr, ok := ss.(string); ok {
					shipSymbol = strings.TrimSpace(ssStr)
				}
			}

			// Parse survey if provided
			if surveyData, exists := argsMap["survey"]; exists {
				if surveyMap, ok := surveyData.(map[string]interface{}); ok {
					survey = &client.Survey{}

					if sig, exists := surveyMap["signature"]; exists {
						if sigStr, ok := sig.(string); ok {
							survey.Signature = sigStr
						}
					}

					if sym, exists := surveyMap["symbol"]; exists {
						if symStr, ok := sym.(string); ok {
							survey.Symbol = symStr
						}
					}

					if exp, exists := surveyMap["expiration"]; exists {
						if expStr, ok := exp.(string); ok {
							survey.Expiration = expStr
						}
					}

					if size, exists := surveyMap["size"]; exists {
						if sizeStr, ok := size.(string); ok {
							survey.Size = sizeStr
						}
					}

					if deposits, exists := surveyMap["deposits"]; exists {
						if depositsArray, ok := deposits.([]interface{}); ok {
							survey.Deposits = make([]client.SurveyDeposit, 0, len(depositsArray))
							for _, dep := range depositsArray {
								if depMap, ok := dep.(map[string]interface{}); ok {
									if symbol, exists := depMap["symbol"]; exists {
										if symbolStr, ok := symbol.(string); ok {
											survey.Deposits = append(survey.Deposits, client.SurveyDeposit{
												Symbol: symbolStr,
											})
										}
									}
								}
							}
						}
					}
				}
			}
		}

		if shipSymbol == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("❌ ship_symbol is required and must be a non-empty string"),
				},
				IsError: true,
			}, nil
		}

		ctxLogger.Info("Attempting to extract resources with ship %s", shipSymbol)
		if survey != nil {
			ctxLogger.Info("Using survey data for %s (expires: %s)", survey.Symbol, survey.Expiration)
		}

		// Extract resources
		start := time.Now()
		resp, err := t.client.ExtractResources(shipSymbol, survey)
		duration := time.Since(start)

		if err != nil {
			ctxLogger.Error("Failed to extract resources: %v", err)
			ctxLogger.APICall(fmt.Sprintf("/my/ships/%s/extract", shipSymbol), 0, duration.String())
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("❌ Failed to extract resources: %s", err.Error())),
				},
				IsError: true,
			}, nil
		}

		cooldown := resp.Data.Cooldown
		extraction := resp.Data.Extraction
		cargo := resp.Data.Cargo
		events := resp.Data.Events

		ctxLogger.APICall(fmt.Sprintf("/my/ships/%s/extract", shipSymbol), 201, duration.String())
		ctxLogger.Info("Successfully extracted %d units of %s", extraction.Yield.Units, extraction.Yield.Symbol)

		// Format the response
		result := map[string]interface{}{
			"success":     true,
			"message":     fmt.Sprintf("Successfully extracted resources with ship %s", shipSymbol),
			"ship_symbol": shipSymbol,
			"extraction": map[string]interface{}{
				"resource_symbol": extraction.Yield.Symbol,
				"units_extracted": extraction.Yield.Units,
			},
			"cargo": map[string]interface{}{
				"capacity": cargo.Capacity,
				"units":    cargo.Units,
				"inventory": func() []map[string]interface{} {
					inventory := make([]map[string]interface{}, len(cargo.Inventory))
					for i, item := range cargo.Inventory {
						inventory[i] = map[string]interface{}{
							"symbol":      item.Symbol,
							"name":        item.Name,
							"description": item.Description,
							"units":       item.Units,
						}
					}
					return inventory
				}(),
			},
			"cooldown": map[string]interface{}{
				"ship_symbol":       cooldown.ShipSymbol,
				"total_seconds":     cooldown.TotalSeconds,
				"remaining_seconds": cooldown.RemainingSeconds,
				"expiration":        cooldown.Expiration,
			},
		}

		// Add events if any occurred
		if len(events) > 0 {
			eventList := make([]map[string]interface{}, len(events))
			for i, event := range events {
				eventList[i] = map[string]interface{}{
					"symbol":      event.Symbol,
					"component":   event.Component,
					"name":        event.Name,
					"description": event.Description,
				}
			}
			result["events"] = eventList
		}

		jsonData := utils.FormatJSON(result)

		// Calculate cargo utilization
		cargoPercent := float64(cargo.Units) / float64(cargo.Capacity) * 100

		// Create formatted text summary
		textSummary := "⛏️ **Resource Extraction Successful!**\n\n"
		textSummary += fmt.Sprintf("**Ship:** %s\n", shipSymbol)
		textSummary += fmt.Sprintf("**Extracted:** %d units of %s\n", extraction.Yield.Units, extraction.Yield.Symbol)
		textSummary += fmt.Sprintf("**Cargo Status:** %d/%d units (%.1f%% full)", cargo.Units, cargo.Capacity, cargoPercent)

		if cargoPercent >= 90 {
			textSummary += " ⚠️ *Nearly full!*"
		} else if cargoPercent >= 75 {
			textSummary += " ⚠️ *Getting full*"
		}
		textSummary += "\n\n"

		if cooldown.RemainingSeconds > 0 {
			textSummary += fmt.Sprintf("**Cooldown:** %d seconds remaining (until %s)\n", cooldown.RemainingSeconds, cooldown.Expiration)
			if cooldown.RemainingSeconds > 60 {
				minutes := cooldown.RemainingSeconds / 60
				seconds := cooldown.RemainingSeconds % 60
				textSummary += fmt.Sprintf("*That's %d minutes and %d seconds*\n", minutes, seconds)
			}
		} else {
			textSummary += "**Status:** Ready for next extraction!\n"
		}

		if survey != nil {
			textSummary += "\n**Survey Used:** Enhanced extraction with survey data\n"
			textSummary += fmt.Sprintf("- Survey ID: %s\n", survey.Signature)
			textSummary += fmt.Sprintf("- Survey Size: %s\n", survey.Size)
		}

		// Show current cargo inventory
		if len(cargo.Inventory) > 0 {
			textSummary += "\n**Current Cargo Inventory:**\n"
			for _, item := range cargo.Inventory {
				textSummary += fmt.Sprintf("- %s: %d units\n", item.Name, item.Units)
			}
		}

		// Add events information
		if len(events) > 0 {
			textSummary += "\n**Mining Events:**\n"
			for _, event := range events {
				textSummary += fmt.Sprintf("- **%s:** %s\n", event.Name, event.Description)
			}
		}

		// Add helpful tips
		textSummary += "\n💡 **Next Steps:**\n"
		if cargoPercent >= 90 {
			textSummary += "• ⚠️ **Cargo nearly full!** Consider using `jettison_cargo` to make space or dock to sell\n"
		}
		if cooldown.RemainingSeconds > 0 {
			textSummary += "• ⏱️ Wait for cooldown to complete before next extraction\n"
		} else {
			textSummary += "• 🔄 Ready to extract again immediately!\n"
		}
		textSummary += "• 📊 Use `get_status_summary` to check all your ships\n"
		textSummary += "• 🏪 Dock at a marketplace to sell extracted resources\n"

		ctxLogger.ToolCall("extract_resources", true)
		ctxLogger.Debug("Extract resources response size: %d bytes", len(jsonData))

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(textSummary),
				mcp.NewTextContent(fmt.Sprintf("```json\n%s\n```", jsonData)),
			},
		}, nil
	}
}
