package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// AegisFlow Prometheus metrics.
// All metrics are registered with the default prometheus registry so they
// are automatically included in the existing /metrics endpoint served by
// the admin server.

var (
	// RequestsTotal counts every request processed by the gateway.
	// Labels: tenant, model, status (200/400/403/500 etc.), stage (gateway/stream).
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "aegisflow_telemetry_requests_total",
		Help: "Total number of requests processed by AegisFlow gateway.",
	}, []string{"tenant", "model", "status", "stage"})

	// RequestDurationSeconds tracks end-to-end request latency.
	// Labels: stage — allows computing P50/P95/P99 per pipeline stage.
	RequestDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "aegisflow_telemetry_request_duration_seconds",
		Help:    "Request latency in seconds by pipeline stage.",
		Buckets: prometheus.DefBuckets, // .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
	}, []string{"stage", "provider"})

	// PolicyDecisionsTotal counts policy engine decisions.
	// Labels: decision (allow/review/block/warn), phase (input/output).
	PolicyDecisionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "aegisflow_telemetry_policy_decisions_total",
		Help: "Total number of policy decisions by outcome.",
	}, []string{"decision", "phase"})

	// ProviderLatencySeconds tracks per-provider LLM call latency.
	// This is usually the dominant latency source (500ms–3000ms for OpenAI).
	ProviderLatencySeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "aegisflow_provider_latency_seconds",
		Help:    "LLM provider call latency in seconds.",
		Buckets: []float64{0.1, 0.25, 0.5, 1.0, 2.0, 3.0, 5.0, 10.0, 30.0},
	}, []string{"provider", "model"})

	// TokensTotal counts tokens processed per request.
	// Labels: tenant, model, type (prompt/completion/total).
	TokensTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "aegisflow_tokens_total",
		Help: "Total tokens processed by model and type.",
	}, []string{"tenant", "model", "type"})

	// EdgeRequestsTotal counts requests processed by the AgentOS Edge proxy.
	// Labels: upstream (gateway/mcpgw/admin), status.
	EdgeRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "aegisflow_edge_requests_total",
		Help: "Total requests processed by the AgentOS Edge proxy.",
	}, []string{"upstream", "status"})

	// EdgeRequestDurationSeconds tracks edge proxy latency (excluding upstream).
	EdgeRequestDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "aegisflow_edge_request_duration_seconds",
		Help:    "AgentOS Edge proxy request processing latency (excluding upstream).",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25},
	}, []string{"upstream"})

	// EdgeActiveConnections is the current number of active edge proxy connections.
	EdgeActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "aegisflow_edge_active_connections",
		Help: "Current number of active connections at the AgentOS Edge proxy.",
	})

	// CircuitBreakerState tracks per-upstream circuit breaker state.
	// 0 = closed (healthy), 1 = open (failing), 2 = half-open (probing).
	CircuitBreakerState = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "aegisflow_circuit_breaker_state",
		Help: "Circuit breaker state per upstream (0=closed, 1=open, 2=half-open).",
	}, []string{"upstream"})

	// EdgeLoadShedTotal counts requests dropped by the adaptive load shedder.
	EdgeLoadShedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "aegisflow_edge_loadshed_total",
		Help: "Total requests dropped by the AgentOS Edge adaptive load shedder.",
	}, []string{"reason"}) // reason: cpu_threshold | queue_full

	// LoadShedRequestsTotal counts requests dropped by the load shedder (both gateway and edge).
	LoadShedRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "aegisflow_loadshed_requests_total",
		Help: "Total requests dropped by the load shedding system.",
	}, []string{"reason", "layer"}) // reason: cpu_threshold | queue_full | capacity | queue_timeout, layer: edge | gateway

	// RateLimitRequestsTotal counts requests rejected by rate limits.
	RateLimitRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "aegisflow_ratelimit_requests_total",
		Help: "Total requests rejected by rate limits.",
	}, []string{"tenant", "type"}) // type: request | token

	// ApprovalQueueDepth is the current depth of the approval queue.
	ApprovalQueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "aegisflow_approval_queue_depth",
		Help: "Current number of items in the approval queue.",
	})

	// EvidenceChainLength tracks the cumulative evidence chain length.
	EvidenceChainLength = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "aegisflow_evidence_chain_length",
		Help: "Current number of entries in the evidence chain.",
	})
)
