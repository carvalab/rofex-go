package rofex

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/carvalab/rofex-go/rofex/model"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// StreamConnection is a managed websocket connection with reconnect support.
//
// All methods are safe for concurrent use.
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

// MarketDataSubscription streams real-time market data events.
//
// Reference: docs/primary-api.md "Suscribirse a MarketData en tiempo real a través de WebSocket"
type MarketDataSubscription struct {
	Events <-chan *model.MarketDataEvent
	Errs   <-chan error
	Close  func() error

	conn    *StreamConnection
	symbols []string
	entries []model.MDEntry
	depth   int
	market  model.Market
}

// OrderReportSubscription streams real-time order reports (execution reports).
//
// Reference: docs/primary-api.md "Suscribirse a Execution Reports a través de WebSocket"
type OrderReportSubscription struct {
	Events <-chan *model.OrderReportEvent
	Errs   <-chan error
	Close  func() error

	conn               *StreamConnection
	account            string
	snapshotOnlyActive bool
}

// NewStreamConnection creates a new managed websocket connection.
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

// Connect establishes the websocket connection.
func (sc *StreamConnection) Connect() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.isConnected {
		return nil
	}

	opts := &websocket.DialOptions{
		HTTPHeader:      sc.headers,
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

// Disconnect closes the websocket connection.
func (sc *StreamConnection) Disconnect() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if !sc.isConnected || sc.conn == nil {
		return nil
	}

	sc.isConnected = false
	sc.cancel()

	err := sc.conn.Close(websocket.StatusNormalClosure, "client disconnect")
	if sc.client.logger != nil {
		sc.client.logger.Debug("websocket disconnected")
	}
	return err
}

// IsConnected returns the current connection status.
func (sc *StreamConnection) IsConnected() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.isConnected
}

// WriteJSON writes a JSON message to the websocket.
func (sc *StreamConnection) WriteJSON(ctx context.Context, v any) error {
	sc.mu.RLock()
	if !sc.isConnected || sc.conn == nil {
		sc.mu.RUnlock()
		return ErrClosed
	}
	conn := sc.conn
	sc.mu.RUnlock()
	return wsjson.Write(ctx, conn, v)
}

// ReadJSON reads a JSON message from the websocket.
func (sc *StreamConnection) ReadJSON(ctx context.Context, v any) error {
	sc.mu.RLock()
	if !sc.isConnected || sc.conn == nil {
		sc.mu.RUnlock()
		return ErrClosed
	}
	conn := sc.conn
	sc.mu.RUnlock()
	return wsjson.Read(ctx, conn, v)
}

// Ping sends a ping frame to keep the connection alive.
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

// SubscribeMarketData subscribes to real-time MarketData via WebSocket.
//
// Depth is the order-book depth requested: 1 = top of book (default), 2..5 = multiple
// levels per side (Primary API limit).
//
// Reference: docs/primary-api.md "Suscribirse a MarketData en tiempo real a través de WebSocket"
func (c *Client) SubscribeMarketData(ctx context.Context, symbols []string, entries []model.MDEntry, depth int, market model.Market) (*MarketDataSubscription, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("validation: symbols: required")
	}
	if market == "" {
		market = model.MarketROFEX
	}
	if depth <= 0 {
		depth = 1
	}

	eventsChan := make(chan *model.MarketDataEvent, c.wsBuf)
	errsChan := make(chan error, 5)

	subMsg := struct {
		Type     model.WSMessageType `json:"type"`
		Level    int                 `json:"level"`
		Depth    int                 `json:"depth"`
		Entries  []model.MDEntry     `json:"entries"`
		Products []map[string]string `json:"products"`
	}{
		Type:    model.WSMessageSubscribeMarketData,
		Level:   1,
		Depth:   depth,
		Entries: entries,
	}
	for _, symbol := range symbols {
		subMsg.Products = append(subMsg.Products, map[string]string{
			"symbol":   symbol,
			"marketId": string(market),
		})
	}

	sub := &MarketDataSubscription{
		Events:  eventsChan,
		Errs:    errsChan,
		symbols: symbols,
		entries: entries,
		depth:   depth,
		market:  market,
	}

	go runSubscription(ctx, c,
		func() (*StreamConnection, any, error) {
			token, err := c.wsAuthToken(ctx)
			if err != nil {
				return nil, nil, fmt.Errorf("auth token error: %w", err)
			}
			headers := http.Header{"X-Auth-Token": []string{token}}
			conn := c.NewStreamConnection(ctx, c.wsURL, headers)
			sub.conn = conn
			return conn, subMsg, nil
		},
		func(ctx context.Context, conn *StreamConnection) error {
			return readEvents(ctx, c, conn, eventsChan, errsChan, model.WSMessageMarketData, func(e *model.MarketDataEvent) {
				e.Type = model.WSMessageType(strings.ToLower(string(e.Type)))
			})
		},
		eventsChan, errsChan,
	)

	sub.Close = func() error {
		if sub.conn != nil {
			return sub.conn.Disconnect()
		}
		return nil
	}
	return sub, nil
}

