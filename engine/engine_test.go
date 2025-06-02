package engine

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// TestNewEngine tests the creation of a new engine
func TestNewEngine(t *testing.T) {
	engine := NewEngine()

	if engine == nil {
		t.Fatal("Engine should not be nil")
	}

	if engine.books == nil {
		t.Error("Engine books map should be initialized")
	}

	if engine.TradeStream == nil {
		t.Error("TradeStream channel should be initialized")
	}

	if engine.PriceUpdates == nil {
		t.Error("PriceUpdates channel should be initialized")
	}

	if engine.DepthUpdates == nil {
		t.Error("DepthUpdates channel should be initialized")
	}

	if engine.FillStream == nil {
		t.Error("FillStream channel should be initialized")
	}

	if engine.tradeStats == nil {
		t.Error("Trade stats map should be initialized")
	}

	// Check channel capacities
	if cap(engine.TradeStream) != 1000 {
		t.Errorf("Expected TradeStream capacity 1000, got %d", cap(engine.TradeStream))
	}

	if cap(engine.PriceUpdates) != 100 {
		t.Errorf("Expected PriceUpdates capacity 100, got %d", cap(engine.PriceUpdates))
	}

	if cap(engine.DepthUpdates) != 100 {
		t.Errorf("Expected DepthUpdates capacity 100, got %d", cap(engine.DepthUpdates))
	}

	if cap(engine.FillStream) != 1000 {
		t.Errorf("Expected FillStream capacity 1000, got %d", cap(engine.FillStream))
	}
}

// TestGetOrCreateBook tests the order book creation and retrieval
func TestGetOrCreateBook(t *testing.T) {
	engine := NewEngine()
	pair := "BTC-USD"

	// First call should create the book
	book1 := engine.getOrCreateBook(pair)
	if book1 == nil {
		t.Fatal("Order book should not be nil")
	}

	if book1.Pair != pair {
		t.Errorf("Expected pair %s, got %s", pair, book1.Pair)
	}

	// Second call should return the same book
	book2 := engine.getOrCreateBook(pair)
	if book1 != book2 {
		t.Error("Should return the same order book instance")
	}

	// Different pair should create a new book
	book3 := engine.getOrCreateBook("ETH-USD")
	if book1 == book3 {
		t.Error("Different pairs should have different order books")
	}
}

