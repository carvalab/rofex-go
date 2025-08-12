# rofex-go - Go SDK for Primary (ROFEX) Trading API

![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)

A complete and modern Go SDK for interacting with the Primary (formerly ROFEX) trading API, Argentina's leading futures and options market.

## 🚀 Features

- **Complete REST and WebSocket APIs**: Full support for all Primary API functionality
- **Strong typing**: All responses are typed with comprehensive validation
- **Automatic token management**: Authentication and automatic token renewal
- **Robust reconnection**: Smart WebSocket reconnection handling with exponential backoff
- **Rate limiting**: Built-in support for rate limiting
- **Structured logging**: Integration with slog for observability
- **Thread-safe**: All methods are safe for concurrent use
- **Multiple environments**: Support for reMarkets (sandbox) and production
- **Bilingual documentation**: Complete documentation in Spanish and English

## 📦 Installation

```bash
go get github.com/carvalab/rofex-go
```

## 🏁 Quick Start

### Basic Setup

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

    // Create client with auto-authentication
    client, err := rofex.NewClient(
        rofex.WithEnvironment(model.EnvironmentRemarket),
        rofex.WithAuth(rofex.NewPasswordAuth(rofex.Credentials{
            Username: "your_username",
            Password: "your_password",
        })),
    )
    if err != nil {
        log.Fatal("Error creating client:", err)
    }

    fmt.Println("Client ready to use!")
}
```

### Get Market Data

```go
// Get market data snapshot
md, err := client.MarketDataSnapshot(ctx, rofex.MDRequest{
    Symbol: "DLR/DIC21",
    Market: model.MarketROFEX,
    Entries: []model.MDEntry{
        model.MDBids,    // Best bid offers
        model.MDOffers,  // Best ask offers  
        model.MDLast,    // Last traded price
    },
    Depth: 5, // Book depth
})
if err != nil {
    log.Fatal("Error getting market data:", err)
}

fmt.Printf("Market Data for %s:\n", md.Instrument.Symbol)
if len(md.MarketData.BI) > 0 {
    fmt.Printf("Best Bid: $%.2f (Size: %d)\n", 
        md.MarketData.BI[0].Price, md.MarketData.BI[0].Size)
}
if len(md.MarketData.OF) > 0 {
    fmt.Printf("Best Offer: $%.2f (Size: %d)\n",
        md.MarketData.OF[0].Price, md.MarketData.OF[0].Size)
}
```

### Send an Order

```go
// Fetch the first available account and send a BUY LIMIT order
accounts, err := client.Accounts(ctx)
if err != nil {
    log.Fatal("Error getting accounts:", err)
}
if len(accounts.Accounts) == 0 {
    log.Fatal("No accounts available for this user")
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
    log.Fatal("Error sending order:", err)
}

fmt.Printf("Order sent. Client Order ID: %s\n", order.Order.ClientID)

// Check order status
status, err := client.OrderStatus(ctx, order.Order.ClientID, "")
if err != nil {
    log.Fatal("Error checking status:", err)
}

fmt.Printf("Order status: %s\n", status.Order.Status)
```

### Cancel an Order

```go
// Cancel order by Client Order ID
cancelResp, err := client.CancelOrder(ctx, "client_order_id", "")
if err != nil {
    log.Fatal("Error canceling order:", err)
}

