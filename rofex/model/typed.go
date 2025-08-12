package model

// Generic typed response containers using RawMessage until full schemas are defined.

type InstrumentsResponse struct {
	Status      string       `json:"status,omitempty"`
	Instruments []Instrument `json:"instruments"`
}

type InstrumentDetailResponse struct {
	Status     string     `json:"status,omitempty"`
	Instrument Instrument `json:"instrument"`
}

type MarketDataSnapshotResponse struct {
	Status     string     `json:"status"`
	MarketData MarketData `json:"marketData"`
	Depth      int        `json:"depth"`
	Aggregated bool       `json:"aggregated"`
}

type TradesResponse struct {
	Status      string `json:"status"`
	Symbol      string `json:"symbol"`
	Market      string `json:"market"`
	Description string `json:"description,omitempty"`
	Message     string `json:"message,omitempty"`
	Trades      []struct {
		Price      float64 `json:"price"`
		Size       float64 `json:"size"` // puede venir como 3.0; usar float64 para tolerar ambos
		Datetime   string  `json:"datetime"`
		ServerTime int64   `json:"servertime"`
		Symbol     string  `json:"symbol"`
	} `json:"trades"`
}

type OrderStatusResponse struct {
	Status string `json:"status"`
	Order  Order  `json:"order"`
}

type AllOrdersStatusResponse struct {
	Status string  `json:"status"`
	Orders []Order `json:"orders"`
}

type AccountPositionResponse struct {
	Status    string     `json:"status"`
	Positions []Position `json:"positions"`
}

type DetailedPositionResponse struct {
	Status           string           `json:"status"`
	DetailedPosition DetailedPosition `json:"detailedPosition"`
}

type AccountReportResponse struct {
	Status      string      `json:"status"`
	AccountData AccountData `json:"accountData"`
}
