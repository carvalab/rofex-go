package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/carvalab/rofex-go/rofex"
	"github.com/carvalab/rofex-go/rofex/model"
	"github.com/joho/godotenv"
)

// Ejemplo básico: cuenta, datos de referencia, instrumentos, snapshot de market data y trades
func main() {
	_ = godotenv.Load()

	user := os.Getenv("PRIMARY_USER")
	pass := os.Getenv("PRIMARY_PASS")
	envVar := os.Getenv("PRIMARY_ENV")

	// Config por código
	symbol := "GGAL/AGO25"

	if user == "" || pass == "" {
		slog.Error("PRIMARY_USER and PRIMARY_PASS are required")
		os.Exit(1)
	}

	env := model.EnvironmentRemarket
	if strings.EqualFold(envVar, "live") {
		env = model.EnvironmentLive
	}

	// Construir opciones; si es LIVE, setear URLs de Eco Valores por defecto
	opts := []rofex.Option{
		rofex.WithEnvironment(env),
		rofex.WithAuth(rofex.NewPasswordAuth(rofex.Credentials{Username: user, Password: pass})),
	}
	if env == model.EnvironmentLive {
		opts = append(opts,
			rofex.WithBaseURL("https://api.eco.xoms.com.ar/"),
			rofex.WithWSURL("wss://api.eco.xoms.com.ar/"),
		)
	}
	c, err := rofex.NewClient(opts...)
	if err != nil {
		slog.Error("new client", slog.Any("err", err))
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	// 1) Cuenta: se obtiene la primera cuenta del usuario desde /rest/accounts
	accounts, err := c.Accounts(ctx)
	if err != nil {
		slog.Error("accounts", slog.Any("err", err))
		os.Exit(1)
	}
	if len(accounts.Accounts) == 0 {
		slog.Error("no accounts found; please verify credentials")
		os.Exit(1)
	}
	account := accounts.Accounts[0].Name
	pos, err := c.AccountPosition(ctx, account)
	if err != nil {
		slog.Error("account position", slog.Any("err", err))
		os.Exit(1)
	}
	b, _ := json.MarshalIndent(pos, "", "  ")
	fmt.Println("Account Position (" + account + "):\n" + string(b))

	// Reporte de cuenta: disponible total y montos por moneda (>0)
	if rep, err := c.AccountReport(ctx, account); err == nil {
		lastKey := ""
		for k := range rep.AccountData.DetailedAccountReports {
			if k > lastKey {
				lastKey = k
			}
		}
		if lastKey != "" {
			dar := rep.AccountData.DetailedAccountReports[lastKey]
			fmt.Printf("Disponible total: %.2f\n", dar.AvailableToOperate.Total)
			for cur, ca := range dar.CurrencyBalance.DetailedCurrencyBalance {
				if ca.Available > 0 || ca.Consumed > 0 {
					fmt.Printf("- %s: available=%.2f, consumed=%.2f\n", cur, ca.Available, ca.Consumed)
				}
			}
			for cur, val := range dar.AvailableToOperate.Cash.DetailedCash {
				if val > 0 {
					fmt.Printf("- %s: cash=%.2f\n", cur, val)
				}
			}
		}
	}

	waitContinue("Continuar a Datos de Referencia (Segments)")

	// 2) Datos de referencia (segmentos)
	seg, err := c.Segments(ctx)
	if err != nil {
		slog.Error("segments", slog.Any("err", err))
		os.Exit(1)
	}
	bSeg, _ := json.MarshalIndent(seg, "", "  ")
	fmt.Println("Segments:\n" + string(bSeg))

	waitContinue("Continuar a Instruments por CFICode")

	// 2.1) Instrumentos por CFICode (ej.:  CEDEAR)
	byCFI, err := c.InstrumentsByCFICode(ctx, []model.CFICode{model.CFICedear})
	if err != nil {
		slog.Error("instruments by CFICode", slog.Any("err", err))
		os.Exit(1)
	}
	bByCFI, _ := json.MarshalIndent(byCFI, "", "  ")
	fmt.Println("Instruments by CFICode (EMXXXX):\n" + string(bByCFI))

	// 2.1.a) Filtro simple: CEDEARs en pesos (solo "24hs").
	// Agrupa por base (quita sufijo final C/D si existe) y prefiere el símbolo sin sufijo.
	getBaseIf24hs := func(sym string) (base string, isBase bool, ok bool) {
		parts := strings.Split(sym, " - ")
		// Formato esperado: ... - <CODE> - 24hs
		if len(parts) < 2 || parts[len(parts)-1] != "24hs" {
			return "", false, false
		}
		code := parts[len(parts)-2]
		isBase = !(strings.HasSuffix(code, "C") || strings.HasSuffix(code, "D"))
		if !isBase && len(code) > 1 {
			code = code[:len(code)-1]
		}
		return code, isBase, true
	}

	type chosen struct{ sym string; base bool }
	selected := map[string]chosen{}
	for _, it := range byCFI.Instruments {
		s := it.InstrumentID.Symbol
		if s == "" {
			s = it.SymbolAlt
		}
		b, isBase, ok := getBaseIf24hs(s)
		if !ok {
			continue
		}
		prev, exists := selected[b]
		if !exists || (isBase && !prev.base) {
			selected[b] = chosen{sym: s, base: isBase}
		}
	}

	if len(selected) > 0 {
		fmt.Println("Filtered CEDEARs (base, 24hs):")
		for b, ch := range selected {
			fmt.Printf("- %s -> %s\n", b, ch.sym)
		}
	}

	waitContinue("Continuar a Instruments All")

	// 3) Instrumentos: listado completo
	intr, err := c.InstrumentsAll(ctx)
	if err != nil {
		slog.Error("instruments all", slog.Any("err", err))
		os.Exit(1)
	}
	bIntr, _ := json.MarshalIndent(intr, "", "  ")
	fmt.Println("Instruments All:\n" + string(bIntr))

	waitContinue("Continuar a Snapshot de Market Data")

	// 4) Snapshot de Market Data
	md, err := c.MarketDataSnapshot(ctx, rofex.MDRequest{
		Symbol:  symbol,
		Market:  model.MarketROFEX,
		Entries: []model.MDEntry{model.MDBids, model.MDOffers, model.MDLast},
		Depth:   2,
	})
	if err != nil {
		slog.Error("market data", slog.Any("err", err))
		os.Exit(1)
	}
	bMD, _ := json.MarshalIndent(md, "", "  ")
	fmt.Println("Market Data:\n" + string(bMD))

	waitContinue("Continuar a Trades (últimas 24 horas)")

	// 5) Trades (últimas 24 horas)
	from := time.Now().AddDate(0, 0, -1)
	to := time.Now()
	tr, err := c.HistoricTrades(ctx, symbol, model.MarketROFEX, from, to)
	if err != nil {
		slog.Error("trades", slog.Any("err", err))
		os.Exit(1)
	}
	bTR, _ := json.MarshalIndent(tr, "", "  ")
	fmt.Println("Trades:\n" + string(bTR))
}

// waitContinue pausa hasta que el usuario presione Enter para continuar.
func waitContinue(msg string) {
	reader := bufio.NewReader(os.Stdin)
	if msg == "" {
		msg = "Presioná Enter para continuar"
	}
	fmt.Printf("\n➡️  %s... ", msg)
	_, _ = reader.ReadString('\n')
	fmt.Println()
}
