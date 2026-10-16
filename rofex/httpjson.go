package rofex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// getTyped issues a GET request to path and decodes the JSON body into T.
// Primary reports business errors as HTTP 200 with "status":"ERROR"; those
// are returned as *APIError.
func getTyped[T any](ctx context.Context, c *Client, path string) (T, error) {
	var zero T
	resp, err := c.doGET(ctx, path)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return zero, &HTTPError{StatusCode: resp.StatusCode, Body: b}
	}

	var head struct {
		Status      string          `json:"status"`
		Message     json.RawMessage `json:"message"`
		Description json.RawMessage `json:"description"`
	}
	if err := json.Unmarshal(b, &head); err == nil && strings.EqualFold(head.Status, "ERROR") {
		return zero, &APIError{Message: rawText(head.Message), Description: rawText(head.Description)}
	}

	var out T
	if err := json.Unmarshal(b, &out); err != nil {
		return zero, fmt.Errorf("decode json: %w", err)
	}
	if c.logger != nil {
		c.logger.Debug("http response decoded", slog.String("path", path), slog.String("type", fmt.Sprintf("%T", out)))
	}
	return out, nil
}

// rawText returns a JSON string value as text, null/absent as "", and any
// other JSON value verbatim.
func rawText(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	if t := strings.TrimSpace(string(raw)); t != "null" {
		return t
	}
	return ""
}
