package rofex

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/carvalab/rofex-go/rofex/model"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// StreamConnection administra una conexión websocket persistente con reconexión automática.
//
// Proporciona una abstracción de alto nivel para manejar conexiones WebSocket de larga duración
// con características avanzadas como:
//   - Reconexión automática con backoff exponencial
//   - Operaciones thread-safe con mutex
//   - Gestión apropiada del contexto y cancelación
//   - Detección inteligente de errores recuperables vs no recuperables
//   - Keepalive automático con ping/pong
//
// Esta estructura encapsula toda la complejidad de mantener conexiones WebSocket estables
// para trading de alta frecuencia donde la pérdida de conectividad debe minimizarse.
type StreamConnection struct {
	client      *Client
	conn        *websocket.Conn
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	isConnected bool
	url         string
	headers     http.Header
}

// MarketDataSubscription gestiona una suscripción a datos de mercado en tiempo real.
//
// Basado en primary-api.md: "Suscribirse a MarketData en tiempo real a través de WebSocket"
//
// Uso:
//
//	sub, _ := client.SubscribeMarketData(ctx, symbols, entries, depth, market)
//	for event := range sub.Events {
//	    fmt.Printf("Símbolo: %s, Precio: %v\n", event.InstrumentID.Symbol, event.MarketData)
//	}
type MarketDataSubscription struct {
	Events <-chan *model.MarketDataEvent // Canal tipado según primary-api.md
	Errs   <-chan error                  // Canal de errores
	Close  func() error                  // Función para cerrar suscripción

	// Campos internos para gestión
	conn    *StreamConnection
	symbols []string
	entries []model.MDEntry
	depth   int
	market  model.Market
}

// OrderReportSubscription gestiona una suscripción a reportes de órdenes en tiempo real.
//
// Basado en primary-api.md: "Suscribirse a Execution Reports a través de WebSocket"
//
// Uso:
//
//	sub, _ := client.SubscribeOrderReport(ctx, account, snapshotOnlyActive)
//	for event := range sub.Events {
//	    fmt.Printf("Orden: %s, Estado: %s\n", event.OrderReport.ClOrdID, event.OrderReport.Status)
//	}
type OrderReportSubscription struct {
	Events <-chan *model.OrderReportEvent // Canal tipado según primary-api.md
	Errs   <-chan error                   // Canal de errores
	Close  func() error                   // Función para cerrar suscripción

	// Campos internos para gestión
	conn               *StreamConnection
	account            string
	snapshotOnlyActive bool
}

// NewStreamConnection creates a new managed websocket connection
func (c *Client) NewStreamConnection(ctx context.Context, url string, headers http.Header) *StreamConnection {
	connCtx, cancel := context.WithCancel(ctx)

	return &StreamConnection{
		client:  c,
		ctx:     connCtx,
		cancel:  cancel,
		url:     url,
		headers: headers,
	}
}

// Connect establishes the websocket connection with proper error handling
func (sc *StreamConnection) Connect() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.isConnected {
		return nil
	}

	opts := &websocket.DialOptions{
		HTTPHeader: sc.headers,
		// Add compression support
		CompressionMode: websocket.CompressionContextTakeover,
	}

	conn, _, err := websocket.Dial(sc.ctx, sc.url, opts)
	if err != nil {
		return fmt.Errorf("websocket dial failed: %w", err)
	}

	sc.conn = conn
	sc.isConnected = true

	if sc.client.logger != nil {
		sc.client.logger.Debug("websocket connected", slog.String("url", sc.url))
	}

	return nil
}

// Disconnect closes the websocket connection gracefully
func (sc *StreamConnection) Disconnect() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if !sc.isConnected || sc.conn == nil {
		return nil
	}

	sc.isConnected = false
	sc.cancel()

	// Close with proper status code and reason
	err := sc.conn.Close(websocket.StatusNormalClosure, "client disconnect")
	if sc.client.logger != nil {
		sc.client.logger.Debug("websocket disconnected")
	}

	return err
}

