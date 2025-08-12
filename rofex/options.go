package rofex

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/carvalab/rofex-go/rofex/model"
)

type Option func(*Client)

func WithBaseURL(u string) Option          { return func(c *Client) { c.baseURL = u } }
func WithWSURL(u string) Option            { return func(c *Client) { c.wsURL = u } }
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.http = h } }
func WithRateLimiter(r RateLimiter) Option { return func(c *Client) { c.limiter = r } }
func WithAuth(a AuthProvider) Option       { return func(c *Client) { c.auth = a } }
func WithLogger(l *slog.Logger) Option     { return func(c *Client) { c.logger = l } }

// WithStaticToken establece un token de autenticación pre-obtenido (no necesita flujo de login).
func WithStaticToken(token string) Option {
	return func(c *Client) { c.auth = NewStaticTokenAuth(token) }
}

// WithWSClient inyecta una implementación personalizada de cliente WS (útil para tests).
func WithWSClient(w WSClient) Option {
	return func(c *Client) {
		if w != nil {
			c.wsClient = w
		}
	}
}
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.timeout = d
		if hc, ok := c.http.(*http.Client); ok && hc != nil {
			hc.Timeout = d
		}
	}
}
func WithUserAgent(ua string) Option  { return func(c *Client) { c.userAgent = ua } }
func WithProprietary(p string) Option { return func(c *Client) { c.proprietary = p } }

// WithWSBuffer sets the buffered channel size for streaming event channels (default 128).
func WithWSBuffer(n int) Option {
	return func(c *Client) {
		if n > 0 {
			c.wsBuf = n
		}
	}
}

// WithWSDropOnFull makes subscriptions drop events when the channel buffer is full (default false: block).
func WithWSDropOnFull(drop bool) Option { return func(c *Client) { c.wsDropOnFull = drop } }

// WithEnvironment sets base and ws URLs and default proprietary based on environment.
func WithEnvironment(env model.Environment) Option {
	return func(c *Client) {
		c.applyEnvironment(env)
	}
}

// WithSandbox is a convenience option to explicitly select REMARKET (sandbox).
func WithSandbox() Option { return func(c *Client) { c.applyEnvironment(model.EnvironmentRemarket) } }

// WithLive is a convenience option to explicitly select LIVE (production).
func WithLive() Option { return func(c *Client) { c.applyEnvironment(model.EnvironmentLive) } }
