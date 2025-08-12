package model

// InstrumentID identifies an instrument in a specific market.
type InstrumentID struct {
	Symbol   string `json:"symbol"`
	MarketID string `json:"marketId"`
}
