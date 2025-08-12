// WebSocket unificado: Market Data + Order Reports
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/carvalab/rofex-go/rofex"
	"github.com/carvalab/rofex-go/rofex/model"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// Sólo se leen del entorno: usuario, contraseña y (opcional) depth
	user := os.Getenv("PRIMARY_USER")
	pass := os.Getenv("PRIMARY_PASS")
	envVar := os.Getenv("PRIMARY_ENV")
	depthVar := os.Getenv("PRIMARY_DEPTH")

	if user == "" || pass == "" {
		slog.Error("PRIMARY_USER and PRIMARY_PASS are required")
		os.Exit(1)
	}
	// Símbolos y entradas definidos en código
	symbols := []string{"DLR/ENE26", "GGAL/AGO25", "MERV - XMEV - GD30 - 24hs"}
	entries := []model.MDEntry{model.MDBids, model.MDOffers, model.MDLast, model.MDOpeningPrice, model.MDClosePrice, model.MDSettlementPrice, model.MDHighPrice, model.MDLowPrice, model.MDTradeVolume, model.MDOpenInterest}

	env := model.EnvironmentRemarket
	if strings.EqualFold(envVar, "live") {
		env = model.EnvironmentLive
	}

	// Si es LIVE, setear URLs Eco por defecto
	wsOpts := []rofex.Option{
		rofex.WithEnvironment(env),
		rofex.WithAuth(rofex.NewPasswordAuth(rofex.Credentials{Username: user, Password: pass})),
	}
	if env == model.EnvironmentLive {
		wsOpts = append(wsOpts,
			rofex.WithBaseURL("https://api.eco.xoms.com.ar/"),
			rofex.WithWSURL("wss://api.eco.xoms.com.ar/"),
		)
	}
	client, err := rofex.NewClient(wsOpts...)
	if err != nil {
		slog.Error("new client", slog.Any("err", err))
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	depth := 1
	if d, err := strconv.Atoi(strings.TrimSpace(depthVar)); err == nil && d > 0 {
		depth = d
	}

	// Obtener la cuenta desde el endpoint (/rest/accounts)
	accountsResp, err := client.Accounts(ctx)
	if err != nil {
		slog.Error("no se pudo obtener la cuenta", slog.Any("err", err))
		os.Exit(1)
	}
	if len(accountsResp.Accounts) == 0 {
		slog.Error("no hay cuentas disponibles para el usuario")
		os.Exit(1)
	}
	account := accountsResp.Accounts[0].Name

	mdSub, err := client.SubscribeMarketData(ctx, symbols, entries, depth, model.MarketROFEX)
	if err != nil {
		slog.Error("subscribe market data", slog.Any("err", err))
		os.Exit(1)
	}
	defer mdSub.Close()

	orSub, err := client.SubscribeOrderReport(ctx, account, true)
	if err != nil {
		slog.Error("subscribe order report", slog.Any("err", err))
		os.Exit(1)
	}
	defer orSub.Close()

	// Suscripciones WebSocket iniciadas
	slog.Info("suscripciones websocket iniciadas",
		slog.String("symbols", strings.Join(symbols, ",")),
		slog.Any("entries", entries),
		slog.Int("depth", depth),
		slog.String("environment", string(env)),
		slog.String("account", account),
	)

	for {
		select {
		case event := <-mdSub.Events:
			if event != nil {
				// Mostrar Market Data en formato tabla para lectura rápida
				printMarketDataTable(event)
			}
		case err := <-mdSub.Errs:
			if err != nil {
				// Error del stream de Market Data
				slog.Error("error websocket market data", slog.Any("err", err))
			} else {
				// Cierre limpio de la conexión de Market Data
				slog.Info("conexión market data cerrada")
				return
			}
		case event := <-orSub.Events:
			if event != nil {
				fmt.Printf("📋 OR %s: %s\n", event.OrderReport.ClOrdID, event.OrderReport.Status)
			}
		case err := <-orSub.Errs:
			if err != nil {
				// Error del stream de Order Reports
				slog.Error("error websocket order report", slog.Any("err", err))
			} else {
				// Cierre limpio de la conexión de Order Reports
				slog.Info("conexión order report cerrada")
				return
			}
		case <-sig:
			// Señal del SO recibida: cerrar ejemplo
			slog.Info("señal recibida, cerrando")
			return
		}
	}
}

