package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"spacetraders-mcp/pkg/client"
	"spacetraders-mcp/pkg/logging"

	"github.com/mark3labs/mcp-go/mcp"
)

// FactionsResource handles the factions resource
type FactionsResource struct {
	client *client.Client
	logger *logging.Logger
}

// NewFactionsResource creates a new factions resource handler
func NewFactionsResource(client *client.Client, logger *logging.Logger) *FactionsResource {
	return &FactionsResource{
		client: client,
		logger: logger,
	}
}

// Resource returns the MCP resource definition
func (r *FactionsResource) Resource() mcp.Resource {
	return mcp.Resource{
		URI:         "spacetraders://factions/*",
		Name:        "Factions Data",
		Description: "Factions information - use 'spacetraders://factions' for all factions or 'spacetraders://factions/{factionSymbol}' for specific faction details",
		MIMEType:    "application/json",
	}
}

// Handler returns the resource handler function
func (r *FactionsResource) Handler() func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// Set up context logger
		ctxLogger := r.logger.WithContext(ctx, "factions-resource")
		ctxLogger.Debug("Processing factions resource request")

		// Check if this is a request for a specific faction or all factions
		if request.Params.URI == "spacetraders://factions" {
			return r.handleFactionsList(ctx, request, ctxLogger)
		} else if strings.HasPrefix(request.Params.URI, "spacetraders://factions/") {
			return r.handleSpecificFaction(ctx, request, ctxLogger)
		} else {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Invalid factions resource URI",
				},
			}, nil
		}
	}
}

// handleFactionsList handles requests for the full factions list
func (r *FactionsResource) handleFactionsList(ctx context.Context, request mcp.ReadResourceRequest, ctxLogger *logging.ContextLogger) ([]mcp.ResourceContents, error) {
	ctxLogger.Debug("Fetching factions list from API")

	// Get factions from the API
	start := time.Now()
	factions, err := r.client.GetAllFactions()
	duration := time.Since(start)

	if err != nil {
		ctxLogger.Error("Failed to fetch factions: %v", err)
		ctxLogger.APICall("/factions", 0, duration.String())
		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "text/plain",
				Text:     "Error fetching factions: " + err.Error(),
			},
		}, nil
	}

	ctxLogger.APICall("/factions", 200, duration.String())
	ctxLogger.Info("Successfully retrieved %d factions", len(factions))

	// Format the response
	result := map[string]interface{}{
		"factions": r.formatFactionsList(factions),
		"meta": map[string]interface{}{
			"total":     len(factions),
			"retrieved": time.Now().UTC().Format(time.RFC3339),
		},
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		ctxLogger.Error("Failed to marshal factions data to JSON: %v", err)
		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "text/plain",
				Text:     "Error formatting factions information",
			},
		}, nil
	}

	ctxLogger.ResourceRead(request.Params.URI, true)
	ctxLogger.Debug("Factions resource response size: %d bytes", len(jsonData))

	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}

// handleSpecificFaction handles requests for a specific faction
func (r *FactionsResource) handleSpecificFaction(ctx context.Context, request mcp.ReadResourceRequest, ctxLogger *logging.ContextLogger) ([]mcp.ResourceContents, error) {
	// Extract faction symbol from URI
	factionSymbol, err := r.parseFactionSymbol(request.Params.URI)
	if err != nil {
		ctxLogger.Error("Failed to parse faction symbol from URI: %v", err)
		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "text/plain",
				Text:     "Invalid faction URI format",
			},
		}, nil
	}

	ctxLogger.Debug("Fetching faction details for: %s", factionSymbol)

	// Get faction details from the API
	start := time.Now()
	faction, err := r.client.GetFaction(factionSymbol)
	duration := time.Since(start)

	if err != nil {
		ctxLogger.Error("Failed to fetch faction %s: %v", factionSymbol, err)
		ctxLogger.APICall(fmt.Sprintf("/factions/%s", factionSymbol), 0, duration.String())
		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "text/plain",
				Text:     fmt.Sprintf("Error fetching faction %s: %s", factionSymbol, err.Error()),
			},
		}, nil
	}

	ctxLogger.APICall(fmt.Sprintf("/factions/%s", factionSymbol), 200, duration.String())
	ctxLogger.Info("Successfully retrieved faction details for: %s", factionSymbol)

	// Format the response
	result := map[string]interface{}{
		"faction": r.formatFactionDetails(faction),
		"meta": map[string]interface{}{
			"retrieved": time.Now().UTC().Format(time.RFC3339),
		},
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		ctxLogger.Error("Failed to marshal faction data to JSON: %v", err)
		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "text/plain",
				Text:     "Error formatting faction information",
			},
		}, nil
	}

	ctxLogger.ResourceRead(request.Params.URI, true)
	ctxLogger.Debug("Faction resource response size: %d bytes", len(jsonData))

	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}

// parseFactionSymbol extracts the faction symbol from the URI
func (r *FactionsResource) parseFactionSymbol(uri string) (string, error) {
	// Expected format: spacetraders://factions/{factionSymbol}
	if !strings.HasPrefix(uri, "spacetraders://factions/") {
		return "", fmt.Errorf("invalid URI format")
	}

	// Remove the protocol prefix
	factionSymbol := strings.TrimPrefix(uri, "spacetraders://factions/")

	// Faction symbol shouldn't be empty
	if factionSymbol == "" {
		return "", fmt.Errorf("faction symbol cannot be empty")
	}

	return factionSymbol, nil
}

// formatFactionsList formats a list of factions for the response
func (r *FactionsResource) formatFactionsList(factions []client.Faction) []map[string]interface{} {
	result := make([]map[string]interface{}, len(factions))

	for i, faction := range factions {
		traits := make([]map[string]interface{}, len(faction.Traits))
		for j, trait := range faction.Traits {
			traits[j] = map[string]interface{}{
				"symbol":      trait.Symbol,
				"name":        trait.Name,
				"description": trait.Description,
			}
		}

		result[i] = map[string]interface{}{
			"symbol":       faction.Symbol,
			"name":         faction.Name,
			"description":  faction.Description,
			"headquarters": faction.Headquarters,
			"traits":       traits,
			"isRecruiting": faction.IsRecruiting,
		}
	}

	return result
}

// formatFactionDetails formats detailed faction information
func (r *FactionsResource) formatFactionDetails(faction *client.Faction) map[string]interface{} {
	// Format traits
	traits := make([]map[string]interface{}, len(faction.Traits))
	for i, trait := range faction.Traits {
		traits[i] = map[string]interface{}{
			"symbol":      trait.Symbol,
			"name":        trait.Name,
			"description": trait.Description,
		}
	}

	return map[string]interface{}{
		"symbol":       faction.Symbol,
		"name":         faction.Name,
		"description":  faction.Description,
		"headquarters": faction.Headquarters,
		"traits":       traits,
		"isRecruiting": faction.IsRecruiting,
	}
}
