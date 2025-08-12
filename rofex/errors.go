package rofex

import (
	"errors"
	"fmt"
)

// HTTPError representa una respuesta HTTP no-2xx.
type HTTPError struct {
	StatusCode int
	Body       []byte
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http error: status=%d body=%s", e.StatusCode, string(e.Body))
}

// ValidationError representa fallas de validación del lado del cliente.
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s: %s", e.Field, e.Msg)
}

// AuthError representa fallas de autenticación.
type AuthError struct{ Msg string }

func (e *AuthError) Error() string { return e.Msg }

// TemporaryError envuelve un error transitorio para sugerir reintentos.
type TemporaryError struct{ Err error }

func (e *TemporaryError) Error() string { return e.Err.Error() }
func (e *TemporaryError) Unwrap() error { return e.Err }

var (
	// ErrUnauthorized indicates missing/expired credentials.
	ErrUnauthorized = &AuthError{Msg: "unauthorized"}
	// ErrClosed indicates a closed resource.
	ErrClosed = errors.New("closed")
)
