# rofex-go - SDK de Go para la API de Primary (ROFEX)

### WebSocket: guía rápida (Primary API)

Esta sección resume los valores clave usados por la API WebSocket de Primary y su significado práctico.

#### Tipos de mensaje (`type`)

| Valor | Significado | Uso |
| :--- | :--- | :--- |
| `smd` | Subscribe Market Data | Suscribirse a cotizaciones en tiempo real |
| `os`  | Order Subscription | Suscribirse a Execution Reports de órdenes |
| `no`  | New Order | Enviar una nueva orden |
| `co`  | Cancel Order | Cancelar una orden |

Ejemplo de suscripción a Market Data:

```json
{
  "type": "smd",
  "level": 1,
  "entries": ["OF"],
  "products": [
    {"symbol": "DLR/DIC23", "marketId": "ROFX"},
    {"symbol": "SOJ.ROS/MAY23", "marketId": "ROFX"}
  ],
  "depth": 2
}
```

#### Entradas de Market Data (`entries`)

| Entry | Significado | Notas |
| :--- | :--- | :--- |
| `BI` | Mejor compra (BIDS) | Lista de niveles si depth > 1 |
| `OF` | Mejor venta (OFFERS) | Lista de niveles si depth > 1 |
| `LA` | Último precio operado (LAST) | Puede incluir size y date |
| `OP` | Precio de apertura (OPEN) | Numérico |
| `CL` | Precio de cierre (CLOSE) | Puede incluir {price,size,date} |
| `SE` | Precio de ajuste (SETTLEMENT) | Futuros |
| `HI` | Máximo de la rueda | Numérico |
| `LO` | Mínimo de la rueda | Numérico |
| `TV` | Volumen operado | Numérico |
| `OI` | Interés abierto | Suele incluir size/date |
| `IV` | Valor de índice | Índices |
| `EV` | Volumen efectivo | ByMA |
| `NV` | Volumen nominal | ByMA |
| `ACP` | Precio de subasta del día | |
| `TC` | Cantidad de trades | |

#### Profundidad del libro (`depth`)

- 1: Top of book (mejor BID/ASK). Menor ancho de banda/latencia. Valor por defecto.
- 2..5: Hasta 5 niveles por lado. Más granularidad del libro, más datos a procesar.

Ordenamiento de niveles (cuando depth > 1):
- `BI` (compras): mejor → peor (precio descendente)
- `OF` (ventas): mejor → peor (precio ascendente)

Para más detalles ver `docs/primary-api.md`.

![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)

Un SDK completo y moderno en Go para interactuar con la API de trading de Primary (anteriormente ROFEX), el mercado de futuros y opciones líder de Argentina.

## 🚀 Características

- **APIs REST y WebSocket completas**: Soporte total para todas las funcionalidades de la API de Primary
- **Tipado fuerte**: Todas las respuestas están tipadas con validación exhaustiva
- **Gestión automática de tokens**: Autenticación y renovación automática de tokens
- **Reconexión robusta**: Manejo inteligente de reconexiones WebSocket con backoff exponencial  
- **Rate limiting**: Soporte integrado para limitación de velocidad
- **Logging estructurado**: Integración con slog para observabilidad
- **Thread-safe**: Todos los métodos son seguros para uso concurrente
- **Entornos múltiples**: Soporte para reMarkets (sandbox) y producción
- **Documentación bilingüe**: Documentación completa en español e inglés

## 📦 Instalación

```bash
go get github.com/carvalab/rofex-go
```

## 🏁 Inicio Rápido

### Configuración Básica

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/carvalab/rofex-go/rofex"
    "github.com/carvalab/rofex-go/rofex/model"
)

