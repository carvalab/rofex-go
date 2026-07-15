package model

// SegmentsResponse contains the list of segments returned by the API.
type SegmentsResponse struct {
	Status   string `json:"status"`
	Segments []struct {
		MarketSegmentID string `json:"marketSegmentId"`
		MarketID        string `json:"marketId"`
	} `json:"segments"`
}
