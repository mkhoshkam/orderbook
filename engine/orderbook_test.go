package engine

import (
    "testing"
    "time"

    "github.com/shopspring/decimal"
)

// TestNewOrderBook tests the creation of a new order book
func TestNewOrderBook(t *testing.T) {
    pair := "BTC-USDT"
    ob := NewOrderBook(pair)

    if ob.Pair != pair {
        t.Errorf("Expected pair %s, got %s", pair, ob.Pair)
    }

    if ob.BestBid() != 0 {
        t.Errorf("Expected empty order book to have 0 best bid, got %f", ob.BestBid())
    }

    if ob.BestAsk() != 0 {
        t.Errorf("Expected empty order book to have 0 best ask, got %f", ob.BestAsk())
    }
}

// TestOrderBookBestPrices tests the BestBid and BestAsk methods
func TestOrderBookBestPrices(t *testing.T) {
    ob := NewOrderBook("BTC-USDT")
    tradeCh := make(chan Trade, 10)
    fillCh := make(chan OrderFill, 10)

    // Add a buy order
    buyOrder := Order{
        ID:    "buy1",
        Side:  Buy,
        Price: decimal.NewFromFloat(100.0),
        Qty:   decimal.NewFromFloat(1.0),
        Time:  time.Now().Unix(),
    }
    ob.Match(buyOrder, tradeCh, fillCh, buyOrder.Qty)

    // Add a sell order
    sellOrder := Order{
        ID:    "sell1",
        Side:  Sell,
        Price: decimal.NewFromFloat(105.0),
        Qty:   decimal.NewFromFloat(1.0),
        Time:  time.Now().Unix(),
    }
    ob.Match(sellOrder, tradeCh, fillCh, sellOrder.Qty)

    // Check best prices
    if ob.BestBid() != 100.0 {
        t.Errorf("Expected best bid 100.0, got %f", ob.BestBid())
    }

    if ob.BestAsk() != 105.0 {
        t.Errorf("Expected best ask 105.0, got %f", ob.BestAsk())
    }
}

// TestOrderBookMatching tests the order matching functionality
func TestOrderBookMatching(t *testing.T) {
    ob := NewOrderBook("BTC-USDT")
    tradeCh := make(chan Trade, 10)
    fillCh := make(chan OrderFill, 10)

    // Add a sell order first
    sellOrder := Order{
        ID:    "sell1",
        Side:  Sell,
        Price: decimal.NewFromFloat(100.0),
        Qty:   decimal.NewFromFloat(1.0),
        Time:  time.Now().Unix(),
    }
    ob.Match(sellOrder, tradeCh, fillCh, sellOrder.Qty)

    // Drain the fill channel
    <-fillCh

    // Add a matching buy order
    buyOrder := Order{
        ID:    "buy1",
        Side:  Buy,
        Price: decimal.NewFromFloat(100.0),
        Qty:   decimal.NewFromFloat(1.0),
        Time:  time.Now().Unix(),
    }
    ob.Match(buyOrder, tradeCh, fillCh, buyOrder.Qty)

    // Check that a trade was generated
    select {
    case trade := <-tradeCh:
        if trade.BuyOrderID != "buy1" {
            t.Errorf("Expected buy order ID 'buy1', got %s", trade.BuyOrderID)
        }
        if trade.SellOrderID != "sell1" {
            t.Errorf("Expected sell order ID 'sell1', got %s", trade.SellOrderID)
        }
        if !trade.Price.Equal(decimal.NewFromFloat(100.0)) {
            t.Errorf("Expected trade price 100.0, got %s", trade.Price.String())
        }
        if !trade.Qty.Equal(decimal.NewFromFloat(1.0)) {
            t.Errorf("Expected trade quantity 1.0, got %s", trade.Qty.String())
        }
    default:
        t.Error("Expected a trade to be generated")
    }

    // Check fill events - should have 2 fills (one for each order)
    fillCount := 0
    for len(fillCh) > 0 {
        <-fillCh
        fillCount++
    }
    if fillCount != 2 {
        t.Errorf("Expected 2 fill events, got %d", fillCount)
    }
}

