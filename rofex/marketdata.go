package rofex

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/carvalab/rofex-go/rofex/model"
)

// MDRequest defines a market data snapshot request.
type MDRequest struct {
	Symbol  string
	Market  model.Market
	Entries []model.MDEntry
	Depth   int
}

// MarketDataSnapshot obtiene datos de mercado en tiempo real según la documentación Primary API.
//
// Este método permite consultar Market Data en tiempo real para un instrumento específico.
// Soporta solicitar diferentes tipos de datos y profundidad del book de órdenes.
//
// Entries disponibles (docs/primary-api.md - "Descripción de MarketData Entries"):
//   - BI: Mejores ofertas de compra en el Book
//   - OF: Mejores ofertas de venta en el Book
//   - LA: Último precio operado
//   - OP: Precio de apertura
//   - CL: Precio de cierre de la rueda anterior
//   - SE: Precio de ajuste (solo futuros)
//   - HI: Precio máximo de la rueda
//   - LO: Precio mínimo de la rueda
//   - TV: Volumen operado (para instrumentos MATBA ROFEX)
//   - OI: Interés abierto (solo futuros)
//   - EV: Volumen efectivo (solo instrumentos ByMA)
//   - NV: Volumen nominal (solo instrumentos ByMA)
//   - ACP: Precio de cierre del día corriente
//
// El parámetro depth indica la profundidad del book (1-5), por defecto 1.
//
// Referencia: docs/primary-api.md - "MarketData en tiempo real a través de REST"
func (c *Client) MarketDataSnapshot(ctx context.Context, req MDRequest) (model.MarketDataSnapshotResponse, error) {
	if req.Symbol == "" {
		return model.MarketDataSnapshotResponse{}, &ValidationError{Field: "symbol", Msg: "required"}
	}
	if req.Market == "" {
		req.Market = model.MarketROFEX
	}
	if req.Depth <= 0 {
		req.Depth = 1
	}
	entries := joinEntries(req.Entries)
	path := fmt.Sprintf(pathMDGet, string(req.Market), req.Symbol, entries, req.Depth)
	return getTyped[model.MarketDataSnapshotResponse](ctx, c, path)
}

// HistoricTrades obtiene datos históricos de trades según la documentación Primary API.
//
// La API para acceder a datos históricos del mercado permite consultar los trades
// que se hayan realizado para un contrato en un rango de fechas específico.
//
// Parámetros:
//   - symbol: Símbolo del contrato (ej: "DLR/DIC23")
//   - market: Mercado (model.MarketROFEX para MATBA ROFEX)
//   - from: Fecha desde (se formatea como YYYY-MM-DD)
//   - to: Fecha hasta (se formatea como YYYY-MM-DD)
//
// Para instrumentos de mercados externos a MATBA ROFEX, usar el parámetro external=true
// en implementaciones futuras.
//
// Referencia: docs/primary-api.md - "MarketData Histórica"
func (c *Client) HistoricTrades(ctx context.Context, symbol string, market model.Market, from, to time.Time) (model.TradesResponse, error) {
	if symbol == "" {
		return model.TradesResponse{}, &ValidationError{Field: "symbol", Msg: "required"}
	}
	if market == "" {
		market = model.MarketROFEX
	}
	df := from.Format("2006-01-02")
	dt := to.Format("2006-01-02")
	// URL-encode symbol to safely handle spaces and special characters
	escSymbol := url.QueryEscape(symbol)
	path := fmt.Sprintf(pathTrades, string(market), escSymbol, df, dt)
	// If market is different from ROFEX (e.g., MERV/ByMA), add external=true
	if market != model.MarketROFEX {
		path += "&external=true"
	}
	// In REMARKET (sandbox), also append environment=REMARKETS as per docs
	if c.env == model.EnvironmentRemarket {
		path += "&environment=REMARKETS"
	}
	return getTyped[model.TradesResponse](ctx, c, path)
}
