package rofex

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/carvalab/rofex-go/rofex/model"
)

func TestSendOrder_PathParams_UppercaseAndOptionals(t *testing.T) {
	// Constantes del test
	okStatus := "OK"
	symbol := "DLR/DIC23"
	account := "REM6771"
	market := string(model.MarketROFEX)
	orderQty := int64(100)
	price := 210.5
	displayQty := int64(5)

	// Servidor de prueba
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/rest/order/newSingleOrder") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		// Validación de parámetros según docs/primary-api.md
		mustEq(t, q, "marketId", market)
		mustEq(t, q, "symbol", symbol)
		mustEq(t, q, "orderQty", "100")
		mustEq(t, q, "ordType", string(model.OrderTypeLimit))
		mustEq(t, q, "side", string(model.Buy))
		mustEq(t, q, "timeInForce", string(model.Day))
		mustEq(t, q, "account", account)
		mustEq(t, q, "cancelPrevious", "false")
		mustEq(t, q, "price", "210.5")
		mustEq(t, q, "iceberg", "true")
		mustEq(t, q, "displayQty", "5")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": okStatus,
			"order": map[string]any{
				"clientId":    "user125469825632595",
				"proprietary": "PBCP",
			},
		})
	}))
	defer ts.Close()

	c, err := NewClient(WithBaseURL(ts.URL + "/"))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	order := NewOrder{
		Symbol:         symbol,
		Market:         model.MarketROFEX,
		Side:           model.Buy,
		Type:           model.OrderTypeLimit,
		Qty:            orderQty,
		Price:          &price,
		TIF:            model.Day,
		Account:        account,
		CancelPrevious: false,
		Iceberg:        true,
		DisplayQty:     &displayQty,
	}

	resp, err := c.SendOrder(ctx, order)
	if err != nil {
		t.Fatalf("SendOrder: %v", err)
	}
	if resp.Status != okStatus {
		t.Fatalf("status: want %s got %s", okStatus, resp.Status)
	}
}

func TestSendOrder_GTD_WithExpireDate(t *testing.T) {
	// Constantes
	okStatus := "OK"
	symbol := "DLR/DIC23"
	account := "REM2747"
	expire := "20230505"
	price := 182.5

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/rest/order/newSingleOrder") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		mustEq(t, q, "timeInForce", string(model.GoodTillDate))
		mustEq(t, q, "expireDate", expire)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": okStatus,
			"order":  map[string]any{"clientId": "utfa3256548752365489", "proprietary": "api"},
		})
	}))
	defer ts.Close()

	c, _ := NewClient(WithBaseURL(ts.URL + "/"))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.SendOrder(ctx, NewOrder{
		Symbol:     symbol,
		Market:     model.MarketROFEX,
		Side:       model.Buy,
		Type:       model.OrderTypeLimit,
		Qty:        100,
		Price:      &price,
		TIF:        model.GoodTillDate,
		ExpireDate: &expire,
		Account:    account,
	})
	if err != nil {
		t.Fatalf("SendOrder GTD: %v", err)
	}
	if resp.Status != okStatus {
		t.Fatalf("status: %s", resp.Status)
	}
}