// SubscribeOrderReport subscribes to Execution Reports via WebSocket.
//
// Reference: docs/primary-api.md "Suscribirse a Execution Reports a través de WebSocket"
func (c *Client) SubscribeOrderReport(ctx context.Context, account string, snapshotOnlyActive bool) (*OrderReportSubscription, error) {
	if account == "" {
		return nil, fmt.Errorf("validation: account: required")
	}

	eventsChan := make(chan *model.OrderReportEvent, c.wsBuf)
	errsChan := make(chan error, 5)

	subMsg := struct {
		Type    model.WSMessageType `json:"type"`
		Account struct {
			ID string `json:"id"`
		} `json:"account"`
		SnapshotOnlyActive bool `json:"snapshotOnlyActive"`
	}{
		Type:               model.WSMessageOrderSubscription,
		SnapshotOnlyActive: snapshotOnlyActive,
		Account: struct {
			ID string `json:"id"`
		}{ID: account},
	}

	sub := &OrderReportSubscription{
		Events:             eventsChan,
		Errs:               errsChan,
		account:            account,
		snapshotOnlyActive: snapshotOnlyActive,
	}

	go runSubscription(ctx, c,
		func() (*StreamConnection, any, error) {
			token, err := c.wsAuthToken(ctx)
			if err != nil {
				return nil, nil, fmt.Errorf("auth token error: %w", err)
			}
			headers := http.Header{"X-Auth-Token": []string{token}}
			conn := c.NewStreamConnection(ctx, c.wsURL, headers)
			sub.conn = conn
			return conn, subMsg, nil
		},
		func(ctx context.Context, conn *StreamConnection) error {
			return readEvents(ctx, c, conn, eventsChan, errsChan, model.WSMessageOrderReport, func(e *model.OrderReportEvent) {
				e.Type = model.WSMessageType(strings.ToLower(string(e.Type)))
			})
		},
		eventsChan, errsChan,
	)

	sub.Close = func() error {
		if sub.conn != nil {
			return sub.conn.Disconnect()
		}
		return nil
	}
	return sub, nil
}

// runSubscription owns the outer connection lifecycle for any subscription:
// auth → connect → write subscription msg → read events → reconnect on
// recoverable errors with exponential backoff (cap 30s).
func runSubscription[T any](
	ctx context.Context,
	c *Client,
	setup func() (*StreamConnection, any, error),
	readLoop func(ctx context.Context, conn *StreamConnection) error,
	events chan *T,
	errs chan error,
) {
	defer close(events)
	defer close(errs)

	const maxBackoff = 30 * time.Second
	backoff := time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, subMsg, err := setup()
		if err != nil {
			c.sendErr(errs, ctx, err)
			return
		}

		if err := conn.Connect(); err != nil {
			conn.Disconnect()
			c.sleepAndBackoff(errs, ctx, err, &backoff, maxBackoff)
			continue
		}

		if err := conn.WriteJSON(ctx, subMsg); err != nil {
			conn.Disconnect()
			c.sleepAndBackoff(errs, ctx, fmt.Errorf("subscription send failed: %w", err), &backoff, maxBackoff)
			continue
		}

		backoff = time.Second
		if c.logger != nil {
			c.logger.Info("subscription established")
		}

		if err := readLoop(ctx, conn); err != nil {
			if c.logger != nil {
				c.logger.Debug("connection lost, attempting reconnect", slog.Any("err", err))
			}
			conn.Disconnect()

			if !isRecoverable(err) {
				c.sendErr(errs, ctx, err)
				return
			}

			time.Sleep(backoff)
			if backoff < maxBackoff {
				backoff *= 2
			}
		}
	}
}

