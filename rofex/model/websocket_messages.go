package model

import (
	"time"
)

// MarketDataEvent representa los datos de mercado recibidos via WebSocket.
//
// Según primary-api.md: "Mensaje de Market Data"
//
// Ejemplo de mensaje recibido:
//
//	{
//	  "type": "md",
//	  "instrumentId": {
//	    "marketId": "ROFX",
//	    "symbol": "DLR/DIC23"
//	  },
//	  "marketData": {
//	    "OF": [
//	      {"price": 189, "size": 21},
//	      {"price": 188, "size": 13}
//	    ]
//	  },
//	  "timestamp": 1669995044232
//	}
type MarketDataEvent struct {
	// Type indica el tipo de mensaje WebSocket ("md" para market data)
	Type WSMessageType `json:"type"`

	// InstrumentID identifica el instrumento del cual se recibe la data
	InstrumentID InstrumentID `json:"instrumentId"`

	// MarketData contiene los datos del mercado solicitados via "entries".
	// Claves según Primary API: BI, OF, LA, OP, CL, SE, HI, LO, TV, OI, IV, EV, NV, ACP, TC.
	MarketData MarketData `json:"marketData"`

	// Timestamp marca de tiempo del mensaje en milisegundos Unix
	Timestamp *int64 `json:"timestamp,omitempty"`
}

// HumanTime convierte el timestamp Unix a time.Time legible
func (d *MarketDataEvent) HumanTime() *time.Time {
	if d.Timestamp == nil {
		return nil
	}
	t := time.Unix(*d.Timestamp/1000, (*d.Timestamp%1000)*1000000)
	return &t
}

// OrderReportEvent representa un Execution Report recibido via WebSocket.
//
// Según primary-api.md: "Mensaje para Execution Reports"
//
// Ejemplo de mensaje recibido:
//
//	{
//	  "type": "or",
//	  "timestamp": 1537212212623,
//	  "orderReport": {
//	    "orderId": "1128056",
//	    "clOrdId": "user14545967430231",
//	    "proprietary": "PBCP",
//	    "execId": "160127155448-fix1-1368",
//	    "accountId": {"id": "30"},
//	    "instrumentId": {"marketId": "ROFX", "symbol": "DLR/DIC23"},
//	    "price": 189,
//	    "orderQty": 10,
//	    "ordType": "LIMIT",
//	    "side": "BUY",
//	    "timeInForce": "DAY",
//	    "transactTime": "20230204-11:41:54",
//	    "avgPx": 0,
//	    "lastPx": 0,
//	    "lastQty": 0,
//	    "cumQty": 0,
//	    "leavesQty": 10,
//	    "status": "CANCELLED",
//	    "text": "Reemplazada"
//	  }
//	}
type OrderReportEvent struct {
	// Type indica el tipo de mensaje WebSocket ("or" para order report)
	Type WSMessageType `json:"type"`

	// Timestamp marca de tiempo del mensaje en milisegundos Unix
	Timestamp *int64 `json:"timestamp,omitempty"`

	// OrderReport contiene los detalles del reporte de ejecución
	OrderReport OrderDetails `json:"orderReport"`
}

// HumanTime convierte el timestamp Unix a time.Time legible
func (r *OrderReportEvent) HumanTime() *time.Time {
	if r.Timestamp == nil {
		return nil
	}
	t := time.Unix(*r.Timestamp/1000, (*r.Timestamp%1000)*1000000)
	return &t
}

// OrderDetails representa los detalles de una orden en un Execution Report.
//
// Según primary-api.md: campos del orderReport
type OrderDetails struct {
	// OrderID identificador único de la orden en el mercado
	OrderID *string `json:"orderId,omitempty"`

	// ClOrdID identificador del request de la orden
	ClOrdID string `json:"clOrdId"`

	// Proprietary usuario FIX que envió la orden
	Proprietary string `json:"proprietary"`

	// ExecID identificador de ejecución
	ExecID *string `json:"execId,omitempty"`

	// AccountID cuenta asociada a la orden
	AccountID *AccountReference `json:"accountId,omitempty"`

	// InstrumentID instrumento de la orden
	InstrumentID InstrumentID `json:"instrumentId"`

	// Price precio de la orden
	Price *float64 `json:"price,omitempty"`

	// OrderQty cantidad de la orden
	OrderQty int `json:"orderQty"`

	// OrdType tipo de orden (LIMIT, MARKET, etc.)
	OrdType string `json:"ordType"`

	// Side lado de la orden (BUY, SELL)
	Side string `json:"side"`

	// TimeInForce tiempo de vida de la orden (DAY, IOC, FOK, GTD)
	TimeInForce string `json:"timeInForce"`

	// TransactTime fecha y hora de la transacción
	TransactTime string `json:"transactTime"`

	// AvgPx precio promedio operado
	AvgPx *float64 `json:"avgPx,omitempty"`

	// LastPx último precio operado
	LastPx *float64 `json:"lastPx,omitempty"`

	// LastQty última cantidad operada
	LastQty *int `json:"lastQty,omitempty"`

	// CumQty cantidad total operada
	CumQty *int `json:"cumQty,omitempty"`

	// LeavesQty cantidad remanente de la orden
	LeavesQty *int `json:"leavesQty,omitempty"`

	// Status estado de la orden (NEW, FILLED, CANCELLED, etc.)
	Status string `json:"status"`

	// Text descripción del estado
	Text *string `json:"text,omitempty"`

	// WSClOrdID identificador de orden enviada via WebSocket (opcional)
	WSClOrdID *string `json:"wsClOrdId,omitempty"`
}

// AccountReference representa una referencia a una cuenta
type AccountReference struct {
	// ID identificador de la cuenta
	ID string `json:"id"`
}
