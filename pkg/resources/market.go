package resources

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"spacetraders-mcp/pkg/logging"
	"spacetraders-mcp/pkg/client"

	"github.com/mark3labs/mcp-go/mcp"
)

// MarketResource handles market data
type MarketResource struct {
	client *client.Client
	logger *logging.Logger
}

// NewMarketResource creates a new market resource
func NewMarketResource(client *client.Client, logger *logging.Logger) *MarketResource {
	return &MarketResource{
		client: client,
		logger: logger,
	}
}

// Resource returns the MCP resource definition
func (r *MarketResource) Resource() mcp.Resource {
	return mcp.Resource{
		URI:         "spacetraders://systems/{systemSymbol}/waypoints/{waypointSymbol}/market",
		Name:        "Market Data",
		Description: "Market prices, trade goods, and trading opportunities at a specific waypoint",
		MIMEType:    "application/json",
	}
}

// Handler returns the resource handler function
func (r *MarketResource) Handler() func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		contextLogger := r.logger.WithContext(ctx, "market-resource")

		// Parse URI to extract system and waypoint symbols
		uri := request.Params.URI
		if !strings.HasPrefix(uri, "spacetraders://systems/") {
			contextLogger.Error(fmt.Sprintf("Invalid URI format: %s", uri))
			return []mcp.ResourceContents{}, fmt.Errorf("invalid URI format")
		}

		// Extract system and waypoint symbols from URI
		// Format: spacetraders://systems/{systemSymbol}/waypoints/{waypointSymbol}/market
		parts := strings.Split(strings.TrimPrefix(uri, "spacetraders://systems/"), "/")
		if len(parts) != 4 || parts[1] != "waypoints" || parts[3] != "market" {
			contextLogger.Error(fmt.Sprintf("Invalid market URI format: %s", uri))
			return []mcp.ResourceContents{}, fmt.Errorf("invalid market URI format")
		}

		systemSymbol := parts[0]
		waypointSymbol := parts[2]

		contextLogger.Debug(fmt.Sprintf("Fetching market data for %s at %s from API", waypointSymbol, systemSymbol))

		// Get market data from the API
		market, err := r.client.GetMarket(systemSymbol, waypointSymbol)
		if err != nil {
			contextLogger.Error(fmt.Sprintf("Failed to fetch market data for %s: %v", waypointSymbol, err))
			return []mcp.ResourceContents{}, fmt.Errorf("failed to fetch market data: %w", err)
		}

		contextLogger.Info(fmt.Sprintf("Successfully retrieved market data for %s at %s", waypointSymbol, systemSymbol))

		// Create the resource content
		content := map[string]interface{}{
			"system":   systemSymbol,
			"waypoint": waypointSymbol,
			"market": map[string]interface{}{
				"symbol":       market.Symbol,
				"exports":      r.formatTradeGoods(market.Exports),
				"imports":      r.formatTradeGoods(market.Imports),
				"exchange":     r.formatTradeGoods(market.Exchange),
				"transactions": r.formatTransactions(market.Transactions),
				"trade_goods":  r.formatTradeGoodsWithPrices(market.TradeGoods),
			},
			"analysis": r.analyzeMarket(market),
		}

		contextLogger.Info("Resource read successful: " + uri)
		contextLogger.Debug(fmt.Sprintf("Market resource response size: %d bytes", len(fmt.Sprintf("%+v", content))))

		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      uri,
				MIMEType: "application/json",
				Text:     r.formatMarketAsText(market, systemSymbol, waypointSymbol),
			},
		}, nil
	}
}

// formatTradeGoods formats trade goods for display
func (r *MarketResource) formatTradeGoods(goods []client.TradeGood) []map[string]interface{} {
	var result []map[string]interface{}
	for _, good := range goods {
		result = append(result, map[string]interface{}{
			"symbol":      good.Symbol,
			"name":        good.Name,
			"description": good.Description,
		})
	}
	return result
}

