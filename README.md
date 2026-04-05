# Auction Simulator

**Backend Engineer — Take Home Assessment**

A Go-based Auction Simulator that runs 40 concurrent auctions with 100 bidders, enforces per-auction timeouts, declares winners, and measures end-to-end execution time under standardized resource constraints.

---

## How It Works

1. **100 Bidders** are created once and shared across all auctions.
2. **40 Auctions** launch concurrently using goroutines. Each auction:
   - Generates **20 random attributes** describing the auctioned object.
   - Broadcasts attributes to all 100 bidders (each bidder runs in its own goroutine).
   - Each bidder may or may not respond (~60% probability) with a random latency (5–250ms).
   - A **`context.WithTimeout`** enforces the auction deadline (200ms). Bids arriving after the timeout are discarded.
   - The **highest bid wins**.
3. **Timing** is measured from the start of the first auction to the completion of the last.
4. **Resource usage** (goroutines, heap, GC) is reported before and after the run.

## Concurrency Model

- **Goroutines**: Each auction runs in its own goroutine. Within each auction, every bidder also runs in a goroutine — up to 4,000 concurrent goroutines at peak.
- **Channels**: Bids are collected via a buffered channel per auction — no shared mutable state.
- **`context.WithTimeout`**: Enforces the auction deadline. Bidders use `select` on the context to bail out if the auction has already closed.
- **`sync.WaitGroup`**: Coordinates auction-level and bidder-level completion.
- **`GOMAXPROCS`**: Set to match the configured vCPU count for deterministic scheduling.

## Quick Start

```bash
go run . # quick start
go run -race .  # to check race condition
```

Or build and run:

```bash
go build -o auction-simulator .
./auction-simulator
```

## Output

Files are written to the `output/` directory:
- `AUCTION-001.json` through `AUCTION-040.json` — full details per auction (attributes, all bids, winner, timing).
- `_SUMMARY.json` — consolidated config, timing, resource snapshot, and per-auction summary.

## Resource Standardization (Docker)

To run with exactly **2 vCPUs** and **512 MB RAM**:

```bash
docker build -t auction-simulator .
docker run --cpus=2 --memory=512m auction-simulator
```

This ensures reproducible resource constraints across any host machine. Inside the application, `GOMAXPROCS` is set to 2 to match.

## Project Structure

```
.
├── main.go                 # Entry point — orchestrates everything
├── config/config.go        # All tunable parameters
├── models/models.go        # Shared data structures
├── auction/auction.go      # Single auction runner (goroutines + channels)
├── bidder/bidder.go        # Bidder logic (context-aware, simulated latency)
├── resource/resource.go    # Runtime resource monitoring
├── Dockerfile              # Resource-standardized execution
├── go.mod
└── README.md
```

## Configuration

All parameters are in `config/config.go`:

| Parameter | Default | Description |
|---|---|---|
| `TotalBidders` | 100 | Number of bidders |
| `ConcurrentAuctions` | 40 | Auctions running simultaneously |
| `AttributesPerAuction` | 20 | Attributes per auction object |
| `AuctionTimeout` | 200ms | Timeout per auction |
| `BidResponseProbability` | 0.6 | Chance a bidder responds |
| `ResourceLimits.VCPUs` | 2 | vCPU limit (Docker + GOMAXPROCS) |
| `ResourceLimits.MemoryMB` | 512 | Memory limit (Docker) |