func main() {
    ctx := context.Background()

    // Crear cliente con auto-autenticación
    client, err := rofex.NewClient(
        rofex.WithEnvironment(model.EnvironmentRemarket),
        rofex.WithAuth(rofex.NewPasswordAuth(rofex.Credentials{
            Username: "tu_usuario",
            Password: "tu_contraseña",
        })),
    )
    if err != nil {
        log.Fatal("Error creando cliente:", err)
    }

    fmt.Println("¡Cliente listo para usar!")
}
```

### Obtener Datos de Mercado

```go
// Obtener snapshot de datos de mercado
md, err := client.MarketDataSnapshot(ctx, rofex.MDRequest{
    Symbol: "DLR/DIC21",
    Market: model.MarketROFEX,
    Entries: []model.MDEntry{
        model.MDBids,    // Mejores ofertas de compra
        model.MDOffers,  // Mejores ofertas de venta  
        model.MDLast,    // Último precio operado
    },
    Depth: 5, // Profundidad del book
})
if err != nil {
    log.Fatal("Error obteniendo market data:", err)
}

fmt.Printf("Market Data para %s:\n", md.Instrument.Symbol)
if len(md.MarketData.BI) > 0 {
    fmt.Printf("Mejor Bid: $%.2f (Cantidad: %d)\n", 
        md.MarketData.BI[0].Price, md.MarketData.BI[0].Size)
}
if len(md.MarketData.OF) > 0 {
    fmt.Printf("Mejor Offer: $%.2f (Cantidad: %d)\n",
        md.MarketData.OF[0].Price, md.MarketData.OF[0].Size)
}
```

### Enviar una Orden

```go
// Obtener la primera cuenta disponible y enviar orden de compra LIMIT
accounts, err := client.Accounts(ctx)
if err != nil {
    log.Fatal("Error obteniendo cuentas:", err)
}
if len(accounts.Accounts) == 0 {
    log.Fatal("No hay cuentas disponibles para el usuario")
}
account := accounts.Accounts[0].Name

price := 18.50
order, err := client.SendOrder(ctx, rofex.NewOrder{
    Symbol:  "DLR/DIC21",
    Market:  model.MarketROFEX,
    Side:    model.Buy,
    Type:    model.OrderTypeLimit,
    Qty:     10,
    Price:   &price,
    Account: account,
    TIF:     model.Day,
})
if err != nil {
    log.Fatal("Error enviando orden:", err)
}

fmt.Printf("Orden enviada. Client Order ID: %s\n", order.Order.ClientID)

// Verificar estado de la orden
status, err := client.OrderStatus(ctx, order.Order.ClientID, "")
if err != nil {
    log.Fatal("Error consultando estado:", err)
}

fmt.Printf("Estado de la orden: %s\n", status.Order.Status)
```

### Cancelar una Orden

```go
// Cancelar orden por Client Order ID
cancelResp, err := client.CancelOrder(ctx, "client_order_id", "")
if err != nil {
    log.Fatal("Error cancelando orden:", err)
}

fmt.Printf("Orden cancelada. Cancel ID: %s\n", cancelResp.Order.ClientID)
```

## 📊 Streaming de Datos en Tiempo Real

### Market Data en Tiempo Real

```go
// Suscribirse a datos de mercado en tiempo real
subscription, err := client.SubscribeMarketData(ctx, 
    []string{"DLR/DIC21", "DOFeb25"}, // Símbolos
    []model.MDEntry{model.MDBids, model.MDOffers, model.MDLast},
    5, // Profundidad
    model.MarketROFEX,
)
if err != nil {
    log.Fatal("Error suscribiéndose:", err)
}
defer subscription.Close()

// Procesar eventos
for {
    select {
    case event := <-subscription.Events:
        fmt.Printf("Market Data: %+v\n", event)
    case err := <-subscription.Errs:
        fmt.Printf("Error: %v\n", err)
    case <-ctx.Done():
        return
    }
}
```

### Reportes de Órdenes en Tiempo Real

```go
// Suscribirse a reportes de órdenes
accounts, err := client.Accounts(ctx)
if err != nil {
    log.Fatal("Error obteniendo cuentas:", err)
}
if len(accounts.Accounts) == 0 {
    log.Fatal("No hay cuentas disponibles para el usuario")
}
account := accounts.Accounts[0].Name

