package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"spacetraders-mcp/pkg/client"
	"spacetraders-mcp/pkg/logging"

	"github.com/mark3labs/mcp-go/mcp"
)

// ShipCooldownResource handles ship cooldown information resources
type ShipCooldownResource struct {
	client *client.Client
	logger *logging.Logger
}

// NewShipCooldownResource creates a new ship cooldown resource handler
func NewShipCooldownResource(client *client.Client, logger *logging.Logger) *ShipCooldownResource {
	return &ShipCooldownResource{
		client: client,
		logger: logger,
	}
}

// Resource returns the MCP resource definition
func (r *ShipCooldownResource) Resource() mcp.Resource {
	return mcp.Resource{
		URI:         "spacetraders://ships/{shipSymbol}/cooldown",
		Name:        "Ship Cooldown Status",
		Description: "Real-time cooldown status for a specific ship, including remaining time and operational availability",
		MIMEType:    "application/json",
	}
}

// Handler returns the resource handler function
func (r *ShipCooldownResource) Handler() func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// Extract ship symbol from URI
		shipSymbol := r.extractShipSymbol(request.Params.URI)
		if shipSymbol == "" {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Invalid ship cooldown resource URI. Expected format: spacetraders://ships/{shipSymbol}/cooldown",
				},
			}, nil
		}

		// Set up context logger
		ctxLogger := r.logger.WithContext(ctx, "ship-cooldown-resource")
		ctxLogger.Debug("Fetching cooldown status for ship %s", shipSymbol)

		// Get cooldown information from the API
		start := time.Now()
		cooldown, err := r.client.GetShipCooldown(shipSymbol)
		duration := time.Since(start)

		if err != nil {
			ctxLogger.Error("Failed to fetch cooldown for ship %s: %v", shipSymbol, err)
			ctxLogger.APICall(fmt.Sprintf("/my/ships/%s/cooldown", shipSymbol), 0, duration.String())
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     fmt.Sprintf("Error fetching cooldown for ship %s: %s", shipSymbol, err.Error()),
				},
			}, nil
		}

		ctxLogger.APICall(fmt.Sprintf("/my/ships/%s/cooldown", shipSymbol), 200, duration.String())

		// Create cooldown analysis
		result := r.createCooldownAnalysis(shipSymbol, cooldown)

		// Convert to JSON for response
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			ctxLogger.Error("Failed to marshal cooldown data to JSON: %v", err)
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Error formatting cooldown information",
				},
			}, nil
		}

		if cooldown != nil {
			ctxLogger.Info("Ship %s has cooldown: %d seconds remaining", shipSymbol, cooldown.RemainingSeconds)
		} else {
			ctxLogger.Info("Ship %s has no active cooldown", shipSymbol)
		}

		ctxLogger.ResourceRead(request.Params.URI, true)
		ctxLogger.Debug("Ship cooldown resource response size: %d bytes", len(jsonData))

		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		}, nil
	}
}

// extractShipSymbol extracts the ship symbol from the URI
func (r *ShipCooldownResource) extractShipSymbol(uri string) string {
	// Match pattern: spacetraders://ships/{shipSymbol}/cooldown
	re := regexp.MustCompile(`^spacetraders://ships/([A-Za-z0-9_-]+)/cooldown$`)
	matches := re.FindStringSubmatch(uri)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}