// TestPartialFill tests partial order filling
func TestPartialFill(t *testing.T) {
    ob := NewOrderBook("BTC-USDT")
    tradeCh := make(chan Trade, 10)
    fillCh := make(chan OrderFill, 10)

    // Add a large sell order
    sellOrder := Order{
        ID:    "sell1",
        Side:  Sell,
        Price: decimal.NewFromFloat(100.0),
        Qty:   decimal.NewFromFloat(5.0),
        Time:  time.Now().Unix(),
    }
    ob.Match(sellOrder, tradeCh, fillCh, sellOrder.Qty)

    // Drain the fill channel
    <-fillCh

    // Add a smaller buy order
    buyOrder := Order{
        ID:    "buy1",
        Side:  Buy,
        Price: decimal.NewFromFloat(100.0),
        Qty:   decimal.NewFromFloat(2.0),
        Time:  time.Now().Unix(),
    }
    ob.Match(buyOrder, tradeCh, fillCh, buyOrder.Qty)

    // Check trade
    select {
    case trade := <-tradeCh:
        if !trade.Qty.Equal(decimal.NewFromFloat(2.0)) {
            t.Errorf("Expected trade quantity 2.0, got %s", trade.Qty.String())
        }
    default:
        t.Error("Expected a trade to be generated")
    }

    // Check that there's still an ask in the book (partially filled sell order)
    if ob.BestAsk() != 100.0 {
        t.Errorf("Expected remaining ask at 100.0, got %f", ob.BestAsk())
    }
}

// TestMultiplePriceLevels tests orders at different price levels
func TestMultiplePriceLevels(t *testing.T) {
    ob := NewOrderBook("BTC-USDT")
    tradeCh := make(chan Trade, 10)
    fillCh := make(chan OrderFill, 10)

    // Add buy orders at different prices
    buyOrder1 := Order{
        ID:    "buy1",
        Side:  Buy,
        Price: decimal.NewFromFloat(100.0),
        Qty:   decimal.NewFromFloat(1.0),
        Time:  time.Now().Unix(),
    }
    ob.Match(buyOrder1, tradeCh, fillCh, buyOrder1.Qty)

    buyOrder2 := Order{
        ID:    "buy2",
        Side:  Buy,
        Price: decimal.NewFromFloat(99.0),
        Qty:   decimal.NewFromFloat(1.0),
        Time:  time.Now().Unix(),
    }
    ob.Match(buyOrder2, tradeCh, fillCh, buyOrder2.Qty)

    // Best bid should be the higher price
    if ob.BestBid() != 100.0 {
        t.Errorf("Expected best bid 100.0, got %f", ob.BestBid())
    }

    // Add sell orders at different prices
    sellOrder1 := Order{
        ID:    "sell1",
        Side:  Sell,
        Price: decimal.NewFromFloat(105.0),
        Qty:   decimal.NewFromFloat(1.0),
        Time:  time.Now().Unix(),
    }
    ob.Match(sellOrder1, tradeCh, fillCh, sellOrder1.Qty)

    sellOrder2 := Order{
        ID:    "sell2",
        Side:  Sell,
        Price: decimal.NewFromFloat(104.0),
        Qty:   decimal.NewFromFloat(1.0),
        Time:  time.Now().Unix(),
    }
    ob.Match(sellOrder2, tradeCh, fillCh, sellOrder2.Qty)

    // Best ask should be the lower price
    if ob.BestAsk() != 104.0 {
        t.Errorf("Expected best ask 104.0, got %f", ob.BestAsk())
    }
}

// TestGetBidDepth tests the bid depth functionality
func TestGetBidDepth(t *testing.T) {
    ob := NewOrderBook("BTC-USDT")
    tradeCh := make(chan Trade, 10)
    fillCh := make(chan OrderFill, 10)

    // Add multiple buy orders at same price
    buyOrder1 := Order{
        ID:    "buy1",
        Side:  Buy,
        Price: decimal.NewFromFloat(100.0),
        Qty:   decimal.NewFromFloat(1.0),
        Time:  time.Now().Unix(),
    }
    ob.Match(buyOrder1, tradeCh, fillCh, buyOrder1.Qty)

    buyOrder2 := Order{
        ID:    "buy2",
        Side:  Buy,
        Price: decimal.NewFromFloat(100.0),
        Qty:   decimal.NewFromFloat(2.0),
        Time:  time.Now().Unix(),
    }
    ob.Match(buyOrder2, tradeCh, fillCh, buyOrder2.Qty)

    // Add buy order at different price
    buyOrder3 := Order{
        ID:    "buy3",
        Side:  Buy,
        Price: decimal.NewFromFloat(99.0),
        Qty:   decimal.NewFromFloat(1.5),
        Time:  time.Now().Unix(),
    }
    ob.Match(buyOrder3, tradeCh, fillCh, buyOrder3.Qty)

    // Test depth
    depth := ob.GetBidDepth(2)
    if len(depth) != 2 {
        t.Errorf("Expected 2 depth levels, got %d", len(depth))
    }

    // First level should be at 100.0 with quantity 3.0
    if !depth[0].Price.Equal(decimal.NewFromFloat(100.0)) {
        t.Errorf("Expected first level price 100.0, got %s", depth[0].Price.String())
    }
    if !depth[0].Quantity.Equal(decimal.NewFromFloat(3.0)) {
        t.Errorf("Expected first level quantity 3.0, got %s", depth[0].Quantity.String())
    }
    if depth[0].TradeCount != 2 {
        t.Errorf("Expected first level trade count 2, got %d", depth[0].TradeCount)
    }
}

