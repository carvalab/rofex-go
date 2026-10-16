package rofex

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/carvalab/rofex-go/rofex/model"
)

// jsonServer answers every request with body and records the last request URL.
func jsonServer(t *testing.T, status int, body string) (*Client, *url.URL) {
	t.Helper()
	var last url.URL
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		last = *r.URL
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(ts.Close)
	c, _ := NewClient(WithBaseURL(ts.URL+"/"), WithStaticToken("tok"))
	return c, &last
}

// Primary reports business errors with HTTP 200 and "status":"ERROR". They must
// surface as errors so `if err != nil` catches rejected orders. Bodies below
// were captured from the LIVE API.
func TestAPIError_StatusErrorWithHTTP200(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name     string
		body     string
		call     func(c *Client) error
		wantMsg  string
		wantDesc string
	}{
		{"rejected order",
			`{"status":"ERROR","message":"Invalid Instrument.","description":"Product NOPE/XX99:ROFX doesn't exist."}`,
			func(c *Client) error {
				_, err := c.SendOrder(ctx, NewOrder{Symbol: "NOPE/XX99", Side: model.Buy, Type: model.OrderTypeLimit, Qty: 1, Price: ptr(1.0), TIF: model.Day, Account: "X"})
				return err
			},
			"Invalid Instrument.", "Product NOPE/XX99:ROFX doesn't exist."},
		{"account without access",
			`{"status":"ERROR","message":"Access Denied.","description":"You don't have access to account NOACCOUNT"}`,
			func(c *Client) error { _, err := c.AccountPosition(ctx, "NOACCOUNT"); return err },
			"Access Denied.", "You don't have access to account NOACCOUNT"},
		{"unknown symbol, null message",
			`{"status":"ERROR","message":null,"description":"Security NOPE/XX99:ROFX doesn't exist"}`,
			func(c *Client) error { _, err := c.MarketDataSnapshot(ctx, MDRequest{Symbol: "NOPE/XX99"}); return err },
			"", "Security NOPE/XX99:ROFX doesn't exist"},
		{"unexpected message shape is still an error",
			`{"status":"ERROR","message":{"code":42}}`,
			func(c *Client) error { _, err := c.Segments(ctx); return err },
			`{"code":42}`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := jsonServer(t, http.StatusOK, tt.body)
			err := tt.call(c)
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("want *APIError, got %v", err)
			}
			if apiErr.Message != tt.wantMsg || apiErr.Description != tt.wantDesc {
				t.Errorf("got %+v", apiErr)
			}
			if !strings.Contains(err.Error(), "primary api error") {
				t.Errorf("error text: %q", err.Error())
			}
		})
	}
}

// Non-JSON success bodies (proxy pages, truncated responses) must not be
// mistaken for an empty successful response.
func TestGetTyped_TransportErrors(t *testing.T) {
	ctx := context.Background()

	c, _ := jsonServer(t, http.StatusOK, `<html>maintenance</html>`)
	if _, err := c.Segments(ctx); err == nil || !strings.Contains(err.Error(), "decode json") {
		t.Errorf("HTML with 200: want decode error, got %v", err)
	}

	c, _ = jsonServer(t, http.StatusBadGateway, `bad gateway`)
	_, err := c.Segments(ctx)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusBadGateway || !strings.Contains(err.Error(), "status=502") {
		t.Errorf("502: want HTTPError, got %v", err)
	}
}

const orderJSON = `{"orderId":"1130835","clOrdId":"user1145712381052053","proprietary":"PBCP","execId":"160229133429-fix1-493",
"accountId":{"id":"10"},"instrumentId":{"marketId":"ROFX","symbol":"DLR/DIC23"},"price":183,"orderQty":10,"ordType":"LIMIT",
"side":"BUY","timeInForce":"DAY","transactTime":"20160304-17:37:35","avgPx":182.5,"lastPx":182,"lastQty":4,"cumQty":4,"leavesQty":6,
"status":"PARTIALLY_FILLED","text":"Aceptada"}`