// IsConnected returns the current connection status
func (sc *StreamConnection) IsConnected() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.isConnected
}

// WriteJSON writes a JSON message to the websocket
func (sc *StreamConnection) WriteJSON(ctx context.Context, v interface{}) error {
	sc.mu.RLock()
	if !sc.isConnected || sc.conn == nil {
		sc.mu.RUnlock()
		return ErrClosed
	}
	conn := sc.conn
	sc.mu.RUnlock()

	return wsjson.Write(ctx, conn, v)
}

// ReadJSON reads a JSON message from the websocket
func (sc *StreamConnection) ReadJSON(ctx context.Context, v interface{}) error {
	sc.mu.RLock()
	if !sc.isConnected || sc.conn == nil {
		sc.mu.RUnlock()
		return ErrClosed
	}
	conn := sc.conn
	sc.mu.RUnlock()

	return wsjson.Read(ctx, conn, v)
}

// Ping sends a ping frame to keep the connection alive
func (sc *StreamConnection) Ping(ctx context.Context) error {
	sc.mu.RLock()
	if !sc.isConnected || sc.conn == nil {
		sc.mu.RUnlock()
		return ErrClosed
	}
	conn := sc.conn
	sc.mu.RUnlock()

	return conn.Ping(ctx)
}

// SubscribeMarketData suscribe a MarketData en tiempo real a través de WebSocket.
//
// Utilizando el protocolo Web Socket es posible recibir Market Data de los instrumentos
// especificados de manera asíncrona cuando esta cambie sin necesidad de hacer un request
// cada vez que necesitemos.
//
// Para recibir este tipo de mensajes hay que suscribirse indicando los instrumentos de
// los cuales queremos recibir MD. El servidor enviara un mensaje de MD por cada instrumento
// al que nos suscribimos cada vez que este cambie.
//
// Con este mensaje nos suscribimos para recibir MD de los instrumentos especificados, el
// servidor solamente enviará los datos especificados en la lista "entries".
//
// Depth: profundidad del libro (order book) solicitada.
//   - 1: Top of book (mejor BID/ASK). Es el valor por defecto y el más liviano.
//   - 2..5: Devuelve múltiples niveles por lado (hasta 5 según Primary API).
//
// Diferencias prácticas entre depth=1 y depth=5:
//   - 1 envía solo el mejor precio/cantidad por lado -> menor ancho de banda/latencia.
//   - 5 envía hasta 5 niveles por lado -> mayor granularidad del libro, más datos a procesar.
//
// Mensaje enviado:
//
//	{
//	  "type":"smd",
//	  "level":1,
//	  "entries":["OF"],
//	  "products":[
//	    {"symbol":"DLR/DIC23", "marketId": "ROFX"},
//	    {"symbol":"SOJ.ROS/MAY23", "marketId": "ROFX"}
//	  ],
//	  "depth":2
//	}
//
// Referencia: docs/primary-api.md - "Suscribirse a MarketData en tiempo real a través de WebSocket"
func (c *Client) SubscribeMarketData(ctx context.Context, symbols []string, entries []model.MDEntry, depth int, market model.Market) (*MarketDataSubscription, error) {
	if len(symbols) == 0 {
		return nil, &ValidationError{Field: "symbols", Msg: "required"}
	}
	if market == "" {
		market = model.MarketROFEX
	}
	if depth <= 0 {
		depth = 1
	}

	// Crear canales tipados según primary-api.md
	eventsChan := make(chan *model.MarketDataEvent, c.wsBuf)
	errorChan := make(chan error, 5)

	// Create subscription message structure
	subscriptionMsg := struct {
		Type     model.WSMessageType `json:"type"`
		Level    int                 `json:"level"`
		Depth    int                 `json:"depth"`
		Entries  []model.MDEntry     `json:"entries"`
		Products []map[string]string `json:"products"`
	}{
		Type:     model.WSMessageSubscribeMarketData,
		Level:    1,
		Depth:    depth,
		Entries:  entries,
		Products: make([]map[string]string, 0, len(symbols)),
	}

	for _, symbol := range symbols {
		subscriptionMsg.Products = append(subscriptionMsg.Products, map[string]string{
			"symbol":   symbol,
			"marketId": string(market),
		})
	}

	subscription := &MarketDataSubscription{
		Events:  eventsChan,
		Errs:    errorChan,
		symbols: symbols,
		entries: entries,
		depth:   depth,
		market:  market,
	}

	// Iniciar gestión de conexión
	go c.manageMarketDataConnection(ctx, subscription, subscriptionMsg, eventsChan, errorChan)

	subscription.Close = func() error {
		if subscription.conn != nil {
			return subscription.conn.Disconnect()
		}
		return nil
	}

	return subscription, nil
}