fmt.Printf("Order canceled. Cancel ID: %s\n", cancelResp.Order.ClientID)
```

## 📊 Real-time Data Streaming

### WebSocket: Quick Reference (Primary API)

This section summarizes the key values used by Primary's WebSocket API and their practical meaning.

#### Message Types (`type`)

| Value | Meaning | Usage |
| :--- | :--- | :--- |
| `smd` | Subscribe Market Data | Subscribe to real-time quotes |
| `os`  | Order Subscription | Subscribe to order Execution Reports |
| `no`  | New Order | Send a new order |
| `co`  | Cancel Order | Cancel an order |

Market Data subscription example:

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

#### Market Data Entries (`entries`)

| Entry | Meaning | Notes |
| :--- | :--- | :--- |
| `BI` | Best bid (BIDS) | List of levels if depth > 1 |
| `OF` | Best ask (OFFERS) | List of levels if depth > 1 |
| `LA` | Last traded price (LAST) | May include size and date |
| `OP` | Opening price (OPEN) | Numeric |
| `CL` | Closing price (CLOSE) | May include {price,size,date} |
| `SE` | Settlement price | Futures only |
| `HI` | Session high | Numeric |
| `LO` | Session low | Numeric |
| `TV` | Traded volume | Numeric |
| `OI` | Open interest | Usually includes size/date |
| `IV` | Index value | Indices only |
| `EV` | Effective volume | ByMA |
| `NV` | Nominal volume | ByMA |
| `ACP` | Auction closing price | Current day |
| `TC` | Trade count | |

#### Book Depth (`depth`)

- 1: Top of book (best BID/ASK). Lower bandwidth/latency. Default value.
- 2..5: Up to 5 levels per side. More book granularity, more data to process.

Level ordering (when depth > 1):
- `BI` (bids): best → worst (descending price)
- `OF` (offers): best → worst (ascending price)

For more details see `docs/primary-api.md`.

### Real-time Market Data

```go
// Subscribe to real-time market data
subscription, err := client.SubscribeMarketData(ctx, 
    []string{"DLR/DIC21", "DOFeb25"}, // Symbols
    []model.MDEntry{model.MDBids, model.MDOffers, model.MDLast},
    5, // Depth
    model.MarketROFEX,
)
if err != nil {
    log.Fatal("Error subscribing:", err)
}
defer subscription.Close()

// Process events
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

### Real-time Order Reports

```go
// Subscribe to order reports (fetch account first)
accounts, err := client.Accounts(ctx)
if err != nil {
    log.Fatal("Error getting accounts:", err)
}
if len(accounts.Accounts) == 0 {
    log.Fatal("No accounts available for this user")
}
account := accounts.Accounts[0].Name

orderSub, err := client.SubscribeOrderReport(ctx, account, true)
if err != nil {
    log.Fatal("Error subscribing to orders:", err)
}
defer orderSub.Close()

// Process order reports
for {
    select {
    case report := <-orderSub.Events:
        fmt.Printf("Order Report: %+v\n", report)
    case err := <-orderSub.Errs:
        fmt.Printf("Order report error: %v\n", err)
    case <-ctx.Done():
        return
    }
}
```

## 🛠️ Advanced Features

### Configuration with Logging

```go
import "log/slog"

client, err := rofex.NewClient(
    rofex.WithEnvironment(model.EnvironmentRemarket),
    rofex.WithAuth(auth),
    rofex.WithLogger(slog.Default()), // Enable logging
)
```

### Custom Rate Limiting

```go
import "golang.org/x/time/rate"

// Create custom limiter
limiter := rate.NewLimiter(rate.Limit(10), 1) // 10 requests/second, burst=1

client, err := rofex.NewClient(
    rofex.WithRateLimit(limiter),
    // ... other options
)
```

### Custom HTTP Client

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
    // ... other options
)
```

## 📖 API Documentation

### Instruments and Reference Data

```go
// Get all instruments
instruments, err := client.InstrumentsAll(ctx)

// Get instruments with full details
detailed, err := client.InstrumentsDetails(ctx)

// Get instruments by segment
bySegment, err := client.InstrumentsBySegment(ctx, 
    model.MarketROFEX, 
    []model.MarketSegment{model.SegmentDDF, model.SegmentDDA},
)

// Get instruments by CFI code
byCFI, err := client.InstrumentsByCFICode(ctx,
    []model.CFICode{model.CFIFuture, model.CFIStock},
)

// Get specific instrument details
detail, err := client.InstrumentDetail(ctx, "DLR/DIC21", model.MarketROFEX)
```

### Order Management

```go
// Query all orders for an account
allOrders, err := client.AllOrdersStatus(ctx, "account")

