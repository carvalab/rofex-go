package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"
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
	account := os.Getenv("PRIMARY_ACCOUNT")

	symbol := os.Getenv("PRIMARY_ORDER_SYMBOL")
	sideVar := os.Getenv("PRIMARY_ORDER_SIDE") // buy|sell
	typeVar := os.Getenv("PRIMARY_ORDER_TYPE") // market|limit|market_to_limit
	qtyVar := os.Getenv("PRIMARY_ORDER_QTY")
	priceVar := os.Getenv("PRIMARY_ORDER_PRICE")

	if user == "" || pass == "" || account == "" || symbol == "" {
		slog.Error("PRIMARY_USER, PRIMARY_PASS, PRIMARY_ACCOUNT, PRIMARY_ORDER_SYMBOL are required")
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

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = ctx // Las operaciones WebSocket usan contextos internos

	sd := model.Buy
	if strings.EqualFold(sideVar, "sell") {
		sd = model.Sell
	}

	ot := model.OrderTypeMarket
	switch strings.ToLower(typeVar) {
	case "limit":
		ot = model.OrderTypeLimit
	case "market_to_limit":
		ot = model.OrderTypeMarketToLimit
	}

	qty := int64(1)
	if q, err := strconv.ParseInt(strings.TrimSpace(qtyVar), 10, 64); err == nil && q > 0 {
		qty = q
	}

	var pricePtr *float64
	if ot == model.OrderTypeLimit {
		if pv, err := strconv.ParseFloat(strings.TrimSpace(priceVar), 64); err == nil && pv > 0 {
			pricePtr = &pv
		} else {
			slog.Error("limit order requires PRIMARY_ORDER_PRICE")
			return
		}
	}

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
		slog.String("type", string(ot)))
}
