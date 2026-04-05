// Package bidder implements the simulated bidder logic.
// Each bidder receives auction attributes via a context-aware call and may
// or may not respond with a bid within the auction timeout.
package bidder

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/auction-simulator/config"
	"github.com/auction-simulator/models"
)

// Bidder represents a single participant in auctions.
type Bidder struct {
	ID   int
	Name string
}

// NewBidderPool creates the shared pool of bidders.
func NewBidderPool(count int) []Bidder {
	pool := make([]Bidder, count)
	for i := 0; i < count; i++ {
		pool[i] = Bidder{
			ID:   i + 1,
			Name: fmt.Sprintf("Bidder-%03d", i+1),
		}
	}
	return pool
}

// GenerateBid simulates a bidder evaluating attributes and optionally returning a bid.
// It respects the context deadline (auction timeout). Returns nil if the bidder
// declines to bid or the context expires before the response is ready.
func (b *Bidder) GenerateBid(ctx context.Context, auctionID string, attrs models.Attributes) *models.Bid {
	// Simulate variable network + processing latency
	latency := config.BidMinLatency + time.Duration(rand.Int63n(int64(config.BidMaxLatency-config.BidMinLatency)))

	select {
	case <-time.After(latency):
		// Bidder finished processing in time
	case <-ctx.Done():
		// Auction timed out before bidder could respond
		return nil
	}

	// Not every bidder responds
	if rand.Float64() > config.BidResponseProbability {
		return nil
	}

	price := config.BidMinPrice + rand.Float64()*(config.BidMaxPrice-config.BidMinPrice)
	price = float64(int(price*100)) / 100 // round to 2 decimals

	return &models.Bid{
		BidderID:    b.ID,
		BidderName:  b.Name,
		AuctionID:   auctionID,
		Price:       price,
		LatencyMs:   latency.Milliseconds(),
		RespondedAt: time.Now(),
	}
}
