package rofex

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/carvalab/rofex-go/rofex/model"
	"github.com/coder/websocket"
)

func ptr[T any](v T) *T { return &v }

// Invalid orders must be rejected locally, before anything reaches the market.
func TestNewOrder_Validate(t *testing.T) {
	valid := NewOrder{Symbol: "DLR/DIC26", Side: model.Buy, Type: model.OrderTypeLimit, Qty: 1, Price: ptr(1000.0), TIF: model.Day}

	tests := []struct {
		name    string
		mutate  func(o *NewOrder)
		wantErr string
	}{
		{"valid limit", func(o *NewOrder) {}, ""},
		{"valid market without price", func(o *NewOrder) { o.Type, o.Price = model.OrderTypeMarket, nil }, ""},
		{"missing symbol", func(o *NewOrder) { o.Symbol = "" }, "symbol"},
		{"zero qty", func(o *NewOrder) { o.Qty = 0 }, "qty"},
		{"negative qty", func(o *NewOrder) { o.Qty = -5 }, "qty"},
		{"missing side", func(o *NewOrder) { o.Side = "" }, "side"},
		{"limit without price", func(o *NewOrder) { o.Price = nil }, "price"},
		{"GTD without expire date", func(o *NewOrder) { o.TIF = model.GoodTillDate }, "expireDate"},
		{"GTD with empty expire date", func(o *NewOrder) { o.TIF, o.ExpireDate = model.GoodTillDate, ptr("") }, "expireDate"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := valid
			tt.mutate(&o)
			err := o.validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("want error about %q, got %v", tt.wantErr, err)
			}
		})
	}

	// An invalid order must not produce any HTTP request.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request to %s", r.URL.Path)
	}))
	defer ts.Close()
	c, _ := NewClient(WithBaseURL(ts.URL+"/"), WithStaticToken("tok"))
	if _, err := c.SendOrder(context.Background(), NewOrder{Symbol: "DLR/DIC26", Side: model.Buy, Type: model.OrderTypeLimit, Qty: 1}); err == nil {
		t.Fatal("SendOrder: want validation error for limit order without price")
	}
}

// A market order must never carry a price, even if the caller set one.
func TestSendOrder_MarketOrderOmitsPrice(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("price") {
			t.Errorf("market order sent price=%s", r.URL.Query().Get("price"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "OK"})
	}))
	defer ts.Close()

	c, _ := NewClient(WithBaseURL(ts.URL+"/"), WithStaticToken("tok"))
	_, err := c.SendOrder(context.Background(), NewOrder{
		Symbol: "DLR/DIC26", Side: model.Sell, Type: model.OrderTypeMarket, Qty: 3, Price: ptr(1000.0), TIF: model.ImmediateOrCancel, Account: "REM1",
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Replacing only the price must not send orderQty (and vice versa): sending a
// zero/garbage value would change the live order's size.
func TestReplaceOrder_OnlySendsChangedFields(t *testing.T) {
	var got []map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := map[string]string{}
		for k := range r.URL.Query() {
			q[k] = r.URL.Query().Get(k)
		}
		got = append(got, q)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "OK"})
	}))
	defer ts.Close()

	c, _ := NewClient(WithBaseURL(ts.URL+"/"), WithStaticToken("tok"))
	ctx := context.Background()
	if _, err := c.ReplaceOrder(ctx, "ID1", "", nil, ptr(1234.5)); err != nil {
		t.Fatal(err)
	}
	if _, err := c.ReplaceOrder(ctx, "ID1", "", ptr(int64(7)), nil); err != nil {
		t.Fatal(err)
	}

	if _, ok := got[0]["orderQty"]; ok || got[0]["price"] != "1234.5" {
		t.Errorf("price-only replace sent %v", got[0])
	}
	if _, ok := got[1]["price"]; ok || got[1]["orderQty"] != "7" {
		t.Errorf("qty-only replace sent %v", got[1])
	}
	// Empty proprietary falls back to the environment default (REMARKET = PBCP).
	if got[0]["proprietary"] != "PBCP" || got[0]["clOrdId"] != "ID1" {
		t.Errorf("replace identity params: %v", got[0])
	}
}

