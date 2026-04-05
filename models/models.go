// Package models defines the data structures used across the simulator.
package models

import "time"

// Attributes represents the 20 attributes of an auction object.
type Attributes struct {
	Category       string  `json:"attr1_category"`
	Quality        string  `json:"attr2_quality"`
	BasePrice      float64 `json:"attr3_base_price"`
	Weight         float64 `json:"attr4_weight"`
	Rating         float64 `json:"attr5_rating"`
	Popularity     int     `json:"attr6_popularity"`
	Stock          int     `json:"attr7_stock"`
	Discount       float64 `json:"attr8_discount"`
	ShippingCost   float64 `json:"attr9_shipping_cost"`
	ReturnPolicy   bool    `json:"attr10_return_policy"`
	Warranty       bool    `json:"attr11_warranty"`
	BrandTier      int     `json:"attr12_brand_tier"`
	Seasonality    float64 `json:"attr13_seasonality"`
	DemandIndex    float64 `json:"attr14_demand_index"`
	SupplyIndex    float64 `json:"attr15_supply_index"`
	Margin         float64 `json:"attr16_margin"`
	ClickRate      float64 `json:"attr17_click_rate"`
	ConversionRate float64 `json:"attr18_conversion_rate"`
	Impressions    int     `json:"attr19_impressions"`
	GeoRegion      string  `json:"attr20_geo_region"`
}

// Bid represents a single bid from a bidder.
type Bid struct {
	BidderID   int       `json:"bidder_id"`
	BidderName string    `json:"bidder_name"`
	AuctionID  string    `json:"auction_id"`
	Price      float64   `json:"price"`
	LatencyMs  int64     `json:"latency_ms"`
	RespondedAt time.Time `json:"responded_at"`
}

// AuctionResult holds the complete outcome of a single auction.
type AuctionResult struct {
	AuctionID             string     `json:"auction_id"`
	Attributes            Attributes `json:"attributes"`
	TotalBidders          int        `json:"total_bidders"`
	BidsReceived          int        `json:"bids_received"`
	BidsTimedOutOrDeclined int       `json:"bids_timed_out_or_declined"`
	Bids                  []Bid      `json:"bids"`
	Winner                *Winner    `json:"winner"`
	Timing                Timing     `json:"timing"`
}

// Winner holds info about the auction winner.
type Winner struct {
	BidderID   int     `json:"bidder_id"`
	BidderName string  `json:"bidder_name"`
	Price      float64 `json:"price"`
	LatencyMs  int64   `json:"latency_ms"`
}

// Timing holds start/end timestamps and duration for an auction.
type Timing struct {
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	DurationMs int64     `json:"duration_ms"`
}

// Summary is the consolidated output written to _SUMMARY.json.
type Summary struct {
	Configuration SummaryConfig       `json:"configuration"`
	Timing        SummaryTiming       `json:"timing"`
	Resources     ResourceSnapshot    `json:"resources"`
	Auctions      []AuctionSummaryRow `json:"auctions"`
}

// SummaryConfig captures the run configuration.
type SummaryConfig struct {
	TotalBidders       int    `json:"total_bidders"`
	ConcurrentAuctions int    `json:"concurrent_auctions"`
	AttributesPerAuction int  `json:"attributes_per_auction"`
	AuctionTimeoutMs   int64  `json:"auction_timeout_ms"`
	VCPUs              int    `json:"vcpus"`
	MemoryMB           int    `json:"memory_mb"`
}

// SummaryTiming captures the global timing.
type SummaryTiming struct {
	FirstAuctionStart time.Time `json:"first_auction_start"`
	LastAuctionEnd    time.Time `json:"last_auction_end"`
	TotalWallClockMs  int64     `json:"total_wall_clock_ms"`
}

// AuctionSummaryRow is a compact per-auction summary for the consolidated output.
type AuctionSummaryRow struct {
	AuctionID    string `json:"auction_id"`
	BidsReceived int    `json:"bids_received"`
	TimedOut     int    `json:"timed_out"`
	Winner       string `json:"winner"`
	DurationMs   int64  `json:"duration_ms"`
}

// ResourceSnapshot captures CPU and memory usage at a point in time.
type ResourceSnapshot struct {
	ConfiguredVCPUs    int     `json:"configured_vcpus"`
	ConfiguredMemoryMB int     `json:"configured_memory_mb"`
	NumGoroutines      int     `json:"num_goroutines"`
	HeapAllocMB        float64 `json:"heap_alloc_mb"`
	HeapSysMB          float64 `json:"heap_sys_mb"`
	TotalAllocMB       float64 `json:"total_alloc_mb"`
	NumGC              uint32  `json:"num_gc"`
	NumCPU             int     `json:"num_cpu"`
}
