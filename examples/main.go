package main

import (
	"fmt"
	"time"

	"github.com/mkhoshkam/orderbook/engine"

	"github.com/shopspring/decimal"
)

func main() {
	e := engine.NewEngine()

	e.StartPriceBroadcaster()
	e.StartDepthStreamer(5)

	go func() {
		for trade := range e.TradeStream {
			fmt.Printf("[TRADE] %s %.4f @ %.2f\n",
				trade.Pair,
				trade.Qty.InexactFloat64(),
				trade.Price.InexactFloat64())
		}
	}()

	go func() {
		for price := range e.PriceUpdates {
			fmt.Printf("[PRICE] %s BID: %.2f / ASK: %.2f / AVG: %.2f\n",
				price.Pair,
				price.BestBid.InexactFloat64(),
				price.BestAsk.InexactFloat64(),
				price.AvgPrice.InexactFloat64())
		}
	}()

	// Listen to order fill events
	go func() {
		for fill := range e.FillStream {
			fmt.Printf("[FILL] Order %s (%s %s) - Status: %s | Executed: %.4f | Remaining: %.4f | Fill Price: %.2f\n",
				fill.OrderID,
				fill.Side,
				fill.Pair,
				fill.Status,
				fill.ExecutedQty.InexactFloat64(),
				fill.RemainingQty.InexactFloat64(),
				fill.FillPrice.InexactFloat64())
		}
	}()

	go func() {
		for depth := range e.DepthUpdates {
			fmt.Printf("\n[DEPTH] %s (Trades: %d)\n", depth.Pair, depth.TradeCount)

			fmt.Println("BIDS:")
			for i, bid := range depth.Bids {
				fmt.Printf("  %d. %.2f | %.4f (%d orders)\n",
					i+1,
					bid.Price.InexactFloat64(),
					bid.Quantity.InexactFloat64(),
					bid.TradeCount)
			}

			fmt.Println("ASKS:")
			for i, ask := range depth.Asks {
				fmt.Printf("  %d. %.2f | %.4f (%d orders)\n",
					i+1,
					ask.Price.InexactFloat64(),
					ask.Quantity.InexactFloat64(),
					ask.TradeCount)
			}
			fmt.Println("---")
		}
	}()

	fmt.Println("Adding orders...")

	e.AddOrder("BTC/USDT", engine.Order{
		ID:    "BUY-1",
		Side:  engine.Buy,
		Price: decimal.NewFromFloat(30000),
		Qty:   decimal.NewFromFloat(0.5),
		Time:  time.Now().Unix(),
	})

	e.AddOrder("BTC/USDT", engine.Order{
		ID:    "BUY-2",
		Side:  engine.Buy,
		Price: decimal.NewFromFloat(29900),
		Qty:   decimal.NewFromFloat(0.3),
		Time:  time.Now().Unix(),
	})

	e.AddOrder("BTC/USDT", engine.Order{
		ID:    "BUY-3",
		Side:  engine.Buy,
		Price: decimal.NewFromFloat(29900),
		Qty:   decimal.NewFromFloat(0.2),
		Time:  time.Now().Unix(),
	})

	e.AddOrder("BTC/USDT", engine.Order{
		ID:    "SELL-1",
		Side:  engine.Sell,
		Price: decimal.NewFromFloat(30100),
		Qty:   decimal.NewFromFloat(0.4),
		Time:  time.Now().Unix(),
	})

	e.AddOrder("BTC/USDT", engine.Order{
		ID:    "SELL-2",
		Side:  engine.Sell,
		Price: decimal.NewFromFloat(30200),
		Qty:   decimal.NewFromFloat(0.6),
		Time:  time.Now().Unix(),
	})

	time.Sleep(2 * time.Second)
	fmt.Println("\n=== Adding market buy order (should trigger fills) ===")

	// This order will match with SELL-1 (0.4) and partially with SELL-2
	e.AddOrder("BTC/USDT", engine.Order{
		ID:    "BUY-MARKET",
		Side:  engine.Buy,
		Price: decimal.NewFromFloat(30300), // High price to trigger market execution
		Qty:   decimal.NewFromFloat(0.8),   // More than SELL-1, will trigger partial fill of SELL-2
		Time:  time.Now().Unix(),
	})

	time.Sleep(2 * time.Second)
	fmt.Println("\n=== Adding large sell order (should partially fill) ===")

	// This will match with remaining buy orders but won't fill completely
	e.AddOrder("BTC/USDT", engine.Order{
		ID:    "SELL-LARGE",
		Side:  engine.Sell,
		Price: decimal.NewFromFloat(29000), // Low price to trigger market execution
		Qty:   decimal.NewFromFloat(2.0),   // Large quantity, will be partially filled
		Time:  time.Now().Unix(),
	})

	time.Sleep(3 * time.Second)
	fmt.Println("\n=== Adding ETH orders ===")

	e.AddOrder("ETH/USDT", engine.Order{
		ID:    "ETH-BUY-1",
		Side:  engine.Buy,
		Price: decimal.NewFromFloat(2000),
		Qty:   decimal.NewFromFloat(1.0),
		Time:  time.Now().Unix(),
	})

	e.AddOrder("ETH/USDT", engine.Order{
		ID:    "ETH-SELL-1",
		Side:  engine.Sell,
		Price: decimal.NewFromFloat(2010),
		Qty:   decimal.NewFromFloat(0.8),
		Time:  time.Now().Unix(),
	})

	select {}
}
