package rofex

import (
	"context"
	"fmt"
	"strings"

	"github.com/carvalab/rofex-go/rofex/model"
)

// NewOrder representa los parámetros para enviar una nueva orden vía REST.
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
		return &ValidationError{Field: "symbol", Msg: "required"}
	}
	if o.Qty <= 0 {
		return &ValidationError{Field: "qty", Msg: "must be > 0"}
	}
	if o.Side == "" {
		return &ValidationError{Field: "side", Msg: "required"}
	}
	if o.Type == model.OrderTypeLimit && o.Price == nil {
		return &ValidationError{Field: "price", Msg: "required for limit"}
	}
	if o.TIF == model.GoodTillDate && (o.ExpireDate == nil || *o.ExpireDate == "") {
		return &ValidationError{Field: "expireDate", Msg: "required for GTD"}
	}
	return nil
}

// SendOrder envía una orden al mercado vía REST según la documentación Primary API.
//
// Funcionalidad para ingresar una orden al mercado. Es importante verificar
// que la orden fue realmente cargada y no fue rechazada según la secuencia
// recomendada en la documentación oficial.
//
// Secuencia recomendada (docs/primary-api.md - "Ingresar una orden"):
//  1. Ingresar la orden a través de la API
//  2. Si la respuesta es "status":"OK", consultar el estado de la orden usando el clientId devuelto
//  3. El estado puede ser: "NEW" (aceptada) o "REJECTED" (rechazada con motivo en el campo text)
//
// Parámetros obligatorios:
//   - Symbol: Símbolo del instrumento (ej: "DLR/DIC23" para futuro de dólar vencimiento Diciembre 2023)
//   - Market: Mercado (model.MarketROFEX para MATBA ROFEX, model.MarketMERV para mercados externos)
//   - Side: Lado de la orden (model.Buy o model.Sell)
//   - Type: Tipo de orden (model.OrderTypeLimit, model.OrderTypeMarket)
//   - Qty: Tamaño de la orden
//   - Account: Número de cuenta
//   - TIF: Tiempo de vida (model.Day, model.IOC, model.FOK, model.GoodTillDate)
//
// Parámetros condicionales:
//   - Price: Requerido para órdenes LIMIT
//   - DisplayQty: Requerido para órdenes Iceberg
//   - ExpireDate: Requerido para órdenes GTD (formato: "20230720")
//
// Tipos de orden soportados:
//   - LIMIT: Orden limitada con precio específico
//   - MARKET: Orden a mercado
//   - MARKET_TO_LIMIT: Orden a mercado convertida a limitada
//
// Modificadores de tiempo de vida:
//   - DAY: Válida solo por el día, se expira al cierre
//   - IOC: Immediate or Cancel
//   - FOK: Fill or Kill
//   - GTD: Good Till Date (requiere expireDate)
//
// Después de enviar una orden, verificar su estado con OrderStatus() ya que
// puede ser rechazada por el mercado.
//
// Example / Ejemplo:
//
//	price := 18.50
//	order, err := client.SendOrder(ctx, rofex.NewOrder{
//		Symbol:  "DLR/DIC21",
//		Side:    model.Buy,
//		Type:    model.OrderTypeLimit,
//		Qty:     10,
//		Price:   &price,
//		Account: "123",
//		TIF:     model.Day,
//	})
//	if err != nil {
//		return fmt.Errorf("send order failed: %w", err)
//	}
//
//	// Check order status / Verificar estado de orden
//	status, err := client.OrderStatus(ctx, order.Order.ClientID, "")
//
// Referencia: docs/primary-api.md - "Ingresar una orden"
func (c *Client) SendOrder(ctx context.Context, o NewOrder) (model.SendOrderResponse, error) {
	if err := o.validate(); err != nil {
		return model.SendOrderResponse{}, err
	}
	if o.Market == "" {
		o.Market = model.MarketROFEX
	}
	path := fmt.Sprintf(pathNewOrder,
		string(o.Market), o.Symbol, o.Qty, string(o.Type), string(o.Side), string(o.TIF), o.Account, o.CancelPrevious,
	)
	if o.Type == model.OrderTypeLimit && o.Price != nil {
		path += fmt.Sprintf("&price=%v", *o.Price)
	}
	if o.TIF == model.GoodTillDate && o.ExpireDate != nil {
		path += fmt.Sprintf("&expireDate=%s", *o.ExpireDate)
	}
	if o.Iceberg && o.DisplayQty != nil {
		path += fmt.Sprintf("&iceberg=true&displayQty=%d", *o.DisplayQty)
	}
	return getTyped[model.SendOrderResponse](ctx, c, path)
}

