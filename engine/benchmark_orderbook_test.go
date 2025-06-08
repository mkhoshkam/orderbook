package engine

import (
	"fmt"
	"log"
	"math/rand/v2"
	"runtime"
	"runtime/debug"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

var book = NewOrderBook("BTC-USDT")
var fillCount = 0
var tradeCount = 0
var orders = make([]Order, 0, 2000000)

func init() {
	// disable garbage collection for benchmark tests
	debug.SetGCPercent(-1)

	log.Println("Generating random order data for benchmark tests")
	for i := 0; i < 2000000; i++ {
		randomPrice := rand.Float64() * 150000.0
		randomQty := rand.Float64() * 100.0

		side := Buy
		if rand.Int32()%2 == 0 {
			side = Sell
		}

		order := Order{
			ID:    fmt.Sprintf("O%d", i),
			Side:  side,
			Price: decimal.NewFromFloat(randomPrice),
			Qty:   decimal.NewFromFloat(randomQty),
			Time:  time.Now().Unix(),
		}
		orders = append(orders, order)
	}

	// Run garbage collection after generating orders to clean up memory
	runtime.GC()
}

func BenchmarkWithRandomData(benchmark *testing.B) {
	tradeCh := make(chan Trade, 100)
	fillCh := make(chan OrderFill, 100)

	go func() {
		for trade := range tradeCh {
			_ = trade
			tradeCount++
		}
	}()
	go func() {
		for fill := range fillCh {
			_ = fill
			if fill.Status == Filled || fill.Status == PartiallyFilled {
				fillCount++
			}
		}
	}()

	// submit orders to the order book
	for i := 0; i < benchmark.N; i++ {
		book.Match(orders[i], tradeCh, fillCh, orders[i].Qty)
	}

	// Wait for all trades and fills to be processed
	for len(tradeCh) > 0 || len(fillCh) > 0 {
		time.Sleep(100 * time.Millisecond)
	}

	close(tradeCh)
	close(fillCh)

	// Run garbage collection after each benchmark run to clean up memory
	runtime.GC()

	fmt.Printf("Total trades processed: %d, Total orders filled: %d\n", tradeCount, fillCount)
}
