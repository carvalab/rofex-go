// Package rofex provides a Go SDK for the Primary (ROFEX) trading API.
//
// It implements a complete client for the Primary API v1.21 (REST + WebSocket):
//
//   - Market data (real-time and historical)
//   - Order management (send, cancel, replace, status)
//   - Account info (positions, reports)
//   - Reference data (instruments, segments)
//
// Reference: docs/primary-api.md
//
// Quickstart:
//
//	client, err := rofex.NewClient(
//	    rofex.WithEnvironment(model.EnvironmentRemarket),
//	    rofex.WithAuth(rofex.NewPasswordAuth(rofex.Credentials{
//	        Username: "user",
//	        Password: "pass",
//	    })),
//	)
package rofex

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/carvalab/rofex-go/rofex/model"
)

// WaitFunc blocks until the rate limiter allows the next request.
type WaitFunc func(ctx context.Context) error

func noWait(ctx context.Context) error { return nil }

// AuthProvider applies authentication to outbound HTTP requests and may refresh tokens.
type AuthProvider interface {
	Apply(req *http.Request) error
	Refresh(ctx context.Context, c *Client) error
}

// Credentials for username/password auth.
type Credentials struct {
	Username string
	Password string
}

// PasswordAuth implements AuthProvider via the Primary X-Username/X-Password login that yields X-Auth-Token.
type PasswordAuth struct {
	cred  Credentials
	token string
}

func NewPasswordAuth(cred Credentials) *PasswordAuth { return &PasswordAuth{cred: cred} }

func (a *PasswordAuth) Apply(req *http.Request) error {
	if a.token == "" {
		return ErrUnauthorized
	}
	req.Header.Set("X-Auth-Token", a.token)
	return nil
}

func (a *PasswordAuth) Refresh(ctx context.Context, c *Client) error {
	token, err := c.login(ctx, a.cred)
	if err != nil {
		return err
	}
	a.token = token
	return nil
}

// StaticTokenAuth uses a pre-obtained token.
type StaticTokenAuth struct{ token string }

func NewStaticTokenAuth(token string) *StaticTokenAuth { return &StaticTokenAuth{token: token} }

func (a *StaticTokenAuth) Apply(req *http.Request) error {
	if a.token == "" {
		return ErrUnauthorized
	}
	req.Header.Set("X-Auth-Token", a.token)
	return nil
}

func (a *StaticTokenAuth) Refresh(ctx context.Context, c *Client) error { return nil }

// Client is the SDK entry point for the Primary (ROFEX) trading API.
//
// All methods are safe for concurrent use.
//
// Reference: docs/primary-api.md
type Client struct {
	baseURL      string
	wsURL        string
	http         *http.Client
	rateLimit    WaitFunc
	auth         AuthProvider
	logger       *slog.Logger
	userAgent    string
	timeout      time.Duration
	proprietary  string
	wsBuf        int
	wsDropOnFull bool
	env          model.Environment
}

// applyEnvironment sets URLs and defaults for the selected environment.
func (c *Client) applyEnvironment(env model.Environment) {
	switch env {
	case model.EnvironmentRemarket:
		c.baseURL = "https://api.remarkets.primary.com.ar/"
		c.wsURL = "wss://api.remarkets.primary.com.ar/"
		c.proprietary = "PBCP"
	case model.EnvironmentLive:
		// Production: provider may vary (Primary, Eco Valores, etc.).
		// User must set BaseURL and WSURL via WithBaseURL/WithWSURL.
		//   - Primary:   https://api.primary.com.ar/   | wss://api.primary.com.ar/
		//   - Eco Valores: https://api.eco.xoms.com.ar/ | wss://api.eco.xoms.com.ar/
		c.proprietary = "api"
	}
	c.env = env
}