// manageMarketDataConnection handles connection lifecycle with exponential backoff
func (c *Client) manageMarketDataConnection(
	ctx context.Context,
	sub *MarketDataSubscription,
	subscriptionMsg interface{},
	eventsChan chan<- *model.MarketDataEvent,
	errorChan chan<- error,
) {
	defer close(eventsChan)
	defer close(errorChan)

	backoff := time.Second
	const maxBackoff = 30 * time.Second
	const maxRetries = 10
	retryCount := 0

	for {
		select {
		case <-ctx.Done():
			if c.logger != nil {
				c.logger.Debug("market data subscription context cancelled")
			}
			return
		default:
		}

		// Get fresh authentication token
		token, err := c.wsAuthToken(ctx)
		if err != nil {
			c.sendError(errorChan, ctx, fmt.Errorf("auth token error: %w", err))
			return
		}

		// Create connection with authentication header
		headers := http.Header{"X-Auth-Token": []string{token}}
		conn := c.NewStreamConnection(ctx, c.wsURL, headers)
		sub.conn = conn

		// Attempt to connect
		if err := conn.Connect(); err != nil {
			c.handleConnectionError(errorChan, ctx, err, &backoff, maxBackoff, &retryCount, maxRetries)
			continue
		}

		// Send subscription message
		if err := conn.WriteJSON(ctx, subscriptionMsg); err != nil {
			c.handleConnectionError(errorChan, ctx, fmt.Errorf("subscription send failed: %w", err), &backoff, maxBackoff, &retryCount, maxRetries)
			conn.Disconnect()
			continue
		}

		// Reset backoff and retry count on successful connection
		backoff = time.Second
		retryCount = 0

		if c.logger != nil {
			c.logger.Info("market data subscription established")
		}

		// Start keepalive and message processing
		if err := c.processMarketDataMessages(ctx, conn, eventsChan, errorChan); err != nil {
			if c.logger != nil {
				c.logger.Debug("connection lost, attempting reconnect", slog.Any("err", err))
			}
			conn.Disconnect()

			// Check if it's a recoverable error
			if !c.isRecoverableError(err) {
				c.sendError(errorChan, ctx, err)
				return
			}

			// Wait before retry
			time.Sleep(backoff)
			if backoff < maxBackoff {
				backoff *= 2
			}
		}
	}
}

// processMarketDataMessages handles message reading and keepalive
func (c *Client) processMarketDataMessages(
	ctx context.Context,
	conn *StreamConnection,
	eventsChan chan<- *model.MarketDataEvent,
	errorChan chan<- error,
) error {
	// Create context for this connection session
	connCtx, connCancel := context.WithCancel(ctx)
	defer connCancel()

	// Start keepalive goroutine
	go c.keepAlive(connCtx, conn)

	// Message processing loop
	for {
		select {
		case <-connCtx.Done():
			return connCtx.Err()
		default:
		}

		// Leer directamente a tipo estructurado según primary-api.md
		var event model.MarketDataEvent
		if err := conn.ReadJSON(connCtx, &event); err != nil {
			return err
		}
		// Normalizar el tipo a minúsculas para ser tolerantes con variantes ("Md" vs "md")
		event.Type = model.WSMessageType(strings.ToLower(string(event.Type)))

		// Enviar solo si es market data tipado
		if event.Type == model.WSMessageMarketData {
			if c.wsDropOnFull {
				select {
				case eventsChan <- &event:
				default:
					if c.logger != nil {
						c.logger.Warn("market data event dropped - channel full")
					}
				}
			} else {
				select {
				case eventsChan <- &event:
				case <-connCtx.Done():
					return connCtx.Err()
				}
			}
		}
	}
}

