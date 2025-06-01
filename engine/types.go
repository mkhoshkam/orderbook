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
