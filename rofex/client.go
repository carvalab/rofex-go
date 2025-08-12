// Package rofex proporciona un SDK de Go para la API de trading de Primary (ROFEX).
//
// Este paquete implementa un cliente completo para la Primary API v1.21 según
// la documentación oficial (docs/primary-api.md).
//
// El SDK soporta tanto APIs REST como WebSocket para:
// - Datos de mercado (tiempo real e históricos)
// - Gestión de órdenes (enviar, cancelar, modificar, estado)
// - Información de cuenta (posiciones, reportes)
// - Datos de referencia (instrumentos, segmentos)
//
// Ejemplo de uso:
//
//	import "github.com/carvalab/rofex-go/rofex"
//	import "github.com/carvalab/rofex-go/rofex/model"
//
//	// Crear cliente
//	client, err := rofex.NewClient(
//		rofex.WithEnvironment(model.EnvironmentRemarket),
//		rofex.WithAuth(rofex.NewPasswordAuth(rofex.Credentials{
//			Username: "tu_usuario",
//			Password: "tu_contraseña",
//		})),
//	)
//
//	// Obtener datos de mercado
//	md, err := client.MarketDataSnapshot(ctx, rofex.MDRequest{
//		Symbol: "DLR/DIC21",
//		Market: model.MarketROFEX,
//		Entries: []model.MDEntry{model.MDBids, model.MDOffers},
//	})
//
//	// Enviar orden
//	order, err := client.SendOrder(ctx, rofex.NewOrder{
//		Symbol:  "DLR/DIC21",
//		Side:    model.Buy,
//		Type:    model.OrderTypeLimit,
//		Qty:     10,
//		Price:   &price,
//		Account: "123",
//	})
//
// Para más ejemplos, ver el directorio examples/.
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
	"github.com/coder/websocket"
)

// RateLimiter is a simple token bucket style limiter interface.
type RateLimiter interface {
	Wait(ctx context.Context) error
}

// applyEnvironment sets URLs and defaults for the selected environment.
func (c *Client) applyEnvironment(env model.Environment) {
	switch env {
	case model.EnvironmentRemarket:
		c.baseURL = "https://api.remarkets.primary.com.ar/"
		c.wsURL = "wss://api.remarkets.primary.com.ar/"
		c.proprietary = "PBCP"
	case model.EnvironmentLive:
		// Producción: el proveedor puede variar (Primary, Eco Valores, etc.).
		// NO seteamos valores por defecto aquí. Es OBLIGATORIO que el usuario
		// establezca baseURL y wsURL explícitamente mediante WithBaseURL/WithWSURL.
		// Ejemplos:
		// - Primary:   https://api.primary.com.ar/   | wss://api.primary.com.ar/
		// - Eco Valores: https://api.eco.xoms.com.ar/ | wss://api.eco.xoms.com.ar/
		c.proprietary = "api"
	}
	c.env = env
	c.envExplicit = true
}

// HTTPDoer abstrae *http.Client para facilitar testing/mocking.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// wsAuthToken extrae un token de autenticación para conexiones WebSocket.
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

// Getters mínimos usados por helpers de stream.
func (c *Client) WSURLString() string        { return c.wsURL }
func (c *Client) AuthProvider() AuthProvider { return c.auth }

// noLimiter es una implementación de limitador sin operación.
type noLimiter struct{}

func (noLimiter) Wait(ctx context.Context) error { return nil }

// AuthProvider aplica autenticación a requests HTTP salientes y puede refrescar tokens si es necesario.
type AuthProvider interface {
	Apply(req *http.Request) error
	Refresh(ctx context.Context, c *Client) error
}

// Credentials para autenticación usuario/contraseña.
type Credentials struct {
	Username string
	Password string
}

// PasswordAuth implementa AuthProvider usando el login X-Username/X-Password de Primary que produce X-Auth-Token.
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
	// Perform login and store token
	token, err := c.login(ctx, a.cred)
	if err != nil {
		return err
	}
	a.token = token
	return nil
}