// TestGetAskDepth tests the ask depth functionality
func TestGetAskDepth(t *testing.T) {
    ob := NewOrderBook("BTC-USDT")
    tradeCh := make(chan Trade, 10)
    fillCh := make(chan OrderFill, 10)

    // Add multiple sell orders
    sellOrder1 := Order{
        ID:    "sell1",
        Side:  Sell,
        Price: decimal.NewFromFloat(105.0),
        Qty:   decimal.NewFromFloat(1.0),
        Time:  time.Now().Unix(),
    }
    ob.Match(sellOrder1, tradeCh, fillCh, sellOrder1.Qty)

    sellOrder2 := Order{
        ID:    "sell2",
        Side:  Sell,
        Price: decimal.NewFromFloat(104.0),
        Qty:   decimal.NewFromFloat(2.0),
        Time:  time.Now().Unix(),
    }
    ob.Match(sellOrder2, tradeCh, fillCh, sellOrder2.Qty)

    // Test depth
    depth := ob.GetAskDepth(2)
    if len(depth) != 2 {
        t.Errorf("Expected 2 depth levels, got %d", len(depth))
    }

    // First level should be the lowest price (104.0)
    if !depth[0].Price.Equal(decimal.NewFromFloat(104.0)) {
        t.Errorf("Expected first level price 104.0, got %s", depth[0].Price.String())
    }
}

// TestEmptyDepth tests depth methods with empty order book
func TestEmptyDepth(t *testing.T) {
    ob := NewOrderBook("BTC-USDT")

    bidDepth := ob.GetBidDepth(5)
    if len(bidDepth) != 0 {
        t.Errorf("Expected empty bid depth, got %d levels", len(bidDepth))
    }

    askDepth := ob.GetAskDepth(5)
    if len(askDepth) != 0 {
        t.Errorf("Expected empty ask depth, got %d levels", len(askDepth))
    }
}

// TestInvalidDepth tests depth methods with invalid parameters
func TestInvalidDepth(t *testing.T) {
    ob := NewOrderBook("BTC-USDT")
    tradeCh := make(chan Trade, 10)
    fillCh := make(chan OrderFill, 10)

    // Add an order
    order := Order{
        ID:    "test1",
        Side:  Buy,
        Price: decimal.NewFromFloat(100.0),
        Qty:   decimal.NewFromFloat(1.0),
        Time:  time.Now().Unix(),
    }
    ob.Match(order, tradeCh, fillCh, order.Qty)

    // Test with invalid depth parameters
    bidDepth := ob.GetBidDepth(0)
    if len(bidDepth) != 0 {
        t.Errorf("Expected empty result for depth 0, got %d levels", len(bidDepth))
    }

    bidDepth = ob.GetBidDepth(-1)
    if len(bidDepth) != 0 {
        t.Errorf("Expected empty result for negative depth, got %d levels", len(bidDepth))
    }
}

// TestOrderFillEvents tests that proper fill events are generated
func TestOrderFillEvents(t *testing.T) {
    ob := NewOrderBook("BTC-USDT")
    tradeCh := make(chan Trade, 10)
    fillCh := make(chan OrderFill, 10)

    // Add a new order that doesn't match
    order := Order{
        ID:    "test1",
        Side:  Buy,
        Price: decimal.NewFromFloat(100.0),
        Qty:   decimal.NewFromFloat(1.0),
        Time:  time.Now().Unix(),
    }
    ob.Match(order, tradeCh, fillCh, order.Qty)

    // Should receive one fill event with status "NEW"
    select {
    case fill := <-fillCh:
        if fill.Status != New {
            t.Errorf("Expected status NEW, got %s", fill.Status)
        }
        if fill.OrderID != "test1" {
            t.Errorf("Expected order ID 'test1', got %s", fill.OrderID)
        }
        if !fill.ExecutedQty.IsZero() {
            t.Errorf("Expected executed quantity 0, got %s", fill.ExecutedQty.String())
        }
    default:
        t.Error("Expected a fill event to be generated")
    }
}