// keepAlive sends periodic ping messages to maintain connection
func (c *Client) keepAlive(ctx context.Context, conn *StreamConnection) {
	ticker := time.NewTicker(25 * time.Second) // Slightly less than 30s timeout
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			if err := conn.Ping(pingCtx); err != nil {
				if c.logger != nil {
					c.logger.Debug("ping failed", slog.Any("err", err))
				}
				cancel()
				return
			}
			cancel()
		}
	}
}

// SubscribeOrderReport suscribe a Execution Reports a través de WebSocket.
//
// Los siguientes mensajes le permiten al usuario suscribirse para recibir mensajes de
// Execution Reports sobre las órdenes asociadas a una cuenta determinada. Se permite
// suscribirse para ordenes asociadas a una, varias o todas las cuentas del usuario.
//
// Para una cuenta:
// Con este mensaje se podrá recibir los Execution Reports de todas las órdenes ingresadas
// con la cuenta indicada en el mensaje.
//
//	{"type":"os", "account": {"id":"40"}}
//
// Para varias cuentas:
// Con este mensaje se podrá recibir los Execution Reports de todas las órdenes ingresadas
// con las cuentas indicadas en el mensaje.
//
//	{"type":"os", "accounts":[{"id":"40"}, {"id":"4000"}]}
//
// Para todas las cuentas:
// Con este mensaje se podrá recibir los Execution Reports de todas las órdenes ingresadas
// con las cuentas asociadas al usuario.
//
//	{"type":"os"}
//
// Se encuentra disponible el parámetro "snapshotOnlyActive" para recibir los Execution
// Reports de las ordenes activas (en estado NEW o PARTIALLY_FILLED) para todas las
// cuentas, mas de una cuenta o para una única cuenta.
//
//	{"type":"os", "snapshotOnlyActive":true}
//
// Referencia: docs/primary-api.md - "Suscribirse a Execution Reports a través de WebSocket"
func (c *Client) SubscribeOrderReport(ctx context.Context, account string, snapshotOnlyActive bool) (*OrderReportSubscription, error) {
	if account == "" {
		return nil, &ValidationError{Field: "account", Msg: "required"}
	}

	eventsChan := make(chan *model.OrderReportEvent, c.wsBuf)
	errorChan := make(chan error, 5)

	subscriptionMsg := struct {
		Type    model.WSMessageType `json:"type"`
		Account struct {
			ID string `json:"id"`
		} `json:"account"`
		SnapshotOnlyActive bool `json:"snapshotOnlyActive"`
	}{
		Type: model.WSMessageOrderSubscription,
		Account: struct {
			ID string `json:"id"`
		}{ID: account},
		SnapshotOnlyActive: snapshotOnlyActive,
	}

	subscription := &OrderReportSubscription{
		Events:             eventsChan,
		Errs:               errorChan,
		account:            account,
		snapshotOnlyActive: snapshotOnlyActive,
	}

	// Iniciar gestión de conexión
	go c.manageOrderReportConnection(ctx, subscription, subscriptionMsg, eventsChan, errorChan)

	subscription.Close = func() error {
		if subscription.conn != nil {
			return subscription.conn.Disconnect()
		}
		return nil
	}

	return subscription, nil
}

