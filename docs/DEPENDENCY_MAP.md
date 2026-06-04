# AgentOS — Package Dependency Map

## Dependency Stability Classification

| Status | Meaning |
|--------|---------|
| 🟢 **NEW** | Created during this upgrade |
| 🟡 **ENHANCED** | Existing package with minimal additions |
| 🔴 **FROZEN** | Zero modifications — only used via call-site wrappers |
| ⚪ **UNCHANGED** | Existing package, no changes needed |

---

## Package Dependency Graph

```
cmd/AgentOS (ENHANCED — main.go startup)
├── internal/edgeproxy  🟢 NEW
│   ├── internal/ratelimit  ⚪ UNCHANGED (reuses Limiter interface)
│   ├── internal/resilience ⚪ UNCHANGED (reuses CircuitBreaker)
│   ├── internal/loadshed   ⚪ UNCHANGED (reuses Loadshed)
│   └── internal/telemetry  🟡 ENHANCED (tracer, metrics, store)
│
├── internal/gateway  🟡 ENHANCED (span wrappers at call sites)
│   ├── internal/policy     🔴 FROZEN  (called, not modified)
│   ├── internal/credential 🔴 FROZEN  (called, not modified)
│   ├── internal/evidence   🔴 FROZEN  (called, not modified)
│   ├── internal/approval   🔴 FROZEN  (called, not modified)
│   ├── internal/behavioral 🔴 FROZEN  (called, not modified)
│   ├── internal/analytics  🔴 FROZEN  (called, not modified)
│   ├── internal/router     ⚪ UNCHANGED
│   └── internal/telemetry  🟡 ENHANCED (StartSpan helper)
│
├── internal/telemetry  🟡 ENHANCED
│   ├── telemetry.go        (add OTLP exporter + TraceStore processor)
│   ├── tracer.go  🟢 NEW   (StartSpan, RecordError, SetDecision)
│   ├── metrics.go 🟢 NEW   (Prometheus counters/histograms)
│   ├── store.go   🟢 NEW   (TraceStore ring buffer)
│   └── attributes.go       (existing — UNCHANGED)
│
├── internal/admin  🟡 ENHANCED
│   ├── tracehandler.go 🟢 NEW  (query TraceStore → HTTP JSON)
│   ├── server.go (add /admin/v1/traces route)
│   └── internal/telemetry (TraceStore injected)
│
├── internal/config  🟡 ENHANCED
│   └── config.go (add EdgeProxyConfig + TraceStoreSize)
│
└── (all other packages: FROZEN or UNCHANGED)
```

---

## Dependency Table

| Package | Depends On | Used By | Status |
|---------|-----------|---------|--------|
| `edgeproxy` | `ratelimit`, `resilience`, `loadshed`, `telemetry`, `config` | `cmd/AgentOS` | 🟢 NEW |
| `telemetry/tracer` | `go.opentelemetry.io/otel` | `edgeproxy`, `gateway` | 🟢 NEW |
| `telemetry/metrics` | `prometheus/client_golang` | `edgeproxy`, `middleware` | 🟢 NEW |
| `telemetry/store` | stdlib only | `telemetry`, `admin` | 🟢 NEW |
| `admin/tracehandler` | `telemetry/store` | `admin/server` | 🟢 NEW |
| `grpcgw` (Phase 8) | `policy`, `credential`, `evidence`, `telemetry` | `cmd/AgentOS` | 🟢 PLANNED |
| `gateway` | `policy`, `credential`, `router`, `analytics`, `evidence`, `behavioral` | `cmd/AgentOS` | 🟡 ENHANCED |
| `telemetry` | `otel`, `otlptrace`, `prometheus` | `gateway`, `edgeproxy`, `admin` | 🟡 ENHANCED |
| `config` | `gopkg.in/yaml.v3` | All | 🟡 ENHANCED |
| `admin` | `telemetry/store`, all adapters | `cmd/AgentOS` | 🟡 ENHANCED |
| `policy` | `wasmfilter` (wazero) | `gateway`, `grpcgw` | 🔴 FROZEN |
| `approval` | `webhook` | `gateway`, `mcpgw` | 🔴 FROZEN |
| `credential` | `vault`, AWS STS | `gateway`, `mcpgw` | 🔴 FROZEN |
| `evidence` | stdlib | `gateway` | 🔴 FROZEN |
| `behavioral` | stdlib | `gateway` | 🔴 FROZEN |
| `analytics` | stdlib | `gateway` | 🔴 FROZEN |
| `audit` | `storage` (postgres) | `gateway` | 🔴 FROZEN |
| `toolpolicy` | stdlib | `mcpgw`, `cmd/AgentOS` | 🔴 FROZEN |
| `router` | `provider`, `resilience` | `gateway` | ⚪ UNCHANGED |
| `provider` | stdlib | `router` | ⚪ UNCHANGED |
| `ratelimit` | `redis/go-redis` | `middleware`, `edgeproxy` | ⚪ UNCHANGED |
| `resilience` | stdlib | `router`, `edgeproxy` | ⚪ UNCHANGED |
| `loadshed` | stdlib | `middleware`, `edgeproxy` | ⚪ UNCHANGED |
| `mcpgw` | `toolpolicy`, `approval` | `cmd/AgentOS` | ⚪ UNCHANGED |
| `middleware` | `ratelimit`, `loadshed` | `cmd/AgentOS` | ⚪ UNCHANGED |

---

## New Go Module Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `go.opentelemetry.io/otel/exporters/otlp/otlptrace` | latest compat | OTLP base |
| `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc` | latest compat | Jaeger via OTLP gRPC |
| `google.golang.org/grpc` | v1.x | gRPC (Phase 8) |
| `github.com/quic-go/quic-go` | v0.x | HTTP/3 benchmarks |

> `golang.org/x/net/http2` — **already in go.mod** (v0.55.0). HTTP/2 = zero new dep.
> `google.golang.org/protobuf` — **already indirect**. Promoted to direct in Phase 8.