// Query active orders
activeOrders, err := client.ActiveOrders(ctx, "account")

// Query filled orders
filledOrders, err := client.FilledOrders(ctx, "account")

// Query order by Order ID
orderByID, err := client.OrderByOrderID(ctx, "order_id")

// Query order by Execution ID
orderByExec, err := client.OrderByExecID(ctx, "exec_id")

// Modify an existing order
newQty := int64(20)
newPrice := 19.00
replaceResp, err := client.ReplaceOrder(ctx, "client_order_id", "", &newQty, &newPrice)
```

### Account Information

```go
// Get accounts associated with user
accounts, err := client.Accounts(ctx)

// Get account positions
positions, err := client.AccountPosition(ctx, "account")

// Get detailed positions
detailedPos, err := client.DetailedPosition(ctx, "account")

// Get account report
report, err := client.AccountReport(ctx, "account")
```

### Historical Data

```go
// Get historical trades
from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

trades, err := client.HistoricTrades(ctx, "DLR/DIC21", model.MarketROFEX, from, to)
```

## 🌐 Environments

### reMarkets (Sandbox)
- **URL**: `https://api.remarkets.primary.com.ar/`
- **WebSocket**: `wss://api.remarkets.primary.com.ar/`  
- **Purpose**: Testing and development
- **Registration**: [remarkets.primary.ventures](https://remarkets.primary.ventures)

### Production (mandatory URLs)
- You must explicitly set BaseURL and WSURL because providers vary.
- Provider examples:
  - Primary: [https://api.primary.com.ar/](https://api.primary.com.ar/) (WS: `wss://api.primary.com.ar/`)
  - Eco Valores SA: [https://api.eco.xoms.com.ar/](https://api.eco.xoms.com.ar/) (WS: `wss://api.eco.xoms.com.ar/`)
- If not set, the client returns an error and refuses to continue.

```go
// Configure for production (mandatory URLs)
client, err := rofex.NewClient(
    rofex.WithEnvironment(model.EnvironmentLive),
    rofex.WithBaseURL("https://api.primary.com.ar/"),
    rofex.WithWSURL("wss://api.primary.com.ar/"),
    rofex.WithAuth(rofex.NewPasswordAuth(rofex.Credentials{Username: user, Password: pass})),
)
```

## 🔧 Important Data Types

### Trading Enums

```go
// Order types
model.OrderTypeLimit         // Limit order
model.OrderTypeMarket        // Market order
model.OrderTypeMarketToLimit // Market to limit

// Order sides
model.Buy   // Buy
model.Sell  // Sell

// Time in Force
model.Day               // Valid during the day
model.ImmediateOrCancel // IOC - Immediate or cancel
model.FillOrKill        // FOK - Fill or kill
model.GoodTillDate      // GTD - Good till date

// Market Data Entries
model.MDBids              // BI - Bid offers
model.MDOffers            // OF - Ask offers
model.MDLast              // LA - Last price
model.MDOpeningPrice      // OP - Opening price
model.MDClosePrice        // CL - Closing price
model.MDSettlementPrice   // SE - Settlement price
model.MDTradeVolume       // TV - Traded volume
model.MDOpenInterest      // OI - Open interest
```

### Order Structure

```go
type NewOrder struct {
    Symbol         string              // Instrument symbol
    Market         model.Market        // Market (ROFX)
    Side           model.Side          // Buy/Sell
    Type           model.OrderType     // Order type
    Qty            int64               // Quantity
    Price          *float64           // Price (required for LIMIT)
    TIF            model.TimeInForce   // Time in Force
    Account        string              // Account
    CancelPrevious bool                // Cancel previous orders
    Iceberg        bool                // Iceberg order
    ExpireDate     *string            // Expiration date (GTD)
    DisplayQty     *int64             // Display quantity (iceberg)
    AllOrNone      bool               // All or none (WebSocket)
    WSClOrdID      *string            // Client Order ID (WebSocket)
}
```

## ⚠️ Best Practices

### 1. Error Handling
```go
if err != nil {
    var httpErr *rofex.HTTPError
    if errors.As(err, &httpErr) {
        fmt.Printf("HTTP Error %d: %s\n", httpErr.StatusCode, string(httpErr.Body))
        return
    }
    var validationErr *rofex.ValidationError  
    if errors.As(err, &validationErr) {
        fmt.Printf("Validation Error in %s: %s\n", validationErr.Field, validationErr.Msg)
        return
    }
    var authErr *rofex.AuthError
    if errors.As(err, &authErr) {
        fmt.Printf("Authentication Error: %s\n", authErr.Msg)
        // Retry with new credentials
        return
    }
    var tempErr *rofex.TemporaryError
    if errors.As(err, &tempErr) {
        fmt.Printf("Temporary Error: %s - retry in a few seconds\n", tempErr.Error())
        time.Sleep(5 * time.Second)
        // Retry the operation
        return
    }
    fmt.Printf("Uncategorized error: %v\n", err)
}
```

### 2. Order Status Verification
```go
// Always verify status after sending an order
order, err := client.SendOrder(ctx, newOrder)
if err != nil {
    return fmt.Errorf("send order failed: %w", err)
}

// Wait for market confirmation
for i := 0; i < 5; i++ {
    status, err := client.OrderStatus(ctx, order.Order.ClientID, "")
    if err != nil {
        return fmt.Errorf("order status failed: %w", err)
    }
    
    switch status.Order.Status {
    case "NEW":
        fmt.Println("Order accepted by market")
        return nil
    case "REJECTED":
        return fmt.Errorf("order rejected: %s", status.Order.Text)
    case "PENDING_NEW":
        time.Sleep(time.Second)
        continue
    }
}
```

### 3. Rate Limiting
```go
// Respect API limits according to documentation
// - Authentication: 1 request/day (token lasts 24hrs)
// - Market Data: Use WebSocket for real-time
// - Orders: Maximum 1 request/second for cancellations
// - Reports: 1 request every 5 seconds
```

### 4. WebSocket Resource Management
```go
// Always close subscriptions
subscription, err := client.SubscribeMarketData(ctx, symbols, entries, depth, market)
if err != nil {
    return err
}
defer subscription.Close() // Important: always close

// Handle contexts for cancellation
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// For graceful shutdown
go func() {
    <-shutdownCh
    cancel() // This will close all WebSocket subscriptions
}()
```

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 📚 Documentation and Resources

### Primary Documentation
- **[Primary API v1.21 (docs/primary-api.md)](docs/primary-api.md)** - Complete official documentation
- **[API Data JSON (docs/api_data.json)](docs/api_data.json)** - Schemas and endpoints
- **[Postman Collection (docs/rest.postman_collection.json)](docs/rest.postman_collection.json)** - API tests

### Primary Online Resources
- **[API Hub - Trading Documentation](https://apihub.primary.com.ar/assets/apidoc/trading/index.html)** - Complete official documentation
- **[Swagger Documentation](https://api.remarkets.primary.com.ar/api-docs/index.html)**

### How to get api_data.json
The `api_data.json` file contains the complete API specification and can be obtained from:
1. **From this repository**: [docs/api_data.json](docs/api_data.json)
2. **From the API**: `GET https://api.remarkets.primary.com.ar/rest/api-data` (requires authentication)
3. **From Swagger**: Export from [api-docs](https://api.remarkets.primary.com.ar/api-docs/index.html)

## 📞 Support

- **Official Documentation**: [apihub.primary.com.ar](https://apihub.primary.com.ar)
- **reMarkets (Sandbox)**: [remarkets.primary.ventures](https://remarkets.primary.ventures)

- **Issues**: [GitHub Issues](https://github.com/carvalab/rofex-go/issues)

## 🙏 Acknowledgments

- **Primary (ROFEX)** for providing a robust and well-documented API
- **pyRofex** for serving as an implementation reference
- The Go community for excellent libraries that make this SDK possible

---

**Note**: This SDK is not officially endorsed by Primary. It's an open-source project maintained by the community.