// StaticTokenAuth usa un token previamente obtenido.
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

// Client es el punto de entrada principal del SDK para acceder a la API de trading de Primary (ROFEX).
//
// El Cliente proporciona métodos para:
// - Autenticación y gestión de sesiones
// - Obtención de datos de mercado (REST y WebSocket)
// - Gestión de órdenes (enviar, cancelar, modificar, consultar)
// - Información de cuenta (posiciones, reportes)
// - Datos de referencia (instrumentos, segmentos)
//
// Todos los métodos son thread-safe y pueden ser llamados concurrentemente.
//
// Referencia: docs/primary-api.md - Documentación completa de Primary API v1.21
type Client struct {
	baseURL      string            // URL base para API REST
	wsURL        string            // URL base para WebSocket
	http         HTTPDoer          // Interfaz de cliente HTTP
	limiter      RateLimiter       // Limitador de velocidad
	auth         AuthProvider      // Proveedor de autenticación
	logger       *slog.Logger      // Logger estructurado
	userAgent    string            // Encabezado user agent
	timeout      time.Duration     // Timeout de requests
	proprietary  string            // Valor proprietary por defecto
	wsBuf        int               // Tamaño del buffer WebSocket
	wsDropOnFull bool              // Descartar mensajes cuando buffer lleno
	wsClient     WSClient          // Interfaz de cliente WebSocket
	env          model.Environment // Entorno actual
	envExplicit  bool              // Si el entorno fue establecido explícitamente
}

// WSClient abstrae websocket connection para facilitar testing/mocking.
type WSClient interface {
	Dial(ctx context.Context, url string, opts *websocket.DialOptions) (*websocket.Conn, *http.Response, error)
}

// coderWSClient es una implementación de WSClient usando coder/websocket.
type coderWSClient struct{}

func (coderWSClient) Dial(ctx context.Context, url string, opts *websocket.DialOptions) (*websocket.Conn, *http.Response, error) {
	return websocket.Dial(ctx, url, opts)
}

