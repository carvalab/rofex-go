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

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrClosed       = errors.New("closed")
)
