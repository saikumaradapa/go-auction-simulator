// Package resource provides runtime resource monitoring and reporting.
//
// Resource Standardization Approach:
// 1. Docker is the primary enforcement mechanism — the Dockerfile and run
//    command use --cpus and --memory flags to cap resources identically
//    across any host machine.
// 2. GOMAXPROCS is set to match the configured vCPU count so the Go
//    scheduler uses exactly that many OS threads for goroutines.
// 3. This module reports heap/GC stats so output includes actual usage
//    relative to the configured limits.
package resource

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/auction-simulator/config"
	"github.com/auction-simulator/models"
)

// peakGoroutines tracks the maximum number of goroutines observed during execution.
var peakGoroutines int64

// GoroutineTracker monitors goroutine count in the background and records the peak.
type GoroutineTracker struct {
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// StartGoroutineTracker begins polling goroutine count every 5ms.
func StartGoroutineTracker() *GoroutineTracker {
	t := &GoroutineTracker{stopCh: make(chan struct{})}
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		for {
			select {
			case <-t.stopCh:
				return
			default:
				current := int64(runtime.NumGoroutine())
				for {
					old := atomic.LoadInt64(&peakGoroutines)
					if current <= old || atomic.CompareAndSwapInt64(&peakGoroutines, old, current) {
						break
					}
				}
				time.Sleep(5 * time.Millisecond)
			}
		}
	}()
	return t
}

// Stop halts the tracker and returns the peak goroutine count observed.
func (t *GoroutineTracker) Stop() int64 {
	close(t.stopCh)
	t.wg.Wait()
	return atomic.LoadInt64(&peakGoroutines)
}

// PeakGoroutines returns the peak goroutine count recorded so far.
func PeakGoroutines() int64 {
	return atomic.LoadInt64(&peakGoroutines)
}

// Snapshot captures current Go runtime resource usage.
func Snapshot() models.ResourceSnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return models.ResourceSnapshot{
		ConfiguredVCPUs:    config.ResourceLimits.VCPUs,
		ConfiguredMemoryMB: config.ResourceLimits.MemoryMB,
		NumGoroutines:      runtime.NumGoroutine(),
		HeapAllocMB:        float64(m.HeapAlloc) / 1024 / 1024,
		HeapSysMB:          float64(m.HeapSys) / 1024 / 1024,
		TotalAllocMB:       float64(m.TotalAlloc) / 1024 / 1024,
		NumGC:              m.NumGC,
		NumCPU:             runtime.NumCPU(),
	}
}

// padRight pads a string with spaces to reach exactly n visible characters.
func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

// PrintComparison prints a before vs after resource comparison with deltas.
func PrintComparison(before, after models.ResourceSnapshot) {
	// Box inner width = 50 visible characters
	const W = 50

	line := strings.Repeat("=", W)
	row := func(content string) {
		fmt.Printf("| %s|\n", padRight(content, W-1))
	}
	sep := func() {
		fmt.Printf("+%s+\n", line)
	}

	fmt.Println()
	sep()
	row("           RESOURCE USAGE REPORT")
	sep()
	row(fmt.Sprintf("Configured limits : %d vCPUs, %d MB RAM", after.ConfiguredVCPUs, after.ConfiguredMemoryMB))
	row(fmt.Sprintf("GOMAXPROCS        : %d", runtime.GOMAXPROCS(0)))
	row(fmt.Sprintf("Num CPUs (host)   : %d", after.NumCPU))
	sep()
	row(fmt.Sprintf("%-20s %-9s %-9s %s", "Metric", "Before", "After", "Delta"))
	sep()
	row(fmt.Sprintf("%-20s %-9d %-9d %s", "Goroutines (now):", before.NumGoroutines, after.NumGoroutines, deltaI(after.NumGoroutines-before.NumGoroutines)))
	row(fmt.Sprintf("%-20s %d", "Goroutines (peak):", PeakGoroutines()))
	row(fmt.Sprintf("%-20s %-9s %-9s %s", "Heap Alloc (MB):", fmtMB(before.HeapAllocMB), fmtMB(after.HeapAllocMB), fmtDeltaMB(after.HeapAllocMB-before.HeapAllocMB)))
	row(fmt.Sprintf("%-20s %-9s %-9s %s", "Heap Sys  (MB):", fmtMB(before.HeapSysMB), fmtMB(after.HeapSysMB), fmtDeltaMB(after.HeapSysMB-before.HeapSysMB)))
	row(fmt.Sprintf("%-20s %-9s %-9s %s", "Total Alloc (MB):", fmtMB(before.TotalAllocMB), fmtMB(after.TotalAllocMB), fmtDeltaMB(after.TotalAllocMB-before.TotalAllocMB)))
	row(fmt.Sprintf("%-20s %-9d %-9d %s", "GC Cycles:", before.NumGC, after.NumGC, deltaI(int(after.NumGC)-int(before.NumGC))))
	sep()
}

func fmtMB(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func fmtDeltaMB(v float64) string {
	if v >= 0 {
		return fmt.Sprintf("+%.2f", v)
	}
	return fmt.Sprintf("%.2f", v)
}

func deltaI(v int) string {
	if v >= 0 {
		return fmt.Sprintf("+%d", v)
	}
	return fmt.Sprintf("%d", v)
}
