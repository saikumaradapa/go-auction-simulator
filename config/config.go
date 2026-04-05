// Package config centralizes all tunable parameters for the Auction Simulator.
package config

import "time"

const (
	// TotalBidders is the number of bidders participating across all auctions.
	TotalBidders = 100

	// ConcurrentAuctions is the number of auctions running at the same time.
	ConcurrentAuctions = 40

	// AttributesPerAuction is the number of attributes describing each auction object.
	AttributesPerAuction = 20

	// AuctionTimeout is the maximum duration an auction waits for bids.
	AuctionTimeout = 200 * time.Millisecond

	// BidResponseProbability is the chance (0.0–1.0) that a bidder responds.
	BidResponseProbability = 0.6

	// BidMinPrice and BidMaxPrice define the bid price range.
	BidMinPrice = 0.01
	BidMaxPrice = 50.0

	// BidMinLatency and BidMaxLatency define simulated bidder response delay.
	BidMinLatency = 5 * time.Millisecond
	BidMaxLatency = 250 * time.Millisecond

	// OutputDir is the directory where per-auction JSON output files are written.
	OutputDir = "output"
)

// ResourceLimits defines the standardized resource constraints.
// Enforced via Docker: docker run --cpus=2 --memory=512m
var ResourceLimits = struct {
	VCPUs    int
	MemoryMB int
}{
	VCPUs:    2,
	MemoryMB: 512,
}