// printMarketDataTable imprime el Market Data en una tabla legible.
func printMarketDataTable(event *model.MarketDataEvent) {
	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	tw.SetStyle(table.StyleLight)
	tw.Style().Options.SeparateRows = true
	tw.Style().Format.Header = text.FormatUpper
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Name: "ENTRY", Align: text.AlignLeft},
		{Name: "PRICE", Align: text.AlignRight},
		{Name: "SIZE", Align: text.AlignRight},
		{Name: "DATE", Align: text.AlignLeft},
	})
	tw.AppendHeader(table.Row{"ENTRY", "PRICE", "SIZE", "DATE"})

	loc := time.Local

	// Helper para fechas (Unix ms -> hora legible en loc)
	formatMillis := func(ms *int64) string {
		if ms == nil {
			return ""
		}
		t := time.Unix((*ms)/1000, ((*ms)%1000)*int64(time.Millisecond)).In(loc)
		return t.Format("2006-01-02 15:04:05")
	}

	md := event.MarketData

	// BI (Compras)
	for i, lvl := range md.Bids {
		tw.AppendRow(table.Row{labelLevel("Compras", i, len(md.Bids)), fmtNumber(lvl.Price), fmtNumber(lvl.Size), ""})
	}
	// OF (Ventas)
	for i, lvl := range md.Offers {
		tw.AppendRow(table.Row{labelLevel("Ventas", i, len(md.Offers)), fmtNumber(lvl.Price), fmtNumber(lvl.Size), ""})
	}
	// LA (Último)
	if md.LA != nil {
		tw.AppendRow(table.Row{"Último", fmtNumberPtr(md.LA.Price), fmtNumberPtr(md.LA.Size), formatMillis(md.LA.Date)})
	}
	// OP (Apertura)
	if md.OpeningPrice != nil {
		tw.AppendRow(table.Row{"Apertura", fmtNumber(*md.OpeningPrice), "", ""})
	}
	// CL (Cierre)
	if md.CL != nil {
		tw.AppendRow(table.Row{"Cierre", fmtNumberPtr(md.CL.Price), fmtNumberPtr(md.CL.Size), formatMillis(md.CL.Date)})
	}
	// SE (Ajuste)
	if md.SE != nil {
		tw.AppendRow(table.Row{"Ajuste", fmtNumberPtr(md.SE.Price), fmtNumberPtr(md.SE.Size), formatMillis(md.SE.Date)})
	}
	if md.HighPrice != nil {
		tw.AppendRow(table.Row{"Máximo", fmtNumber(*md.HighPrice), "", ""})
	}
	if md.LowPrice != nil {
		tw.AppendRow(table.Row{"Mínimo", fmtNumber(*md.LowPrice), "", ""})
	}
	if md.TradeVolume != nil {
		tw.AppendRow(table.Row{"Volumen", "", fmtNumber(*md.TradeVolume), ""})
	}
	if md.OpenInterest != nil {
		tw.AppendRow(table.Row{"Interés Abierto", fmtNumberPtr(md.OpenInterest.Price), fmtNumberPtr(md.OpenInterest.Size), formatMillis(md.OpenInterest.Date)})
	}
	if md.IndexValue != nil {
		tw.AppendRow(table.Row{"Índice", fmtNumber(*md.IndexValue), "", ""})
	}
	if md.EffectiveVolume != nil {
		tw.AppendRow(table.Row{"Volumen Efectivo", "", fmtNumber(*md.EffectiveVolume), ""})
	}
	if md.NominalVolume != nil {
		tw.AppendRow(table.Row{"Volumen Nominal", "", fmtNumber(*md.NominalVolume), ""})
	}
	if md.ACP != nil {
		tw.AppendRow(table.Row{"Subasta", fmtNumber(*md.ACP), "", ""})
	}
	if md.TradeCount != nil {
		tw.AppendRow(table.Row{"Trades", "", fmtNumber(*md.TradeCount), ""})
	}

	// Título: símbolo y hora del evento (en loc)
	ts := ""
	if event.Timestamp != nil {
		ts = formatMillis(event.Timestamp)
	}
	fmt.Printf("\n📈 %s @ %s\n", event.InstrumentID.Symbol, ts)
	tw.Render()
}

// fmtNumber formatea números: sin decimales si es entero, con 2 decimales si no.
func fmtNumber[N ~float64 | ~int | ~int64](v N) string {
	switch any(v).(type) {
	case float64:
		f := float64(v)
		if f == float64(int64(f)) {
			return fmt.Sprintf("%d", int64(f))
		}
		return fmt.Sprintf("%.2f", f)
	case int:
		return strconv.Itoa(any(v).(int))
	case int64:
		return strconv.FormatInt(any(v).(int64), 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func fmtNumberPtr[N ~float64 | ~int | ~int64](p *N) any {
	if p == nil {
		return ""
	}
	return fmtNumber(*p)
}

// labelLevel devuelve una etiqueta amigable, agregando índice solo si hay múltiples niveles.
func labelLevel(base string, idx, total int) string {
	if total > 1 {
		return fmt.Sprintf("%s (%d)", base, idx+1)
	}
	return base
}
