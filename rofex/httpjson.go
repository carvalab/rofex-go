package rofex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
)

// getTyped issues a GET request to path and decodes the JSON body into dest.
func getTyped[T any](ctx context.Context, c *Client, path string) (T, error) {
	var zero T
	resp, err := c.doGET(ctx, path)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Only read body for errors to avoid double read
		b, _ := io.ReadAll(resp.Body)
		return zero, &HTTPError{StatusCode: resp.StatusCode, Body: b}
	}

	// For successful responses, decode directly from stream for better performance
	dec := json.NewDecoder(resp.Body)
	var out T
	if err := dec.Decode(&out); err != nil {
		return zero, fmt.Errorf("decode json: %w", err)
	}

	// Optional debug logging - only read body if logger is configured and debug level
	if c.logger != nil {
		// Note: Body is already consumed at this point, so we can't log it efficiently
		// This is a trade-off between performance and debug visibility
		c.logger.Debug("http response decoded", slog.String("path", path), slog.String("type", fmt.Sprintf("%T", out)))
	}

	return out, nil
}

// getTypedStrict is like getTyped but fails on unknown fields.
func getTypedStrict[T any](ctx context.Context, c *Client, path string) (T, error) {
	var zero T
	resp, err := c.doGET(ctx, path)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Only read body for errors to avoid double read
		b, _ := io.ReadAll(resp.Body)
		return zero, &HTTPError{StatusCode: resp.StatusCode, Body: b}
	}

	// For successful responses, decode directly from stream for better performance
	dec := json.NewDecoder(resp.Body)
	dec.DisallowUnknownFields()
	var out T
	if err := dec.Decode(&out); err != nil {
		return zero, fmt.Errorf("decode json: %w", err)
	}

	if c.logger != nil {
		c.logger.Debug("http response decoded (strict)", slog.String("path", path), slog.String("type", fmt.Sprintf("%T", out)))
	}

	return out, nil
}
