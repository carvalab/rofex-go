package model

import (
	"encoding/json"
)

// Segment describe el segmento del instrumento (marketSegmentId, marketId).
type Segment struct {
	MarketSegmentID string `json:"marketSegmentId"`
	MarketID        string `json:"marketId"`
}

// TickRange representa un rango de ticks de precio dinámico.
type TickRange struct {
	LowerLimit *float64 `json:"lowerLimit"`
	UpperLimit *float64 `json:"upperLimit"`
	Tick       float64  `json:"tick"`
}

// Instrument representa un instrumento negociable según Primary API.
//
// Referencia: docs/primary-api.md - "Instrumentos (Securities)"
// Soporta tanto las respuestas detalladas (con "instrumentId" y más campos)
// como las respuestas simples de listados por CFI o segmento que devuelven
// "marketId" y "symbol" a nivel superior.
type Instrument struct {
	InstrumentID             InstrumentID         `json:"instrumentId"`
	CFICode                  string               `json:"cficode,omitempty"`
	Segment                  *Segment             `json:"segment,omitempty"`
	LowLimitPrice            *float64             `json:"lowLimitPrice,omitempty"`
	HighLimitPrice           *float64             `json:"highLimitPrice,omitempty"`
	MinPriceIncrement        *float64             `json:"minPriceIncrement,omitempty"`
	MinTradeVol              *float64             `json:"minTradeVol,omitempty"`
	MaxTradeVol              *float64             `json:"maxTradeVol,omitempty"`
	TickSize                 *float64             `json:"tickSize,omitempty"`
	ContractMultiplier       *float64             `json:"contractMultiplier,omitempty"`
	RoundLot                 *float64             `json:"roundLot,omitempty"`
	PriceConvertionFactor    *float64             `json:"priceConvertionFactor,omitempty"`
	MaturityDate             *string              `json:"maturityDate,omitempty"`
	Currency                 string               `json:"currency,omitempty"`
	OrderTypes               []string             `json:"orderTypes,omitempty"`
	TimesInForce             []string             `json:"timesInForce,omitempty"`
	SecurityType             *string              `json:"securityType,omitempty"`
	SettlType                *string              `json:"settlType,omitempty"`
	InstrumentPricePrecision *int                 `json:"instrumentPricePrecision,omitempty"`
	InstrumentSizePrecision  *int                 `json:"instrumentSizePrecision,omitempty"`
	SecurityId               *string              `json:"securityId,omitempty"`
	SecurityIdSource         *string              `json:"securityIdSource,omitempty"`
	SecurityDescription      string               `json:"securityDescription,omitempty"`
	TickPriceRanges          map[string]TickRange `json:"tickPriceRanges,omitempty"`

	// Forma alternativa (byCFICode / bySegment): campos a nivel superior
	// que se utilizan para poblar InstrumentID cuando instrumentId no está.
	MarketIDAlt string `json:"marketId,omitempty"`
	SymbolAlt   string `json:"symbol,omitempty"`
}

// UnmarshalJSON permite decodificar tanto la forma detallada (con instrumentId)
// como la forma simple (marketId+symbol a nivel superior).
func (i *Instrument) UnmarshalJSON(b []byte) error {
	// Definimos un alias para decodificar todos los posibles campos
	type Alias Instrument
	var aux Alias
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	// Si no vino instrumentId pero sí marketId/symbol a nivel superior,
	// los usamos para completar InstrumentID.
	if aux.InstrumentID.MarketID == "" && aux.InstrumentID.Symbol == "" {
		if aux.MarketIDAlt != "" || aux.SymbolAlt != "" {
			aux.InstrumentID = InstrumentID{
				MarketID: aux.MarketIDAlt,
				Symbol:   aux.SymbolAlt,
			}
		}
	}
	*i = Instrument(aux)
	return nil
}