// CancelOrder cancela una orden vía REST según la documentación Primary API.
//
// Funcionalidad que permite cancelar una orden ingresada en el mercado.
// Requiere el clientOrderID devuelto al enviar la orden original.
//
// ⚠️ Rate Limit: Máximo 1 request por segundo para cancelaciones según documentación oficial.
//
// Parámetros:
//   - clientOrderID: ID del request devuelto por la API al ingresar la orden
//   - proprietary: ID que identifica al participante (opcional, usa valor por defecto)
//
// Referencia: docs/primary-api.md - "Cancelar una orden"
func (c *Client) CancelOrder(ctx context.Context, clientOrderID, proprietary string) (model.CancelOrderResponse, error) {
	if clientOrderID == "" {
		return model.CancelOrderResponse{}, &ValidationError{Field: "clientOrderID", Msg: "required"}
	}
	if strings.TrimSpace(proprietary) == "" {
		proprietary = c.proprietary
	}
	path := fmt.Sprintf(pathCancelOrder, clientOrderID, proprietary)
	return getTyped[model.CancelOrderResponse](ctx, c, path)
}

// ReplaceOrder reemplaza una orden existente según la documentación Primary API.
//
// Permite reemplazar una orden ingresada al mercado modificando cantidad y/o precio.
// Solo los campos no-nil son enviados en la modificación.
//
// Parámetros:
//   - clOrdID: ID del request de la orden original
//   - proprietary: ID del participante (opcional)
//   - newQty: Nueva cantidad (opcional)
//   - newPrice: Nuevo precio (opcional)
//
// Referencia: docs/primary-api.md - "Reemplazar una orden"
func (c *Client) ReplaceOrder(ctx context.Context, clOrdID, proprietary string, newQty *int64, newPrice *float64) (model.ReplaceOrderResponse, error) {
	if clOrdID == "" {
		return model.ReplaceOrderResponse{}, &ValidationError{Field: "clOrdID", Msg: "required"}
	}
	if strings.TrimSpace(proprietary) == "" {
		proprietary = c.proprietary
	}
	path := fmt.Sprintf(pathOrderReplace, clOrdID, proprietary)
	if newQty != nil {
		path += fmt.Sprintf("&orderQty=%d", *newQty)
	}
	if newPrice != nil {
		path += fmt.Sprintf("&price=%v", *newPrice)
	}
	return getTyped[model.ReplaceOrderResponse](ctx, c, path)
}

// OrderStatus consulta el estado de una orden según la documentación Primary API.
//
// Este método permite consultar el último estado de una orden específica
// usando su Client Order ID.
//
// Parámetros:
//   - clientOrderID: ID del request devuelto al enviar la orden
//   - proprietary: ID del participante (opcional)
//
// Estados posibles:
//   - PENDING_NEW: orden enviada, pendiente de confirmación
//   - NEW: orden aceptada por el mercado
//   - PARTIALLY_FILLED: orden parcialmente operada
//   - FILLED: orden completamente operada
//   - CANCELLED: orden cancelada
//   - REJECTED: orden rechazada (campo text indica motivo)
//
// Referencia: docs/primary-api.md - "Consultar último estado por Client Order ID"
func (c *Client) OrderStatus(ctx context.Context, clientOrderID, proprietary string) (model.OrderStatusResponse, error) {
	if clientOrderID == "" {
		return model.OrderStatusResponse{}, &ValidationError{Field: "clientOrderID", Msg: "required"}
	}
	if strings.TrimSpace(proprietary) == "" {
		proprietary = c.proprietary
	}
	path := fmt.Sprintf(pathOrderStatus, clientOrderID, proprietary)
	return getTyped[model.OrderStatusResponse](ctx, c, path)
}

