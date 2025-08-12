package model

import (
	"net/http"
)

// JSON es un objeto JSON decodificado genérico.
type JSON map[string]any

// APIResponse envuelve una respuesta HTTP de la API con status, headers y body.
type APIResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte // raw JSON body; caller can unmarshal
}

// SegmentsResponse contiene la lista de segmentos devuelta por la API.
type SegmentsResponse struct {
	Status   string `json:"status"`
	Segments []struct {
		MarketSegmentID string `json:"marketSegmentId"`
		MarketID        string `json:"marketId"`
	} `json:"segments"`
}
