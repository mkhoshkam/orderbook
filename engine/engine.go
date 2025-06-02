package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// Package engine provides a high-performance order matching engine with real-time
// market data streaming capabilities. It implements a multi-asset trading engine
// that manages order books for different trading pairs and provides comprehensive
// trade execution, fill tracking, and market data distribution.
//
// The engine supports:
//   - Multi-asset order book management
//   - Real-time trade execution with price-time priority matching
//   - Live price updates and depth streaming
//   - Order fill tracking and status updates
//   - Trade statistics and analytics
//   - Thread-safe operations with concurrent access
//
// Example usage:
//
//	engine := NewEngine()
//
//	// Start real-time data streams
//	engine.StartPriceBroadcaster()
//	engine.StartDepthStreamer(10)
//
//	// Listen for trades
//	go func() {
//	    for trade := range engine.TradeStream {
//	        fmt.Printf("Trade: %+v\n", trade)
//	    }
//	}()
//
//	// Add orders
//	buyOrder := Order{
//	    ID:    "order1",
//	    Side:  Buy,
//	    Price: decimal.NewFromFloat(50000),
//	    Qty:   decimal.NewFromFloat(0.1),
//	    Time:  time.Now().Unix(),
//	}
//	engine.AddOrder("BTC-USD", buyOrder)

// TradeStats holds aggregate trading statistics for a trading pair.
// It tracks cumulative trading activity including total volume, value, and trade count.
type TradeStats struct {
	TotalQty   decimal.Decimal // Cumulative quantity of all trades
	TotalValue decimal.Decimal // Cumulative value of all trades (qty * price)
	TradeCount int64           // Total number of trades executed
}

// Engine is the core trading engine that manages multiple order books and provides
// real-time market data streaming. It coordinates order matching across different
// trading pairs and distributes trade events, price updates, and market depth information.
//
// The engine is thread-safe and supports concurrent operations across multiple
// trading pairs. It maintains separate order books for each pair and provides
// real-time data feeds through Go channels.
type Engine struct {
	books        map[string]*OrderBook  // Order books indexed by trading pair
	mutex        sync.Mutex             // Protects concurrent access to engine state
	TradeStream  chan Trade             // Stream of executed trades
	PriceUpdates chan PriceUpdate       // Stream of best bid/ask price updates
	DepthUpdates chan DepthUpdate       // Stream of order book depth snapshots
	FillStream   chan OrderFill         // Stream of order fill events
	tradeStats   map[string]*TradeStats // Trading statistics by pair
	tradeCounter int64                  // Global trade counter for unique IDs
}

// NewEngine creates and initializes a new trading engine with default channel capacities.
// The engine is ready to accept orders and start data streaming immediately after creation.
//
// Channel capacities:
//   - TradeStream: 1000 (high capacity for trade events)
//   - PriceUpdates: 100 (moderate capacity for price updates)
//   - DepthUpdates: 100 (moderate capacity for depth updates)
//   - FillStream: 1000 (high capacity for fill events)
//
// Returns a fully initialized engine ready for trading operations.
func NewEngine() *Engine {
	return &Engine{
		books:        make(map[string]*OrderBook),
		TradeStream:  make(chan Trade, 1000),
		PriceUpdates: make(chan PriceUpdate, 100),
		DepthUpdates: make(chan DepthUpdate, 100),
		FillStream:   make(chan OrderFill, 1000),
		tradeStats:   make(map[string]*TradeStats),
		tradeCounter: 0,
	}
}

// getOrCreateBook retrieves an existing order book for the specified trading pair
// or creates a new one if it doesn't exist. This method is thread-safe and ensures
// that each trading pair has exactly one order book instance.
//
// Parameters:
//   - pair: Trading pair identifier (e.g., "BTC-USD", "ETH-BTC")
//
// Returns the order book for the specified pair, creating it if necessary.
func (e *Engine) getOrCreateBook(pair string) *OrderBook {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	book, exists := e.books[pair]
	if !exists {
		book = NewOrderBook(pair)
		e.books[pair] = book
	}
	return book
}

// AddOrder processes a new order for the specified trading pair. The order will be
// matched against existing orders in the order book, potentially generating trades
// and fill events. Any unmatched portion of the order will be added to the order book.
//
// This method is the primary entry point for order processing and handles:
//   - Order validation and matching
//   - Trade generation and broadcasting
//   - Fill event creation and distribution
//   - Trade statistics updates
//   - Order book maintenance
//
// The method operates asynchronously for trade and fill event processing to ensure
// low-latency order processing.
//
// Parameters:
//   - pair: Trading pair identifier (e.g., "BTC-USD")
//   - order: The order to process
//
// Events generated:
//   - Trade events sent to TradeStream channel
//   - OrderFill events sent to FillStream channel
//   - Updated trade statistics
func (e *Engine) AddOrder(pair string, order Order) {
	book := e.getOrCreateBook(pair)
	tradeCh := make(chan Trade, 10)
	fillCh := make(chan OrderFill, 10)

	go func() {
		for trade := range tradeCh {
			e.TradeStream <- trade

			e.mutex.Lock()
			stats := e.tradeStats[pair]
			if stats == nil {
				stats = &TradeStats{}
				e.tradeStats[pair] = stats
			}
			stats.TotalQty = stats.TotalQty.Add(trade.Qty)
			stats.TotalValue = stats.TotalValue.Add(trade.Qty.Mul(trade.Price))
			stats.TradeCount++
			e.mutex.Unlock()
		}
	}()

	go func() {
		for fill := range fillCh {
			e.FillStream <- fill
		}
	}()

	originalQty := order.Qty
	book.Match(order, tradeCh, fillCh, originalQty)
	close(tradeCh)
	close(fillCh)
}

