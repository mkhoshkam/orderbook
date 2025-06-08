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
			fmt.Printf("[PRICE] %s BID: %.4f / ASK: %.4f / AVG: %.4f\n",
				price.Pair,
				price.BestBid.InexactFloat64(),
				price.BestAsk.InexactFloat64(),
				price.AvgPrice.InexactFloat64())
		}
	}()

	// Listen to order fill events
	go func() {
		for fill := range e.FillStream {
			fmt.Printf("[FILL] Order %s (%s %s) - Status: %s | Executed: %.4f | Remaining: %.4f | Price: %.4f | Fill Price: %.4f\n",
				fill.OrderID,
				fill.Side,
				fill.Pair,
				fill.Status,
				fill.ExecutedQty.InexactFloat64(),
				fill.RemainingQty.InexactFloat64(),
				fill.Price.InexactFloat64(),
				fill.FillPrice.InexactFloat64())
		}
	}()

	go func() {
		for depth := range e.DepthUpdates {
			fmt.Printf("\n[DEPTH] %s (Trades: %d)\n", depth.Pair, depth.TradeCount)

			fmt.Println("BIDS")
			for i, bid := range depth.Bids {
				fmt.Printf("  %d. %.4f | %.4f (%d orders)\n",
					i+1,
					bid.Price.InexactFloat64(),
					bid.Quantity.InexactFloat64(),
					bid.TradeCount)
			}
			fmt.Println("---")

			for i, ask := range depth.Asks {
				fmt.Printf("  %d. %.4f | %.4f (%d orders)\n",
					i+1,
					ask.Price.InexactFloat64(),
					ask.Quantity.InexactFloat64(),
					ask.TradeCount)
			}
			fmt.Println("ASKS")
			time.Sleep(2 * time.Second) // Sleep to avoid flooding the console
		}
	}()

	fmt.Println("Adding orders...")

	// Bulk order generation for testing

	// for i := 0; i < 10000; i++ {
	//     time.Sleep(200 * time.Millisecond)
	//     randomNumber := rand.Intn(999) // Random number between 0 and 999
	//     randomPair := randomNumber % 4 // 0 for BTC/USDT, 1 for ETH/USDT
	//
	//     var pair string
	//     switch randomPair {
	//     case 0:
	//         pair = "BTC/USDT"
	//     case 1:
	//         pair = "ETH/USDT"
	//     case 2:
	//         pair = "LTC/USDT"
	//     case 3:
	//         pair = "XRP/USDT"
	//     }
	//
	//     // random orders
	//     side := engine.Buy
	//     if !(rand.Intn(2) == 0) {
	//         side = engine.Sell
	//     }
	//
	//     order := engine.Order{
	//         ID:    fmt.Sprintf("ORDER-%d", i),
	//         Side:  side,
	//         Price: decimal.NewFromFloat(rand.Float64() * 100000),
	//         Qty:   decimal.NewFromFloat(rand.Float64() * 10),
	//         Time:  time.Now().Unix(),
	//     }
	//     e.AddOrder(pair, order)
	//
	// }

	e.AddOrder("BTC/USDT", engine.Order{
		ID:    "BUY-2",
		Side:  engine.Buy,
		Price: decimal.NewFromFloat(2),
		Qty:   decimal.NewFromFloat(1),
		Time:  time.Now().Unix(),
	})

	e.AddOrder("BTC/USDT", engine.Order{
		ID:    "SELL-1",
		Side:  engine.Sell,
		Price: decimal.NewFromFloat(1),
		Qty:   decimal.NewFromFloat(1),
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
