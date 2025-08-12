package model

// BookLevel representa un nivel de precio en el libro de órdenes.
type BookLevel struct {
	Price float64 `json:"price"`
	Size  float64 `json:"size"`
}

// Entry representa un registro de MarketData con precio, size y fecha opcionales.
// Se usa para entries como LA (Last), CL (Close) y SE (Settlement) que pueden
// incluir size y date según la documentación de Primary API.
type Entry struct {
	Price *float64 `json:"price,omitempty"`
	Size  *float64 `json:"size,omitempty"`
	Date  *int64   `json:"date,omitempty"`
}

// MarketData aggregates selected entries for a symbol.
// Only populated keys requested via entries will be present.
//
// Depth and ordering notes (per docs/primary-api.md):
// - When depth = 1 (default), BI and OF return only top-of-book (best level per side).
// - When depth > 1 (up to 5), BI and OF contain multiple levels ordered by price.
//   - BI (bids) sorted best to worst buy price (descending).
//   - OF (offers) sorted best to worst sell price (ascending).
type MarketData struct {
	// Bids (BI): buy-side levels; sorted from best to worse.
	Bids []BookLevel `json:"BI,omitempty"`
	// Offers (OF): sell-side levels; sorted from best to worse.
	Offers []BookLevel `json:"OF,omitempty"`
	LA     *Entry      `json:"LA,omitempty"`

	// OP in documentation can come as a plain number
	OpeningPrice *float64 `json:"OP,omitempty"`
	// CL and SE can bring structure {price,size,date}
	CL              *Entry   `json:"CL,omitempty"`
	SE              *Entry   `json:"SE,omitempty"`
	HighPrice       *float64 `json:"HI,omitempty"`
	LowPrice        *float64 `json:"LO,omitempty"`
	TradeVolume     *float64 `json:"TV,omitempty"`
	OpenInterest    *Entry   `json:"OI,omitempty"`
	IndexValue      *float64 `json:"IV,omitempty"`
	EffectiveVolume *float64 `json:"EV,omitempty"`
	NominalVolume   *float64 `json:"NV,omitempty"`
	ACP             *float64 `json:"ACP,omitempty"`
	TradeCount      *int64   `json:"TC,omitempty"`
}
