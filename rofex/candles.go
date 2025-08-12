package rofex

import (
	"context"
	"sort"
	"time"

	"github.com/carvalab/rofex-go/rofex/model"
)

// HistoricCandles obtiene velas OHLCV agregadas a partir de trades históricos.
//
// Parámetros:
//   - symbol: Símbolo del instrumento (ej.: "DLR/DIC23")
//   - market: Mercado (por defecto model.MarketROFEX si viene vacío)
//   - from, to: Rango temporal. Se recomienda usar UTC. Se filtrará por servertime del trade
//   - resolution: Resolución de la vela (1,5,15,30,1h,4h,D,W,M,3M)
//
// Devuelve una serie de velas ordenadas por tiempo ascendente.
// La agregación usa el timestamp del servidor (servertime, milisegundos) y bucketiza en UTC.
func (c *Client) HistoricCandles(
	ctx context.Context,
	symbol string,
	market model.Market,
	from, to time.Time,
	resolution model.CandleResolution,
) ([]model.OHLCV, error) {
	if symbol == "" {
		return nil, &ValidationError{Field: "symbol", Msg: "required"}
	}
	if market == "" {
		market = model.MarketROFEX
	}
	// Obtener trades del rango (API por fecha). Luego filtramos por timestamp exacto.
	tr, err := c.HistoricTrades(ctx, symbol, market, from, to)
	if err != nil {
		return nil, err
	}

	fromUTC := from.UTC()
	toUTC := to.UTC()

	type bucket struct {
		open, high, low, close float64
		vol                    float64
		t                      time.Time
		init                   bool
	}

	byTime := make(map[time.Time]*bucket)
	times := make([]time.Time, 0)

	// Para garantizar open/cierre correctos por bucket, iteramos en orden cronológico.
	// Ordenamos por servertime ascendente.
	idx := make([]int, len(tr.Trades))
	for i := range tr.Trades {
		idx[i] = i
	}
	sort.Slice(idx, func(i, j int) bool {
		ti := tr.Trades[idx[i]].ServerTime
		tj := tr.Trades[idx[j]].ServerTime
		if ti == tj {
			return idx[i] < idx[j]
		}
		return ti < tj
	})

	for _, k := range idx {
		trow := tr.Trades[k]
		ts := time.Unix(trow.ServerTime/1000, (trow.ServerTime%1000)*int64(time.Millisecond)).UTC()
		if ts.Before(fromUTC) || ts.After(toUTC) {
			continue
		}
		key := floorTimeByResolution(ts, resolution)
		b := byTime[key]
		if b == nil {
			b = &bucket{open: trow.Price, high: trow.Price, low: trow.Price, close: trow.Price, vol: trow.Size, t: key, init: true}
			byTime[key] = b
			times = append(times, key)
			continue
		}
		if trow.Price > b.high {
			b.high = trow.Price
		}
		if trow.Price < b.low {
			b.low = trow.Price
		}
		b.close = trow.Price
		b.vol += trow.Size
	}

	sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
	out := make([]model.OHLCV, 0, len(times))
	for _, t0 := range times {
		b := byTime[t0]
		// Filtrar velas no reales (todos los campos en cero)
		if b.open == 0 && b.high == 0 && b.low == 0 && b.close == 0 && b.vol == 0 {
			continue
		}
		out = append(out, model.OHLCV{
			Open:       b.open,
			High:       b.high,
			Low:        b.low,
			Close:      b.close,
			Volume:     b.vol,
			Time:       t0,
			Resolution: string(resolution),
			SecurityID: symbol,
		})
	}
	return out, nil
}

// floorTimeByResolution devuelve el comienzo del bucket para el timestamp dado
// en función de la resolución. Todo en UTC.
func floorTimeByResolution(t time.Time, r model.CandleResolution) time.Time {
	t = t.UTC()
	switch r {
	case model.Resolution1m:
		return t.Truncate(time.Minute)
	case model.Resolution5m:
		return t.Truncate(5 * time.Minute)
	case model.Resolution15m:
		return t.Truncate(15 * time.Minute)
	case model.Resolution30m:
		return t.Truncate(30 * time.Minute)
	case model.Resolution1h:
		return t.Truncate(time.Hour)
	case model.Resolution4h:
		return t.Truncate(4 * time.Hour)
	case model.Resolution1D:
		y, m, d := t.Date()
		return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	case model.Resolution1W:
		// Truncar al lunes 00:00:00 UTC de la semana ISO
		// Go: Weekday() devuelve Sunday=0...Saturday=6. Usamos Monday=1.
		wd := int(t.Weekday())
		if wd == 0 { // Sunday
			wd = 7
		}
		// Retroceder (wd-1) días hasta lunes
		monday := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(wd - 1))
		return monday
	case model.Resolution1M:
		y, m, _ := t.Date()
		return time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
	case model.Resolution3M:
		y, m, _ := t.Date()
		// Trimestre que contiene m: Q1 (1), Q2 (4), Q3 (7), Q4 (10)
		qm := time.Month(((int(m)-1)/3)*3 + 1)
		return time.Date(y, qm, 1, 0, 0, 0, 0, time.UTC)
	default:
		// Fallback seguro: 1 minuto
		return t.Truncate(time.Minute)
	}
}