// formatTradeGoodsWithPrices formats trade goods with pricing information
func (r *MarketResource) formatTradeGoodsWithPrices(goods []client.MarketTradeGood) []map[string]interface{} {
	var result []map[string]interface{}
	for _, good := range goods {
		goodData := map[string]interface{}{
			"symbol":         good.Symbol,
			"type":           good.Type,
			"trade_volume":   good.TradeVolume,
			"supply":         good.Supply,
			"activity":       good.Activity,
			"purchase_price": good.PurchasePrice,
			"sell_price":     good.SellPrice,
		}
		result = append(result, goodData)
	}
	return result
}

// formatTransactions formats recent transactions
func (r *MarketResource) formatTransactions(transactions []client.MarketTransaction) []map[string]interface{} {
	var result []map[string]interface{}
	for _, transaction := range transactions {
		result = append(result, map[string]interface{}{
			"waypoint_symbol": transaction.WaypointSymbol,
			"ship_symbol":     transaction.ShipSymbol,
			"trade_symbol":    transaction.TradeSymbol,
			"type":            transaction.Type,
			"units":           transaction.Units,
			"price_per_unit":  transaction.PricePerUnit,
			"total_price":     transaction.TotalPrice,
			"timestamp":       transaction.Timestamp,
		})
	}
	return result
}

// analyzeMarket provides market analysis and insights
func (r *MarketResource) analyzeMarket(market *client.Market) map[string]interface{} {
	analysis := map[string]interface{}{
		"total_exports":       len(market.Exports),
		"total_imports":       len(market.Imports),
		"total_exchange":      len(market.Exchange),
		"total_trade_goods":   len(market.TradeGoods),
		"recent_transactions": len(market.Transactions),
	}

	// Analyze trade opportunities
	var highValueGoods []string
	var lowSupplyGoods []string
	var activeGoods []string

	for _, good := range market.TradeGoods {
		if good.SellPrice > 100 {
			highValueGoods = append(highValueGoods, good.Symbol)
		}
		if good.Supply == "SCARCE" || good.Supply == "LIMITED" {
			lowSupplyGoods = append(lowSupplyGoods, good.Symbol)
		}
		if good.Activity == "STRONG" || good.Activity == "GROWING" {
			activeGoods = append(activeGoods, good.Symbol)
		}
	}

	analysis["high_value_goods"] = highValueGoods
	analysis["low_supply_goods"] = lowSupplyGoods
	analysis["active_goods"] = activeGoods

	// Market activity classification
	if len(market.Transactions) > 10 {
		analysis["activity_level"] = "HIGH"
	} else if len(market.Transactions) > 5 {
		analysis["activity_level"] = "MODERATE"
	} else {
		analysis["activity_level"] = "LOW"
	}

	return analysis
}

