# Auction Simulator

**Backend Engineer — Take Home Assessment**

A Go-based Auction Simulator that runs 40 concurrent auctions with 100 bidders, enforces per-auction timeouts, declares winners, and measures end-to-end execution time under standardized resource constraints.

---

## Demo Execution

https://github.com/user-attachments/assets/go-auction-simulator-demo-output.mp4

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

## Sample Output

### Individual Auction (`output/AUCTION-001.json`)

Each auction file contains the full attributes, all received bids, the winner, and timing:

```json
{
  "auction_id": "AUCTION-001",
  "attributes": {
    "attr1_category": "health",
    "attr2_quality": "used-good",
    "attr3_base_price": 25.82,
    "attr4_weight": 46.79,
    "attr5_rating": 4.6,
    "attr6_popularity": 660,
    "attr7_stock": 402,
    "attr8_discount": 0.47,
    "attr9_shipping_cost": 9.47,
    "attr10_return_policy": false,
    "attr11_warranty": true,
    "attr12_brand_tier": 3,
    "attr13_seasonality": 0.73,
    "attr14_demand_index": 0.29,
    "attr15_supply_index": 0.06,
    "attr16_margin": 0.33,
    "attr17_click_rate": 0.1554,
    "attr18_conversion_rate": 0.0078,
    "attr19_impressions": 64999,
    "attr20_geo_region": "EU"
  },
  "total_bidders": 100,
  "bids_received": 54,
  "bids_timed_out_or_declined": 46,
  "bids": [
    {
      "bidder_id": 90,
      "bidder_name": "Bidder-090",
      "auction_id": "AUCTION-001",
      "price": 32.99,
      "latency_ms": 6,
      "responded_at": "2026-04-05T14:12:43.6765685+05:30"
    },
    "... (54 bids total)"
  ],
  "winner": {
    "bidder_id": 83,
    "bidder_name": "Bidder-083",
    "price": 49.76,
    "latency_ms": 81
  },
  "timing": {
    "start_time": "2026-04-05T14:12:43.6648335+05:30",
    "end_time": "2026-04-05T14:12:43.8655758+05:30",
    "duration_ms": 200
  }
}
```

### Summary (`output/_SUMMARY.json`)

The summary file consolidates configuration, wall-clock timing, resource usage, and per-auction results:

```json
{
  "configuration": {
    "total_bidders": 100,
    "concurrent_auctions": 40,
    "attributes_per_auction": 20,
    "auction_timeout_ms": 200,
    "vcpus": 2,
    "memory_mb": 512
  },
  "timing": {
    "first_auction_start": "2026-04-05T14:12:43.6648335+05:30",
    "last_auction_end": "2026-04-05T14:12:43.9513344+05:30",
    "total_wall_clock_ms": 286
  },
  "resources": {
    "configured_vcpus": 2,
    "configured_memory_mb": 512,
    "num_goroutines": 1,
    "heap_alloc_mb": 6.07,
    "heap_sys_mb": 7.41,
    "total_alloc_mb": 6.55,
    "num_gc": 1,
    "num_cpu": 18
  },
  "auctions": [
    {
      "auction_id": "AUCTION-001",
      "bids_received": 54,
      "timed_out": 46,
      "winner": "Bidder-083 @ $49.76 (81ms)",
      "duration_ms": 200
    },
    {
      "auction_id": "AUCTION-002",
      "bids_received": 44,
      "timed_out": 56,
      "winner": "Bidder-006 @ $49.36 (68ms)",
      "duration_ms": 200
    },
    "... (40 auctions total)"
  ]
}
```

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