// StartPriceBroadcaster starts a background goroutine that continuously broadcasts
// price updates for all active trading pairs. The broadcaster sends periodic updates
// containing best bid/ask prices and average trade prices.
//
// Update frequency: 500ms
// Channel: PriceUpdates
//
// Price updates include:
//   - Best bid price (highest buy order)
//   - Best ask price (lowest sell order)
//   - Volume-weighted average price (if trades have occurred)
//
// The broadcaster runs indefinitely until the program terminates. If the PriceUpdates
// channel is full, updates are skipped to prevent blocking.
func (e *Engine) StartPriceBroadcaster() {
	go func() {
		for {
			var updates []PriceUpdate

			e.mutex.Lock()
			for pair, book := range e.books {
				update := PriceUpdate{
					Pair:    pair,
					BestBid: decimal.NewFromFloat(book.BestBid()),
					BestAsk: decimal.NewFromFloat(book.BestAsk()),
				}
				stats := e.tradeStats[pair]
				if stats != nil && !stats.TotalQty.IsZero() {
					update.AvgPrice = stats.TotalValue.Div(stats.TotalQty)
				}
				updates = append(updates, update)
			}
			e.mutex.Unlock()

			for _, update := range updates {
				select {
				case e.PriceUpdates <- update:
				default:
					// Skip if channel is full
				}
			}

			time.Sleep(500 * time.Millisecond)
		}
	}()
}

// StartDepthStreamer starts a background goroutine that continuously broadcasts
// order book depth updates for all active trading pairs. The streamer provides
// real-time snapshots of market depth at the specified number of price levels.
//
// Update frequency: 100ms
// Channel: DepthUpdates
//
// Parameters:
//   - depth: Number of price levels to include on each side (bids/asks)
//
// Depth updates include:
//   - Top N bid levels (buy orders)
//   - Top N ask levels (sell orders)
//   - Timestamp of the snapshot
//   - Total trade count for the pair
//
// The streamer runs indefinitely until the program terminates. If the DepthUpdates
// channel is full, updates are skipped to prevent blocking.
func (e *Engine) StartDepthStreamer(depth int) {
	go func() {
		for {
			var updates []DepthUpdate

			e.mutex.Lock()
			for pair, book := range e.books {
				stats := e.tradeStats[pair]
				tradeCount := int64(0)
				if stats != nil {
					tradeCount = stats.TradeCount
				}

				update := DepthUpdate{
					Pair:       pair,
					Bids:       book.GetBidDepth(depth),
					Asks:       book.GetAskDepth(depth),
					Timestamp:  time.Now().Unix(),
					TradeCount: tradeCount,
				}
				updates = append(updates, update)
			}
			e.mutex.Unlock()

			for _, update := range updates {
				select {
				case e.DepthUpdates <- update:
				default:
					// Skip if channel is full
				}
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()
}

// GetOrderBookDepth returns a snapshot of the current order book depth for the
// specified trading pair. This method provides on-demand access to market depth
// information without subscribing to the continuous depth stream.
//
// Parameters:
//   - pair: Trading pair identifier
//   - depth: Number of price levels to include on each side
//
// Returns:
//   - DepthUpdate containing current market depth, or nil if pair doesn't exist
//
// The returned snapshot includes:
//   - Current bid levels (highest to lowest price)
//   - Current ask levels (lowest to highest price)
//   - Current timestamp
//   - Total trade count for the pair
func (e *Engine) GetOrderBookDepth(pair string, depth int) *DepthUpdate {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	book, exists := e.books[pair]
	if !exists {
		return nil
	}

	stats := e.tradeStats[pair]
	tradeCount := int64(0)
	if stats != nil {
		tradeCount = stats.TradeCount
	}

	return &DepthUpdate{
		Pair:       pair,
		Bids:       book.GetBidDepth(depth),
		Asks:       book.GetAskDepth(depth),
		Timestamp:  time.Now().Unix(),
		TradeCount: tradeCount,
	}
}

// GetNextTradeID generates a unique identifier for trade events. Trade IDs are
// sequential and globally unique across all trading pairs.
//
// Returns:
//   - Unique trade ID in format "T{number}" (e.g., "T1", "T2", "T123")
//
// This method is thread-safe and ensures no duplicate trade IDs are generated
// even under high concurrency.
func (e *Engine) GetNextTradeID() string {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.tradeCounter++
	return fmt.Sprintf("T%d", e.tradeCounter)
}
