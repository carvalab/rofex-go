package rofex

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/carvalab/rofex-go/rofex/model"
	"github.com/coder/websocket"
)

// fakePrimaryWS simulates the Primary WebSocket endpoint. For each connection
// it records the subscription message and then runs script(n, conn), where n
// is the 1-based connection number.
type fakePrimaryWS struct {
	conns   atomic.Int32
	subs    chan map[string]any
	script  func(n int32, ctx context.Context, conn *websocket.Conn)
	wsURL   string
	httpURL string
}

func newFakePrimaryWS(t *testing.T, script func(n int32, ctx context.Context, conn *websocket.Conn)) *fakePrimaryWS {
	t.Helper()
	f := &fakePrimaryWS{subs: make(chan map[string]any, 16), script: script}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()
		n := f.conns.Add(1)
		_, b, err := conn.Read(r.Context())
		if err != nil {
			return
		}
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		f.subs <- m
		// Like the real server, keep reading so close frames are answered.
		f.script(n, conn.CloseRead(r.Context()), conn)
	}))
	t.Cleanup(ts.Close)
	f.httpURL = ts.URL
	f.wsURL = "ws" + strings.TrimPrefix(ts.URL, "http") + "/"
	return f
}

func send(ctx context.Context, conn *websocket.Conn, s string) {
	_ = conn.Write(ctx, websocket.MessageText, []byte(s))
}

func mdMsg(symbol string, bid float64) string {
	b, _ := json.Marshal(bid)
	return `{"type":"Md","instrumentId":{"marketId":"ROFX","symbol":"` + symbol + `"},"marketData":{"BI":[{"price":` +
		string(b) + `,"size":10}]}}`
}

func waitEvent[T any](t *testing.T, ch <-chan *T) *T {
	t.Helper()
	select {
	case ev, ok := <-ch:
		if !ok {
			t.Fatal("events channel closed")
		}
		return ev
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for event")
		return nil
	}
}

// Only "md" messages reach Events. Other server messages (errors, other types)
// must not show up as empty quotes.
func TestSubscribeMarketData_ForwardsOnlyMarketData(t *testing.T) {
	f := newFakePrimaryWS(t, func(n int32, ctx context.Context, conn *websocket.Conn) {
		send(ctx, conn, `{"status":"ERROR","description":"invalid symbol"}`)
		send(ctx, conn, `{"type":"or","orderReport":{"clOrdId":"x"}}`)
		send(ctx, conn, mdMsg("DLR/DIC26", 1500.5))
		<-ctx.Done()
	})
	c, _ := NewClient(WithWSURL(f.wsURL), WithStaticToken("tok"))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := c.SubscribeMarketData(ctx, []string{"DLR/DIC26"}, []model.MDEntry{model.MDBids}, 2, "")
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Close()

	m := <-f.subs
	if m["type"] != "smd" || m["depth"] != 2.0 {
		t.Errorf("subscription message: %v", m)
	}
	if p := m["products"].([]any)[0].(map[string]any); p["symbol"] != "DLR/DIC26" || p["marketId"] != "ROFX" {
		t.Errorf("products: %v", m["products"])
	}

	ev := waitEvent(t, sub.Events)
	if ev.InstrumentID.Symbol != "DLR/DIC26" || len(ev.MarketData.Bids) != 1 || ev.MarketData.Bids[0].Price != 1500.5 {
		t.Fatalf("first event is not the md message: %+v", ev)
	}
}