// manageOrderReportConnection handles order report connection lifecycle
func (c *Client) manageOrderReportConnection(
	ctx context.Context,
	sub *OrderReportSubscription,
	subscriptionMsg interface{},
	eventsChan chan<- *model.OrderReportEvent,
	errorChan chan<- error,
) {
	defer close(eventsChan)
	defer close(errorChan)

	backoff := time.Second
	const maxBackoff = 30 * time.Second
	const maxRetries = 10
	retryCount := 0

	for {
		select {
		case <-ctx.Done():
			if c.logger != nil {
				c.logger.Debug("order report subscription context cancelled")
			}
			return
		default:
		}

		token, err := c.wsAuthToken(ctx)
		if err != nil {
			c.sendError(errorChan, ctx, fmt.Errorf("auth token error: %w", err))
			return
		}

		headers := http.Header{"X-Auth-Token": []string{token}}
		conn := c.NewStreamConnection(ctx, c.wsURL, headers)
		sub.conn = conn

		if err := conn.Connect(); err != nil {
			c.handleConnectionError(errorChan, ctx, err, &backoff, maxBackoff, &retryCount, maxRetries)
			continue
		}

		if err := conn.WriteJSON(ctx, subscriptionMsg); err != nil {
			c.handleConnectionError(errorChan, ctx, fmt.Errorf("subscription send failed: %w", err), &backoff, maxBackoff, &retryCount, maxRetries)
			conn.Disconnect()
			continue
		}

		backoff = time.Second
		retryCount = 0

		if c.logger != nil {
			c.logger.Info("order report subscription established")
		}

		if err := c.processOrderReportMessages(ctx, conn, eventsChan, errorChan); err != nil {
			if c.logger != nil {
				c.logger.Debug("connection lost, attempting reconnect", slog.Any("err", err))
			}
			conn.Disconnect()

			if !c.isRecoverableError(err) {
				c.sendError(errorChan, ctx, err)
				return
			}

			time.Sleep(backoff)
			if backoff < maxBackoff {
				backoff *= 2
			}
		}
	}
}

// processOrderReportMessages handles order report message reading
func (c *Client) processOrderReportMessages(
	ctx context.Context,
	conn *StreamConnection,
	eventsChan chan<- *model.OrderReportEvent,
	errorChan chan<- error,
) error {
	connCtx, connCancel := context.WithCancel(ctx)
	defer connCancel()

	go c.keepAlive(connCtx, conn)

	for {
		select {
		case <-connCtx.Done():
			return connCtx.Err()
		default:
		}

		// Leer directamente a tipo estructurado según primary-api.md
		var event model.OrderReportEvent
		if err := conn.ReadJSON(connCtx, &event); err != nil {
			return err
		}
		// Normalizar el tipo a minúsculas para ser tolerantes con variantes ("Or" vs "or")
		event.Type = model.WSMessageType(strings.ToLower(string(event.Type)))

		// Enviar solo si es order report tipado
		if event.Type == model.WSMessageOrderReport {
			if c.wsDropOnFull {
				select {
				case eventsChan <- &event:
				default:
					if c.logger != nil {
						c.logger.Warn("order report event dropped - channel full")
					}
				}
			} else {
				select {
				case eventsChan <- &event:
				case <-connCtx.Done():
					return connCtx.Err()
				}
			}
		}
	}
}

// Helper functions for better error handling and management

// isRecoverableError determines if an error should trigger reconnection
func (c *Client) isRecoverableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for websocket close status
	if status := websocket.CloseStatus(err); status != -1 {
		switch status {
		case websocket.StatusNormalClosure, websocket.StatusGoingAway:
			return false // Normal closure, don't reconnect
		case websocket.StatusAbnormalClosure:
			return true // Abnormal closure, reconnect
		default:
			return true // Other close statuses, attempt reconnect
		}
	}

	// Check for network errors
	errStr := err.Error()
	recoverableErrors := []string{
		"i/o timeout",
		"connection reset",
		"EOF",
		"broken pipe",
		"network is unreachable",
		"connection refused",
	}

	for _, recoverable := range recoverableErrors {
		if strings.Contains(errStr, recoverable) {
			return true
		}
	}

	return false
}

