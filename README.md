# Orderbook Engine

A high-performance order matching engine with real-time market data streaming capabilities. This engine implements a price-time priority matching algorithm using heap-based data structures for efficient order management and execution.

## Features

- **Multi-asset order book management** - Support for multiple trading pairs
- **Real-time trade execution** - Price-time priority matching algorithm
- **Live market data streaming** - Continuous price updates and depth streaming
- **Order fill tracking** - Comprehensive order status updates
- **Trade statistics** - Real-time analytics and trade metrics
- **Thread-safe operations** - Concurrent access support across multiple trading pairs
- **High-performance matching** - Heap-based data structures for O(log n) operations

## Architecture

The engine consists of several key components:

### Core Components

- **Engine** - Main coordinator managing multiple order books and data streams
- **OrderBook** - Individual trading pair order books with bid/ask heaps
- **Order Matching** - Price-time priority matching with automatic execution
- **Data Streaming** - Real-time feeds for trades, prices, and market depth

### Data Structures

- **Orders** - Individual buy/sell orders with price, quantity, and timing
- **Trades** - Executed matches between buy and sell orders
- **Fills** - Order execution status updates and partial fill tracking
- **Market Data** - Price updates and order book depth snapshots

## Installation

```bash
go mod tidy
```

## Usage

### Basic Engine Setup

```go
package main

import (
    "fmt"
    "time"
    "github.com/shopspring/decimal"
    "orderbook/engine"
)

func main() {
    // Create new engine
    e := engine.NewEngine()
    
    // Start real-time data streams
    e.StartPriceBroadcaster()
    e.StartDepthStreamer(10)
    
    // Listen for trades
    go func() {
        for trade := range e.TradeStream {
            fmt.Printf("Trade: %+v\n", trade)
        }
    }()
    
    // Listen for fills
    go func() {
        for fill := range e.FillStream {
            fmt.Printf("Fill: %+v\n", fill)
        }
    }()
    
    // Add orders
    buyOrder := engine.Order{
        ID:    "buy1",
        Side:  engine.Buy,
        Price: decimal.NewFromFloat(50000),
        Qty:   decimal.NewFromFloat(0.1),
        Time:  time.Now().Unix(),
    }
    
    sellOrder := engine.Order{
        ID:    "sell1",
        Side:  engine.Sell,
        Price: decimal.NewFromFloat(50000),
        Qty:   decimal.NewFromFloat(0.1),
        Time:  time.Now().Unix(),
    }
    
    // Process orders (will generate trade since prices match)
    e.AddOrder("BTC-USD", sellOrder)
    e.AddOrder("BTC-USD", buyOrder)
    
    // Get market depth
    depth := e.GetOrderBookDepth("BTC-USD", 5)
    if depth != nil {
        fmt.Printf("Market Depth: %+v\n", depth)
    }
}
```

### Channel-based Event Handling

The engine provides four main event channels:

```go
// Trade events - when orders are matched
for trade := range engine.TradeStream {
    // Handle trade execution
}

// Fill events - order status updates
for fill := range engine.FillStream {
    // Handle order fills and status changes
}

// Price updates - best bid/ask and average prices
for update := range engine.PriceUpdates {
    // Handle price movements
}

// Depth updates - order book snapshots
for depth := range engine.DepthUpdates {
    // Handle market depth changes
}
```

## Testing

The project includes comprehensive test coverage for all major components:

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./engine

# Run specific test
go test -v ./engine -run TestNewEngine

