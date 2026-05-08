package main

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/carvalab/rofex-go/rofex"
	"github.com/carvalab/rofex-go/rofex/model"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	user := os.Getenv("PRIMARY_USER")
	pass := os.Getenv("PRIMARY_PASS")
	envVar := os.Getenv("PRIMARY_ENV")
	// Config por código (simplificado)
	var symbol string
	qty := int64(1) // cantidad
	price := 10.0   // precio para limit

	if user == "" || pass == "" {
		slog.Error("PRIMARY_USER and PRIMARY_PASS are required")
		return
	}

	env := model.EnvironmentRemarket
	if strings.EqualFold(envVar, "live") {
		env = model.EnvironmentLive
	}

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
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	_ = ctx // Las operaciones WebSocket usan contextos internos

	intr, err := c.InstrumentsAll(ctx)
	if err != nil {
		slog.Error("instruments", slog.Any("err", err))
		os.Exit(1)
	}
	for _, it := range intr.Instruments {
		s := it.InstrumentID.Symbol
		if s == "" {
			continue
		}
		if strings.HasPrefix(s, "DLR/") {
			symbol = s
			break
		}
		if symbol == "" && !strings.Contains(s, " ") {
			symbol = s
		}
	}
	if symbol == "" {
		slog.Error("no symbol available")
		os.Exit(1)
	}
	slog.Info("using symbol", slog.String("symbol", symbol))

	// Obtener la primera cuenta desde la API
	accounts, err := c.Accounts(ctx)
	if err != nil {
		slog.Error("accounts", slog.Any("err", err))
		return
	}
	if len(accounts.Accounts) == 0 {
		slog.Error("no accounts found; please verify credentials")
		return
	}
	account := accounts.Accounts[0].Name

	// Side y Tipo usando constantes
	sd := model.Buy
	ot := model.OrderTypeLimit
	var pricePtr *float64 = &price

	ord := rofex.NewOrder{
		Symbol:  symbol,
		Market:  model.MarketROFEX,
		Side:    sd,
		Type:    ot,
		Qty:     qty,
		Price:   pricePtr,
		TIF:     model.Day,
		Account: account,
	}
	if err := c.SendOrderWS(context.Background(), ord); err != nil {
		slog.Error("send ws order", slog.Any("err", err))
		return
	}
	slog.Info("ws order sent",
		slog.String("symbol", symbol),
		slog.Int64("qty", qty),
		slog.String("side", string(sd)),
		slog.String("type", string(ot)),
		slog.String("account", account))
}