// handleConnectionError manages connection errors with backoff and retry logic
func (c *Client) handleConnectionError(
	errorChan chan<- error,
	ctx context.Context,
	err error,
	backoff *time.Duration,
	maxBackoff time.Duration,
	retryCount *int,
	maxRetries int,
) {
	*retryCount++

	if *retryCount >= maxRetries {
		c.sendError(errorChan, ctx, fmt.Errorf("max retries exceeded: %w", err))
		return
	}

	if c.logger != nil {
		c.logger.Warn("connection error, will retry",
			slog.Any("err", err),
			slog.Duration("backoff", *backoff),
			slog.Int("retry", *retryCount))
	}

	time.Sleep(*backoff)
	if *backoff < maxBackoff {
		*backoff *= 2
	}
}

// sendError safely sends an error to the error channel
func (c *Client) sendError(errorChan chan<- error, ctx context.Context, err error) {
	select {
	case errorChan <- err:
	case <-ctx.Done():
	default:
		if c.logger != nil {
			c.logger.Warn("error channel full, dropping error", slog.Any("err", err))
		}
	}
}

// SendOrderWS permite ingresar una orden a través de WebSocket.
//
// Con este mensaje se envía una orden al mercado. Para poder saber que ocurrió con la
// orden hay que estar suscripto a los Execution Report para la cuenta con la que
// mandamos la orden, de lo contrario no recibiremos ningún mensaje sobre el estado
// de la orden.
//
// Mensaje enviado:
//
//	{
//	  "type":"no",
//	  "product":{"marketId":"ROFX", "symbol":"DLR/DIC23"},
//	  "price":185,
//	  "quantity":23,
//	  "side": "BUY",
//	  "account":"20",
//	  "iceberg":false
//	}
//
// Mensaje para el envío de una orden para un contrato Todo o Nada:
//
//	{
//	  "type":"no",
//	  "product":{"marketId":"ROFX", "symbol":"DLR/DIC23A"},
//	  "price":185,
//	  "quantity":3000,
//	  "side": "BUY",
//	  "account":"20",
//	  "iceberg":false
//	}
//
// Mensaje para el envío de una orden con identificador (wsClOrdId):
//
//	{
//	  "type":"no",
//	  "product":{"marketId":"ROFX", "symbol":"DLR/DIC23"},
//	  "price":185,
//	  "quantity":23,
//	  "side":"BUY",
//	  "account":"20",
//	  "iceberg":false,
//	  "wsClOrdId":"asdjuej213n1"
//	}
//
// El campo wsClOrdId se utiliza para identificar la orden enviada. Este campo va a
// venir solamente en el primer Execution Report (con estado PENDING_NEW o REJECT).
// En el primer Execution Report recibido el campo wsClOrdId se debe referenciar con
// el campo clOrdId para poder seguir los diferentes estados de la orden.
//
// Mensaje para el envío de una orden Iceberg vía WebSocket:
//
//	{
//	  "type":"no",
//	  "product":{"marketId":"ROFX", "symbol": "DLR/DIC22"},
//	  "price":185,
//	  "quantity":20,
//	  "side":"BUY",
//	  "account":"20",
//	  "iceberg":true,
//	  "displayQuantity":8
//	}
//
// Mensaje para el envío de una orden GTD:
//
//	{
//	  "type":"no",
//	  "product":{"marketId":"ROFX", "symbol":"DLR/DIC23"},
//	  "price":185,
//	  "quantity":23,
//	  "side": "BUY",
//	  "account":"20",
//	  "timeInForce":"GTD",
//	  "expireDate":"20231010"
//	}
//
// Referencia: docs/primary-api.md - "Ingresar una orden a través de WebSocket"
func (c *Client) SendOrderWS(ctx context.Context, o NewOrder) error {
	if err := o.validate(); err != nil {
		return err
	}
	if o.Market == "" {
		o.Market = model.MarketROFEX
	}

	token, err := c.wsAuthToken(ctx)
	if err != nil {
		return fmt.Errorf("auth token error: %w", err)
	}

	headers := http.Header{"X-Auth-Token": []string{token}}
	conn := c.NewStreamConnection(ctx, c.wsURL, headers)

	if err := conn.Connect(); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Disconnect()

	orderMsg := struct {
		Type        model.WSMessageType `json:"type"`
		Product     map[string]string   `json:"product"`
		Quantity    string              `json:"quantity"`
		OrdType     string              `json:"ordType"`
		Side        string              `json:"side"`
		Account     string              `json:"account"`
		AllOrNone   string              `json:"allOrNone"`
		TimeInForce string              `json:"timeInForce"`
		Price       *string             `json:"price,omitempty"`
		Iceberg     *string             `json:"iceberg,omitempty"`
		DisplayQty  *string             `json:"displayQuantity,omitempty"`
		ExpireDate  *string             `json:"expireDate,omitempty"`
		WSClOrdID   *string             `json:"wsClOrdId,omitempty"`
	}{
		Type: model.WSMessageNewOrder,
		Product: map[string]string{
			"marketId": string(o.Market),
			"symbol":   o.Symbol,
		},
		Quantity:    fmt.Sprintf("%d", o.Qty),
		OrdType:     string(o.Type),
		Side:        strings.ToUpper(string(o.Side)),
		Account:     o.Account,
		AllOrNone:   fmt.Sprintf("%t", o.AllOrNone),
		TimeInForce: strings.ToUpper(string(o.TIF)),
	}

	// Add optional fields
	if o.Price != nil && o.Type == model.OrderTypeLimit {
		priceStr := fmt.Sprintf("%v", *o.Price)
		orderMsg.Price = &priceStr
	}
	if o.Iceberg && o.DisplayQty != nil {
		icebergStr := "true"
		displayStr := fmt.Sprintf("%d", *o.DisplayQty)
		orderMsg.Iceberg = &icebergStr
		orderMsg.DisplayQty = &displayStr
	}
	if o.TIF == model.GoodTillDate && o.ExpireDate != nil {
		orderMsg.ExpireDate = o.ExpireDate
	}
	if o.WSClOrdID != nil && *o.WSClOrdID != "" {
		orderMsg.WSClOrdID = o.WSClOrdID
	}

	return conn.WriteJSON(ctx, orderMsg)
}

