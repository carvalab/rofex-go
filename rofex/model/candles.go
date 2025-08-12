package model

import "time"

type CandleResolution string

const (
	Resolution1m  CandleResolution = "1"
	Resolution5m  CandleResolution = "5"
	Resolution15m CandleResolution = "15"
	Resolution30m CandleResolution = "30"
	Resolution1h  CandleResolution = "1h"
	Resolution4h  CandleResolution = "4h"
	Resolution1D  CandleResolution = "D"
	Resolution1W  CandleResolution = "W"
	Resolution1M  CandleResolution = "M"
	Resolution3M  CandleResolution = "3M"
)

type OHLCV struct {
	Open       float64   `json:"o"`
	High       float64   `json:"h"`
	Low        float64   `json:"l"`
	Close      float64   `json:"c"`
	Volume     float64   `json:"v"`
	Time       time.Time `json:"d"`
	Resolution string    `json:"r"`
	SecurityID string    `json:"sid"`
}
