package rofex

import (
	"errors"
	"fmt"
)

// HTTPError wraps a non-2xx HTTP response.
type HTTPError struct {
	StatusCode int
	Body       []byte
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http error: status=%d body=%s", e.StatusCode, string(e.Body))
}

// APIError is returned when Primary answers a request with "status":"ERROR"
// (it does so with HTTP 200), e.g. a rejected order, an unknown symbol or an
// account without access.
type APIError struct {
	Message     string // short reason, e.g. "Access Denied." (may be empty)
	Description string // details, e.g. "You don't have access to account X"
}

func (e *APIError) Error() string {
	switch {
	case e.Message != "" && e.Description != "":
		return "primary api error: " + e.Message + " " + e.Description
	case e.Description != "":
		return "primary api error: " + e.Description
	case e.Message != "":
		return "primary api error: " + e.Message
	default:
		return "primary api error"
	}
}

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrClosed       = errors.New("closed")
)
