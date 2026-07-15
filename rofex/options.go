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
func WithRateLimiter(wait WaitFunc) Option { return func(c *Client) { c.rateLimit = wait } }
func WithAuth(a AuthProvider) Option       { return func(c *Client) { c.auth = a } }
func WithLogger(l *slog.Logger) Option     { return func(c *Client) { c.logger = l } }

// WithStaticToken sets a pre-obtained auth token (no login flow).
func WithStaticToken(token string) Option {
	return func(c *Client) { c.auth = NewStaticTokenAuth(token) }
}

func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.timeout = d
		if c.http != nil {
			c.http.Timeout = d
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

// WithEnvironment sets base/ws URLs and default proprietary based on environment.
func WithEnvironment(env model.Environment) Option {
	return func(c *Client) { c.applyEnvironment(env) }
}