func TestInstrumentDetailResponse_Decode(t *testing.T) {
	// Constantes
	okStatus := "OK"
	symbol := "DLR/NOV23"
	market := string(model.MarketROFEX)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verificar query string
		q := r.URL.Query()
		mustEq(t, q, "symbol", symbol)
		mustEq(t, q, "marketId", market)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": okStatus,
			"instrument": map[string]any{
				"segment":                  map[string]any{"marketSegmentId": "DDF", "marketId": "ROFX"},
				"lowLimitPrice":            321,
				"highLimitPrice":           370,
				"minPriceIncrement":        0.05,
				"minTradeVol":              1,
				"maxTradeVol":              10000,
				"tickSize":                 1,
				"contractMultiplier":       1000,
				"roundLot":                 1,
				"priceConvertionFactor":    1,
				"maturityDate":             "20231130",
				"currency":                 "ARS",
				"orderTypes":               []any{"STOP LIMIT", "MARKET TO LIMIT", "MARKET", "LIMIT"},
				"timesInForce":             []any{"IOC", "DAY", "GTD"},
				"instrumentPricePrecision": 2,
				"instrumentSizePrecision":  0,
				"securityDescription":      symbol,
				"tickPriceRanges":          map[string]any{"0": map[string]any{"lowerLimit": 0, "upperLimit": nil, "tick": 0.05}},
				"cficode":                  "FXXXSX",
				"instrumentId":             map[string]any{"marketId": market, "symbol": symbol},
			},
		})
	}))
	defer ts.Close()

	c, _ := NewClient(WithBaseURL(ts.URL + "/"))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.InstrumentDetail(ctx, symbol, model.MarketROFEX)
	if err != nil {
		t.Fatalf("InstrumentDetail: %v", err)
	}
	if res.Status != okStatus {
		t.Fatalf("status: %s", res.Status)
	}
	if res.Instrument.InstrumentID.Symbol != symbol {
		t.Fatalf("symbol decode")
	}
}

func TestInstrumentsByCFICode_TopLevelID_Decode(t *testing.T) {
	// Constantes
	okStatus := "OK"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		mustEq(t, q, "CFICode", string(model.CFIFuture))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": okStatus,
			"instruments": []any{
				map[string]any{"marketId": "ROFX", "symbol": "DLR/DIC23"},
				map[string]any{"marketId": "ROFX", "symbol": "TRI.ROS/DIC23"},
			},
		})
	}))
	defer ts.Close()

	c, _ := NewClient(WithBaseURL(ts.URL + "/"))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.InstrumentsByCFICode(ctx, []model.CFICode{model.CFIFuture})
	if err != nil {
		t.Fatalf("InstrumentsByCFICode: %v", err)
	}
	if res.Status != okStatus && res.Status != "" {
		t.Fatalf("status: %s", res.Status)
	}
	if len(res.Instruments) == 0 {
		t.Fatalf("no instruments")
	}
	if res.Instruments[0].InstrumentID.MarketID != string(model.MarketROFEX) {
		t.Fatalf("instrumentId fallback not set")
	}
}

func TestMarketDataSnapshot_Decode_LA_CL_SE_Object(t *testing.T) {
	// Constantes
	okStatus := "OK"
	symbol := "DLR/DIC23"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		// entries=BI,OF,LA,OP,CL,SE,OI&depth=2
		if q.Get("symbol") == "" || q.Get("entries") == "" {
			t.Fatalf("missing query")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": okStatus,
			"marketData": map[string]any{
				"SE": map[string]any{"price": 180.3, "size": nil, "date": 1669852800000},
				"LA": map[string]any{"price": 179.85, "size": 4, "date": 1669995044232},
				"OI": map[string]any{"price": nil, "size": 217596, "date": 1664150400000},
				"OF": []any{map[string]any{"price": 179.8, "size": 1000}},
				"OP": 180.35,
				"CL": map[string]any{"price": 180.35, "size": nil, "date": 1669852800000},
				"BI": []any{map[string]any{"price": 179.75, "size": 275}},
			},
			"depth":        2,
			"aggregated":   true,
			"instrumentId": map[string]any{"marketId": "ROFX", "symbol": symbol},
		})
	}))
	defer ts.Close()

	c, _ := NewClient(WithBaseURL(ts.URL + "/"))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.MarketDataSnapshot(ctx, MDRequest{
		Symbol:  symbol,
		Market:  model.MarketROFEX,
		Entries: []model.MDEntry{model.MDBids, model.MDOffers, model.MDLast, model.MDOpeningPrice, model.MDClosePrice, model.MDSettlementPrice, model.MDOpenInterest},
		Depth:   2,
	})
	if err != nil {
		t.Fatalf("MarketDataSnapshot: %v", err)
	}
	if res.Status != okStatus {
		t.Fatalf("status: %s", res.Status)
	}
	if res.MarketData.LA == nil || res.MarketData.LA.Price == nil {
		t.Fatalf("LA not decoded as object")
	}
	if res.MarketData.CL == nil || res.MarketData.SE == nil {
		t.Fatalf("CL/SE not decoded as object")
	}
}

