package rofex_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/carvalab/rofex-go/rofex"
	"github.com/carvalab/rofex-go/rofex/model"
)

// Connect to the reMarkets sandbox with username and password.
func ExampleNewClient() {
	client, err := rofex.NewClient(
		rofex.WithEnvironment(model.EnvironmentRemarket),
		rofex.WithAuth(rofex.NewPasswordAuth(rofex.Credentials{
			Username: os.Getenv("PRIMARY_USER"),
			Password: os.Getenv("PRIMARY_PASS"),
		})),
	)
	if err != nil {
		log.Fatal(err)
	}

	accounts, err := client.Accounts(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, a := range accounts.Accounts {
		fmt.Println(a.Name)
	}
}

// Connect to production. LIVE URLs depend on your broker.
func ExampleNewClient_live() {
	client, err := rofex.NewClient(
		rofex.WithEnvironment(model.EnvironmentLive),
		rofex.WithBaseURL("https://api.eco.xoms.com.ar/"),
		rofex.WithWSURL("wss://api.eco.xoms.com.ar/"),
		rofex.WithAuth(rofex.NewPasswordAuth(rofex.Credentials{
			Username: os.Getenv("PRIMARY_USER"),
			Password: os.Getenv("PRIMARY_PASS"),
		})),
	)
	if err != nil {
		log.Fatal(err)
	}
	_ = client
}

// Get the top of book and last price for a BYMA stock.
func ExampleClient_MarketDataSnapshot() {
	var client *rofex.Client // see ExampleNewClient

	md, err := client.MarketDataSnapshot(context.Background(), rofex.MDRequest{
		Symbol:  "MERV - XMEV - GGAL - 24hs",
		Market:  model.MarketROFEX,
		Entries: []model.MDEntry{model.MDBids, model.MDOffers, model.MDLast},
		Depth:   1,
	})
	if err != nil {
		log.Fatal(err)
	}
	if len(md.MarketData.Bids) > 0 {
		fmt.Println("bid:", md.MarketData.Bids[0].Price, "x", md.MarketData.Bids[0].Size)
	}
	if len(md.MarketData.Offers) > 0 {
		fmt.Println("ask:", md.MarketData.Offers[0].Price, "x", md.MarketData.Offers[0].Size)
	}
	if md.MarketData.LA != nil && md.MarketData.LA.Price != nil {
		fmt.Println("last:", *md.MarketData.LA.Price)
	}
}

// Stream real-time market data over WebSocket.
func ExampleClient_SubscribeMarketData() {
	var client *rofex.Client // see ExampleNewClient

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	sub, err := client.SubscribeMarketData(ctx,
		[]string{"MERV - XMEV - GGAL - 24hs", "MERV - XMEV - AL30 - 24hs"},
		[]model.MDEntry{model.MDBids, model.MDOffers, model.MDLast},
		1, model.MarketROFEX,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Close()

	for {
		select {
		case ev := <-sub.Events:
			fmt.Println(ev.InstrumentID.Symbol, ev.MarketData.Bids, ev.MarketData.Offers)
		case err := <-sub.Errs:
			log.Println("stream error:", err)
		case <-ctx.Done():
			return
		}
	}
}

// Send a limit buy order and check its status.
func ExampleClient_SendOrder() {
	var client *rofex.Client // see ExampleNewClient
	ctx := context.Background()

	price := 1000.0
	resp, err := client.SendOrder(ctx, rofex.NewOrder{
		Symbol:  "DLR/DIC26",
		Market:  model.MarketROFEX,
		Side:    model.Buy,
		Type:    model.OrderTypeLimit,
		Qty:     1,
		Price:   &price,
		TIF:     model.Day,
		Account: "REM1234",
	})
	if err != nil {
		log.Fatal(err)
	}

	status, err := client.OrderStatus(ctx, resp.Order.ClientID, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(status.Order.Status)
}
