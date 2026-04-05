// Package auction implements the single-auction runner.
// It broadcasts attributes to all bidders using goroutines, collects bids
// via a channel, and enforces the auction timeout using context.WithTimeout.
package auction

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/auction-simulator/bidder"
	"github.com/auction-simulator/config"
	"github.com/auction-simulator/models"
)

// GenerateAttributes creates the 20 random attributes for an auction object.
func GenerateAttributes() models.Attributes {
	categories := []string{"electronics", "fashion", "home", "sports", "automotive",
		"books", "toys", "health", "garden", "food"}
	qualities := []string{"new", "refurbished", "used-good", "used-fair"}
	regions := []string{"NA", "EU", "APAC", "LATAM", "MEA"}

	return models.Attributes{
		Category:       categories[rand.Intn(len(categories))],
		Quality:        qualities[rand.Intn(len(qualities))],
		BasePrice:      roundF(1 + rand.Float64()*99),
		Weight:         roundF(0.1 + rand.Float64()*49.9),
		Rating:         float64(int((1+rand.Float64()*4)*10)) / 10,
		Popularity:     rand.Intn(1000) + 1,
		Stock:          rand.Intn(501),
		Discount:       roundF(rand.Float64() * 0.5),
		ShippingCost:   roundF(rand.Float64() * 20),
		ReturnPolicy:   rand.Float64() > 0.5,
		Warranty:       rand.Float64() > 0.5,
		BrandTier:      rand.Intn(5) + 1,
		Seasonality:    roundF(rand.Float64()),
		DemandIndex:    roundF(rand.Float64()),
		SupplyIndex:    roundF(rand.Float64()),
		Margin:         roundF(0.05 + rand.Float64()*0.55),
		ClickRate:      float64(int(rand.Float64()*0.3*10000)) / 10000,
		ConversionRate: float64(int(rand.Float64()*0.1*10000)) / 10000,
		Impressions:    rand.Intn(99901) + 100,
		GeoRegion:      regions[rand.Intn(len(regions))],
	}
}

func roundF(v float64) float64 {
	return float64(int(v*100)) / 100
}

// Run executes a single auction: fans out bid requests to all bidders,
// collects results via a channel, and picks the highest bidder as winner.
func Run(auctionNumber int, bidderPool []bidder.Bidder) models.AuctionResult {
	auctionID := fmt.Sprintf("AUCTION-%03d", auctionNumber)
	attrs := GenerateAttributes()
	startTime := time.Now()

	// Context enforces the auction timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.AuctionTimeout)
	defer cancel()

	// Channel to collect bids from all bidders
	bidCh := make(chan *models.Bid, len(bidderPool))

	var wg sync.WaitGroup
	for i := range bidderPool {
		wg.Add(1)
		go func(b *bidder.Bidder) {
			defer wg.Done()
			bid := b.GenerateBid(ctx, auctionID, attrs)
			if bid != nil {
				bidCh <- bid
			}
		}(&bidderPool[i])
	}

	// Close channel once all bidders finish (or time out)
	go func() {
		wg.Wait()
		close(bidCh)
	}()

	// Collect all valid bids
	var validBids []models.Bid
	for bid := range bidCh {
		validBids = append(validBids, *bid)
	}

	endTime := time.Now()

	// Determine winner: highest price
	var winner *models.Winner
	for _, b := range validBids {
		if winner == nil || b.Price > winner.Price {
			winner = &models.Winner{
				BidderID:   b.BidderID,
				BidderName: b.BidderName,
				Price:      b.Price,
				LatencyMs:  b.LatencyMs,
			}
		}
	}

	return models.AuctionResult{
		AuctionID:              auctionID,
		Attributes:             attrs,
		TotalBidders:           len(bidderPool),
		BidsReceived:           len(validBids),
		BidsTimedOutOrDeclined: len(bidderPool) - len(validBids),
		Bids:                   validBids,
		Winner:                 winner,
		Timing: models.Timing{
			StartTime:  startTime,
			EndTime:    endTime,
			DurationMs: endTime.Sub(startTime).Milliseconds(),
		},
	}
}