// formatMarketAsText creates a human-readable text representation
func (r *MarketResource) formatMarketAsText(market *client.Market, systemSymbol, waypointSymbol string) string {
	var text strings.Builder

	text.WriteString(fmt.Sprintf("# Market Data: %s\n\n", waypointSymbol))
	text.WriteString(fmt.Sprintf("**System:** %s\n", systemSymbol))
	text.WriteString(fmt.Sprintf("**Waypoint:** %s\n\n", waypointSymbol))

	// Exports
	if len(market.Exports) > 0 {
		text.WriteString("## 📦 Exports (What this market sells)\n")
		for _, export := range market.Exports {
			text.WriteString(fmt.Sprintf("- **%s** - %s\n", export.Name, export.Description))
		}
		text.WriteString("\n")
	}

	// Imports
	if len(market.Imports) > 0 {
		text.WriteString("## 📥 Imports (What this market buys)\n")
		for _, import_ := range market.Imports {
			text.WriteString(fmt.Sprintf("- **%s** - %s\n", import_.Name, import_.Description))
		}
		text.WriteString("\n")
	}

	// Exchange
	if len(market.Exchange) > 0 {
		text.WriteString("## 🔄 Exchange (Goods traded here)\n")
		for _, exchange := range market.Exchange {
			text.WriteString(fmt.Sprintf("- **%s** - %s\n", exchange.Name, exchange.Description))
		}
		text.WriteString("\n")
	}

	// Trade Goods with Prices
	if len(market.TradeGoods) > 0 {
		text.WriteString("## 💰 Current Prices\n\n")

		// Sort by sell price descending for better readability
		sortedGoods := make([]client.MarketTradeGood, len(market.TradeGoods))
		copy(sortedGoods, market.TradeGoods)
		sort.Slice(sortedGoods, func(i, j int) bool {
			return sortedGoods[i].SellPrice > sortedGoods[j].SellPrice
		})

		text.WriteString("| Good | Buy Price | Sell Price | Supply | Activity |\n")
		text.WriteString("|------|-----------|------------|---------|----------|\n")

		for _, good := range sortedGoods {
			supplyIcon := r.getSupplyIcon(good.Supply)
			activityIcon := r.getActivityIcon(good.Activity)

			text.WriteString(fmt.Sprintf("| %s | %d | %d | %s %s | %s %s |\n",
				good.Symbol,
				good.PurchasePrice,
				good.SellPrice,
				supplyIcon,
				good.Supply,
				activityIcon,
				good.Activity))
		}
		text.WriteString("\n")
	}

	// Recent Transactions
	if len(market.Transactions) > 0 {
		text.WriteString("## 📊 Recent Transactions\n\n")
		text.WriteString("| Ship | Good | Type | Units | Price/Unit | Total |\n")
		text.WriteString("|------|------|------|-------|------------|-------|\n")

		// Show last 10 transactions
		transactionCount := len(market.Transactions)
		if transactionCount > 10 {
			transactionCount = 10
		}

		for i := 0; i < transactionCount; i++ {
			transaction := market.Transactions[i]
			text.WriteString(fmt.Sprintf("| %s | %s | %s | %d | %d | %d |\n",
				transaction.ShipSymbol,
				transaction.TradeSymbol,
				transaction.Type,
				transaction.Units,
				transaction.PricePerUnit,
				transaction.TotalPrice))
		}
		text.WriteString("\n")
	}

	// Trading Opportunities
	text.WriteString("## 🎯 Trading Opportunities\n\n")

	var opportunities []string
	for _, good := range market.TradeGoods {
		if good.Supply == "SCARCE" && good.SellPrice > 50 {
			opportunities = append(opportunities, fmt.Sprintf("🔥 **%s** is SCARCE - high sell price of %d credits", good.Symbol, good.SellPrice))
		} else if good.Activity == "STRONG" {
			opportunities = append(opportunities, fmt.Sprintf("📈 **%s** has STRONG activity - good for trading", good.Symbol))
		} else if good.Supply == "ABUNDANT" && good.PurchasePrice < 20 {
			opportunities = append(opportunities, fmt.Sprintf("💰 **%s** is ABUNDANT - low buy price of %d credits", good.Symbol, good.PurchasePrice))
		}
	}

	if len(opportunities) > 0 {
		for _, opportunity := range opportunities {
			text.WriteString(fmt.Sprintf("- %s\n", opportunity))
		}
	} else {
		text.WriteString("- No significant trading opportunities identified at this time\n")
	}

	text.WriteString("\n")

	// Market Summary
	text.WriteString("## 📋 Market Summary\n\n")
	text.WriteString(fmt.Sprintf("- **Total Goods Available:** %d\n", len(market.TradeGoods)))
	text.WriteString(fmt.Sprintf("- **Exports:** %d goods\n", len(market.Exports)))
	text.WriteString(fmt.Sprintf("- **Imports:** %d goods\n", len(market.Imports)))
	text.WriteString(fmt.Sprintf("- **Recent Transactions:** %d\n", len(market.Transactions)))

	return text.String()
}

// getSupplyIcon returns an icon for supply level
func (r *MarketResource) getSupplyIcon(supply string) string {
	switch supply {
	case "SCARCE":
		return "🔴"
	case "LIMITED":
		return "🟡"
	case "MODERATE":
		return "🟢"
	case "HIGH":
		return "🔵"
	case "ABUNDANT":
		return "🟣"
	default:
		return "⚪"
	}
}

// getActivityIcon returns an icon for activity level
func (r *MarketResource) getActivityIcon(activity string) string {
	switch activity {
	case "WEAK":
		return "📉"
	case "GROWING":
		return "📈"
	case "STRONG":
		return "🚀"
	case "RESTRICTED":
		return "🚫"
	default:
		return "➖"
	}
}