// When the server drops the connection (network error or server restart), the
// subscription reconnects and re-subscribes on its own.
func TestSubscribeMarketData_ReconnectsAndResubscribes(t *testing.T) {
	for _, tc := range []struct {
		name string
		drop func(conn *websocket.Conn)
	}{
		{"abrupt drop", func(conn *websocket.Conn) { _ = conn.CloseNow() }},
		{"server restart (going away)", func(conn *websocket.Conn) { _ = conn.Close(websocket.StatusGoingAway, "restart") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFakePrimaryWS(t, func(n int32, ctx context.Context, conn *websocket.Conn) {
				send(ctx, conn, mdMsg("DLR/DIC26", float64(n)))
				if n == 1 {
					tc.drop(conn)
					return
				}
				<-ctx.Done()
			})
			c, _ := NewClient(WithWSURL(f.wsURL), WithStaticToken("tok"))
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sub, err := c.SubscribeMarketData(ctx, []string{"DLR/DIC26"}, []model.MDEntry{model.MDBids}, 1, model.MarketROFEX)
			if err != nil {
				t.Fatal(err)
			}
			defer sub.Close()

			if ev := waitEvent(t, sub.Events); ev.MarketData.Bids[0].Price != 1 {
				t.Fatalf("first event from connection %v", ev.MarketData.Bids[0].Price)
			}
			if ev := waitEvent(t, sub.Events); ev.MarketData.Bids[0].Price != 2 {
				t.Fatalf("second event from connection %v", ev.MarketData.Bids[0].Price)
			}
			<-f.subs
			if m := <-f.subs; m["type"] != "smd" {
				t.Errorf("re-subscription message: %v", m)
			}
		})
	}
}

// Close stops the subscription for good: channels close and no reconnect happens.
func TestSubscribeMarketData_CloseStops(t *testing.T) {
	f := newFakePrimaryWS(t, func(n int32, ctx context.Context, conn *websocket.Conn) {
		send(ctx, conn, mdMsg("DLR/DIC26", 1))
		<-ctx.Done()
	})
	c, _ := NewClient(WithWSURL(f.wsURL), WithStaticToken("tok"))

	sub, err := c.SubscribeMarketData(context.Background(), []string{"DLR/DIC26"}, nil, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	waitEvent(t, sub.Events)
	if err := sub.Close(); err != nil {
		t.Fatal(err)
	}

	deadline := time.After(5 * time.Second)
	for {
		select {
		case _, ok := <-sub.Events:
			if !ok {
				time.Sleep(1500 * time.Millisecond) // longer than the first reconnect backoff
				if n := f.conns.Load(); n != 1 {
					t.Fatalf("reconnected after Close: %d connections", n)
				}
				return
			}
		case <-deadline:
			t.Fatal("events channel not closed after Close")
		}
	}
}

func TestSubscribeOrderReport_SubscribesAccountAndForwardsReports(t *testing.T) {
	f := newFakePrimaryWS(t, func(n int32, ctx context.Context, conn *websocket.Conn) {
		send(ctx, conn, `{"type":"Or","orderReport":{"clOrdId":"ID1","status":"FILLED"}}`)
		<-ctx.Done()
	})
	c, _ := NewClient(WithWSURL(f.wsURL), WithStaticToken("tok"))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := c.SubscribeOrderReport(ctx, "REM1", true)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Close()

	m := <-f.subs
	if acc, _ := m["account"].(map[string]any); m["type"] != "os" || acc["id"] != "REM1" || m["snapshotOnlyActive"] != true {
		t.Errorf("subscription message: %v", m)
	}
	ev := waitEvent(t, sub.Events)
	if ev.OrderReport.ClOrdID != "ID1" {
		t.Errorf("order report: %+v", ev.OrderReport)
	}
}

// Subscribing before any REST call must log in first and use that token.
// The order report timestamp (ms) must convert to the right instant.
func TestSubscribe_LogsInWithPasswordAuth(t *testing.T) {
	var gotToken atomic.Value
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/getToken", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Auth-Token", "fresh-token")
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		gotToken.Store(r.Header.Get("X-Auth-Token"))
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()
		ctx := r.Context()
		_, _, _ = conn.Read(ctx)
		ctx = conn.CloseRead(ctx)
		send(ctx, conn, `{"type":"or","timestamp":1760630400123,"orderReport":{"clOrdId":"ID1"}}`)
		<-ctx.Done()
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	c, _ := NewClient(WithBaseURL(ts.URL+"/"), WithWSURL("ws"+strings.TrimPrefix(ts.URL, "http")+"/"),
		WithAuth(NewPasswordAuth(Credentials{Username: "u", Password: "p"})))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sub, err := c.SubscribeOrderReport(ctx, "REM1", false)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Close()

	ev := waitEvent(t, sub.Events)
	if gotToken.Load() != "fresh-token" {
		t.Errorf("ws token = %v", gotToken.Load())
	}
	if ht := ev.HumanTime(); ht == nil || ht.UnixMilli() != 1760630400123 {
		t.Errorf("HumanTime = %v", ht)
	}
}

