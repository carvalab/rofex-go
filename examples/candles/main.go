package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/carvalab/rofex-go/rofex"
	"github.com/carvalab/rofex-go/rofex/model"
	"github.com/joho/godotenv"
)

// Este ejemplo obtiene velas OHLCV oficiales a partir de trades históricos.

func main() {
	ctx := context.Background()

	// Cargar variables de entorno desde .env si está presente.
	_ = godotenv.Load()

	// Configuración básica del cliente.
	// - En LIVE, el proveedor cambia. Como ejemplo usamos Eco Valores.
	env := os.Getenv("PRIMARY_ENV")
	opts := []rofex.Option{}
	if env == "live" || env == "LIVE" {
		opts = append(opts,
			rofex.WithEnvironment(model.EnvironmentLive),
			rofex.WithBaseURL("https://api.eco.xoms.com.ar/"),
			rofex.WithWSURL("wss://api.eco.xoms.com.ar/"),
		)
	} else {
		opts = append(opts, rofex.WithEnvironment(model.EnvironmentRemarket))
	}
	opts = append(opts, rofex.WithAuth(rofex.NewPasswordAuth(rofex.Credentials{
		Username: os.Getenv("PRIMARY_USER"),
		Password: os.Getenv("PRIMARY_PASS"),
	})))

	client, err := rofex.NewClient(opts...)
	if err != nil {
		log.Fatal(err)
	}

	// Parámetros de ejemplo (1 minuto). Usamos UTC para alinear con timestamps ...Z del JSON.
	to := time.Now().UTC()
	from := to.AddDate(0, 0, -2)
	res := model.Resolution1m

	// 1) Candles para un instrumento ROFEX (interno)
	rofexSymbol := "GGAL/AGO25"
	rofexMarket := model.MarketROFEX
	candlesROFEX, err := client.HistoricCandles(ctx, rofexSymbol, rofexMarket, from, to, res)
	if err != nil {
		log.Fatalf("candles ROFEX (%s): %v", rofexSymbol, err)
	}
	fmt.Printf("# ROFEX %s\n", rofexSymbol)
	fmt.Println("timestamp,open,high,low,close,volume")
	for _, c := range candlesROFEX {
		fmt.Printf("%s,%.4f,%.4f,%.4f,%.4f,%.4f\n", c.Time.Local().Format(time.RFC3339), c.Open, c.High, c.Low, c.Close, c.Volume)
	}

	// 2) Candles para un instrumento externo MERV (ByMA)
	mervSymbol := "MERV - XMEV - ARKK - 24hs"
	mervMarket := model.MarketMERV
	candlesMERV, err := client.HistoricCandles(ctx, mervSymbol, mervMarket, from, to, res)
	if err != nil {
		log.Printf("candles MERV (%s): %v", mervSymbol, err)
		return
	}
	fmt.Println("--------------------------------")
	fmt.Printf("\n# MERV %s\n", mervSymbol)
	fmt.Println("timestamp,open,high,low,close,volume")
	for _, c := range candlesMERV {
		fmt.Printf("%s,%.4f,%.4f,%.4f,%.4f,%.4f\n", c.Time.Local().Format(time.RFC3339), c.Open, c.High, c.Low, c.Close, c.Volume)
	}
}