// Every order query must hit its endpoint with the right identifiers and
// decode the fill state (cumQty/leavesQty/avgPx) that trading logic relies on.
func TestOrderQueries(t *testing.T) {
	ctx := context.Background()
	single := `{"status":"OK","order":` + orderJSON + `}`
	list := `{"status":"OK","orders":[{"orderId":null,"clOrdId":"user1145712381052053","status":"PENDING_NEW","text":"Enviada"},` + orderJSON + `]}`

	type result struct {
		orders []model.Order
		err    error
	}
	one := func(r model.OrderStatusResponse, err error) result { return result{[]model.Order{r.Order}, err} }
	many := func(r model.AllOrdersStatusResponse, err error) result { return result{r.Orders, err} }

	tests := []struct {
		name      string
		body      string
		call      func(c *Client) result
		wantPath  string
		wantQuery url.Values
	}{
		{"status by clOrdId", single, func(c *Client) result { return one(c.OrderStatus(ctx, "user1145712381052053", "")) },
			"/rest/order/id", url.Values{"clOrdId": {"user1145712381052053"}, "proprietary": {"PBCP"}}},
		{"history by clOrdId", list, func(c *Client) result { return many(c.OrderHistoryByClOrdID(ctx, "user1145712381052053", "api")) },
			"/rest/order/allById", url.Values{"clOrdId": {"user1145712381052053"}, "proprietary": {"api"}}},
		{"by orderId", single, func(c *Client) result { return one(c.OrderByOrderID(ctx, "1130835")) },
			"/rest/order/byOrderId", url.Values{"orderId": {"1130835"}}},
		{"by execId", single, func(c *Client) result { return one(c.OrderByExecID(ctx, "160229133429-fix1-493")) },
			"/rest/order/byExecId", url.Values{"execId": {"160229133429-fix1-493"}}},
		{"filled by account", list, func(c *Client) result { return many(c.FilledOrders(ctx, "REM1")) },
			"/rest/order/filleds", url.Values{"accountId": {"REM1"}}},
		{"active by account", list, func(c *Client) result { return many(c.ActiveOrders(ctx, "REM1")) },
			"/rest/order/actives", url.Values{"accountId": {"REM1"}}},
		{"all by account", list, func(c *Client) result { return many(c.AllOrdersStatus(ctx, "REM1")) },
			"/rest/order/all", url.Values{"accountId": {"REM1"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, last := jsonServer(t, http.StatusOK, tt.body)
			res := tt.call(c)
			if res.err != nil {
				t.Fatal(res.err)
			}
			if last.Path != tt.wantPath {
				t.Errorf("path %s, want %s", last.Path, tt.wantPath)
			}
			for k := range tt.wantQuery {
				if got := last.Query().Get(k); got != tt.wantQuery.Get(k) {
					t.Errorf("query %s=%q, want %q", k, got, tt.wantQuery.Get(k))
				}
			}
			o := res.orders[len(res.orders)-1]
			if o.OrderID != "1130835" || o.Status != "PARTIALLY_FILLED" || *o.CumQty != 4 || *o.LeavesQty != 6 ||
				*o.AvgPx != 182.5 || *o.Price != 183 || o.Side != model.Buy || o.InstrumentID.Symbol != "DLR/DIC23" || o.AccountID.ID != "10" {
				t.Errorf("decoded order: %+v", o)
			}
		})
	}

	// Empty identifiers are rejected locally instead of querying "all orders".
	c, _ := jsonServer(t, http.StatusOK, single)
	for name, err := range map[string]error{
		"OrderStatus":    func() error { _, err := c.OrderStatus(ctx, "", ""); return err }(),
		"OrderHistory":   func() error { _, err := c.OrderHistoryByClOrdID(ctx, "", ""); return err }(),
		"OrderByOrderID": func() error { _, err := c.OrderByOrderID(ctx, ""); return err }(),
		"OrderByExecID":  func() error { _, err := c.OrderByExecID(ctx, ""); return err }(),
		"FilledOrders":   func() error { _, err := c.FilledOrders(ctx, ""); return err }(),
		"ActiveOrders":   func() error { _, err := c.ActiveOrders(ctx, ""); return err }(),
		"AllOrders":      func() error { _, err := c.AllOrdersStatus(ctx, ""); return err }(),
		"CancelOrder":    func() error { _, err := c.CancelOrder(ctx, "", ""); return err }(),
		"ReplaceOrder":   func() error { _, err := c.ReplaceOrder(ctx, "", "", nil, nil); return err }(),
		"Position":       func() error { _, err := c.AccountPosition(ctx, ""); return err }(),
		"Detailed":       func() error { _, err := c.DetailedPosition(ctx, ""); return err }(),
		"Report":         func() error { _, err := c.AccountReport(ctx, ""); return err }(),
	} {
		if err == nil || !strings.Contains(err.Error(), "validation") {
			t.Errorf("%s with empty id: want validation error, got %v", name, err)
		}
	}
}

// Detailed positions drive P&L: sizes, prices and daily differences per
// contract type and symbol (sample from docs/primary-api.md).
func TestDetailedPosition_Decode(t *testing.T) {
	c, last := jsonServer(t, http.StatusOK, `{"status":"OK","detailedPosition":{"account":"REM7374","totalDailyDiffPlain":-184777,
"totalMarketValue":60240,"lastCalculation":1669994833000,"report":{"FUTURE OPTION CALL":{"SOJ.ROS/MAY23 380 C":{
"detailedPositions":[{"symbolReference":"SOJ.ROS/MAY23 380 C","contractType":"FUTURE_OPTION_CALL","marketPrice":380.5,
"currency":"USD G","totalCurrentSize":-2,"sellInitialPrice":383,"detailedDailyDiff":{"totalDailyDiffPlain":-16777}}],
"instrumentInitialSize":-2,"instrumentFilledSize":0,"instrumentCurrentSize":-2}}}}}`)

	res, err := c.DetailedPosition(context.Background(), "REM7374")
	if err != nil {
		t.Fatal(err)
	}
	if last.Path != "/rest/risk/detailedPosition/REM7374" {
		t.Errorf("path %s", last.Path)
	}
	dp := res.DetailedPosition
	inst, ok := dp.Report["FUTURE OPTION CALL"]["SOJ.ROS/MAY23 380 C"]
	if dp.Account != "REM7374" || dp.TotalDailyDiffPlain != -184777 || dp.TotalMarketValue != 60240 || !ok ||
		inst.InstrumentCurrentSize != -2 || len(inst.DetailedPositions) != 1 {
		t.Fatalf("decoded: %+v", dp)
	}
	if p := inst.DetailedPositions[0]; p.MarketPrice != 380.5 || p.TotalCurrentSize != -2 || p.DetailedDailyDiff.TotalDailyDiffPlain != -16777 {
		t.Errorf("position: %+v", p)
	}
}

// Instrument listings come in two shapes (detailed with instrumentId, or flat
// marketId/symbol); both must yield a usable InstrumentID. Multi-segment
// queries must merge every segment.
func TestInstrumentListings(t *testing.T) {
	ctx := context.Background()
	body := `{"status":"OK","instruments":[
{"instrumentId":{"marketId":"ROFX","symbol":"DLR/DIC26"},"cficode":"FXXXSX","segment":{"marketSegmentId":"DDF","marketId":"ROFX"},"minPriceIncrement":0.5,"maturityDate":"20261231"},
{"marketId":"ROFX","symbol":"MERV - XMEV - GGAL - 24hs"}]}`

	for name, call := range map[string]func(c *Client) (model.InstrumentsResponse, error){
		"all":     func(c *Client) (model.InstrumentsResponse, error) { return c.InstrumentsAll(ctx) },
		"details": func(c *Client) (model.InstrumentsResponse, error) { return c.InstrumentsDetails(ctx) },
	} {
		c, _ := jsonServer(t, http.StatusOK, body)
		res, err := call(c)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(res.Instruments) != 2 || res.Instruments[0].InstrumentID.Symbol != "DLR/DIC26" ||
			*res.Instruments[0].MinPriceIncrement != 0.5 || res.Instruments[1].InstrumentID.Symbol != "MERV - XMEV - GGAL - 24hs" ||
			res.Instruments[1].InstrumentID.MarketID != "ROFX" {
			t.Errorf("%s decoded: %+v", name, res.Instruments)
		}
	}

	var segs []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seg := r.URL.Query().Get("MarketSegmentID")
		segs = append(segs, seg+"@"+r.URL.Query().Get("MarketID"))
		_, _ = w.Write([]byte(`{"status":"OK","instruments":[{"marketId":"ROFX","symbol":"` + seg + `-1"}]}`))
	}))
	defer ts.Close()
	c, _ := NewClient(WithBaseURL(ts.URL+"/"), WithStaticToken("tok"))
	res, err := c.InstrumentsBySegment(ctx, model.MarketROFEX, []model.MarketSegment{"DDF", "DDA"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(segs, ",") != "DDF@ROFX,DDA@ROFX" || len(res.Instruments) != 2 || res.Instruments[1].InstrumentID.Symbol != "DDA-1" {
		t.Errorf("segments queried %v, got %+v", segs, res.Instruments)
	}
}