orderSub, err := client.SubscribeOrderReport(ctx, account, true)
if err != nil {
    log.Fatal("Error suscribiéndose a órdenes:", err)
}
defer orderSub.Close()

// Procesar reportes de órdenes
for {
    select {
    case report := <-orderSub.Events:
        fmt.Printf("Order Report: %+v\n", report)
    case err := <-orderSub.Errs:
        fmt.Printf("Error en order report: %v\n", err)
    case <-ctx.Done():
        return
    }
}
```

## 🛠️ Funcionalidades Avanzadas

### Configuración con Logging

```go
import "log/slog"

client, err := rofex.NewClient(
    rofex.WithEnvironment(model.EnvironmentRemarket),
    rofex.WithAuth(auth),
    rofex.WithLogger(slog.Default()), // Habilitar logging
)
```

### Rate Limiting Personalizado

```go
import "golang.org/x/time/rate"

// Crear limitador personalizado
limiter := rate.NewLimiter(rate.Limit(10), 1) // 10 requests/segundo, burst=1

client, err := rofex.NewClient(
    rofex.WithRateLimit(limiter),
    // ... otras opciones
)
```

### Cliente HTTP Personalizado

```go
import "net/http"

httpClient := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
    },
}

client, err := rofex.NewClient(
    rofex.WithHTTPClient(httpClient),
    // ... otras opciones
)
```

## 📖 Documentación de la API

### Instrumentos y Datos de Referencia

```go
// Obtener todos los instrumentos
instruments, err := client.InstrumentsAll(ctx)

// Obtener instrumentos con detalles completos
detailed, err := client.InstrumentsDetails(ctx)

// Obtener instrumentos por segmento
bySegment, err := client.InstrumentsBySegment(ctx, 
    model.MarketROFEX, 
    []model.MarketSegment{model.SegmentDDF, model.SegmentDDA},
)

// Obtener instrumentos por código CFI
byCFI, err := client.InstrumentsByCFICode(ctx,
    []model.CFICode{model.CFIFuture, model.CFIStock},
)

// Obtener detalles de un instrumento específico
detail, err := client.InstrumentDetail(ctx, "DLR/DIC21", model.MarketROFEX)
```

### Gestión de Órdenes

```go
// Consultar todas las órdenes de una cuenta
allOrders, err := client.AllOrdersStatus(ctx, "cuenta")

// Consultar órdenes activas
activeOrders, err := client.ActiveOrders(ctx, "cuenta")

// Consultar órdenes ejecutadas
filledOrders, err := client.FilledOrders(ctx, "cuenta")

// Consultar orden por Order ID
orderByID, err := client.OrderByOrderID(ctx, "order_id")

// Consultar orden por Execution ID
orderByExec, err := client.OrderByExecID(ctx, "exec_id")

// Modificar una orden existente
newQty := int64(20)
newPrice := 19.00
replaceResp, err := client.ReplaceOrder(ctx, "client_order_id", "", &newQty, &newPrice)
```

### Información de Cuenta

```go
// Obtener cuentas asociadas al usuario
accounts, err := client.Accounts(ctx)

// Obtener posiciones de la cuenta
positions, err := client.AccountPosition(ctx, "cuenta")

// Obtener posiciones detalladas
detailedPos, err := client.DetailedPosition(ctx, "cuenta")

// Obtener reporte de cuenta
report, err := client.AccountReport(ctx, "cuenta")
```

### Datos Históricos

```go
// Obtener trades históricos
from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

