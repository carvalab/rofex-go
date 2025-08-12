package model

// Environment representa el entorno objetivo.
type Environment string

const (
	EnvironmentRemarket Environment = "remarket"
	EnvironmentLive     Environment = "live"
)

// Market representa el identificador del mercado.
type Market string

const (
	MarketROFEX Market = "ROFX"
	MarketMERV  Market = "MERV"
)

// MarketSegment identifies the market segment associated to instruments.
type MarketSegment string

const (
	SegmentDDF   MarketSegment = "DDF"
	SegmentDDA   MarketSegment = "DDA"
	SegmentDUAL  MarketSegment = "DUAL"
	SegmentUDDF  MarketSegment = "U-DDF"
	SegmentUDDA  MarketSegment = "U-DDA"
	SegmentUDUAL MarketSegment = "U-DUAL"
	SegmentMERV  MarketSegment = "MERV"
)

// CFICode identifies the type of instrument.
type CFICode string

const (
	CFIStock      CFICode = "ESXXXX"
	CFIBond       CFICode = "DBXXXX"
	CFICallStock  CFICode = "OCASPS"
	CFIPutStock   CFICode = "OPASPS"
	CFIFuture     CFICode = "FXXXSX"
	CFIPutFuture  CFICode = "OPAFXS"
	CFICallFuture CFICode = "OCAFXS"
	CFICedear     CFICode = "EMXXXX"
	CFIOn         CFICode = "DBXXFR"
)

// TimeInForce identifies the active time of an order.
//
// Primary API (docs/primary-api.md) definitions:
// - DAY: Orden válida solo por la rueda del día; expira al cierre
// - IOC (Immediate Or Cancel): Ejecuta inmediatamente la parte disponible y cancela el resto
// - FOK (Fill Or Kill): Debe ejecutarse completamente de inmediato o se cancela
// - GTD (Good Till Date): Válida hasta la fecha indicada (require expireDate)
type TimeInForce string

const (
	// Day: Orden válida solo por la rueda del día; expira al cierre.
	Day TimeInForce = "DAY"
	// ImmediateOrCancel (IOC): Ejecuta inmediatamente la parte disponible y cancela el resto.
	ImmediateOrCancel TimeInForce = "IOC"
	// FillOrKill (FOK): Debe ejecutarse completamente de inmediato o se cancela.
	FillOrKill TimeInForce = "FOK"
	// GoodTillDate (GTD): Válida hasta la fecha indicada (requiere expireDate).
	GoodTillDate TimeInForce = "GTD"
)

// Side identifies the side of an order.
//
// Primary API (docs/primary-api.md):
// - BUY: Compra
// - SELL: Venta
type Side string

const (
	// Buy: Compra.
	Buy Side = "BUY"
	// Sell: Venta.
	Sell Side = "SELL"
)

// OrderType identifies the different order types.
//
// Primary API (docs/primary-api.md):
// - LIMIT: Orden con precio límite
// - MARKET: Orden a mercado
// - MARKET_TO_LIMIT: Orden que intenta a mercado y, si hay remanente, queda como LIMIT al mejor precio
type OrderType string

const (
	// OrderTypeLimit: Orden con precio límite.
	OrderTypeLimit OrderType = "LIMIT"
	// OrderTypeMarket: Orden a mercado.
	OrderTypeMarket OrderType = "MARKET"
	// OrderTypeMarketToLimit: Intenta a mercado y, si hay remanente, queda LIMIT al mejor precio.
	OrderTypeMarketToLimit OrderType = "MARKET_TO_LIMIT"
)

