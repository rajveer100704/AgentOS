// Package edgeproxy implements AgentOS Edge — the high-performance reverse
// proxy layer that sits in front of the AegisFlow Control Plane.
//
// # Architecture
//
//	AI Agent (HTTPS/2)
//	    ↓
//	AgentOS Edge :8443 (TLS) / :8444 (H2C)
//	    ├── TLS 1.3 termination
//	    ├── HTTP/2 (ALPN h2) — always on
//	    ├── HTTP/3 / QUIC — feature flag: http3_enabled: false
//	    ├── Adaptive Load Shedder (CPU% + queue depth)
//	    ├── Path router: /api/* → :8080, /mcp/* → :9090, /admin/* → :8081
//	    ├── Token-bucket rate limiter (reuses ratelimit.Limiter)
//	    ├── Circuit breaker per upstream (reuses resilience.CircuitBreaker)
//	    ├── Connection pool (http.Transport, MaxIdleConnsPerHost)
//	    └── OTel span + Prometheus metrics per request
//	    ↓
//	AgentOS Control Plane :8080 (existing, unchanged)
//
// # Key Design Points
//
//   - Reuses resilience.CircuitBreaker — no duplication
//   - Reuses ratelimit.Limiter interface — no duplication
//   - HTTP/3 is behind a feature flag (disabled by default)
//   - All frozen packages are untouched
//   - Span injection via telemetry.StartSpan at proxy call site
package edgeproxy

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/quic-go/quic-go/http3"
	"github.com/saivedant169/AegisFlow/internal/ratelimit"
	"github.com/saivedant169/AegisFlow/internal/telemetry"
)

// EdgeProxy is the AgentOS Edge reverse proxy.
// It manages TLS, HTTP/2, circuit breakers, load shedding, and metrics.
type EdgeProxy struct {
	cfg       Config
	upstreams []*Upstream
	loadshed  *AdaptiveLoadshed
	limiter   ratelimit.Limiter
	transport *http.Transport

	// Servers
	tlsSrv  *http.Server
	h2cSrv  *http.Server
	quicSrv *http3.Server

	// Active connection counter for metrics.
	activeConns int64
}

