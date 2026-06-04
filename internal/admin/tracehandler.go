package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/saivedant169/AegisFlow/internal/telemetry"
)

// TraceHandler exposes the in-process TraceStore over HTTP for the
// AgentOS Observatory dashboard.
//
// It is intentionally lightweight — it queries the ring buffer directly,
// without connecting to Jaeger. This means the Trace Explorer tab works
// even when no external tracing backend is configured.
type TraceHandler struct {
	store *telemetry.TraceStore
}

// NewTraceHandler creates a TraceHandler backed by the given TraceStore.
// The store is shared with the telemetry SpanProcessor that populates it.
func NewTraceHandler(store *telemetry.TraceStore) *TraceHandler {
	return &TraceHandler{store: store}
}

// ListTraces handles GET /admin/v1/traces?n=50
// Returns up to n most-recent trace summaries in JSON.
//
// Response shape:
//
//	{
//	  "count": 47,
//	  "traces": [
//	    {
//	      "trace_id": "abc123",
//	      "span_id": "def456",
//	      "operation": "gateway.chat_completion",
//	      "duration_ms": 1248,
//	      "status": "ok",
//	      "attrs": { "aegisflow.tenant.id": "dev", "aegisflow.model": "gpt-4" },
//	      "start_time": "2026-06-04T06:00:00Z"
//	    }
//	  ]
//	}
func (h *TraceHandler) ListTraces(w http.ResponseWriter, r *http.Request) {
	n := 50
	if q := r.URL.Query().Get("n"); q != "" {
		if v, err := strconv.Atoi(q); err == nil && v > 0 {
			n = v
		}
	}

	traces := h.store.Recent(n)

	// Convert Duration to milliseconds for JSON readability.
	type wireTrace struct {
		TraceID    string            `json:"trace_id"`
		SpanID     string            `json:"span_id"`
		Operation  string            `json:"operation"`
		DurationMs int64             `json:"duration_ms"`
		Status     string            `json:"status"`
		Attrs      map[string]string `json:"attrs,omitempty"`
		StartTime  string            `json:"start_time"`
	}

	wire := make([]wireTrace, 0, len(traces))
	for _, t := range traces {
		wire = append(wire, wireTrace{
			TraceID:    t.TraceID,
			SpanID:     t.SpanID,
			Operation:  t.Operation,
			DurationMs: t.Duration.Milliseconds(),
			Status:     t.Status,
			Attrs:      t.Attrs,
			StartTime:  t.StartTime.UTC().Format("2006-01-02T15:04:05.000Z"),
		})
	}

	resp := struct {
		Count  int         `json:"count"`
		Total  int         `json:"total"`
		Traces []wireTrace `json:"traces"`
	}{
		Count:  len(wire),
		Total:  h.store.Len(),
		Traces: wire,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// TraceStats handles GET /admin/v1/traces/stats
// Returns aggregate statistics over the stored traces.
func (h *TraceHandler) TraceStats(w http.ResponseWriter, r *http.Request) {
	traces := h.store.Recent(h.store.Len())

	var totalDuration int64
	errorCount := 0
	opCounts := make(map[string]int)

	for _, t := range traces {
		totalDuration += t.Duration.Milliseconds()
		if t.Status == "error" {
			errorCount++
		}
		opCounts[t.Operation]++
	}

	avgMs := int64(0)
	if len(traces) > 0 {
		avgMs = totalDuration / int64(len(traces))
	}

	resp := struct {
		Total      int            `json:"total"`
		Errors     int            `json:"errors"`
		ErrorRate  float64        `json:"error_rate_pct"`
		AvgMs      int64          `json:"avg_duration_ms"`
		Operations map[string]int `json:"operations"`
	}{
		Total:      len(traces),
		Errors:     errorCount,
		ErrorRate:  0,
		AvgMs:      avgMs,
		Operations: opCounts,
	}
	if len(traces) > 0 {
		resp.ErrorRate = float64(errorCount) / float64(len(traces)) * 100.0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// EdgeMetrics handles GET /admin/v1/edge/metrics
// Gathers AgentOS Edge metrics from the default Prometheus registry
// and formats them as JSON.
func (h *TraceHandler) EdgeMetrics(w http.ResponseWriter, r *http.Request) {
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	type cbState struct {
		Upstream string  `json:"upstream"`
		State    float64 `json:"state"`
	}

	type loadshedState struct {
		Reason string  `json:"reason"`
		Count  float64 `json:"count"`
	}

	type upstreamRequest struct {
		Upstream string  `json:"upstream"`
		Status   string  `json:"status"`
		Count    float64 `json:"count"`
	}

	resp := struct {
		ActiveConnections float64           `json:"active_connections"`
		CircuitBreakers   []cbState         `json:"circuit_breakers"`
		LoadShed          []loadshedState   `json:"load_shed"`
		EdgeRequests      []upstreamRequest `json:"edge_requests"`
	}{
		CircuitBreakers: []cbState{},
		LoadShed:        []loadshedState{},
		EdgeRequests:    []upstreamRequest{},
	}

	for _, fam := range families {
		name := fam.GetName()
		for _, m := range fam.GetMetric() {
			switch name {
			case "aegisflow_edge_active_connections":
				if m.Gauge != nil {
					resp.ActiveConnections = m.Gauge.GetValue()
				}
			case "aegisflow_circuit_breaker_state":
				if m.Gauge != nil {
					upstream := ""
					for _, lp := range m.Label {
						if lp.GetName() == "upstream" {
							upstream = lp.GetValue()
						}
					}
					resp.CircuitBreakers = append(resp.CircuitBreakers, cbState{
						Upstream: upstream,
						State:    m.Gauge.GetValue(),
					})
				}
			case "aegisflow_edge_loadshed_total":
				if m.Counter != nil {
					reason := ""
					for _, lp := range m.Label {
						if lp.GetName() == "reason" {
							reason = lp.GetValue()
						}
					}
					resp.LoadShed = append(resp.LoadShed, loadshedState{
						Reason: reason,
						Count:  m.Counter.GetValue(),
					})
				}
			case "aegisflow_edge_requests_total":
				if m.Counter != nil {
					upstream := ""
					status := ""
					for _, lp := range m.Label {
						if lp.GetName() == "upstream" {
							upstream = lp.GetValue()
						}
						if lp.GetName() == "status" {
							status = lp.GetValue()
						}
					}
					resp.EdgeRequests = append(resp.EdgeRequests, upstreamRequest{
						Upstream: upstream,
						Status:   status,
						Count:    m.Counter.GetValue(),
					})
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