func waitErr(t *testing.T, ch <-chan error) error {
	t.Helper()
	select {
	case err := <-ch:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for error")
		return nil
	}
}

func waitClosed[T any](t *testing.T, ch <-chan *T) {
	t.Helper()
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("unexpected event")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("events channel not closed")
	}
}

// Wrong credentials end the subscription with a visible error instead of
// retrying forever.
func TestSubscribe_AuthFailureIsReported(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	c, _ := NewClient(WithBaseURL(ts.URL+"/"), WithWSURL("ws"+strings.TrimPrefix(ts.URL, "http")+"/"),
		WithAuth(NewPasswordAuth(Credentials{Username: "u", Password: "bad"})))
	sub, err := c.SubscribeMarketData(context.Background(), []string{"DLR/DIC26"}, nil, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	var httpErr *HTTPError
	if err := waitErr(t, sub.Errs); !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("want 401 error, got %v", err)
	}
	waitClosed(t, sub.Events)
}

// If the server is down when the bot starts, the subscription reports the
// failure and keeps retrying until it can connect.
func TestSubscribe_RetriesUntilServerIsUp(t *testing.T) {
	var attempts atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()
		_, _, _ = conn.Read(r.Context())
		ctx := conn.CloseRead(r.Context())
		send(ctx, conn, mdMsg("DLR/DIC26", 1))
		<-ctx.Done()
	}))
	defer ts.Close()

	c, _ := NewClient(WithWSURL("ws"+strings.TrimPrefix(ts.URL, "http")+"/"), WithStaticToken("tok"))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sub, err := c.SubscribeMarketData(ctx, []string{"DLR/DIC26"}, nil, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Close()

	if err := waitErr(t, sub.Errs); !strings.Contains(err.Error(), "dial failed") {
		t.Errorf("want dial error, got %v", err)
	}
	waitEvent(t, sub.Events)
}

// A normal close from the server (session ended) is reported and not retried.
func TestSubscribe_ServerEndsSession(t *testing.T) {
	f := newFakePrimaryWS(t, func(n int32, ctx context.Context, conn *websocket.Conn) {
		_ = conn.Close(websocket.StatusNormalClosure, "session ended")
	})
	c, _ := NewClient(WithWSURL(f.wsURL), WithStaticToken("tok"))
	sub, err := c.SubscribeMarketData(context.Background(), []string{"DLR/DIC26"}, nil, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := waitErr(t, sub.Errs); websocket.CloseStatus(err) != websocket.StatusNormalClosure {
		t.Errorf("want normal closure error, got %v", err)
	}
	waitClosed(t, sub.Events)
	if n := f.conns.Load(); n != 1 {
		t.Errorf("reconnected %d times after session end", n-1)
	}
}

// With drop-on-full, a slow consumer loses old quotes but never stalls the
// feed: the next quote still arrives.
func TestSubscribe_SlowConsumerDropOnFull(t *testing.T) {
	more := make(chan struct{})
	f := newFakePrimaryWS(t, func(n int32, ctx context.Context, conn *websocket.Conn) {
		for i := range 20 {
			send(ctx, conn, mdMsg("DLR/DIC26", float64(i)))
		}
		<-more
		send(ctx, conn, mdMsg("DLR/DIC26", 999))
		<-ctx.Done()
	})
	c, _ := NewClient(WithWSURL(f.wsURL), WithStaticToken("tok"), WithWSBuffer(1), WithWSDropOnFull(true))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sub, err := c.SubscribeMarketData(ctx, []string{"DLR/DIC26"}, nil, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Close()

	time.Sleep(300 * time.Millisecond) // let the 20 quotes arrive while nobody reads
	if ev := waitEvent(t, sub.Events); ev.MarketData.Bids[0].Price != 0 {
		t.Errorf("buffered quote %v, want the first one", ev.MarketData.Bids[0].Price)
	}
	close(more)
	if ev := waitEvent(t, sub.Events); ev.MarketData.Bids[0].Price != 999 {
		t.Errorf("after draining got %v, want 999 (feed stalled or old quotes kept)", ev.MarketData.Bids[0].Price)
	}
}
