package rofex

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/carvalab/rofex-go/rofex/model"
)

func ms(s string) int64 {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t.UnixMilli()
}

// Candles are built from raw trades: open/close must follow trade time (not the
// order the API returns them in), and trades outside [from, to] are excluded.
func TestHistoricCandles_OHLCV(t *testing.T) {
	trades := []map[string]any{
		// returned out of order on purpose
		{"price": 105.0, "size": 2.0, "servertime": ms("2026-10-16T14:00:40Z")},
		{"price": 100.0, "size": 1.0, "servertime": ms("2026-10-16T14:00:05Z")}, // open of 14:00
		{"price": 110.0, "size": 3.0, "servertime": ms("2026-10-16T14:00:20Z")}, // high
		{"price": 98.0, "size": 4.0, "servertime": ms("2026-10-16T14:00:30Z")},  // low
		{"price": 120.0, "size": 5.0, "servertime": ms("2026-10-16T14:01:10Z")}, // single-trade candle
		{"price": 999.0, "size": 9.0, "servertime": ms("2026-10-16T13:59:59Z")}, // before from
		{"price": 999.0, "size": 9.0, "servertime": ms("2026-10-16T14:05:01Z")}, // after to
	}
	var query map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = map[string]string{}
		for k := range r.URL.Query() {
			query[k] = r.URL.Query().Get(k)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "OK", "trades": trades})
	}))
	defer ts.Close()

	c, _ := NewClient(WithBaseURL(ts.URL+"/"), WithStaticToken("tok"))
	from, _ := time.Parse(time.RFC3339, "2026-10-16T14:00:00Z")
	to, _ := time.Parse(time.RFC3339, "2026-10-16T14:05:00Z")

	got, err := c.HistoricCandles(context.Background(), "DLR/DIC26", "", from, to, model.Resolution1m)
	if err != nil {
		t.Fatal(err)
	}
	want := []model.OHLCV{
		{Open: 100, High: 110, Low: 98, Close: 105, Volume: 10, Time: from},
		{Open: 120, High: 120, Low: 120, Close: 120, Volume: 5, Time: from.Add(time.Minute)},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d candles, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		g := got[i]
		if g.Open != w.Open || g.High != w.High || g.Low != w.Low || g.Close != w.Close || g.Volume != w.Volume || !g.Time.Equal(w.Time) {
			t.Errorf("candle %d = %+v, want %+v", i, g, w)
		}
	}

	// Sandbox trades live in a separate history; ROFX is the default market.
	if query["environment"] != "REMARKETS" || query["marketId"] != "ROFX" || query["dateFrom"] != "2026-10-16" {
		t.Errorf("trades query: %v", query)
	}
}

// BYMA (MERV) instruments are "external" to MATBA ROFEX history.
func TestHistoricTrades_ExternalMarket(t *testing.T) {
	var external string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		external = r.URL.Query().Get("external")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "OK"})
	}))
	defer ts.Close()

	c, _ := NewClient(WithBaseURL(ts.URL+"/"), WithStaticToken("tok"))
	now := time.Now()
	if _, err := c.HistoricTrades(context.Background(), "GGAL - 24hs", model.MarketMERV, now, now); err != nil {
		t.Fatal(err)
	}
	if external != "true" {
		t.Errorf("MERV trades: external=%q, want true", external)
	}
}

func TestFloorTimeByResolution(t *testing.T) {
	ts := func(s string) time.Time {
		v, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	tests := []struct {
		res  model.CandleResolution
		in   string
		want string
	}{
		{model.Resolution5m, "2026-10-16T14:07:59Z", "2026-10-16T14:05:00Z"},
		{model.Resolution4h, "2026-10-16T15:30:00Z", "2026-10-16T12:00:00Z"},
		{model.Resolution1D, "2026-10-16T23:59:59Z", "2026-10-16T00:00:00Z"},
		{model.Resolution1W, "2026-10-18T10:00:00Z", "2026-10-12T00:00:00Z"}, // Sunday -> previous Monday
		{model.Resolution1W, "2026-10-12T00:00:00Z", "2026-10-12T00:00:00Z"}, // Monday stays
		{model.Resolution1M, "2026-10-31T23:00:00Z", "2026-10-01T00:00:00Z"},
		{model.Resolution3M, "2026-12-15T00:00:00Z", "2026-10-01T00:00:00Z"},
		{model.Resolution3M, "2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z"},
		// Non-UTC input is bucketed in UTC.
		{model.Resolution1D, "2026-10-16T22:00:00-03:00", "2026-10-17T00:00:00Z"},
	}
	for _, tt := range tests {
		if got := floorTimeByResolution(ts(tt.in), tt.res); !got.Equal(ts(tt.want)) {
			t.Errorf("floor(%s, %s) = %s, want %s", tt.in, tt.res, got.Format(time.RFC3339), tt.want)
		}
	}
}