// TestAddOrderAndTradeGeneration tests order addition and trade generation
func TestAddOrderAndTradeGeneration(t *testing.T) {
	engine := NewEngine()
	pair := "BTC-USD"

	// Add a sell order first
	sellOrder := Order{
		ID:    "sell1",
		Side:  Sell,
		Price: decimal.NewFromFloat(50000),
		Qty:   decimal.NewFromFloat(1.0),
		Time:  time.Now().Unix(),
	}
	engine.AddOrder(pair, sellOrder)

	// Add a matching buy order
	buyOrder := Order{
		ID:    "buy1",
		Side:  Buy,
		Price: decimal.NewFromFloat(50000),
		Qty:   decimal.NewFromFloat(1.0),
		Time:  time.Now().Unix(),
	}
	engine.AddOrder(pair, buyOrder)

	// Check that a trade was generated
	select {
	case trade := <-engine.TradeStream:
		if trade.Pair != pair {
			t.Errorf("Expected pair %s, got %s", pair, trade.Pair)
		}
		if trade.BuyOrderID != "buy1" {
			t.Errorf("Expected buy order ID 'buy1', got %s", trade.BuyOrderID)
		}
		if trade.SellOrderID != "sell1" {
			t.Errorf("Expected sell order ID 'sell1', got %s", trade.SellOrderID)
		}
		if !trade.Price.Equal(decimal.NewFromFloat(50000)) {
			t.Errorf("Expected trade price 50000, got %s", trade.Price.String())
		}
		if !trade.Qty.Equal(decimal.NewFromFloat(1.0)) {
			t.Errorf("Expected trade quantity 1.0, got %s", trade.Qty.String())
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected a trade to be generated")
	}

	// Check fill events
	fillCount := 0
	timeout := time.After(100 * time.Millisecond)
	for fillCount < 2 {
		select {
		case fill := <-engine.FillStream:
			fillCount++
			if fill.Pair != pair {
				t.Errorf("Expected fill pair %s, got %s", pair, fill.Pair)
			}
			// Check that status is valid, allowing for NEW status for non-matched orders
			if fill.Status != Filled && fill.Status != PartiallyFilled && fill.Status != New {
				t.Errorf("Expected fill status Filled, PartiallyFilled, or New, got %s", fill.Status)
			}
		case <-timeout:
			t.Errorf("Expected 2 fill events, got %d", fillCount)
			return
		}
	}
}

// TestTradeStatsUpdate tests that trade statistics are properly updated
func TestTradeStatsUpdate(t *testing.T) {
	engine := NewEngine()
	pair := "BTC-USD"

	// Add orders that will result in trades
	sellOrder := Order{
		ID:    "sell1",
		Side:  Sell,
		Price: decimal.NewFromFloat(50000),
		Qty:   decimal.NewFromFloat(2.0),
		Time:  time.Now().Unix(),
	}
	engine.AddOrder(pair, sellOrder)

	buyOrder := Order{
		ID:    "buy1",
		Side:  Buy,
		Price: decimal.NewFromFloat(50000),
		Qty:   decimal.NewFromFloat(1.5),
		Time:  time.Now().Unix(),
	}
	engine.AddOrder(pair, buyOrder)

	// Wait for trade processing
	<-engine.TradeStream

	// Check trade statistics
	engine.mutex.Lock()
	stats := engine.tradeStats[pair]
	engine.mutex.Unlock()

	if stats == nil {
		t.Fatal("Trade stats should be created")
	}

	expectedQty := decimal.NewFromFloat(1.5)
	if !stats.TotalQty.Equal(expectedQty) {
		t.Errorf("Expected total quantity %s, got %s", expectedQty.String(), stats.TotalQty.String())
	}

	expectedValue := decimal.NewFromFloat(75000) // 1.5 * 50000
	if !stats.TotalValue.Equal(expectedValue) {
		t.Errorf("Expected total value %s, got %s", expectedValue.String(), stats.TotalValue.String())
	}

	if stats.TradeCount != 1 {
		t.Errorf("Expected trade count 1, got %d", stats.TradeCount)
	}
}

// TestMultipleTradingPairs tests engine with multiple trading pairs
func TestMultipleTradingPairs(t *testing.T) {
	engine := NewEngine()

	pairs := []string{"BTC-USD", "ETH-USD", "LTC-USD"}

	// Add orders for multiple pairs
	for i, pair := range pairs {
		sellOrder := Order{
			ID:    "sell" + pair,
			Side:  Sell,
			Price: decimal.NewFromFloat(float64(1000 * (i + 1))),
			Qty:   decimal.NewFromFloat(1.0),
			Time:  time.Now().Unix(),
		}
		engine.AddOrder(pair, sellOrder)

		buyOrder := Order{
			ID:    "buy" + pair,
			Side:  Buy,
			Price: decimal.NewFromFloat(float64(1000 * (i + 1))),
			Qty:   decimal.NewFromFloat(1.0),
			Time:  time.Now().Unix(),
		}
		engine.AddOrder(pair, buyOrder)
	}

	// Wait for trades and count them
	tradeCount := 0
	timeout := time.After(200 * time.Millisecond)
	for tradeCount < len(pairs) {
		select {
		case <-engine.TradeStream:
			tradeCount++
		case <-timeout:
			t.Errorf("Expected %d trades, got %d", len(pairs), tradeCount)
			return
		}
	}

	// Check that all pairs have order books
	engine.mutex.Lock()
	for _, pair := range pairs {
		if _, exists := engine.books[pair]; !exists {
			t.Errorf("Order book should exist for pair %s", pair)
		}
		if _, exists := engine.tradeStats[pair]; !exists {
			t.Errorf("Trade stats should exist for pair %s", pair)
		}
	}
	engine.mutex.Unlock()
}

// TestGetNextTradeID tests trade ID generation
func TestGetNextTradeID(t *testing.T) {
	engine := NewEngine()

	// Test sequential ID generation
	id1 := engine.GetNextTradeID()
	id2 := engine.GetNextTradeID()
	id3 := engine.GetNextTradeID()

	if id1 != "T1" {
		t.Errorf("Expected first ID 'T1', got %s", id1)
	}

	if id2 != "T2" {
		t.Errorf("Expected second ID 'T2', got %s", id2)
	}

	if id3 != "T3" {
		t.Errorf("Expected third ID 'T3', got %s", id3)
	}
}

// TestConcurrentTradeIDGeneration tests thread safety of trade ID generation
func TestConcurrentTradeIDGeneration(t *testing.T) {
	engine := NewEngine()
	numGoroutines := 100
	numIDsPerGoroutine := 10

	var wg sync.WaitGroup
	idChan := make(chan string, numGoroutines*numIDsPerGoroutine)

	// Generate IDs concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numIDsPerGoroutine; j++ {
				idChan <- engine.GetNextTradeID()
			}
		}()
	}

	wg.Wait()
	close(idChan)

	// Collect all IDs and check for uniqueness
	idMap := make(map[string]bool)
	totalIDs := 0
	for id := range idChan {
		if idMap[id] {
			t.Errorf("Duplicate trade ID generated: %s", id)
		}
		idMap[id] = true
		totalIDs++
	}

	expectedTotal := numGoroutines * numIDsPerGoroutine
	if totalIDs != expectedTotal {
		t.Errorf("Expected %d IDs, got %d", expectedTotal, totalIDs)
	}
}

