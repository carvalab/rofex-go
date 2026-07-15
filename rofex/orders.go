package rofex

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/carvalab/rofex-go/rofex/model"
)

// NewOrder represents the parameters for sending a new order via REST.
type NewOrder struct {
	Symbol         string
	Market         model.Market
	Side           model.Side
	Type           model.OrderType
	Qty            int64
	Price          *float64
	TIF            model.TimeInForce
	Account        string
	CancelPrevious bool
	Iceberg        bool
	ExpireDate     *string // yyyy-MM-dd for GTD
	DisplayQty     *int64
	// WS-only optional fields
	AllOrNone bool
	WSClOrdID *string
}

func (o NewOrder) validate() error {
	if o.Symbol == "" {
		return fmt.Errorf("validation: symbol: required")
	}
	if o.Qty <= 0 {
		return fmt.Errorf("validation: qty: must be > 0")
	}
	if o.Side == "" {
		return fmt.Errorf("validation: side: required")
	}
	if o.Type == model.OrderTypeLimit && o.Price == nil {
		return fmt.Errorf("validation: price: required for limit")
	}
	if o.TIF == model.GoodTillDate && (o.ExpireDate == nil || *o.ExpireDate == "") {
		return fmt.Errorf("validation: expireDate: required for GTD")
	}
	return nil
}

// SendOrder sends an order to the market via REST (Primary API).
//
// After sending, verify status with OrderStatus() — the order may be rejected.
//
// Reference: docs/primary-api.md "Ingresar una orden"
func (c *Client) SendOrder(ctx context.Context, o NewOrder) (model.SendOrderResponse, error) {
	if err := o.validate(); err != nil {
		return model.SendOrderResponse{}, err
	}
	if o.Market == "" {
		o.Market = model.MarketROFEX
	}
	q := url.Values{
		"marketId":       {string(o.Market)},
		"symbol":         {o.Symbol},
		"orderQty":       {strconv.FormatInt(o.Qty, 10)},
		"ordType":        {string(o.Type)},
		"side":           {string(o.Side)},
		"timeInForce":    {string(o.TIF)},
		"account":        {o.Account},
		"cancelPrevious": {strconv.FormatBool(o.CancelPrevious)},
	}
	if o.Type == model.OrderTypeLimit && o.Price != nil {
		q.Set("price", strconv.FormatFloat(*o.Price, 'f', -1, 64))
	}
	if o.TIF == model.GoodTillDate && o.ExpireDate != nil {
		q.Set("expireDate", *o.ExpireDate)
	}
	if o.Iceberg && o.DisplayQty != nil {
		q.Set("iceberg", "true")
		q.Set("displayQty", strconv.FormatInt(*o.DisplayQty, 10))
	}
	return getTyped[model.SendOrderResponse](ctx, c, pathNewOrder+"?"+q.Encode())
}

// CancelOrder cancels an order via REST (Primary API).
//
// Reference: docs/primary-api.md "Cancelar una orden"
func (c *Client) CancelOrder(ctx context.Context, clientOrderID, proprietary string) (model.CancelOrderResponse, error) {
	if clientOrderID == "" {
		return model.CancelOrderResponse{}, fmt.Errorf("validation: clientOrderID: required")
	}
	if strings.TrimSpace(proprietary) == "" {
		proprietary = c.proprietary
	}
	q := url.Values{"clOrdId": {clientOrderID}, "proprietary": {proprietary}}
	return getTyped[model.CancelOrderResponse](ctx, c, pathCancelOrder+"?"+q.Encode())
}

// ReplaceOrder replaces an existing order (Primary API).
//
// Only non-nil fields are sent.
func (c *Client) ReplaceOrder(ctx context.Context, clOrdID, proprietary string, newQty *int64, newPrice *float64) (model.ReplaceOrderResponse, error) {
	if clOrdID == "" {
		return model.ReplaceOrderResponse{}, fmt.Errorf("validation: clOrdID: required")
	}
	if strings.TrimSpace(proprietary) == "" {
		proprietary = c.proprietary
	}
	q := url.Values{"clOrdId": {clOrdID}, "proprietary": {proprietary}}
	if newQty != nil {
		q.Set("orderQty", strconv.FormatInt(*newQty, 10))
	}
	if newPrice != nil {
		q.Set("price", strconv.FormatFloat(*newPrice, 'f', -1, 64))
	}
	return getTyped[model.ReplaceOrderResponse](ctx, c, pathOrderReplace+"?"+q.Encode())
}

