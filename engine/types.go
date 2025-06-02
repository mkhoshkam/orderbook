// Package engine provides types and data structures for the order matching engine.
package engine

import "github.com/shopspring/decimal"

// Side represents the direction of a trading order (buy or sell).
type Side string

const (
	// Buy represents a buy order (bid) - an order to purchase an asset.
	Buy Side = "buy"
	// Sell represents a sell order (ask) - an order to sell an asset.
	Sell Side = "sell"
)

// Order represents a trading order with all necessary information for matching.
// Orders are the fundamental unit of trading in the engine and contain all
// details needed for price-time priority matching.
type Order struct {
	ID    string          // Unique identifier for the order
	Side  Side            // Direction of the order (Buy or Sell)
	Price decimal.Decimal // Price per unit for the order
	Qty   decimal.Decimal // Quantity/amount to trade
	Time  int64           // Unix timestamp when the order was created
}

// Trade represents a successful match between two orders resulting in an execution.
// Trades are generated when buy and sell orders are matched at a specific price and quantity.
type Trade struct {
	Pair        string          // Trading pair identifier (e.g., "BTC-USD")
	BuyOrderID  string          // ID of the buy order involved in the trade
	SellOrderID string          // ID of the sell order involved in the trade
	Price       decimal.Decimal // Execution price of the trade
	Qty         decimal.Decimal // Quantity traded
}

// PriceUpdate contains current best bid/ask prices and average price information
// for a trading pair. These updates are broadcast periodically to provide
// real-time price information to market participants.
type PriceUpdate struct {
	Pair     string          // Trading pair identifier
	BestBid  decimal.Decimal // Highest bid (buy) price currently available
	BestAsk  decimal.Decimal // Lowest ask (sell) price currently available
	AvgPrice decimal.Decimal // Volume-weighted average price of recent trades
}

// DepthLevel represents a single price level in the order book with aggregated
// quantity and order count information. Multiple orders at the same price are
// aggregated into a single depth level.
type DepthLevel struct {
	Price      decimal.Decimal // Price level
	Quantity   decimal.Decimal // Total quantity available at this price level
	TradeCount int             // Number of individual orders at this price level
}

// DepthUpdate provides a snapshot of the order book depth showing the best
// bid and ask levels. This gives market participants visibility into market
// liquidity and potential support/resistance levels.
type DepthUpdate struct {
	Pair       string       // Trading pair identifier
	Bids       []DepthLevel // Bid (buy) levels ordered from highest to lowest price
	Asks       []DepthLevel // Ask (sell) levels ordered from lowest to highest price
	Timestamp  int64        // Unix timestamp of the snapshot
	TradeCount int64        // Total number of trades executed for this pair
}

// FillStatus represents the current execution status of an order.
// Orders progress through different states as they are processed and matched.
type FillStatus string

const (
	// PartiallyFilled indicates the order has been partially executed but still
	// has remaining quantity to be filled.
	PartiallyFilled FillStatus = "PARTIALLY_FILLED"

	// Filled indicates the order has been completely executed with no remaining quantity.
	Filled FillStatus = "FILLED"

	// New indicates the order has been accepted but not yet executed.
	New FillStatus = "NEW"
)

// OrderFill represents the execution details of an order or part of an order.
// Fill events provide detailed information about order execution status and
// are essential for order management and trade reporting.
type OrderFill struct {
	OrderID      string          // Unique identifier of the order being filled
	Pair         string          // Trading pair identifier
	Side         Side            // Direction of the order (Buy or Sell)
	OriginalQty  decimal.Decimal // Original quantity when the order was placed
	ExecutedQty  decimal.Decimal // Quantity executed in this fill event
	RemainingQty decimal.Decimal // Quantity remaining to be filled
	Price        decimal.Decimal // Original order price
	FillPrice    decimal.Decimal // Actual execution price for this fill
	Status       FillStatus      // Current status of the order after this fill
	Timestamp    int64           // Unix timestamp when the fill occurred
}
