# Orderbook Engine

[![Go Reference](https://pkg.go.dev/badge/github.com/mkhoshkam/orderbook.svg)](https://pkg.go.dev/github.com/mkhoshkam/orderbook)
[![Go Report Card](https://goreportcard.com/badge/github.com/mkhoshkam/orderbook)](https://goreportcard.com/report/github.com/mkhoshkam/orderbook)

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
go get github.com/mkhoshkam/orderbook@v1.0.0
```

Or to get the latest version:

```bash
go get github.com/mkhoshkam/orderbook
```

## Usage

### Basic Engine Setup

```go
package main

import (
    "fmt"
    "time"
    "github.com/shopspring/decimal"
    "github.com/mkhoshkam/orderbook/engine"
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

## Code Quality

This project follows Go standards for code formatting and quality with automated checks.

### Formatting and Linting

The project uses standard Go tools for maintaining code quality:

- **`gofmt`** - Standard Go code formatting
- **`goimports`** - Import organization and formatting
- **`go vet`** - Static analysis for common issues
- **`golangci-lint`** - Comprehensive linting with multiple analyzers

### Development Commands

Use the provided Makefile for consistent code quality:

```bash
# Format code
make fmt

# Check formatting (without changes)
make fmt-check

# Run linter
make lint

# Run static analysis
make vet

# Run tests
make test

# Run all quality checks
make check

# Install required tools
make install-tools
```

### GitHub Actions Integration

Code quality is automatically enforced via GitHub Actions on every push and pull request:

- ✅ **Code formatting** - Ensures consistent style with `gofmt` and `goimports`
- ✅ **Static analysis** - Checks for common issues with `go vet`
- ✅ **Linting** - Comprehensive code quality checks with `golangci-lint`
- ✅ **Test execution** - Runs full test suite with coverage reporting

### Local Development Setup

1. **Install tools:**
   ```bash
   make install-tools
   ```

2. **Format and check code before committing:**
   ```bash
   make check
   ```

3. **Set up pre-commit hook (optional):**
   ```bash
   echo "make check" > .git/hooks/pre-commit
   chmod +x .git/hooks/pre-commit
   ```

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

## CI/CD

This project uses GitHub Actions for continuous integration and deployment with automated testing and coverage reporting.

### GitHub Actions Workflow

The workflow automatically:

- ✅ **Builds** the project on every push and pull request
- ✅ **Runs all tests** with comprehensive coverage reporting
- ✅ **Generates coverage reports** in HTML format
- ✅ **Updates coverage badges** on the main branch
- ✅ **Posts coverage information** to pull request comments
- ✅ **Uploads coverage to Codecov** (if configured)

### Setting Up Repository Permissions

To enable pull request comments, configure your repository settings:

1. Go to **Settings → Actions → General**
2. Under **Workflow permissions**, select **"Read and write permissions"**
3. Enable **"Allow GitHub Actions to create and approve pull requests"**

### Optional: Codecov Integration

To enable Codecov uploads:

1. Sign up at [codecov.io](https://codecov.io)
2. Add your repository
3. Get your upload token
4. Add it as a repository secret: **Settings → Secrets → Actions → New repository secret**
   - Name: `CODECOV_TOKEN`
   - Value: Your Codecov upload token

### Coverage Reports

Coverage information is available in multiple formats:

- **GitHub Actions Summary** - Detailed coverage breakdown in the Actions tab
- **Pull Request Comments** - Coverage percentage posted to PRs automatically
- **HTML Reports** - Generated as artifacts in the Actions runs
- **Codecov Dashboard** - If Codecov integration is configured

### Workflow Status

Current build status: [![Go](../../actions/workflows/go.yml/badge.svg)](../../actions/workflows/go.yml)

### Manual Testing

You can also run the same commands locally:

```bash
# Build the project
go build -v ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out -covermode=atomic ./...

# Generate HTML coverage report
go tool cover -html=coverage.out -o=coverage.html

# View coverage summary
go tool cover -func=coverage.out
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Update documentation
6. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

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