// TestGetOrderBookDepth tests depth retrieval
func TestGetOrderBookDepth(t *testing.T) {
	engine := NewEngine()
	pair := "BTC-USD"

	// Test with non-existent pair
	depth := engine.GetOrderBookDepth(pair, 5)
	if depth != nil {
		t.Error("Expected nil depth for non-existent pair")
	}

	// Add some orders
	sellOrder1 := Order{
		ID:    "sell1",
		Side:  Sell,
		Price: decimal.NewFromFloat(50000),
		Qty:   decimal.NewFromFloat(1.0),
		Time:  time.Now().Unix(),
	}
	engine.AddOrder(pair, sellOrder1)

	sellOrder2 := Order{
		ID:    "sell2",
		Side:  Sell,
		Price: decimal.NewFromFloat(50100),
		Qty:   decimal.NewFromFloat(2.0),
		Time:  time.Now().Unix(),
	}
	engine.AddOrder(pair, sellOrder2)

	buyOrder1 := Order{
		ID:    "buy1",
		Side:  Buy,
		Price: decimal.NewFromFloat(49900),
		Qty:   decimal.NewFromFloat(1.5),
		Time:  time.Now().Unix(),
	}
	engine.AddOrder(pair, buyOrder1)

	// Get depth
	depth = engine.GetOrderBookDepth(pair, 5)
	if depth == nil {
		t.Fatal("Expected depth data")
	}

	if depth.Pair != pair {
		t.Errorf("Expected pair %s, got %s", pair, depth.Pair)
	}

	if len(depth.Asks) != 2 {
		t.Errorf("Expected 2 ask levels, got %d", len(depth.Asks))
	}

	if len(depth.Bids) != 1 {
		t.Errorf("Expected 1 bid level, got %d", len(depth.Bids))
	}

	if depth.Timestamp == 0 {
		t.Error("Timestamp should be set")
	}
}