// readEvents drives the websocket read loop for a single connection session:
// lowercases the message type, drops events whose type doesn't match the
// expected one, and forwards matching events to eventsChan (drop-on-full or block).
func readEvents[T any](
	ctx context.Context,
	c *Client,
	conn *StreamConnection,
	eventsChan chan<- *T,
	errs chan<- error,
	wantType model.WSMessageType,
	normalize func(*T),
) error {
	connCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go keepAlive(connCtx, c, conn)

	for {
		if connCtx.Err() != nil {
			return connCtx.Err()
		}
		var ev T
		if err := conn.ReadJSON(connCtx, &ev); err != nil {
			return err
		}
		if normalize != nil {
			normalize(&ev)
		}
		// Skip non-matching event types without breaking the loop.
		if !eventMatches(&ev, wantType) {
			continue
		}
		if c.wsDropOnFull {
			select {
			case eventsChan <- &ev:
			default:
				if c.logger != nil {
					c.logger.Warn("event dropped - channel full")
				}
			}
		} else {
			select {
			case eventsChan <- &ev:
			case <-connCtx.Done():
				return connCtx.Err()
			}
		}
	}
}

// eventMatches returns true if the WS event's Type equals wantType.
// T must be a struct with a `Type model.WSMessageType` field (MarketDataEvent,
// OrderReportEvent). We use a type assertion to read it; if T doesn't match,
// we forward the event (fail-open) to preserve generic semantics.
func eventMatches[T any](ev *T, want model.WSMessageType) bool {
	type t interface{ Type() model.WSMessageType }
	if e, ok := any(ev).(t); ok {
		return e.Type() == want
	}
	return true
}

// keepAlive sends periodic pings to keep the connection alive.
func keepAlive(ctx context.Context, c *Client, conn *StreamConnection) {
	ticker := time.NewTicker(25 * time.Second) // Slightly less than 30s server timeout
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

// isRecoverable returns true if a connection error should trigger a reconnect.
func isRecoverable(err error) bool {
	if err == nil {
		return false
	}
	if status := websocket.CloseStatus(err); status != -1 {
		switch status {
		case websocket.StatusNormalClosure, websocket.StatusGoingAway:
			return false
		default:
			return true
		}
	}
	return errors.Is(err, io.EOF) ||
		errors.Is(err, net.ErrClosed) ||
		isNetOpError(err)
}

func isNetOpError(err error) bool {
	var op *net.OpError
	return errors.As(err, &op)
}

func (c *Client) sendErr(errs chan<- error, ctx context.Context, err error) {
	select {
	case errs <- err:
	case <-ctx.Done():
	default:
		if c.logger != nil {
			c.logger.Warn("error channel full, dropping error", slog.Any("err", err))
		}
	}
}

func (c *Client) sleepAndBackoff(errs chan<- error, ctx context.Context, err error, backoff *time.Duration, max time.Duration) {
	if c.logger != nil {
		c.logger.Warn("connection error, will retry",
			slog.Any("err", err),
			slog.Duration("backoff", *backoff))
	}
	c.sendErr(errs, ctx, err)
	time.Sleep(*backoff)
	if *backoff < max {
		*backoff *= 2
	}
}

// SendOrderWS sends a new order via WebSocket.
//
// Reference: docs/primary-api.md "Ingresar una orden a través de WebSocket"
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

// CancelOrderWS cancels an existing order via WebSocket.
//
// Reference: docs/primary-api.md "Cancelar una Orden a través de WebSocket"
func (c *Client) CancelOrderWS(ctx context.Context, clientOrderID, proprietary string) error {
	if clientOrderID == "" {
		return fmt.Errorf("validation: clientOrderID: required")
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