// NewClient creates a new Primary (ROFEX) API client with sensible defaults.
//
// For LIVE you must set BaseURL and WSURL via WithBaseURL/WithWSURL; the SDK
// does not default them because production providers differ. See the package
// doc for examples.
func NewClient(opts ...Option) (*Client, error) {
	c := &Client{
		baseURL:      "https://api.remarkets.primary.com.ar/",
		wsURL:        "wss://api.remarkets.primary.com.ar/",
		http:         &http.Client{Timeout: 15 * time.Second},
		rateLimit:    noWait,
		userAgent:    "rofex-go/0.1.0 (+https://github.com/carvalab/rofex-go)",
		timeout:      15 * time.Second,
		proprietary:  "PBCP",
		wsBuf:        128,
		wsDropOnFull: false,
		env:          model.EnvironmentRemarket,
		logger:       slog.Default(),
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.http == nil {
		return nil, errors.New("http client is nil")
	}
	if c.env == model.EnvironmentLive {
		if c.baseURL == "" || c.wsURL == "" ||
			strings.Contains(c.baseURL, "remarkets") || strings.Contains(c.wsURL, "remarkets") {
			return nil, errors.New("EnvironmentLive requires BaseURL and WSURL (e.g. https://api.primary.com.ar/ or https://api.eco.xoms.com.ar/). Configure WithBaseURL and WithWSURL")
		}
	}
	if !strings.HasSuffix(c.baseURL, "/") {
		c.baseURL += "/"
	}
	if !strings.HasSuffix(c.wsURL, "/") {
		c.wsURL += "/"
	}
	return c, nil
}

// Login authenticates with username/password and stores the token.
func (c *Client) Login(ctx context.Context, cred Credentials) error {
	pa, ok := c.auth.(*PasswordAuth)
	if !ok {
		pa = NewPasswordAuth(cred)
		c.auth = pa
	}
	return pa.Refresh(ctx, c)
}

func (c *Client) login(ctx context.Context, cred Credentials) (string, error) {
	endpoint := c.baseURL + pathAuth
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-Username", cred.Username)
	req.Header.Set("X-Password", cred.Password)
	req.Header.Set("User-Agent", c.userAgent)
	if err := c.rateLimit(ctx); err != nil {
		return "", err
	}
	start := time.Now()
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if c.logger != nil {
			c.logger.Error("login failed", slog.Int("status", resp.StatusCode), slog.Duration("dur", time.Since(start)))
		}
		return "", &HTTPError{StatusCode: resp.StatusCode}
	}
	token := resp.Header.Get("X-Auth-Token")
	if token == "" {
		return "", fmt.Errorf("missing X-Auth-Token header in response")
	}
	if c.logger != nil {
		c.logger.Info("login ok", slog.Duration("dur", time.Since(start)))
	}
	return token, nil
}

// doGET issues a GET with auth, rate limiting, and 401-driven token refresh.
func (c *Client) doGET(ctx context.Context, path string) (*http.Response, error) {
	endpoint := c.baseURL + strings.TrimPrefix(path, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	if c.auth != nil {
		if err := c.auth.Apply(req); err != nil {
			if err := c.auth.Refresh(ctx, c); err != nil {
				return nil, err
			}
			if err := c.auth.Apply(req); err != nil {
				return nil, err
			}
		}
	}
	if err := c.rateLimit(ctx); err != nil {
		return nil, err
	}
	start := time.Now()
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized && c.auth != nil {
		_ = resp.Body.Close()
		if c.logger != nil {
			c.logger.Warn("get unauthorized, refreshing token", slog.String("path", path))
		}
		if err := c.auth.Refresh(ctx, c); err != nil {
			return nil, err
		}
		if err := c.rateLimit(ctx); err != nil {
			return nil, err
		}
		if err := c.auth.Apply(req); err != nil {
			return nil, err
		}
		resp, err = c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("http request retry failed: %w", err)
		}
	}
	if c.logger != nil {
		c.logger.Debug("http get",
			slog.String("path", path),
			slog.Int("status", resp.StatusCode),
			slog.Duration("dur", time.Since(start)),
		)
	}
	return resp, nil
}

// wsAuthToken returns an auth token for WebSocket connections.
func (c *Client) wsAuthToken(ctx context.Context) (string, error) {
	switch a := c.auth.(type) {
	case *PasswordAuth:
		if a.token == "" {
			if err := a.Refresh(ctx, c); err != nil {
				return "", err
			}
		}
		return a.token, nil
	case *StaticTokenAuth:
		if a.token == "" {
			return "", ErrUnauthorized
		}
		return a.token, nil
	default:
		return "", ErrUnauthorized
	}
}

// WSURLString returns the WebSocket base URL.
func (c *Client) WSURLString() string { return c.wsURL }

// AuthProvider returns the configured auth provider.
func (c *Client) AuthProvider() AuthProvider { return c.auth }