trades, err := client.HistoricTrades(ctx, "DLR/DIC21", model.MarketROFEX, from, to)
```

## 🌐 Entornos

### reMarkets (Sandbox)
- **URL**: `https://api.remarkets.primary.com.ar/`
- **WebSocket**: `wss://api.remarkets.primary.com.ar/`  
- **Propósito**: Testing y desarrollo
- **Registro**: [remarkets.primary.ventures](https://remarkets.primary.ventures)

### Producción (URLs obligatorias)
- Debés especificar explícitamente BaseURL y WSURL porque dependen del proveedor.
- Ejemplos de proveedores:
  - Primary: [https://api.primary.com.ar/](https://api.primary.com.ar/) (WS: `wss://api.primary.com.ar/`)
  - Eco Valores SA: [https://api.eco.xoms.com.ar/](https://api.eco.xoms.com.ar/) (WS: `wss://api.eco.xoms.com.ar/`)
- Si no se setean, el cliente devuelve error y no continúa.

```go
// Configurar para producción (URLs obligatorias)
client, err := rofex.NewClient(
    rofex.WithEnvironment(model.EnvironmentLive),
    rofex.WithBaseURL("https://api.primary.com.ar/"),
    rofex.WithWSURL("wss://api.primary.com.ar/"),
    rofex.WithAuth(rofex.NewPasswordAuth(rofex.Credentials{Username: user, Password: pass})),
)
```

## 🔧 Tipos de Datos Importantes

### Enums de Trading

```go
// Tipos de órdenes
model.OrderTypeLimit         // Orden limitada
model.OrderTypeMarket        // Orden de mercado
model.OrderTypeMarketToLimit // Mercado a límite

// Lados de orden
model.Buy   // Compra
model.Sell  // Venta

// Time in Force
model.Day               // Válida durante el día
model.ImmediateOrCancel // IOC - Inmediata o cancela
model.FillOrKill        // FOK - Ejecuta todo o cancela
model.GoodTillDate      // GTD - Válida hasta fecha

// Entradas de Market Data
model.MDBids              // BI - Ofertas de compra
model.MDOffers            // OF - Ofertas de venta
model.MDLast              // LA - Último precio
model.MDOpeningPrice      // OP - Precio de apertura
model.MDClosePrice        // CL - Precio de cierre
model.MDSettlementPrice   // SE - Precio de ajuste
model.MDTradeVolume       // TV - Volumen operado
model.MDOpenInterest      // OI - Interés abierto
```

### Estructura de Orden

```go
type NewOrder struct {
    Symbol         string              // Símbolo del instrumento
    Market         model.Market        // Mercado (ROFX)
    Side           model.Side          // Compra/Venta
    Type           model.OrderType     // Tipo de orden
    Qty            int64               // Cantidad
    Price          *float64           // Precio (requerido para LIMIT)
    TIF            model.TimeInForce   // Time in Force
    Account        string              // Cuenta
    CancelPrevious bool                // Cancelar órdenes previas
    Iceberg        bool                // Orden iceberg
    ExpireDate     *string            // Fecha de vencimiento (GTD)
    DisplayQty     *int64             // Cantidad a mostrar (iceberg)
    AllOrNone      bool               // Todo o nada (WebSocket)
    WSClOrdID      *string            // Client Order ID (WebSocket)
}
```

## ⚠️ Buenas Prácticas

### 1. Manejo de Errores
```go
if err != nil {
    var httpErr *rofex.HTTPError
    if errors.As(err, &httpErr) {
        fmt.Printf("Error HTTP %d: %s\n", httpErr.StatusCode, string(httpErr.Body))
        return
    }
    var validationErr *rofex.ValidationError  
    if errors.As(err, &validationErr) {
        fmt.Printf("Error de validación en %s: %s\n", validationErr.Field, validationErr.Msg)
        return
    }
    var authErr *rofex.AuthError
    if errors.As(err, &authErr) {
        fmt.Printf("Error de autenticación: %s\n", authErr.Msg)
        // Reintenta con credenciales nuevas
        return
    }
    var tempErr *rofex.TemporaryError
    if errors.As(err, &tempErr) {
        fmt.Printf("Error temporal: %s - reintenta en unos segundos\n", tempErr.Error())
        time.Sleep(5 * time.Second)
        // Reintenta la operación
        return
    }
    fmt.Printf("Error no categorizado: %v\n", err)
}
```

### 2. Verificación de Estado de Órdenes
```go
// Siempre verificar el estado después de enviar una orden
order, err := client.SendOrder(ctx, newOrder)
if err != nil {
    return fmt.Errorf("send order failed: %w", err)
}

// Esperar confirmación del mercado
for i := 0; i < 5; i++ {
    status, err := client.OrderStatus(ctx, order.Order.ClientID, "")
    if err != nil {
        return fmt.Errorf("order status failed: %w", err)
    }
    
    switch status.Order.Status {
    case "NEW":
        fmt.Println("Orden aceptada por el mercado")
        return nil
    case "REJECTED":
        return fmt.Errorf("orden rechazada: %s", status.Order.Text)
    case "PENDING_NEW":
        time.Sleep(time.Second)
        continue
    }
}
```

### 3. Rate Limiting
```go
// Respetar los límites de la API según documentación
// - Autenticación: 1 request/día (token dura 24hs)
// - Market Data: Usar WebSocket para tiempo real
// - Órdenes: Máximo 1 request/segundo para cancelaciones
// - Reportes: 1 request cada 5 segundos
```

### 4. Gestión de Recursos WebSocket
```go
// Siempre cerrar las suscripciones
subscription, err := client.SubscribeMarketData(ctx, symbols, entries, depth, market)
if err != nil {
    return err
}
defer subscription.Close() // Importante: siempre cerrar

// Manejar contextos para cancelación
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// En caso de shutdown graceful
go func() {
    <-shutdownCh
    cancel() // Esto cerrará todas las suscripciones WebSocket
}()
```

## 🤝 Contribuir

Las contribuciones son bienvenidas! Por favor:

1. Fork el repositorio
2. Crear una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abrir un Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT. Ver el archivo [LICENSE](LICENSE) para más detalles.

## 📚 Documentación y Recursos

### Documentación Principal
- **[Primary API v1.21 (docs/primary-api.md)](docs/primary-api.md)** - Documentación oficial completa
- **[API Data JSON (docs/api_data.json)](docs/api_data.json)** - Esquemas y endpoints
- **[Colección Postman (docs/rest.postman_collection.json)](docs/rest.postman_collection.json)** - Tests de API

### Recursos Online de Primary
- **[API Hub - Documentación Trading](https://apihub.primary.com.ar/assets/apidoc/trading/index.html)** - Documentación oficial completa
- **[Documentación Swagger](https://api.remarkets.primary.com.ar/api-docs/index.html)**


### Cómo obtener api_data.json
El archivo `api_data.json` contiene la especificación completa de la API y se puede obtener de:
1. **Desde este repositorio**: [docs/api_data.json](docs/api_data.json)
2. **Desde la API**: `GET https://api.remarkets.primary.com.ar/rest/api-data` (requiere autenticación)
3. **Desde Swagger**: Exportar desde [api-docs](https://api.remarkets.primary.com.ar/api-docs/index.html)

## 📞 Soporte

- **Documentación Oficial**: [apihub.primary.com.ar](https://apihub.primary.com.ar)
- **reMarkets (Sandbox)**: [remarkets.primary.ventures](https://remarkets.primary.ventures)
- **Issues**: [GitHub Issues](https://github.com/carvalab/rofex-go/issues)

## 🙏 Agradecimientos

- **Primary (ROFEX)** por proporcionar una API robusta y bien documentada
- **pyRofex** por servir como referencia de implementación
- La comunidad de Go por las excelentes librerías que hacen posible este SDK

---

**Nota**: Este SDK no está oficialmente respaldado por Primary. Es un proyecto de código abierto mantenido por la comunidad.
