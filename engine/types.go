package engine

import "github.com/shopspring/decimal"

type Side string

const (
	Buy  Side = "buy"
	Sell Side = "sell"
)

type Order struct {
	ID    string
	Side  Side
	Price decimal.Decimal
	Qty   decimal.Decimal
	Time  int64
}

type Trade struct {
	Pair        string
	BuyOrderID  string
	SellOrderID string
	Price       decimal.Decimal
	Qty         decimal.Decimal
}

type PriceUpdate struct {
	Pair     string
	BestBid  decimal.Decimal
	BestAsk  decimal.Decimal
	AvgPrice decimal.Decimal
}

type DepthLevel struct {
	Price      decimal.Decimal
	Quantity   decimal.Decimal
	TradeCount int
}

type DepthUpdate struct {
	Pair       string
	Bids       []DepthLevel
	Asks       []DepthLevel
	Timestamp  int64
	TradeCount int64
}

type FillStatus string

const (
	PartiallyFilled FillStatus = "PARTIALLY_FILLED"
	Filled          FillStatus = "FILLED"
	New             FillStatus = "NEW"
)

type OrderFill struct {
	OrderID      string
	Pair         string
	Side         Side
	OriginalQty  decimal.Decimal
	ExecutedQty  decimal.Decimal
	RemainingQty decimal.Decimal
	Price        decimal.Decimal
	FillPrice    decimal.Decimal
	Status       FillStatus
	Timestamp    int64
}
