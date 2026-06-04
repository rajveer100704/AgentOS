package edgeproxy

import (
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/saivedant169/AegisFlow/internal/telemetry"
)

// AdaptiveLoadshed implements adaptive load shedding at the AgentOS Edge layer.
//
// It drops low-priority requests under two overload conditions:
//  1. CPU utilisation exceeds CPUThresholdPct
//  2. Active request count exceeds MaxQueueDepth
//
// "Critical" requests (identified by X-Priority: critical header) are never
// dropped — they bypass the shedder unconditionally.
//
// This is the key Cloudflare-style feature: maintaining SLOs for high-priority
// traffic under overload by sacrificing low-priority requests early.
//
// Resume bullet:
// "Implemented adaptive load shedding with CPU-aware traffic prioritisation,
// dropping low-priority requests under overload to maintain SLOs."
type AdaptiveLoadshed struct {
	mu             sync.Mutex
	cpuPct         float64  // current CPU estimate (0–100)
	thresholdPct   float64  // drop low-priority above this CPU%
	maxQueueDepth  int64    // drop all non-critical when active > this
	active         int64    // atomic active request counter
	priorityHeader string   // header name, e.g. "X-Priority"
	cpuSamples     []uint64 // ring of goroutine counts for CPU approximation
	sampleHead     int
}

const cpuSampleRing = 10

// NewAdaptiveLoadshed creates an AdaptiveLoadshed with the given configuration.
func NewAdaptiveLoadshed(cfg LoadShedConfig) *AdaptiveLoadshed {
	header := cfg.PriorityHeader
	if header == "" {
		header = "X-Priority"
	}
	maxQueue := int64(cfg.MaxQueueDepth)
	if maxQueue <= 0 {
		maxQueue = 500
	}
	threshold := cfg.CPUThresholdPct
	if threshold <= 0 {
		threshold = 80.0
	}
	return &AdaptiveLoadshed{
		thresholdPct:   threshold,
		maxQueueDepth:  maxQueue,
		priorityHeader: header,
		cpuSamples:     make([]uint64, cpuSampleRing),
	}
}

// Middleware returns an http.Handler middleware that applies adaptive load
// shedding before forwarding the request to the next handler.
func (a *AdaptiveLoadshed) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Critical requests always pass through — no shedding.
		if r.Header.Get(a.priorityHeader) == "critical" {
			next.ServeHTTP(w, r)
			return
		}

		active := atomic.AddInt64(&a.active, 1)
		defer atomic.AddInt64(&a.active, -1)

		// Shed if queue depth exceeded.
		if active > a.maxQueueDepth {
			telemetry.EdgeLoadShedTotal.WithLabelValues("queue_full").Inc()
			http.Error(w, `{"error":"service_overloaded","message":"request queue full — please retry"}`, http.StatusServiceUnavailable)
			return
		}

		// Shed if estimated CPU is above threshold.
		if a.estimatedCPU() > a.thresholdPct {
			telemetry.EdgeLoadShedTotal.WithLabelValues("cpu_threshold").Inc()
			http.Error(w, `{"error":"service_overloaded","message":"cpu threshold exceeded — shedding low priority traffic"}`, http.StatusServiceUnavailable)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// estimatedCPU returns an approximation of CPU load based on the ratio of
// goroutines to GOMAXPROCS. This is a lightweight heuristic that works
// without cgo or OS-level CPU sampling — suitable for all platforms.
//
// Formula: (goroutines / (GOMAXPROCS * 100)) clamped to 0–100
// Under normal load this is near 0. Under heavy parallel work it climbs.
func (a *AdaptiveLoadshed) estimatedCPU() float64 {
	procs := float64(runtime.GOMAXPROCS(0))
	goroutines := float64(runtime.NumGoroutine())

	// Normalise: 100 goroutines per proc ≈ 100% CPU (heuristic).
	pct := (goroutines / (procs * 100.0)) * 100.0
	if pct > 100.0 {
		pct = 100.0
	}
	return pct
}

// ActiveRequests returns the number of requests currently being processed.
func (a *AdaptiveLoadshed) ActiveRequests() int64 {
	return atomic.LoadInt64(&a.active)
}
