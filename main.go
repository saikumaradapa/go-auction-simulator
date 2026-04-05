// Auction Simulator — Entry Point
//
// Runs 40 concurrent auctions with 100 bidders each.
// Each auction enforces a timeout; the highest bid wins.
// Measures wall-clock time from first auction start to last completion.
// Writes a separate JSON output file per auction plus a consolidated summary.
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/auction-simulator/auction"
	"github.com/auction-simulator/bidder"
	"github.com/auction-simulator/config"
	"github.com/auction-simulator/resource"
	"github.com/auction-simulator/models"
)

func main() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Set GOMAXPROCS to match configured vCPU limit for resource standardization
	runtime.GOMAXPROCS(config.ResourceLimits.VCPUs)

	fmt.Println("+==================================================+")
	fmt.Println("|    AUCTION SIMULATOR [Goroutines + Channels]     |")
	fmt.Println("+==================================================+")
	fmt.Println()
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Bidders             : %d\n", config.TotalBidders)
	fmt.Printf("  Concurrent Auctions : %d\n", config.ConcurrentAuctions)
	fmt.Printf("  Attributes/Auction  : %d\n", config.AttributesPerAuction)
	fmt.Printf("  Auction Timeout     : %v\n", config.AuctionTimeout)
	fmt.Printf("  GOMAXPROCS          : %d (standardized to %d vCPUs)\n", runtime.GOMAXPROCS(0), config.ResourceLimits.VCPUs)
	fmt.Printf("  Memory Limit        : %d MB (enforced via Docker)\n", config.ResourceLimits.MemoryMB)

	// Resource snapshot before auctions
	snapBefore := resource.Snapshot()

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output dir: %v\n", err)
		os.Exit(1)
	}

	// Create shared bidder pool
	bidderPool := bidder.NewBidderPool(config.TotalBidders)
	fmt.Printf("\nCreated %d bidders.\n", len(bidderPool))

	// ── Run all 40 auctions concurrently ──
	fmt.Printf("\nStarting %d auctions concurrently...\n\n", config.ConcurrentAuctions)

	// Start goroutine tracker to capture peak concurrency
	tracker := resource.StartGoroutineTracker()

	globalStart := time.Now()

	results := make([]models.AuctionResult, config.ConcurrentAuctions)
	var wg sync.WaitGroup

	for i := 0; i < config.ConcurrentAuctions; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = auction.Run(idx+1, bidderPool)
		}(i)
	}

	wg.Wait()
	globalEnd := time.Now()
	totalDuration := globalEnd.Sub(globalStart)

	// Stop goroutine tracker
	peakGoroutines := tracker.Stop()
	fmt.Printf("  Peak goroutines during auctions: %d\n\n", peakGoroutines)

	// ── Print per-auction results and write output files ──
	summaryRows := make([]models.AuctionSummaryRow, 0, len(results))

	for _, r := range results {
		// Write per-auction JSON file
		filePath := filepath.Join(config.OutputDir, r.AuctionID+".json")
		data, _ := json.MarshalIndent(r, "", "  ")
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to write %s: %v\n", filePath, err)
		}

		winnerStr := "No winner (no bids)"
		if r.Winner != nil {
			winnerStr = fmt.Sprintf("%s @ $%.2f (%dms)", r.Winner.BidderName, r.Winner.Price, r.Winner.LatencyMs)
		}

		fmt.Printf("  %s: %d bids | Winner: %s | %dms\n",
			r.AuctionID, r.BidsReceived, winnerStr, r.Timing.DurationMs)

		summaryRows = append(summaryRows, models.AuctionSummaryRow{
			AuctionID:    r.AuctionID,
			BidsReceived: r.BidsReceived,
			TimedOut:     r.BidsTimedOutOrDeclined,
			Winner:       winnerStr,
			DurationMs:   r.Timing.DurationMs,
		})
	}

	// ── Timing summary ──
	fmt.Println()
	fmt.Println("+==================================================+")
	fmt.Println("|                TIMING SUMMARY                    |")
	fmt.Println("+==================================================+")
	fmt.Printf("| First auction started at : %-23s|\n", globalStart.Format("2006-01-02 15:04:05.000"))
	fmt.Printf("| Last auction completed at: %-23s|\n", globalEnd.Format("2006-01-02 15:04:05.000"))
	fmt.Printf("| Total wall-clock time    : %-23s|\n", fmt.Sprintf("%dms (%.3fs)", totalDuration.Milliseconds(), totalDuration.Seconds()))
	fmt.Printf("| Avg time per auction     : %-23s|\n", fmt.Sprintf("%.2fms", float64(totalDuration.Milliseconds())/float64(config.ConcurrentAuctions)))
	fmt.Println("+==================================================+")

	// Resource comparison: before vs after with deltas
	snapAfter := resource.Snapshot()
	resource.PrintComparison(snapBefore, snapAfter)

	// ── Write consolidated summary ──
	summary := models.Summary{
		Configuration: models.SummaryConfig{
			TotalBidders:         config.TotalBidders,
			ConcurrentAuctions:   config.ConcurrentAuctions,
			AttributesPerAuction: config.AttributesPerAuction,
			AuctionTimeoutMs:     config.AuctionTimeout.Milliseconds(),
			VCPUs:                config.ResourceLimits.VCPUs,
			MemoryMB:             config.ResourceLimits.MemoryMB,
		},
		Timing: models.SummaryTiming{
			FirstAuctionStart: globalStart,
			LastAuctionEnd:    globalEnd,
			TotalWallClockMs:  totalDuration.Milliseconds(),
		},
		Resources: snapAfter,
		Auctions:  summaryRows,
	}

	summaryData, _ := json.MarshalIndent(summary, "", "  ")
	summaryPath := filepath.Join(config.OutputDir, "_SUMMARY.json")
	if err := os.WriteFile(summaryPath, summaryData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to write summary: %v\n", err)
	}

	fmt.Printf("\nOutput files written to: %s/\n", config.OutputDir)
	fmt.Printf("Summary: %s\n", summaryPath)
	fmt.Println("\nCompleted.")
}
