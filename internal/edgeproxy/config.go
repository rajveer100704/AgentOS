package edgeproxy

import (
	"time"
)

// Config holds all configuration for the AgentOS Edge proxy.
// Loaded from the yaml edge_proxy section of aegisflow.yaml.
type Config struct {
	// Enabled controls whether the edge proxy starts. Default: false.
	Enabled bool `yaml:"enabled"`

	// ListenAddr is the TLS listener address. Default: ":8443".
	ListenAddr string `yaml:"listen_addr"`

	// H2CAddr is the HTTP/2 cleartext listener for internal traffic.
	// Useful for service-mesh deployments where TLS is terminated upstream.
	// Default: ":8444".
	H2CAddr string `yaml:"h2c_addr"`

	// HTTP3Enabled enables QUIC/HTTP3 on HTTP3Addr.
	// Default: false — enable only after HTTP/1.1 and HTTP/2 are stable.
	HTTP3Enabled bool `yaml:"http3_enabled"`

	// HTTP3Addr is the UDP address for QUIC/HTTP3. Default: ":8445".
	HTTP3Addr string `yaml:"http3_addr"`

	// TLS configuration.
	TLS TLSConfig `yaml:"tls"`

	// Upstreams defines the path-prefix → upstream mapping.
	Upstreams []UpstreamConfig `yaml:"upstreams"`

	// MaxIdleConns is the maximum number of idle keep-alive connections
	// in the connection pool per upstream. Default: 100.
	MaxIdleConns int `yaml:"max_idle_conns"`

	// IdleConnTimeout is how long idle connections stay in the pool.
	// Default: 90s.
	IdleConnTimeout time.Duration `yaml:"idle_conn_timeout"`

	// DialTimeout is the maximum time to establish a new TCP connection.
	// Default: 10s.
	DialTimeout time.Duration `yaml:"dial_timeout"`

	// LoadShed controls the adaptive load shedder.
	LoadShed LoadShedConfig `yaml:"load_shed"`

	// TraceStoreSize is the maximum number of edge trace summaries kept
	// in memory for the Observatory dashboard. Default: 500.
	TraceStoreSize int `yaml:"trace_store_size"`
}

// TLSConfig holds TLS certificate paths.
// If CertFile and KeyFile are both empty, the edge proxy auto-generates
// a self-signed certificate valid for 10 years (suitable for development).
type TLSConfig struct {
	// CertFile is the path to the TLS certificate PEM file.
	// Leave empty to auto-generate a self-signed cert.
	CertFile string `yaml:"cert_file"`

	// KeyFile is the path to the TLS private key PEM file.
	// Leave empty to auto-generate a self-signed key.
	KeyFile string `yaml:"key_file"`
}

// UpstreamConfig defines one backend upstream that the edge proxy routes to.
type UpstreamConfig struct {
	// Name is a unique label for this upstream (e.g., "gateway", "mcpgw").
	Name string `yaml:"name"`

	// Prefix is the URL path prefix that routes to this upstream.
	// Example: "/api" matches /api/v1/chat/completions.
	// The prefix is stripped before forwarding if StripPrefix is true.
	Prefix string `yaml:"prefix"`

	// Target is the full URL of the upstream server.
	// Example: "http://localhost:8080"
	Target string `yaml:"target"`

	// StripPrefix removes the prefix from the forwarded request path.
	// Default: false (pass path as-is).
	StripPrefix bool `yaml:"strip_prefix"`

	// CircuitBreaker configures the per-upstream circuit breaker.
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`

	// Timeout is the maximum time to wait for the upstream to respond.
	// Default: 120s.
	Timeout time.Duration `yaml:"timeout"`
}

// CircuitBreakerConfig configures the circuit breaker for an upstream.
// Uses the existing resilience.CircuitBreaker implementation.
type CircuitBreakerConfig struct {
	// Threshold is the number of consecutive failures before the circuit opens.
	// Default: 5.
	Threshold int `yaml:"threshold"`

	// ResetAfter is how long the circuit stays open before entering half-open.
	// Default: 30s.
	ResetAfter time.Duration `yaml:"reset_after"`
}

// LoadShedConfig configures adaptive load shedding at the edge proxy.
// The shedder drops low-priority requests when the system is under stress.
type LoadShedConfig struct {
	// Enabled controls whether the adaptive load shedder runs. Default: true.
	Enabled bool `yaml:"enabled"`

	// CPUThresholdPct is the CPU utilisation percentage above which
	// low-priority requests are dropped. Default: 80.0.
	CPUThresholdPct float64 `yaml:"cpu_threshold_pct"`

	// MaxQueueDepth is the maximum request queue depth. Requests arriving
	// when the queue is full are dropped immediately. Default: 500.
	MaxQueueDepth int `yaml:"max_queue_depth"`

	// PriorityHeader is the request header name used to identify priority.
	// Requests with this header value of "critical" are never dropped.
	// Default: "X-Priority".
	PriorityHeader string `yaml:"priority_header"`
}

// DefaultConfig returns a Config with all defaults pre-populated.
// Used when the edge_proxy section is absent from aegisflow.yaml.
func DefaultConfig() Config {
	return Config{
		Enabled:         false,
		ListenAddr:      ":8443",
		H2CAddr:         ":8444",
		HTTP3Enabled:    false,
		HTTP3Addr:       ":8445",
		MaxIdleConns:    100,
		IdleConnTimeout: 90 * time.Second,
		DialTimeout:     10 * time.Second,
		TraceStoreSize:  500,
		Upstreams: []UpstreamConfig{
			{
				Name:   "gateway",
				Prefix: "/api",
				Target: "http://localhost:8080",
				CircuitBreaker: CircuitBreakerConfig{
					Threshold:  5,
					ResetAfter: 30 * time.Second,
				},
				Timeout: 120 * time.Second,
			},
			{
				Name:   "mcpgw",
				Prefix: "/mcp",
				Target: "http://localhost:9090",
				CircuitBreaker: CircuitBreakerConfig{
					Threshold:  5,
					ResetAfter: 30 * time.Second,
				},
				Timeout: 60 * time.Second,
			},
			{
				Name:   "admin",
				Prefix: "/admin",
				Target: "http://localhost:8081",
				CircuitBreaker: CircuitBreakerConfig{
					Threshold:  10,
					ResetAfter: 15 * time.Second,
				},
				Timeout: 30 * time.Second,
			},
		},
		LoadShed: LoadShedConfig{
			Enabled:         true,
			CPUThresholdPct: 80.0,
			MaxQueueDepth:   500,
			PriorityHeader:  "X-Priority",
		},
	}
}