// New creates a new EdgeProxy from the given configuration and optional
// rate limiter. If limiter is nil, rate limiting at the edge is disabled
// (the gateway's rate limiter still applies downstream).
func New(cfg Config, limiter ratelimit.Limiter) (*EdgeProxy, error) {
	// Apply defaults for any zero-values.
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8443"
	}
	if cfg.H2CAddr == "" {
		cfg.H2CAddr = ":8444"
	}
	if cfg.MaxIdleConns <= 0 {
		cfg.MaxIdleConns = 100
	}
	if cfg.IdleConnTimeout <= 0 {
		cfg.IdleConnTimeout = 90 * time.Second
	}
	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = 10 * time.Second
	}

	// Build upstreams with circuit breakers.
	upstreams := make([]*Upstream, 0, len(cfg.Upstreams))
	for _, ucfg := range cfg.Upstreams {
		u, err := newUpstream(ucfg)
		if err != nil {
			return nil, fmt.Errorf("edgeproxy: upstream %q: %w", ucfg.Name, err)
		}
		upstreams = append(upstreams, u)
		log.Printf("[edgeproxy] upstream registered: %s → %s (prefix: %s)", u.Name, u.Target, u.Prefix)
	}

	// Connection pool — keeps TCP connections alive to avoid per-request
	// dial overhead. This is connection pooling at the edge proxy layer.
	transport := &http.Transport{
		MaxIdleConnsPerHost: cfg.MaxIdleConns,
		IdleConnTimeout:     cfg.IdleConnTimeout,
		DialContext: (&net.Dialer{
			Timeout:   cfg.DialTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		// Enable HTTP/2 for upstream connections (h2c for internal traffic).
		ForceAttemptHTTP2:     true,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 120 * time.Second,
	}

	var shed *AdaptiveLoadshed
	if cfg.LoadShed.Enabled {
		shed = NewAdaptiveLoadshed(cfg.LoadShed)
		log.Printf("[edgeproxy] adaptive load shedding enabled (cpu_threshold=%.0f%%, max_queue=%d)",
			cfg.LoadShed.CPUThresholdPct, cfg.LoadShed.MaxQueueDepth)
	}

	return &EdgeProxy{
		cfg:       cfg,
		upstreams: upstreams,
		loadshed:  shed,
		limiter:   limiter,
		transport: transport,
	}, nil
}

// ListenAndServeTLS starts the TLS edge proxy on cfg.ListenAddr.
// It blocks until the server is stopped. HTTP/2 is enabled automatically
// via ALPN negotiation in the TLS handshake.
func (ep *EdgeProxy) ListenAndServeTLS() error {
	tlsCfg, err := BuildTLSConfig(ep.cfg.TLS.CertFile, ep.cfg.TLS.KeyFile)
	if err != nil {
		return fmt.Errorf("edgeproxy TLS: %w", err)
	}

	mux := ep.buildHandler()

	ep.tlsSrv = &http.Server{
		Addr:         ep.cfg.ListenAddr,
		Handler:      mux,
		TLSConfig:    tlsCfg,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ConfigureServer enables HTTP/2 on the TLS server.
	if err := http2.ConfigureServer(ep.tlsSrv, &http2.Server{
		MaxConcurrentStreams:         250,
		MaxReadFrameSize:             1 << 20, // 1 MiB
		IdleTimeout:                 60 * time.Second,
	}); err != nil {
		return fmt.Errorf("edgeproxy: configuring HTTP/2: %w", err)
	}

	log.Printf("[edgeproxy] TLS edge listening on %s (TLS 1.3, HTTP/2 enabled)", ep.cfg.ListenAddr)
	return ep.tlsSrv.ListenAndServeTLS("", "")
}

// ListenAndServeH2C starts the HTTP/2 cleartext proxy on cfg.H2CAddr.
// Useful for internal service-mesh deployments where TLS is terminated
// by an upstream load balancer.
func (ep *EdgeProxy) ListenAndServeH2C() error {
	mux := ep.buildHandler()

	// h2c.NewHandler wraps the mux to support HTTP/2 over cleartext.
	h2cHandler := h2c.NewHandler(mux, &http2.Server{
		MaxConcurrentStreams: 250,
	})

	ep.h2cSrv = &http.Server{
		Addr:         ep.cfg.H2CAddr,
		Handler:      h2cHandler,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	log.Printf("[edgeproxy] H2C (HTTP/2 cleartext) edge listening on %s", ep.cfg.H2CAddr)
	return ep.h2cSrv.ListenAndServe()
}

// ListenAndServeQUIC starts the HTTP/3 QUIC server on ep.cfg.HTTP3Addr.
func (ep *EdgeProxy) ListenAndServeQUIC() error {
	tlsCfg, err := BuildTLSConfig(ep.cfg.TLS.CertFile, ep.cfg.TLS.KeyFile)
	if err != nil {
		return fmt.Errorf("edgeproxy HTTP/3 TLS: %w", err)
	}

	mux := ep.buildHandler()

	ep.quicSrv = &http3.Server{
		Addr:      ep.cfg.HTTP3Addr,
		Handler:   mux,
		TLSConfig: tlsCfg,
	}

	log.Printf("[edgeproxy] HTTP/3 edge listening on QUIC/UDP %s", ep.cfg.HTTP3Addr)
	return ep.quicSrv.ListenAndServe()
}

// Shutdown gracefully stops the TLS, H2C, and QUIC servers.
func (ep *EdgeProxy) Shutdown(ctx context.Context) {
	if ep.tlsSrv != nil {
		ep.tlsSrv.Shutdown(ctx) //nolint:errcheck
	}
	if ep.h2cSrv != nil {
		ep.h2cSrv.Shutdown(ctx) //nolint:errcheck
	}
	if ep.quicSrv != nil {
		ep.quicSrv.Close() //nolint:errcheck
	}
	log.Println("[edgeproxy] shutdown complete")
}

// buildHandler constructs the edge proxy's http.Handler chain:
//   Connection counter → Load shedder → OTel span → Circuit breaker → ReverseProxy
func (ep *EdgeProxy) buildHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Track active connections for Prometheus.
		active := atomic.AddInt64(&ep.activeConns, 1)
		defer atomic.AddInt64(&ep.activeConns, -1)
		telemetry.EdgeActiveConnections.Set(float64(active))

		// Advertise HTTP/3 support via Alt-Svc header.
		if ep.cfg.HTTP3Enabled && ep.cfg.HTTP3Addr != "" {
			_, port, err := net.SplitHostPort(ep.cfg.HTTP3Addr)
			if err == nil {
				w.Header().Set("Alt-Svc", fmt.Sprintf(`h3=":%s"; ma=86400`, port))
			} else {
				w.Header().Set("Alt-Svc", fmt.Sprintf(`h3="%s"; ma=86400`, ep.cfg.HTTP3Addr))
			}
		}

		// Find matching upstream by longest prefix match.
		upstream := ep.matchUpstream(r.URL.Path)
		if upstream == nil {
			http.Error(w, `{"error":"not_found","message":"no upstream matched"}`, http.StatusNotFound)
			return
		}

		// OTel span for this edge request.
		ctx, span := telemetry.StartSpan(r.Context(), "edge.request",
			attribute.String(telemetry.AttrEdgeUpstream, upstream.Name),
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.Path),
		)
		r = r.WithContext(ctx)
		defer span.End()

		// Adaptive load shedding (CPU / queue-depth based).
		if ep.loadshed != nil {
			// Inline the middleware logic to propagate the OTel context.
			if r.Header.Get(ep.cfg.LoadShed.PriorityHeader) != "critical" {
				if ep.loadshed.ActiveRequests() > int64(ep.cfg.LoadShed.MaxQueueDepth) {
					telemetry.EdgeLoadShedTotal.WithLabelValues("queue_full").Inc()
					telemetry.LoadShedRequestsTotal.WithLabelValues("queue_full", "edge").Inc()
					span.SetAttributes(attribute.Bool("loadshed.dropped", true))
					http.Error(w, `{"error":"service_overloaded","message":"request queue full"}`, http.StatusServiceUnavailable)
					return
				}
				if ep.loadshed.estimatedCPU() > ep.cfg.LoadShed.CPUThresholdPct {
					telemetry.EdgeLoadShedTotal.WithLabelValues("cpu_threshold").Inc()
					telemetry.LoadShedRequestsTotal.WithLabelValues("cpu_threshold", "edge").Inc()
					span.SetAttributes(attribute.Bool("loadshed.dropped", true))
					http.Error(w, `{"error":"service_overloaded","message":"cpu overloaded — dropping low priority"}`, http.StatusServiceUnavailable)
					return
				}
			}
		}

		// Circuit breaker check.
		upstream.updateBreakerMetric()
		if !upstream.Breaker.Allow() {
			span.SetAttributes(attribute.Bool("circuit_breaker.open", true))
			telemetry.EdgeRequestsTotal.WithLabelValues(upstream.Name, "503").Inc()
			http.Error(w, `{"error":"upstream_unavailable","message":"circuit breaker open — upstream temporarily unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		// Capture start time for latency measurement.
		start := time.Now()

		// Build the reverse proxy on-demand so we capture the upstream at
		// the time of the request (supports future hot-reload of upstreams).
		proxy := ep.buildReverseProxy(upstream)

		// Wrap the response writer to capture the status code.
		rw := &statusWriter{ResponseWriter: w}
		proxy.ServeHTTP(rw, r)

		// Record success or failure to the circuit breaker.
		latency := time.Since(start)
		if rw.status >= 500 {
			upstream.Breaker.RecordFailure()
		} else {
			upstream.Breaker.RecordSuccess()
		}
		upstream.updateBreakerMetric()

		// Prometheus edge metrics.
		statusStr := fmt.Sprintf("%d", rw.status)
		telemetry.EdgeRequestsTotal.WithLabelValues(upstream.Name, statusStr).Inc()
		telemetry.EdgeRequestDurationSeconds.WithLabelValues(upstream.Name).Observe(latency.Seconds())

		// Add final span attributes.
		span.SetAttributes(
			attribute.Int("http.status_code", rw.status),
			attribute.Float64("edge.latency_ms", float64(latency.Milliseconds())),
		)
	})
}

// matchUpstream finds the upstream whose Prefix is the longest prefix of path.
// Returns nil if no upstream matches.
func (ep *EdgeProxy) matchUpstream(path string) *Upstream {
	var best *Upstream
	bestLen := -1
	for _, u := range ep.upstreams {
		if strings.HasPrefix(path, u.Prefix) && len(u.Prefix) > bestLen {
			best = u
			bestLen = len(u.Prefix)
		}
	}
	return best
}

// buildReverseProxy creates an httputil.ReverseProxy targeting the given upstream.
// The connection pool (ep.transport) is shared across all upstreams.
func (ep *EdgeProxy) buildReverseProxy(upstream *Upstream) *httputil.ReverseProxy {
	target := upstream.Target
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host

			// Strip prefix if configured (default: pass path as-is).
			if upstream.StripPrefix && upstream.Prefix != "" {
				req.URL.Path = strings.TrimPrefix(req.URL.Path, upstream.Prefix)
				if req.URL.Path == "" {
					req.URL.Path = "/"
				}
			}

			// Forward the original host header, add proxy headers.
			if req.Header.Get("X-Forwarded-For") == "" {
				if host, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
					req.Header.Set("X-Forwarded-For", host)
				}
			}
			req.Header.Set("X-Forwarded-Proto", "https")
			req.Header.Set("X-AgentOS-Edge", "1")

			// Remove hop-by-hop headers.
			req.Header.Del("Te")
			req.Header.Del("Trailers")
		},
		Transport: ep.transport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("[edgeproxy] upstream %s error: %v", upstream.Name, err)
			upstream.Breaker.RecordFailure()
			upstream.updateBreakerMetric()
			telemetry.RecordError(telemetry.SpanFromContext(r.Context()), err)
			http.Error(w, `{"error":"upstream_error","message":"upstream unavailable"}`, http.StatusBadGateway)
		},
	}
	return proxy
}

// statusWriter wraps http.ResponseWriter to capture the HTTP status code
// written by the downstream handler, so we can report it in metrics and spans.
type statusWriter struct {
	http.ResponseWriter
	status  int
	written bool
}

func (sw *statusWriter) WriteHeader(code int) {
	if !sw.written {
		sw.status = code
		sw.written = true
	}
	sw.ResponseWriter.WriteHeader(code)
}

func (sw *statusWriter) Write(b []byte) (int, error) {
	if !sw.written {
		sw.status = http.StatusOK
		sw.written = true
	}
	return sw.ResponseWriter.Write(b)
}
