package rofex

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/carvalab/rofex-go/rofex/model"
)

// helper to join entries to comma list for query
func joinEntries(entries []model.MDEntry) string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, string(e))
	}
	return strings.Join(out, ",")
}

// Segments obtiene la lista de segmentos disponibles según Primary API.
//
// Los segmentos son los distintos ambientes o ruedas de negociación en las que
// está organizada la operatoria de MATBA ROFEX.
//
// Segmentos disponibles:
//   - DDF: Instrumentos de la División Derivados Financieros
//   - DDA: Instrumentos de la División Derivados Agropecuarios
//   - DUAL: Instrumentos listados en ambas divisiones
//   - MERV: Instrumentos de mercados externos a Matba Rofex
//
// Referencia: docs/primary-api.md - "Lista de Segmentos disponibles"
func (c *Client) Segments(ctx context.Context) (model.SegmentsResponse, error) {
	return getTyped[model.SegmentsResponse](ctx, c, pathSegments)
}

// InstrumentsAll obtiene todos los instrumentos disponibles según Primary API.
//
// Devuelve una lista con todos los instrumentos disponibles para negociarse
// en MATBA ROFEX. Por cada instrumento devuelve el símbolo, ID del mercado
// y el código CFI del instrumento.
//
// Referencia: docs/primary-api.md - "Lista de Segmentos disponibles" (instrumentos)
func (c *Client) InstrumentsAll(ctx context.Context) (model.InstrumentsResponse, error) {
	return getTyped[model.InstrumentsResponse](ctx, c, pathInstrAll)
}

// InstrumentsDetails obtiene instrumentos con detalles completos según Primary API.
//
// Similar al método anterior pero agrega una descripción detallada de cada
// instrumento. Devuelve datos de segmento, precio mínimo/máximo, vencimiento, etc.
//
// Referencia: docs/primary-api.md - "Lista detallada de Instrumentos disponibles"
func (c *Client) InstrumentsDetails(ctx context.Context) (model.InstrumentsResponse, error) {
	return getTyped[model.InstrumentsResponse](ctx, c, pathInstrDetails)
}

// InstrumentDetail obtiene la descripción detallada de un instrumento específico.
//
// Devuelve información completa de un solo instrumento incluyendo:
//   - Límites de precio (lowLimitPrice, highLimitPrice)
//   - Incrementos mínimos (minPriceIncrement, tickSize)
//   - Volúmenes de trading (minTradeVol, maxTradeVol)
//   - Información del contrato (contractMultiplier, maturityDate)
//   - Tipos de orden y tiempos de vida soportados
//
// Referencia: docs/primary-api.md - "Descripción detallada de un Instrumento"
func (c *Client) InstrumentDetail(ctx context.Context, symbol string, market model.Market) (model.InstrumentDetailResponse, error) {
	escSymbol := url.QueryEscape(symbol)
	path := fmt.Sprintf(pathInstrDetail, escSymbol, string(market))
	return getTyped[model.InstrumentDetailResponse](ctx, c, path)
}

// InstrumentsByCFICode obtiene instrumentos filtrados por código CFI.
//
// Permite listar todos los instrumentos que pertenezcan al mismo tipo
// identificado por código CFI.
//
// Códigos CFI disponibles:
//   - ESXXXX: Acción
//   - DBXXXX: Bono
//   - OCASPS: Opción Call sobre Acción
//   - OPASPS: Opción Put sobre Acción
//   - FXXXSX: Futuro
//   - OPAFXS: Opción Put sobre Futuro
//   - OCAFXS: Opción Call sobre Futuro
//   - EMXXXX: CEDEAR
//   - DBXXFR: Obligaciones Negociables
//
// Referencia: docs/primary-api.md - "Lista de Instrumentos por Código CFI"
func (c *Client) InstrumentsByCFICode(ctx context.Context, codes []model.CFICode) (model.InstrumentsResponse, error) {
	agg := model.InstrumentsResponse{}
	for _, code := range codes {
		path := fmt.Sprintf(pathInstrByCFI, string(code))
		res, err := getTyped[model.InstrumentsResponse](ctx, c, path)
		if err != nil {
			return model.InstrumentsResponse{}, err
		}
		agg.Instruments = append(agg.Instruments, res.Instruments...)
	}
	return agg, nil
}

// InstrumentsBySegment obtiene instrumentos filtrados por segmento de mercado.
//
// Permite listar todos los instrumentos que pertenezcan al mismo segmento
// de mercado.
//
// Segmentos disponibles: DDF, DDA, DUAL, U-DDF, U-DDA, U-DUAL, MERV
//
// Referencia: docs/primary-api.md - "Lista de Instrumentos por Segmento"
func (c *Client) InstrumentsBySegment(ctx context.Context, market model.Market, segs []model.MarketSegment) (model.InstrumentsResponse, error) {
	agg := model.InstrumentsResponse{}
	for _, seg := range segs {
		path := fmt.Sprintf(pathInstrBySeg, string(seg), string(market))
		res, err := getTyped[model.InstrumentsResponse](ctx, c, path)
		if err != nil {
			return model.InstrumentsResponse{}, err
		}
		agg.Instruments = append(agg.Instruments, res.Instruments...)
	}
	return agg, nil
}