// MDEntry identifies the market data entries for an instrument.
//
// Primary API (docs/primary-api.md - "Descripción de MarketData Entries"):
// - BI: Mejor oferta de compra (BIDS). Puede ser una lista de niveles si depth>1
// - OF: Mejor oferta de venta (OFFERS). Puede ser una lista de niveles si depth>1
// - LA: Último precio operado (LAST)
// - OP: Precio de apertura (OPENING PRICE)
// - CL: Precio de cierre de la rueda anterior (CLOSING PRICE)
// - SE: Precio de ajuste para futuros (SETTLEMENT)
// - HI: Precio máximo de la rueda (HIGH)
// - LO: Precio mínimo de la rueda (LOW)
// - TV: Volumen operado (TRADE VOLUME)
// - OI: Interés abierto (OPEN INTEREST, suele incluir size/date)
// - IV: Valor del índice (INDEX VALUE)
// - EV: Volumen efectivo (ByMA)
// - NV: Volumen nominal (ByMA)
// - ACP: Precio de subasta del día corriente (AUCTION PRICE)
// - TC: Cantidad de trades (TRADE COUNT)
type MDEntry string

const (
	// MDBids (BI): Mejor oferta de compra; lista de niveles si depth > 1.
	MDBids MDEntry = "BI"
	// MDOffers (OF): Mejor oferta de venta; lista de niveles si depth > 1.
	MDOffers MDEntry = "OF"
	// MDLast (LA): Último precio operado.
	MDLast MDEntry = "LA"
	// MDOpeningPrice (OP): Precio de apertura.
	MDOpeningPrice MDEntry = "OP"
	// MDClosePrice (CL): Precio de cierre de la rueda anterior.
	MDClosePrice MDEntry = "CL"
	// MDSettlementPrice (SE): Precio de ajuste (solo futuros).
	MDSettlementPrice MDEntry = "SE"
	// MDHighPrice (HI): Precio máximo de la rueda.
	MDHighPrice MDEntry = "HI"
	// MDLowPrice (LO): Precio mínimo de la rueda.
	MDLowPrice MDEntry = "LO"
	// MDTradeVolume (TV): Volumen operado.
	MDTradeVolume MDEntry = "TV"
	// MDOpenInterest (OI): Interés abierto (puede incluir size/date).
	MDOpenInterest MDEntry = "OI"
	// MDIndexValue (IV): Valor del índice (solo índices).
	MDIndexValue MDEntry = "IV"
	// MDTradeEffectiveVol (EV): Volumen efectivo (ByMA).
	MDTradeEffectiveVol MDEntry = "EV"
	// MDNominalVolume (NV): Volumen nominal (ByMA).
	MDNominalVolume MDEntry = "NV"
	// MDACP (ACP): Precio de subasta del día corriente.
	MDACP MDEntry = "ACP"
	// MDTradeCount (TC): Cantidad de trades.
	MDTradeCount MDEntry = "TC"
)

// WSMessageType identifies the WebSocket message "type" field as defined by Primary API.
//
// Values according to docs/primary-api.md:
//
//	Client-to-server (sent by client):
//	- "smd": Subscribe to Market Data updates
//	- "os": Subscribe to Execution Reports (Order Subscription)
//	- "no": Send a New Order over WebSocket
//	- "co": Cancel an existing Order over WebSocket
//	Server-to-client (received from server):
//	- "md": Market Data message
//	- "or": Order Report (Execution Report) message
type WSMessageType string

const (
	// Client-to-server message types (sent by client)

	// WSMessageSubscribeMarketData ("smd"): Suscribe a Market Data en tiempo real.
	WSMessageSubscribeMarketData WSMessageType = "smd"

	// WSMessageOrderSubscription ("os"): Suscribe a Execution Reports (órdenes).
	WSMessageOrderSubscription WSMessageType = "os"

	// WSMessageNewOrder ("no"): Envía una nueva orden vía WebSocket.
	WSMessageNewOrder WSMessageType = "no"

	// WSMessageCancelOrder ("co"): Cancela una orden vía WebSocket.
	WSMessageCancelOrder WSMessageType = "co"

	// Server-to-client message types (received from server)

	// WSMessageMarketData ("md"): Market Data recibido del servidor.
	// Este mensaje se recibe cuando hay cambios en los datos de mercado
	// de los instrumentos suscritos.
	WSMessageMarketData WSMessageType = "md"

	// WSMessageOrderReport ("or"): Execution Report recibido del servidor.
	// Este mensaje se recibe cuando hay cambios de estado en las órdenes
	// (confirmaciones, ejecuciones, cancelaciones, rechazos).
	WSMessageOrderReport WSMessageType = "or"
)
