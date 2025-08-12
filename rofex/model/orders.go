package model

// SendOrderResponse coincide con APIDoc para /rest/order/newSingleOrder
// { "status":"OK", "order": { "clientId": "...", "proprietary": "api" } }
type SendOrderResponse struct {
	Status string `json:"status"`
	Order  struct {
		ClientID    string `json:"clientId"`
		Proprietary string `json:"proprietary"`
	} `json:"order"`
}

// CancelOrderResponse coincide con APIDoc para /rest/order/cancelById
type CancelOrderResponse struct {
	Status string `json:"status"`
	Order  struct {
		ClientID    string `json:"clientId"`
		Proprietary string `json:"proprietary"`
	} `json:"order"`
}

// ReplaceOrderResponse coincide con APIDoc para /rest/order/replaceById
type ReplaceOrderResponse struct {
	Status string `json:"status"`
	Order  struct {
		ClientID    string `json:"clientId"`
		Proprietary string `json:"proprietary"`
	} `json:"order"`
}

// Order representa un objeto de estado de orden devuelto por la API.
// Los campos son conservadores y pueden no incluir todos los campos posibles. Campos desconocidos son ignorados.
type Order struct {
	OrderID      string       `json:"orderId,omitempty"`
	InstrumentID InstrumentID `json:"instrumentId"`
	ClOrdID      string       `json:"clOrdId"`
	Proprietary  string       `json:"proprietary"`
	ExecID       string       `json:"execId,omitempty"`
	AccountID    *struct {
		ID string `json:"id"`
	} `json:"accountId,omitempty"`
	Status       string      `json:"status,omitempty"`
	Text         string      `json:"text,omitempty"`
	Side         Side        `json:"side,omitempty"`
	OrdType      OrderType   `json:"ordType,omitempty"`
	TimeInForce  TimeInForce `json:"timeInForce,omitempty"`
	Price        *float64    `json:"price,omitempty"`
	OrderQty     *int64      `json:"orderQty,omitempty"`
	LeavesQty    *int64      `json:"leavesQty,omitempty"`
	CumQty       *int64      `json:"cumQty,omitempty"`
	AvgPx        *float64    `json:"avgPx,omitempty"`
	LastPx       *float64    `json:"lastPx,omitempty"`
	LastQty      *int64      `json:"lastQty,omitempty"`
	TransactTime string      `json:"transactTime,omitempty"`
}
