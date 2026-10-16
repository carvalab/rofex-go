//go:build integration

// Tests de integración contra la API real de Primary (read-only: no envía órdenes).
//
// Ejecutar:
//
//	PRIMARY_USER=... PRIMARY_PASS=... go test -tags=integration -v ./rofex/
//
// Variables de entorno:
//   - PRIMARY_USER / PRIMARY_PASS: credenciales (si faltan, los tests se saltean).
//   - PRIMARY_ENV: "live" o vacío (REMARKET por defecto).
//   - PRIMARY_BASE_URL / PRIMARY_WS_URL: URLs para LIVE (default Eco Valores).
//   - PRIMARY_IT_SYMBOL: símbolo para market data (opcional; se elige uno automáticamente).
package rofex

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/carvalab/rofex-go/rofex/model"
	"github.com/joho/godotenv"
)

func integrationClient(t *testing.T) (*Client, model.Environment) {
	t.Helper()
	_ = godotenv.Load("../.env")

	user, pass := os.Getenv("PRIMARY_USER"), os.Getenv("PRIMARY_PASS")
	if user == "" || pass == "" {
		t.Skip("PRIMARY_USER/PRIMARY_PASS not set; skipping integration test")
	}

	env := model.EnvironmentRemarket
	if strings.EqualFold(os.Getenv("PRIMARY_ENV"), "live") {
		env = model.EnvironmentLive
	}
	opts := []Option{
		WithEnvironment(env),
		WithAuth(NewPasswordAuth(Credentials{Username: user, Password: pass})),
		WithTimeout(30 * time.Second),
	}
	if env == model.EnvironmentLive {
		opts = append(opts,
			WithBaseURL(envOr("PRIMARY_BASE_URL", "https://api.eco.xoms.com.ar/")),
			WithWSURL(envOr("PRIMARY_WS_URL", "wss://api.eco.xoms.com.ar/")),
		)
	}
	c, err := NewClient(opts...)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	return c, env
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// pickSymbol elige un instrumento existente para pedir market data.
func pickSymbol(t *testing.T, ctx context.Context, c *Client, env model.Environment) model.InstrumentID {
	t.Helper()
	if s := os.Getenv("PRIMARY_IT_SYMBOL"); s != "" {
		return model.InstrumentID{Symbol: s, MarketID: string(model.MarketROFEX)}
	}
	all, err := c.InstrumentsAll(ctx)
	if err != nil {
		t.Fatalf("instruments all: %v", err)
	}
	if len(all.Instruments) == 0 {
		t.Fatal("instruments all: empty list")
	}
	prefs := []string{"DLR/"} // futuro de dólar: siempre listado en REMARKET
	if env == model.EnvironmentLive {
		prefs = []string{"MERV - XMEV - GGAL - 24hs", "MERV - XMEV - AL30 - 24hs", "DLR/"}
	}
	for _, p := range prefs {
		for _, in := range all.Instruments {
			s := in.InstrumentID.Symbol
			if s == p || (strings.HasSuffix(p, "/") && strings.HasPrefix(s, p) && !strings.Contains(s, " ")) {
				return in.InstrumentID
			}
		}
	}
	return all.Instruments[0].InstrumentID
}

func TestIntegration_AccountAndRisk(t *testing.T) {
	c, _ := integrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	accs, err := c.Accounts(ctx)
	if err != nil {
		t.Fatalf("accounts: %v", err)
	}
	if len(accs.Accounts) == 0 {
		t.Fatal("accounts: empty list")
	}
	acc := accs.Accounts[0].Name

	if _, err := c.AccountPosition(ctx, acc); err != nil {
		t.Errorf("account position: %v", err)
	}
	if _, err := c.DetailedPosition(ctx, acc); err != nil {
		t.Errorf("detailed position: %v", err)
	}
	if _, err := c.AccountReport(ctx, acc); err != nil {
		t.Errorf("account report: %v", err)
	}
	if _, err := c.AllOrdersStatus(ctx, acc); err != nil {
		t.Errorf("all orders status: %v", err)
	}
}

func TestIntegration_ReferenceData(t *testing.T) {
	c, env := integrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	segs, err := c.Segments(ctx)
	if err != nil {
		t.Fatalf("segments: %v", err)
	}
	if len(segs.Segments) == 0 {
		t.Error("segments: empty list")
	}

	id := pickSymbol(t, ctx, c, env)
	t.Logf("using symbol %q (%s)", id.Symbol, id.MarketID)

	det, err := c.InstrumentDetail(ctx, id.Symbol, model.Market(id.MarketID))
	if err != nil {
		t.Fatalf("instrument detail: %v", err)
	}
	if det.Instrument.InstrumentID.Symbol != id.Symbol {
		t.Errorf("instrument detail: got symbol %q, want %q", det.Instrument.InstrumentID.Symbol, id.Symbol)
	}
}

func TestIntegration_MarketData(t *testing.T) {
	c, env := integrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	id := pickSymbol(t, ctx, c, env)
	t.Logf("using symbol %q (%s)", id.Symbol, id.MarketID)
	entries := []model.MDEntry{model.MDBids, model.MDOffers, model.MDLast, model.MDClosePrice}

	snap, err := c.MarketDataSnapshot(ctx, MDRequest{
		Symbol: id.Symbol, Market: model.Market(id.MarketID), Entries: entries,
	})
	if err != nil {
		t.Fatalf("market data snapshot: %v", err)
	}
	if !strings.EqualFold(snap.Status, "OK") {
		t.Fatalf("market data snapshot: status %q", snap.Status)
	}

	// WebSocket: el servidor envía un primer mensaje "md" al suscribirse,
	// incluso con el mercado cerrado.
	sub, err := c.SubscribeMarketData(ctx, []string{id.Symbol}, entries, 1, model.Market(id.MarketID))
	if err != nil {
		t.Fatalf("subscribe market data: %v", err)
	}
	defer func() { _ = sub.Close() }()

	timeout := time.After(45 * time.Second)
	for {
		select {
		case ev := <-sub.Events:
			if ev == nil {
				continue
			}
			if ev.InstrumentID.Symbol != id.Symbol {
				t.Errorf("ws event: got symbol %q, want %q", ev.InstrumentID.Symbol, id.Symbol)
			}
			return
		case err := <-sub.Errs:
			t.Fatalf("ws error: %v", err)
		case <-timeout:
			t.Fatal("ws: no market data event within 45s")
		}
	}
}