// NewClient crea un nuevo cliente de la API de Primary (ROFEX) con valores por defecto sensatos.
//
// Entornos:
//   - REMARKET (sandbox): URLs por defecto a reMarkets.
//   - LIVE (producción): URLs OBLIGATORIAS. Debe especificar BaseURL y WSURL porque dependen del proveedor.
//     Ejemplos:
//   - Primary: https://api.primary.com.ar/  | wss://api.primary.com.ar/
//   - Eco Valores SA: https://api.eco.xoms.com.ar/ | wss://api.eco.xoms.com.ar/
//     Si no se setean, NewClient devuelve error y no continúa.
//
// Opciones de configuración comunes:
//   - WithEnvironment(env): Establecer entorno objetivo
//   - WithAuth(auth): Establecer proveedor de autenticación
//   - WithLogger(logger): Habilitar logging estructurado
//   - WithRateLimit(limiter): Configurar limitación de velocidad
//   - WithHTTPClient(client): Usar cliente HTTP personalizado
//
// Ejemplo:
//
//	client, err := rofex.NewClient(
//		rofex.WithEnvironment(model.EnvironmentRemarket),
//		rofex.WithAuth(rofex.NewPasswordAuth(rofex.Credentials{
//			Username: "tu_usuario",
//			Password: "tu_contraseña",
//		})),
//		rofex.WithLogger(slog.Default()),
//	)
//
// Referencia: docs/primary-api.md - "Conectándose a la API por token de autenticación"
func NewClient(opts ...Option) (*Client, error) {
	c := &Client{
		baseURL:      "https://api.remarkets.primary.com.ar/",
		wsURL:        "wss://api.remarkets.primary.com.ar/",
		http:         &http.Client{Timeout: 15 * time.Second},
		limiter:      noLimiter{},
		userAgent:    "rofex-go/0.1.0 (+https://github.com/carvalab/rofex-go)",
		timeout:      15 * time.Second,
		proprietary:  "PBCP",
		wsBuf:        128,
		wsDropOnFull: false,
		wsClient:     &coderWSClient{},
		env:          model.EnvironmentRemarket,
		logger:       slog.Default(),
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.http == nil {
		return nil, errors.New("http client is nil")
	}
	// Validación obligatoria para producción: URLs deben ser provistas por el usuario
	if c.env == model.EnvironmentLive {
		if c.baseURL == "" || c.wsURL == "" ||
			strings.Contains(c.baseURL, "remarkets") || strings.Contains(c.wsURL, "remarkets") {
			return nil, errors.New("EnvironmentLive requiere BaseURL y WSURL explícitos (ej.: https://api.primary.com.ar/ o https://api.eco.xoms.com.ar/). Configure WithBaseURL y WithWSURL")
		}
	}
	if !strings.HasSuffix(c.baseURL, "/") {
		c.baseURL += "/"
	}
	if !strings.HasSuffix(c.wsURL, "/") {
		c.wsURL += "/"
	}

	// Log sandbox info if configured
	if c.logger != nil && c.env == model.EnvironmentRemarket {
		c.logger.Info("running in sandbox environment", slog.String("env", "REMARKET"))
	}
	return c, nil
}

// Login autentica con usuario/contraseña y almacena el token.
//
// Este método crea o actualiza un proveedor PasswordAuth y realiza la autenticación
// con la API de Primary. El token devuelto se usa automáticamente para solicitudes posteriores.
//
// El token típicamente expira después de 24 horas y será refrescado automáticamente
// en respuestas 401 si usa PasswordAuth.
//
// ⚠️ Rate Limit: Este endpoint tiene límite de 1 request por día. El token dura 24 horas.
//
// Ejemplo:
//
//	creds := rofex.Credentials{
//		Username: "tu_usuario",
//		Password: "tu_contraseña",
//	}
//	if err := client.Login(ctx, creds); err != nil {
//		log.Fatal("Error de autenticación:", err)
//	}
//
// Referencia: docs/primary-api.md - "Conectándose a la API por token de autenticación"
func (c *Client) Login(ctx context.Context, cred Credentials) error {
	pa, ok := c.auth.(*PasswordAuth)
	if !ok {
		// set or replace auth with password auth
		pa = NewPasswordAuth(cred)
		c.auth = pa
	}
	return pa.Refresh(ctx, c)
}

// login realiza la llamada HTTP para obtener un nuevo token.
func (c *Client) login(ctx context.Context, cred Credentials) (string, error) {
	endpoint := c.baseURL + pathAuth
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-Username", cred.Username)
	req.Header.Set("X-Password", cred.Password)
	req.Header.Set("User-Agent", c.userAgent)
	if err := c.limiter.Wait(ctx); err != nil {
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

// doGET realiza un GET con autenticación, limitación de velocidad y manejo mínimo de 401-refresh, devolviendo la respuesta raw.
func (c *Client) doGET(ctx context.Context, path string) (*http.Response, error) {
	endpoint := c.baseURL + strings.TrimPrefix(path, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	if c.auth != nil {
		if err := c.auth.Apply(req); err != nil {
			// Token may be empty on first use, try to refresh
			if err := c.auth.Refresh(ctx, c); err != nil {
				return nil, err
			}
			// Apply again after refresh
			if err := c.auth.Apply(req); err != nil {
				return nil, err
			}
		}
	}
	if err := c.limiter.Wait(ctx); err != nil {
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
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, err
		}
		if c.auth != nil {
			if err := c.auth.Apply(req); err != nil {
				return nil, err
			}
		}
		resp, err = c.http.Do(req)
		if err != nil {
			// Network errors during retry should be marked as temporary
			return nil, &TemporaryError{Err: fmt.Errorf("http request retry failed: %w", err)}
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