// CancelOrderWS permite cancelar una orden a través de WebSocket.
//
// Mensaje que permite cancelar una orden ingresada en el mercado vía WebSocket.
//
// Mensaje enviado:
//
//	{
//	  "type":"co",
//	  "clientId":"user114121092035207",
//	  "proprietary":"PBCP"
//	}
//
// Referencia: docs/primary-api.md - "Cancelar una Orden a través de WebSocket"
func (c *Client) CancelOrderWS(ctx context.Context, clientOrderID, proprietary string) error {
	if clientOrderID == "" {
		return &ValidationError{Field: "clientOrderID", Msg: "required"}
	}
	if proprietary == "" {
		proprietary = c.proprietary
	}

	token, err := c.wsAuthToken(ctx)
	if err != nil {
		return fmt.Errorf("auth token error: %w", err)
	}

	headers := http.Header{"X-Auth-Token": []string{token}}
	conn := c.NewStreamConnection(ctx, c.wsURL, headers)

	if err := conn.Connect(); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Disconnect()

	cancelMsg := struct {
		Type        model.WSMessageType `json:"type"`
		ClientID    string              `json:"clientId"`
		Proprietary string              `json:"proprietary"`
	}{
		Type:        model.WSMessageCancelOrder,
		ClientID:    clientOrderID,
		Proprietary: proprietary,
	}

	return conn.WriteJSON(ctx, cancelMsg)
}