# Run tests with coverage
go test -cover ./engine
```

### Test Categories

1. **Unit Tests** - Individual component functionality
2. **Integration Tests** - Order matching and trade generation
3. **Concurrency Tests** - Thread safety and concurrent operations
4. **Performance Tests** - High-load scenarios and timing
5. **Edge Case Tests** - Boundary conditions and error handling

### Test Coverage

The test suite covers:

- ✅ Engine initialization and configuration
- ✅ Order book creation and management
- ✅ Order processing and matching
- ✅ Trade generation and statistics
- ✅ Fill event generation
- ✅ Price update broadcasting
- ✅ Depth streaming
- ✅ Concurrent order processing
- ✅ Trade ID generation
- ✅ Market data retrieval
- ✅ Multi-pair trading
- ✅ Thread safety

## Documentation

### Godoc Documentation

The project includes comprehensive godoc documentation for all public APIs.

#### Viewing Documentation

1. **Start the godoc server:**
   ```bash
   godoc -http=:6060
   ```

2. **Access in browser:**
   Open http://localhost:6060/pkg/orderbook/engine/

3. **Command line documentation:**
   ```bash
   # View package documentation
   go doc ./engine
   
   # View all documentation
   go doc -all ./engine
   
   # View specific type/function
   go doc ./engine.Engine
   go doc ./engine.NewEngine
   ```

#### Documentation Features

- **Package Overview** - High-level architecture and usage
- **Type Documentation** - Detailed struct and interface docs
- **Method Documentation** - Function parameters, returns, and behavior
- **Examples** - Code samples and usage patterns
- **Cross-references** - Links between related types and methods

### API Reference

#### Core Types

- **Engine** - Main trading engine coordinator
- **OrderBook** - Individual trading pair order book
- **Order** - Trading order representation
- **Trade** - Executed trade between orders
- **OrderFill** - Order execution status and details
- **PriceUpdate** - Market price information
- **DepthUpdate** - Order book depth snapshot

#### Key Methods

- `NewEngine()` - Create new engine instance
- `AddOrder(pair, order)` - Process new trading order
- `StartPriceBroadcaster()` - Begin price update streaming
- `StartDepthStreamer(depth)` - Begin depth update streaming
- `GetOrderBookDepth(pair, depth)` - Get current market depth
- `GetNextTradeID()` - Generate unique trade identifier

## Performance

The engine is designed for high-performance trading applications:

- **O(log n)** order insertion and matching via heaps
- **Concurrent processing** across multiple trading pairs
- **Asynchronous event processing** for low-latency operations
- **Efficient memory usage** with pooled data structures
- **Configurable channel capacities** for different load scenarios

### Benchmarks

```bash
# Run performance benchmarks
go test -bench=. ./engine

# Run with memory profiling
go test -bench=. -memprofile=mem.prof ./engine

# Run with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./engine
```

## Examples

See the `examples/` directory for complete usage examples:

- **Basic Trading** - Simple buy/sell order processing
- **Market Making** - Continuous bid/ask order placement
- **Data Streaming** - Real-time market data consumption
- **Multi-pair Trading** - Managing multiple trading pairs
- **High Frequency** - Low-latency order processing

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Update documentation
6. Submit a pull request

## License

[License information here]

## Architecture Diagram

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Trading       │    │   Order Books   │    │   Data Streams  │
│   Engine        │────│                 │────│                 │
│                 │    │  BTC-USD        │    │  Trades         │
│ ┌─────────────┐ │    │  ETH-USD        │    │  Fills          │
│ │  Order      │ │    │  ...            │    │  Prices         │
│ │  Processing │ │    │                 │    │  Depth          │
│ └─────────────┘ │    │ ┌─────────────┐ │    │                 │
│                 │    │ │  Bid Heap   │ │    └─────────────────┘
│ ┌─────────────┐ │    │ │  Ask Heap   │ │    
│ │  Statistics │ │    │ │  Matching   │ │    ┌─────────────────┐
│ │  Tracking   │ │    │ └─────────────┘ │    │   External      │
│ └─────────────┘ │    └─────────────────┘    │   Consumers     │
└─────────────────┘                           │                 │
                                              │  Price Feeds    │
                                              │  Risk Mgmt      │
                                              │  Settlement     │
                                              └─────────────────┘
``` 