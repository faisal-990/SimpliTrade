package dto

// MarketStatusDTO tells the UI whether US markets are open and when the next
// transition happens — powering the live "markets open / opens in…" banner.
type MarketStatusDTO struct {
	Open       bool  `json:"open"`
	NextChange int64 `json:"next_change"` // unix seconds; 0 if unknown
}