func TestCancelOrder_DefaultsProprietaryByEnvironment(t *testing.T) {
	var proprietary string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proprietary = r.URL.Query().Get("proprietary")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "OK"})
	}))
	defer ts.Close()

	ctx := context.Background()
	live, _ := NewClient(WithEnvironment(model.EnvironmentLive), WithBaseURL(ts.URL+"/"), WithWSURL("ws://unused/"), WithStaticToken("tok"))
	if _, err := live.CancelOrder(ctx, "ID1", ""); err != nil {
		t.Fatal(err)
	}
	if proprietary != "api" {
		t.Errorf("LIVE default proprietary = %q, want api", proprietary)
	}
	if _, err := live.CancelOrder(ctx, "ID1", "BROKER"); err != nil {
		t.Fatal(err)
	}
	if proprietary != "BROKER" {
		t.Errorf("explicit proprietary = %q, want BROKER", proprietary)
	}
}

// LIVE must never silently fall back to the sandbox URLs.
func TestNewClient_LiveRequiresBrokerURLs(t *testing.T) {
	if _, err := NewClient(WithEnvironment(model.EnvironmentLive)); err == nil {
		t.Fatal("want error for LIVE without base/ws URLs")
	}
}

// wsCapture is a WebSocket server that records every message it receives.
func wsCapture(t *testing.T) (url string, msgs <-chan []byte) {
	t.Helper()
	ch := make(chan []byte, 16)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Auth-Token") == "" {
			t.Errorf("websocket dial without X-Auth-Token")
		}
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()
		for {
			_, b, err := conn.Read(r.Context())
			if err != nil {
				return
			}
			ch <- b
		}
	}))
	t.Cleanup(ts.Close)
	return "ws" + strings.TrimPrefix(ts.URL, "http") + "/", ch
}

func nextMsg(t *testing.T, ch <-chan []byte) map[string]any {
	t.Helper()
	select {
	case b := <-ch:
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("invalid json %s: %v", b, err)
		}
		return m
	case <-time.After(5 * time.Second):
		t.Fatal("no websocket message received")
		return nil
	}
}

// The WebSocket order must carry the exact price: prices are written as plain
// decimals, never in scientific notation (1e+06), which the market rejects.
func TestSendOrderWS_Message(t *testing.T) {
	url, msgs := wsCapture(t)
	c, _ := NewClient(WithWSURL(url), WithStaticToken("tok"))
	ctx := context.Background()

	tests := []struct {
		price float64
		want  string
	}{
		{1234.5, "1234.5"},
		{1500000, "1500000"},
		{0.00001, "0.00001"},
	}
	for _, tt := range tests {
		err := c.SendOrderWS(ctx, NewOrder{
			Symbol: "MERV - XMEV - GGAL - 24hs", Market: model.MarketROFEX, Side: model.Buy, Type: model.OrderTypeLimit,
			Qty: 10, Price: ptr(tt.price), TIF: model.GoodTillDate, ExpireDate: ptr("20261231"), Account: "REM1",
			WSClOrdID: ptr("my-id"),
		})
		if err != nil {
			t.Fatal(err)
		}
		m := nextMsg(t, msgs)
		if m["price"] != tt.want {
			t.Errorf("price %v sent as %q, want %q", tt.price, m["price"], tt.want)
		}
		if m["type"] != "no" || m["quantity"] != "10" || m["side"] != "BUY" || m["timeInForce"] != "GTD" ||
			m["account"] != "REM1" || m["expireDate"] != "20261231" || m["wsClOrdId"] != "my-id" {
			t.Errorf("unexpected order message: %v", m)
		}
		if p, _ := m["product"].(map[string]any); p["symbol"] != "MERV - XMEV - GGAL - 24hs" || p["marketId"] != "ROFX" {
			t.Errorf("product: %v", m["product"])
		}
	}
}

func TestCancelOrderWS_Message(t *testing.T) {
	url, msgs := wsCapture(t)
	c, _ := NewClient(WithWSURL(url), WithStaticToken("tok"))
	if err := c.CancelOrderWS(context.Background(), "ID1", ""); err != nil {
		t.Fatal(err)
	}
	m := nextMsg(t, msgs)
	if m["type"] != "co" || m["clientId"] != "ID1" || m["proprietary"] != "PBCP" {
		t.Errorf("unexpected cancel message: %v", m)
	}
}
