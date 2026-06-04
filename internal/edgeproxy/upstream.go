package edgeproxy

import (
	"net/url"
	"time"

	"github.com/saivedant169/AegisFlow/internal/resilience"
	"github.com/saivedant169/AegisFlow/internal/telemetry"
)

// Upstream represents one backend destination that the edge proxy can route to.
// Each upstream has its own circuit breaker, Prometheus metrics, and timeout.
type Upstream struct {
	Name        string
	Prefix      string
	Target      *url.URL
	StripPrefix bool
	Timeout     time.Duration
	Breaker     *resilience.CircuitBreaker
}

// newUpstream constructs an Upstream from its UpstreamConfig, creating
// a fresh CircuitBreaker with the configured threshold and reset period.
func newUpstream(cfg UpstreamConfig) (*Upstream, error) {
	target, err := url.Parse(cfg.Target)
	if err != nil {
		return nil, err
	}

	threshold := cfg.CircuitBreaker.Threshold
	if threshold <= 0 {
		threshold = 5
	}
	resetAfter := cfg.CircuitBreaker.ResetAfter
	if resetAfter <= 0 {
		resetAfter = 30 * time.Second
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 120 * time.Second
	}

	return &Upstream{
		Name:        cfg.Name,
		Prefix:      cfg.Prefix,
		Target:      target,
		StripPrefix: cfg.StripPrefix,
		Timeout:     timeout,
		Breaker:     resilience.NewCircuitBreaker(cfg.Name, threshold, resetAfter),
	}, nil
}

// updateBreakerMetric updates the Prometheus gauge for this upstream's
// circuit breaker state: 0=closed, 1=open, 2=half-open.
func (u *Upstream) updateBreakerMetric() {
	var state float64
	switch u.Breaker.State() {
	case resilience.CircuitClosed:
		state = 0
	case resilience.CircuitOpen:
		state = 1
	case resilience.CircuitHalfOpen:
		state = 2
	}
	telemetry.CircuitBreakerState.WithLabelValues(u.Name).Set(state)
}