// createCooldownAnalysis creates detailed cooldown analysis
func (r *ShipCooldownResource) createCooldownAnalysis(shipSymbol string, cooldown *client.Cooldown) map[string]interface{} {
	now := time.Now()

	if cooldown == nil || cooldown.RemainingSeconds <= 0 {
		return map[string]interface{}{
			"ship_symbol": shipSymbol,
			"cooldown": map[string]interface{}{
				"active":            false,
				"remaining_seconds": 0,
				"total_seconds":     0,
				"expiration":        nil,
			},
			"status": map[string]interface{}{
				"operational": true,
				"message":     "Ship is ready for all actions",
				"priority":    "ready",
				"icon":        "🟢",
			},
			"timing": map[string]interface{}{
				"ready_at":     now.Format(time.RFC3339),
				"time_display": "Ready now",
			},
			"actions": map[string]interface{}{
				"can_extract":  true,
				"can_scan":     true,
				"can_jump":     true,
				"can_navigate": true,
				"can_survey":   true,
			},
			"meta": map[string]interface{}{
				"last_checked": now.Format(time.RFC3339),
				"ship_symbol":  shipSymbol,
			},
		}
	}

	// Calculate timing information
	expirationTime, _ := time.Parse(time.RFC3339, cooldown.Expiration)
	readyAt := now.Add(time.Duration(cooldown.RemainingSeconds) * time.Second)

	// Create time display
	timeDisplay := r.formatTimeRemaining(cooldown.RemainingSeconds)

	// Determine status and priority
	status := r.determineCooldownStatus(cooldown.RemainingSeconds)

	// Determine what actions are blocked
	blockedActions := r.analyzeBlockedActions(cooldown.RemainingSeconds)

	return map[string]interface{}{
		"ship_symbol": shipSymbol,
		"cooldown": map[string]interface{}{
			"active":            true,
			"remaining_seconds": cooldown.RemainingSeconds,
			"total_seconds":     cooldown.TotalSeconds,
			"expiration":        cooldown.Expiration,
		},
		"status": status,
		"timing": map[string]interface{}{
			"ready_at":         readyAt.Format(time.RFC3339),
			"expiration_time":  expirationTime.Format(time.RFC3339),
			"time_display":     timeDisplay,
			"elapsed_seconds":  cooldown.TotalSeconds - cooldown.RemainingSeconds,
			"progress_percent": float64(cooldown.TotalSeconds-cooldown.RemainingSeconds) / float64(cooldown.TotalSeconds) * 100,
		},
		"actions": map[string]interface{}{
			"can_extract":     false,
			"can_scan":        false,
			"can_jump":        false,
			"can_navigate":    true, // Navigation is generally not blocked by cooldown
			"can_survey":      false,
			"blocked_actions": blockedActions,
		},
		"recommendations": r.generateCooldownRecommendations(shipSymbol, cooldown.RemainingSeconds),
		"meta": map[string]interface{}{
			"last_checked": now.Format(time.RFC3339),
			"ship_symbol":  shipSymbol,
		},
	}
}

// formatTimeRemaining formats remaining seconds into human-readable format
func (r *ShipCooldownResource) formatTimeRemaining(seconds int) string {
	if seconds <= 0 {
		return "Ready now"
	}

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	remainingSeconds := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, remainingSeconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, remainingSeconds)
	} else {
		return fmt.Sprintf("%ds", remainingSeconds)
	}
}

// determineCooldownStatus determines the status based on remaining time
func (r *ShipCooldownResource) determineCooldownStatus(remainingSeconds int) map[string]interface{} {
	if remainingSeconds <= 0 {
		return map[string]interface{}{
			"operational": true,
			"message":     "Ship is ready for all actions",
			"priority":    "ready",
			"icon":        "🟢",
		}
	}

	var priority, message, icon string

	if remainingSeconds > 600 { // 10 minutes
		priority = "long"
		message = "Long cooldown active - consider using other ships"
		icon = "🔴"
	} else if remainingSeconds > 300 { // 5 minutes
		priority = "medium"
		message = "Moderate cooldown active - plan ahead"
		icon = "🟡"
	} else if remainingSeconds > 60 { // 1 minute
		priority = "short"
		message = "Short cooldown active - almost ready"
		icon = "🟠"
	} else {
		priority = "finishing"
		message = "Cooldown finishing soon"
		icon = "🟡"
	}

	return map[string]interface{}{
		"operational": false,
		"message":     message,
		"priority":    priority,
		"icon":        icon,
	}
}

// analyzeBlockedActions determines which actions are blocked
func (r *ShipCooldownResource) analyzeBlockedActions(remainingSeconds int) []string {
	if remainingSeconds <= 0 {
		return []string{}
	}

	// Most actions are blocked during cooldown
	blocked := []string{
		"extract_resources",
		"scan_systems",
		"scan_waypoints",
		"scan_ships",
		"create_survey",
		"jump_ship",
		"siphon_resources",
	}

	return blocked
}

// generateCooldownRecommendations generates recommendations based on cooldown status
func (r *ShipCooldownResource) generateCooldownRecommendations(shipSymbol string, remainingSeconds int) []string {
	if remainingSeconds <= 0 {
		return []string{
			"Ship is ready for all actions",
			"Good time to plan next operation",
		}
	}

	recommendations := []string{}

	if remainingSeconds > 600 { // 10 minutes
		recommendations = append(recommendations,
			"Long cooldown - switch to other ships for immediate actions",
			"Use this time to plan fleet operations",
			"Consider checking markets or contracts while waiting",
		)
	} else if remainingSeconds > 300 { // 5 minutes
		recommendations = append(recommendations,
			"Moderate cooldown - good time to manage other ships",
			"Plan next action to execute when cooldown expires",
			"Check cargo and fuel levels while waiting",
		)
	} else if remainingSeconds > 60 { // 1 minute
		recommendations = append(recommendations,
			"Short cooldown - prepare for next action",
			"Ship will be ready shortly",
		)
	} else {
		recommendations = append(recommendations,
			"Cooldown finishing soon - prepare immediate action",
			"Ship will be operational in moments",
		)
	}

	// Add general recommendations
	recommendations = append(recommendations,
		"Navigation and trading are still available during cooldown",
		"Use 'get_ship_details' to check full ship status",
	)

	return recommendations
}
