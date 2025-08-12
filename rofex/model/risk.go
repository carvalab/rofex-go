package model

// Modelos tipados para Primary Risk API según docs/primary-api.md
// Endpoints:
// - GET /rest/risk/position/getPositions/{accountName}
// - GET /rest/risk/detailedPosition/{accountName}
// - GET /rest/risk/accountReport/{accountName}

// PositionInstrument identifica el instrumento dentro de una posición de cuenta.
type PositionInstrument struct {
	SymbolReference string `json:"symbolReference"`
	SettlType       *int   `json:"settlType,omitempty"`
}

// Position representa una posición consolidada por instrumento.
type Position struct {
	Instrument        PositionInstrument `json:"instrument"`
	Symbol            string             `json:"symbol"`
	BuySize           float64            `json:"buySize"`
	BuyPrice          float64            `json:"buyPrice"`
	SellSize          float64            `json:"sellSize"`
	SellPrice         float64            `json:"sellPrice"`
	TotalDailyDiff    float64            `json:"totalDailyDiff"`
	TotalDiff         float64            `json:"totalDiff"`
	TradingSymbol     string             `json:"tradingSymbol"`
	OriginalBuyPrice  float64            `json:"originalBuyPrice"`
	OriginalSellPrice float64            `json:"originalSellPrice"`
}

// ----- Detailed Position -----

// DetailedDailyDiff detalla diferencias diarias de precios y montos.
type DetailedDailyDiff struct {
	BuyPricePPPDiff     float64 `json:"buyPricePPPDiff"`
	SellPricePPPDiff    float64 `json:"sellPricePPPDiff"`
	TotalDailyDiff      float64 `json:"totalDailyDiff"`
	BuyDailyDiff        float64 `json:"buyDailyDiff"`
	SellDailyDiff       float64 `json:"sellDailyDiff"`
	TotalDailyDiffPlain float64 `json:"totalDailyDiffPlain"`
	BuyDailyDiffPlain   float64 `json:"buyDailyDiffPlain"`
	SellDailyDiffPlain  float64 `json:"sellDailyDiffPlain"`
}

// DetailedPositionItem representa el detalle por símbolo (size, precios, etc.).
type DetailedPositionItem struct {
	SymbolReference       string  `json:"symbolReference"`
	ContractType          string  `json:"contractType"`
	PriceConversionFactor float64 `json:"priceConversionFactor"`
	ContractSize          float64 `json:"contractSize"`
	MarketPrice           float64 `json:"marketPrice"`
	Currency              string  `json:"currency"`
	ExchangeRate          float64 `json:"exchangeRate"`
	ContractMultiplier    float64 `json:"contractMultiplier"`

	TotalInitialSize float64 `json:"totalInitialSize"`
	BuyInitialSize   float64 `json:"buyInitialSize"`
	SellInitialSize  float64 `json:"sellInitialSize"`
	BuyInitialPrice  float64 `json:"buyInitialPrice"`
	SellInitialPrice float64 `json:"sellInitialPrice"`

	TotalFilledSize float64 `json:"totalFilledSize"`
	BuyFilledSize   float64 `json:"buyFilledSize"`
	SellFilledSize  float64 `json:"sellFilledSize"`
	BuyFilledPrice  float64 `json:"buyFilledPrice"`
	SellFilledPrice float64 `json:"sellFilledPrice"`

	TotalCurrentSize float64 `json:"totalCurrentSize"`
	BuyCurrentSize   float64 `json:"buyCurrentSize"`
	SellCurrentSize  float64 `json:"sellCurrentSize"`

	DetailedDailyDiff DetailedDailyDiff `json:"detailedDailyDiff"`
}

// DetailedInstrument agrupa los ítems detallados por instrumento y tamaños agregados.
type DetailedInstrument struct {
	DetailedPositions     []DetailedPositionItem `json:"detailedPositions"`
	InstrumentInitialSize float64                `json:"instrumentInitialSize"`
	InstrumentFilledSize  float64                `json:"instrumentFilledSize"`
	InstrumentCurrentSize float64                `json:"instrumentCurrentSize"`
}

// DetailedPosition representa el payload de detailedPosition.
type DetailedPosition struct {
	Account             string                                   `json:"account"`
	TotalDailyDiffPlain float64                                  `json:"totalDailyDiffPlain"`
	TotalMarketValue    float64                                  `json:"totalMarketValue"`
	Report              map[string]map[string]DetailedInstrument `json:"report"`
	LastCalculation     int64                                    `json:"lastCalculation"`
}

// ----- Account Report -----

// CurrencyAmount representa un monto por moneda con consumido y disponible.
type CurrencyAmount struct {
	Consumed  float64 `json:"consumed"`
	Available float64 `json:"available"`
}

// CurrencyBalance detalla saldos por moneda.
type CurrencyBalance struct {
	DetailedCurrencyBalance map[string]CurrencyAmount `json:"detailedCurrencyBalance"`
}

// Cash detalla el efectivo total y por moneda.
type Cash struct {
	TotalCash    float64            `json:"totalCash"`
	DetailedCash map[string]float64 `json:"detailedCash"`
}

// AvailableToOperate agrupa montos disponibles para operar.
type AvailableToOperate struct {
	Cash             Cash     `json:"cash"`
	Movements        float64  `json:"movements"`
	Credit           *float64 `json:"credit"`
	Total            float64  `json:"total"`
	PendingMovements float64  `json:"pendingMovements"`
}

// DetailedAccountReport compone saldos y disponibles por fecha de liquidación.
type DetailedAccountReport struct {
	CurrencyBalance    CurrencyBalance    `json:"currencyBalance"`
	AvailableToOperate AvailableToOperate `json:"availableToOperate"`
	SettlementDate     int64              `json:"settlementDate"`
}

// AccountData es el reporte de cuenta principal.
type AccountData struct {
	AccountName            string                           `json:"accountName"`
	MarketMember           string                           `json:"marketMember"`
	MarketMemberIdentity   string                           `json:"marketMemberIdentity"`
	Collateral             float64                          `json:"collateral"`
	Margin                 float64                          `json:"margin"`
	AvailableToCollateral  float64                          `json:"availableToCollateral"`
	DetailedAccountReports map[string]DetailedAccountReport `json:"detailedAccountReports"`

	HasError        bool    `json:"hasError"`
	ErrorDetail     any     `json:"errorDetail"`
	LastCalculation int64   `json:"lastCalculation"`
	Portfolio       float64 `json:"portfolio"`
	OrdersMargin    float64 `json:"ordersMargin"`
	CurrentCash     float64 `json:"currentCash"`
	DailyDiff       float64 `json:"dailyDiff"`
	UncoveredMargin float64 `json:"uncoveredMargin"`
}