// TestOrderFillAtTopPriceSell tests that SELL orders are filled at the top BUY price, not SELL order price
func TestOrderFillAtTopPriceSell(t *testing.T) {
    ob := NewOrderBook("BTC-USDT")
    tradeCh := make(chan Trade, 10)
    fillCh := make(chan OrderFill, 10)

    // Add a buy order at the top price
    buyOrder := Order{
        ID:    "BUY-1",
        Side:  Buy,
        Price: decimal.NewFromFloat(2),
        Qty:   decimal.NewFromFloat(1),
        Time:  time.Now().Unix(),
    }

    ob.Match(buyOrder, tradeCh, fillCh, buyOrder.Qty)

    // Skip the NEW fill event for BUY-1
    <-fillCh

    sellOrder := Order{
        ID:    "SELL-1",
        Side:  Sell,
        Price: decimal.NewFromFloat(1),
        Qty:   decimal.NewFromFloat(1),
        Time:  time.Now().Unix(),
    }

    ob.Match(sellOrder, tradeCh, fillCh, sellOrder.Qty)

    for i := 0; i < 2; i++ {
        select {
        case fill := <-fillCh:
            t.Logf("Received fill: %+v", fill)
            if fill.OrderID == "BUY-1" {
                if !fill.ExecutedQty.Equal(decimal.NewFromFloat(1)) {
                    t.Errorf("Expected 'BUY-1' executed quantity 1, got %s", fill.ExecutedQty.String())
                }
                if !fill.Price.Equal(decimal.NewFromFloat(2)) {
                    t.Errorf("Expected 'BUY-1' fill price 2, got %s", fill.Price.String())
                }

                if fill.Status != Filled {
                    t.Errorf("Expected 'BUY-1' status Filled, got %s", fill.Status)
                }
            }

            if fill.OrderID == "SELL-1" {
                if !fill.ExecutedQty.Equal(decimal.NewFromFloat(1)) {
                    t.Errorf("Expected 'SELL-1' executed quantity 1, got %s", fill.ExecutedQty.String())
                }
                if !fill.Price.Equal(decimal.NewFromFloat(2)) {
                    t.Errorf("Expected 'SELL-1' fill price 2, got %s", fill.Price.String())
                }

                if fill.Status != Filled {
                    t.Errorf("Expected 'SELL-1' status Filled, got %s", fill.Status)
                }
            }

            if fill.Status != Filled {
                t.Errorf("Expected status Filled, got %s", fill.Status)
            }
        default:
            time.Sleep(100 * time.Millisecond)
        }
    }
}

// TestOrderFillAtTopPriceBuy tests that BUY orders are filled at the BOTTOM SELL price, not BUY order price
func TestOrderFillAtTopPriceBuy(t *testing.T) {
    ob := NewOrderBook("BTC-USDT")
    tradeCh := make(chan Trade, 10)
    fillCh := make(chan OrderFill, 10)

    sellOrder := Order{
        ID:    "SELL-1",
        Side:  Sell,
        Price: decimal.NewFromFloat(1),
        Qty:   decimal.NewFromFloat(1),
        Time:  time.Now().Unix(),
    }

    ob.Match(sellOrder, tradeCh, fillCh, sellOrder.Qty)

    // Skip the NEW fill event for SELL-1
    <-fillCh

    buyOrder := Order{
        ID:    "BUY-1",
        Side:  Buy,
        Price: decimal.NewFromFloat(2),
        Qty:   decimal.NewFromFloat(1),
        Time:  time.Now().Unix(),
    }

    ob.Match(buyOrder, tradeCh, fillCh, buyOrder.Qty)

    for i := 0; i < 2; i++ {
        select {
        case fill := <-fillCh:
            t.Logf("Received fill: %+v", fill)
            if fill.OrderID == "BUY-1" {
                if !fill.ExecutedQty.Equal(decimal.NewFromFloat(1)) {
                    t.Errorf("Expected 'BUY-1' executed quantity 1, got %s", fill.ExecutedQty.String())
                }
                if !fill.Price.Equal(decimal.NewFromFloat(1)) {
                    t.Errorf("Expected 'BUY-1' fill price 2, got %s", fill.Price.String())
                }

                if fill.Status != Filled {
                    t.Errorf("Expected 'BUY-1' status Filled, got %s", fill.Status)
                }
            }

            if fill.OrderID == "SELL-1" {
                if !fill.ExecutedQty.Equal(decimal.NewFromFloat(1)) {
                    t.Errorf("Expected 'SELL-1' executed quantity 1, got %s", fill.ExecutedQty.String())
                }
                if !fill.Price.Equal(decimal.NewFromFloat(1)) {
                    t.Errorf("Expected 'SELL-1' fill price 2, got %s", fill.Price.String())
                }

                if fill.Status != Filled {
                    t.Errorf("Expected 'SELL-1' status Filled, got %s", fill.Status)
                }
            }

            if fill.Status != Filled {
                t.Errorf("Expected status Filled, got %s", fill.Status)
            }
        default:
            time.Sleep(100 * time.Millisecond)
        }
    }
}