// OrderStatus returns the latest status of an order by Client Order ID.
//
// Reference: docs/primary-api.md "Consultar último estado por Client Order ID"
func (c *Client) OrderStatus(ctx context.Context, clientOrderID, proprietary string) (model.OrderStatusResponse, error) {
	if clientOrderID == "" {
		return model.OrderStatusResponse{}, fmt.Errorf("validation: clientOrderID: required")
	}
	if strings.TrimSpace(proprietary) == "" {
		proprietary = c.proprietary
	}
	q := url.Values{"clOrdId": {clientOrderID}, "proprietary": {proprietary}}
	return getTyped[model.OrderStatusResponse](ctx, c, pathOrderStatus+"?"+q.Encode())
}

// OrderHistoryByClOrdID returns all states an order went through.
//
// Reference: docs/primary-api.md "Consultar todos los estados por Client Order ID"
func (c *Client) OrderHistoryByClOrdID(ctx context.Context, clOrdID, proprietary string) (model.AllOrdersStatusResponse, error) {
	if clOrdID == "" {
		return model.AllOrdersStatusResponse{}, fmt.Errorf("validation: clOrdID: required")
	}
	if strings.TrimSpace(proprietary) == "" {
		proprietary = c.proprietary
	}
	q := url.Values{"clOrdId": {clOrdID}, "proprietary": {proprietary}}
	return getTyped[model.AllOrdersStatusResponse](ctx, c, pathOrderAllByID+"?"+q.Encode())
}

// OrderByOrderID returns the status of an order by its exchange Order ID.
//
// Reference: docs/primary-api.md "Consultar Order por OrderID"
func (c *Client) OrderByOrderID(ctx context.Context, orderID string) (model.OrderStatusResponse, error) {
	if strings.TrimSpace(orderID) == "" {
		return model.OrderStatusResponse{}, fmt.Errorf("validation: orderID: required")
	}
	q := url.Values{"orderId": {orderID}}
	return getTyped[model.OrderStatusResponse](ctx, c, pathOrderByOrder+"?"+q.Encode())
}

// OrderByExecID returns the order associated with an Execution ID.
//
// Reference: docs/primary-api.md "Estado de orden por Execution ID"
func (c *Client) OrderByExecID(ctx context.Context, execID string) (model.OrderStatusResponse, error) {
	if strings.TrimSpace(execID) == "" {
		return model.OrderStatusResponse{}, fmt.Errorf("validation: execID: required")
	}
	q := url.Values{"execId": {execID}}
	return getTyped[model.OrderStatusResponse](ctx, c, pathOrderByExecID+"?"+q.Encode())
}

// FilledOrders returns filled/partially-filled orders for an account.
//
// Reference: docs/primary-api.md "Consultar Ordenes Operadas"
func (c *Client) FilledOrders(ctx context.Context, account string) (model.AllOrdersStatusResponse, error) {
	if account == "" {
		return model.AllOrdersStatusResponse{}, fmt.Errorf("validation: account: required")
	}
	q := url.Values{"accountId": {account}}
	return getTyped[model.AllOrdersStatusResponse](ctx, c, pathOrderFilleds+"?"+q.Encode())
}

// ActiveOrders returns active orders for an account.
//
// Reference: docs/primary-api.md "Consultar órdenes activas"
func (c *Client) ActiveOrders(ctx context.Context, account string) (model.AllOrdersStatusResponse, error) {
	if account == "" {
		return model.AllOrdersStatusResponse{}, fmt.Errorf("validation: account: required")
	}
	q := url.Values{"accountId": {account}}
	return getTyped[model.AllOrdersStatusResponse](ctx, c, pathOrderActives+"?"+q.Encode())
}

// AllOrdersStatus returns the latest status of every request for an account.
//
// Reference: docs/primary-api.md "Estado de orden por ID Cuenta"
func (c *Client) AllOrdersStatus(ctx context.Context, account string) (model.AllOrdersStatusResponse, error) {
	if account == "" {
		return model.AllOrdersStatusResponse{}, fmt.Errorf("validation: account: required")
	}
	q := url.Values{"accountId": {account}}
	return getTyped[model.AllOrdersStatusResponse](ctx, c, pathAllOrders+"?"+q.Encode())
}