// OrderHistoryByClOrdID consulta todos los estados de una orden según Primary API.
//
// Devuelve todos los estados por los que pasó una orden asociada a un request
// específico al mercado.
//
// Referencia: docs/primary-api.md - "Consultar todos los estados por Client Order ID"
func (c *Client) OrderHistoryByClOrdID(ctx context.Context, clOrdID, proprietary string) (model.AllOrdersStatusResponse, error) {
	if clOrdID == "" {
		return model.AllOrdersStatusResponse{}, &ValidationError{Field: "clOrdID", Msg: "required"}
	}
	if strings.TrimSpace(proprietary) == "" {
		proprietary = c.proprietary
	}
	path := fmt.Sprintf(pathOrderAllByID, clOrdID, proprietary)
	return getTyped[model.AllOrdersStatusResponse](ctx, c, path)
}

// OrderByOrderID consulta el estado de una orden por su Order ID.
//
// Consulta que devuelve el estado de una orden utilizando su ID de intercambio.
//
// Referencia: docs/primary-api.md - "Consultar Order por OrderID"
func (c *Client) OrderByOrderID(ctx context.Context, orderID string) (model.OrderStatusResponse, error) {
	if strings.TrimSpace(orderID) == "" {
		return model.OrderStatusResponse{}, &ValidationError{Field: "orderID", Msg: "required"}
	}
	path := fmt.Sprintf(pathOrderByOrder, orderID)
	return getTyped[model.OrderStatusResponse](ctx, c, path)
}

// OrderByExecID consulta el estado de una orden por Execution ID.
//
// Devuelve el estado de la orden asociada a un Execution ID específico,
// permitiendo identificar qué orden está involucrada en una operación.
//
// Referencia: docs/primary-api.md - "Estado de orden por Execution ID"
func (c *Client) OrderByExecID(ctx context.Context, execID string) (model.OrderStatusResponse, error) {
	if strings.TrimSpace(execID) == "" {
		return model.OrderStatusResponse{}, &ValidationError{Field: "execID", Msg: "required"}
	}
	path := fmt.Sprintf(pathOrderByExecID, execID)
	return getTyped[model.OrderStatusResponse](ctx, c, path)
}

// FilledOrders consulta las órdenes operadas según Primary API.
//
// Devuelve todas las órdenes que están total o parcialmente operadas
// para una cuenta específica.
//
// Referencia: docs/primary-api.md - "Consultar Ordenes Operadas"
func (c *Client) FilledOrders(ctx context.Context, account string) (model.AllOrdersStatusResponse, error) {
	if account == "" {
		return model.AllOrdersStatusResponse{}, &ValidationError{Field: "account", Msg: "required"}
	}
	path := fmt.Sprintf(pathOrderFilleds, account)
	return getTyped[model.AllOrdersStatusResponse](ctx, c, path)
}

// ActiveOrders consulta las órdenes activas según Primary API.
//
// Devuelve todas las órdenes activas de una cuenta específica.
//
// Referencia: docs/primary-api.md - "Consultar órdenes activas"
func (c *Client) ActiveOrders(ctx context.Context, account string) (model.AllOrdersStatusResponse, error) {
	if account == "" {
		return model.AllOrdersStatusResponse{}, &ValidationError{Field: "account", Msg: "required"}
	}
	path := fmt.Sprintf(pathOrderActives, account)
	return getTyped[model.AllOrdersStatusResponse](ctx, c, path)
}

// AllOrdersStatus consulta el estado de todas las órdenes por ID de cuenta.
//
// Devuelve el último estado de todos los requests (client order ID)
// asociados a una cuenta específica.
//
// Referencia: docs/primary-api.md - "Estado de orden por ID Cuenta"
func (c *Client) AllOrdersStatus(ctx context.Context, account string) (model.AllOrdersStatusResponse, error) {
	if account == "" {
		return model.AllOrdersStatusResponse{}, &ValidationError{Field: "account", Msg: "required"}
	}
	path := fmt.Sprintf(pathAllOrders, account)
	return getTyped[model.AllOrdersStatusResponse](ctx, c, path)
}
