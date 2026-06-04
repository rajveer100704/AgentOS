package telemetry

import (
	"context"
	"sync"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// TraceSummary is a lightweight snapshot of a completed span.
// Stored in the TraceStore ring buffer for the Observatory dashboard.
type TraceSummary struct {
	TraceID   string            `json:"trace_id"`
	SpanID    string            `json:"span_id"`
	Operation string            `json:"operation"`
	Duration  time.Duration     `json:"duration_ms"`
	Status    string            `json:"status"` // "ok" | "error"
	Attrs     map[string]string `json:"attrs,omitempty"`
	StartTime time.Time         `json:"start_time"`
}

// TraceStore is a thread-safe ring buffer of the last N trace summaries.
// It is fed by a custom OTel SpanProcessor and queried by the admin dashboard
// at GET /admin/v1/traces — without needing a live Jaeger connection.
type TraceStore struct {
	mu   sync.RWMutex
	ring []TraceSummary
	size int
	head int
	full bool
}

// NewTraceStore creates a TraceStore with the given ring buffer capacity.
// A size of 500 is recommended for the Observatory dashboard.
func NewTraceStore(size int) *TraceStore {
	if size <= 0 {
		size = 500
	}
	return &TraceStore{
		ring: make([]TraceSummary, size),
		size: size,
	}
}

// Add inserts a TraceSummary into the ring buffer, overwriting the oldest
// entry when the buffer is full.
func (s *TraceStore) Add(t TraceSummary) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ring[s.head] = t
	s.head = (s.head + 1) % s.size
	if s.head == 0 {
		s.full = true
	}
}

// Recent returns up to n most-recent TraceSummaries in reverse chronological
// order (newest first). n is clamped to the buffer size.
func (s *TraceStore) Recent(n int) []TraceSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n <= 0 || n > s.size {
		n = s.size
	}

	var result []TraceSummary

	if !s.full {
		// Buffer not yet wrapped — entries are 0..head-1
		count := s.head
		if count > n {
			count = n
		}
		for i := count - 1; i >= 0; i-- {
			result = append(result, s.ring[i])
		}
		return result
	}

	// Buffer has wrapped — newest entry is at head-1 (mod size)
	for i := 0; i < n; i++ {
		idx := (s.head - 1 - i + s.size) % s.size
		result = append(result, s.ring[idx])
	}
	return result
}

// Len returns the number of entries currently in the store.
func (s *TraceStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.full {
		return s.size
	}
	return s.head
}

// storeProcessor is a sdktrace.SpanProcessor that extracts span data and
// adds it to the TraceStore when a span ends.
type storeProcessor struct {
	store *TraceStore
}

// NewProcessor returns an OTel SpanProcessor backed by this TraceStore.
// Register it with sdktrace.WithSpanProcessor() during Init.
func (s *TraceStore) NewProcessor() sdktrace.SpanProcessor {
	return &storeProcessor{store: s}
}

// OnStart is a no-op — we only care about completed spans.
func (p *storeProcessor) OnStart(_ context.Context, _ sdktrace.ReadWriteSpan) {}

// OnEnd extracts key attributes and records the summary in the ring buffer.
func (p *storeProcessor) OnEnd(span sdktrace.ReadOnlySpan) {
	status := "ok"
	if span.Status().Code.String() == "Error" {
		status = "error"
	}

	attrs := make(map[string]string)
	for _, kv := range span.Attributes() {
		attrs[string(kv.Key)] = kv.Value.AsString()
	}

	summary := TraceSummary{
		TraceID:   span.SpanContext().TraceID().String(),
		SpanID:    span.SpanContext().SpanID().String(),
		Operation: span.Name(),
		Duration:  span.EndTime().Sub(span.StartTime()),
		Status:    status,
		Attrs:     attrs,
		StartTime: span.StartTime(),
	}

	p.store.Add(summary)
}

// Shutdown is a no-op for an in-memory processor.
func (p *storeProcessor) Shutdown(_ context.Context) error   { return nil }
func (p *storeProcessor) ForceFlush(_ context.Context) error { return nil }