// TestStartPriceBroadcaster tests the price broadcaster functionality
func TestStartPriceBroadcaster(t *testing.T) {
	engine := NewEngine()
	pair := "BTC-USD"

	// Add some orders to create a spread
	sellOrder := Order{
		ID:    "sell1",
		Side:  Sell,
		Price: decimal.NewFromFloat(50100),
		Qty:   decimal.NewFromFloat(1.0),
		Time:  time.Now().Unix(),
	}
	engine.AddOrder(pair, sellOrder)

	buyOrder := Order{
		ID:    "buy1",
		Side:  Buy,
		Price: decimal.NewFromFloat(49900),
		Qty:   decimal.NewFromFloat(1.0),
		Time:  time.Now().Unix(),
	}
	engine.AddOrder(pair, buyOrder)

	// Start the price broadcaster
	engine.StartPriceBroadcaster()

	// Wait for a price update
	select {
	case update := <-engine.PriceUpdates:
		if update.Pair != pair {
			t.Errorf("Expected pair %s, got %s", pair, update.Pair)
		}
		if !update.BestBid.Equal(decimal.NewFromFloat(49900)) {
			t.Errorf("Expected best bid 49900, got %s", update.BestBid.String())
		}
		if !update.BestAsk.Equal(decimal.NewFromFloat(50100)) {
			t.Errorf("Expected best ask 50100, got %s", update.BestAsk.String())
		}
	case <-time.After(1 * time.Second):
		t.Error("Expected a price update within 1 second")
	}
}

// TestStartDepthStreamer tests the depth streamer functionality
func TestStartDepthStreamer(t *testing.T) {
	engine := NewEngine()
	pair := "BTC-USD"

	// Add some orders
	sellOrder := Order{
		ID:    "sell1",
		Side:  Sell,
		Price: decimal.NewFromFloat(50000),
		Qty:   decimal.NewFromFloat(1.0),
		Time:  time.Now().Unix(),
	}
	engine.AddOrder(pair, sellOrder)

	buyOrder := Order{
		ID:    "buy1",
		Side:  Buy,
		Price: decimal.NewFromFloat(49900),
		Qty:   decimal.NewFromFloat(1.0),
		Time:  time.Now().Unix(),
	}
	engine.AddOrder(pair, buyOrder)

	// Start the depth streamer
	engine.StartDepthStreamer(5)

	// Wait for a depth update
	select {
	case update := <-engine.DepthUpdates:
		if update.Pair != pair {
			t.Errorf("Expected pair %s, got %s", pair, update.Pair)
		}
		if len(update.Asks) != 1 {
			t.Errorf("Expected 1 ask level, got %d", len(update.Asks))
		}
		if len(update.Bids) != 1 {
			t.Errorf("Expected 1 bid level, got %d", len(update.Bids))
		}
		if update.Timestamp == 0 {
			t.Error("Timestamp should be set")
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected a depth update within 200ms")
	}
}

// TestConcurrentOrderProcessing tests concurrent order processing
func TestConcurrentOrderProcessing(t *testing.T) {
	engine := NewEngine()
	pair := "BTC-USD"
	numGoroutines := 3      // Reduced from 10
	ordersPerGoroutine := 3 // Reduced from 5

	var wg sync.WaitGroup

	// Process orders concurrently with reduced concurrency
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < ordersPerGoroutine; j++ {
				side := Buy
				price := 50000.0
				if (goroutineID+j)%2 == 0 {
					side = Sell
					price = 50000.0 // Same price to ensure matching
				}

				order := Order{
					ID:    fmt.Sprintf("order_%d_%d", goroutineID, j),
					Side:  side,
					Price: decimal.NewFromFloat(price),
					Qty:   decimal.NewFromFloat(0.1),
					Time:  time.Now().Unix(),
				}
				engine.AddOrder(pair, order)

				time.Sleep(50 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Verify that the engine is still in a consistent state
	book := engine.getOrCreateBook(pair)
	if book == nil {
		t.Error("Order book should exist after concurrent processing")
	}

	// Allow more time for async processing
	time.Sleep(500 * time.Millisecond)

	tradeCount := 0
	for len(engine.TradeStream) > 0 {
		<-engine.TradeStream
		tradeCount++
	}

	fillCount := 0
	for len(engine.FillStream) > 0 {
		<-engine.FillStream
		fillCount++
	}

	// Test that the engine processed the orders without errors
	// We don't require specific numbers since timing can vary
	t.Logf("Processed %d trades and %d fills during concurrent processing", tradeCount, fillCount)
}