func TestRisk_GetPositions_Typed(t *testing.T) {
	// Constantes
	okStatus := "OK"
	account := "REM7374"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/rest/risk/position/getPositions/") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": okStatus,
			"positions": []any{
				map[string]any{
					"instrument": map[string]any{"symbolReference": "AAPL", "settlType": 0},
					"symbol":     "MERV XMEV AAPL CI",
					"buySize":    5, "buyPrice": 3092,
					"sellSize": 0, "sellPrice": 0,
					"totalDailyDiff": 0, "totalDiff": 15460,
					"tradingSymbol":    "MERV XMEV AAPL CI",
					"originalBuyPrice": 0, "originalSellPrice": 0,
				},
			},
		})
	}))
	defer ts.Close()

	c, _ := NewClient(WithBaseURL(ts.URL + "/"))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.AccountPosition(ctx, account)
	if err != nil {
		t.Fatalf("AccountPosition: %v", err)
	}
	if res.Status != okStatus {
		t.Fatalf("status: %s", res.Status)
	}
	if len(res.Positions) == 0 {
		t.Fatalf("no positions")
	}
	if res.Positions[0].Instrument.SymbolReference == "" {
		t.Fatalf("instrument.symbolReference empty")
	}
}

func TestRisk_AccountReport_Typed(t *testing.T) {
	// Constantes
	okStatus := "OK"
	account := "REM7374"

	// Payload reducido del doc para validar estructura principal
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": okStatus,
			"accountData": map[string]any{
				"accountName":           account,
				"marketMember":          "PrimaryVenture",
				"marketMemberIdentity":  "PMYVTR",
				"collateral":            0,
				"margin":                2923811.299985,
				"availableToCollateral": 100202251.700015,
				"detailedAccountReports": map[string]any{
					"0": map[string]any{
						"currencyBalance": map[string]any{
							"detailedCurrencyBalance": map[string]any{"ARS": map[string]any{"consumed": 0, "available": 100000000}},
						},
						"availableToOperate": map[string]any{
							"cash": map[string]any{
								"totalCash":    103250600,
								"detailedCash": map[string]any{"ARS": 100000000},
							},
							"movements":        0,
							"credit":           nil,
							"total":            103065823,
							"pendingMovements": 0,
						},
						"settlementDate": 1669950000000,
					},
				},
				"hasError":        false,
				"errorDetail":     nil,
				"lastCalculation": 1669996836647,
				"portfolio":       60240,
				"ordersMargin":    0,
				"currentCash":     103065823,
				"dailyDiff":       -184777,
				"uncoveredMargin": 0,
			},
		})
	}))
	defer ts.Close()

	c, _ := NewClient(WithBaseURL(ts.URL + "/"))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.AccountReport(ctx, account)
	if err != nil {
		t.Fatalf("AccountReport: %v", err)
	}
	if res.Status != okStatus {
		t.Fatalf("status: %s", res.Status)
	}
	if res.AccountData.AccountName != account {
		t.Fatalf("accountName mismatch")
	}
	if len(res.AccountData.DetailedAccountReports) == 0 {
		t.Fatalf("empty detailedAccountReports")
	}
}

// Helper asserts
func mustEq(t *testing.T, q url.Values, key, want string) {
	t.Helper()
	if got := q.Get(key); got != want {
		t.Fatalf("query %s: want %s got %s", key, want, got)
	}
}
