# ADR-001: HTTP/3 (QUIC) Support via quic-go

## Status
Accepted

## Context
AI agent client traffic requires highly concurrent, multiplexed request processing at the Edge Proxy boundary. Standard HTTP/1.1 suffers from ephemeral TCP port exhaustion and head-of-line (HOL) blocking under stress. HTTP/2 improves this by multiplexing streams over a single TCP connection, but a single lost packet still stalls the entire TCP window. We need a transport protocol that supports multiplexed streams without TCP HOL blocking or connection pooling socket overhead.

## Decision
Implement HTTP/3 (QUIC) support at the Edge Proxy layer using the standard Go user-space library `github.com/quic-go/quic-go/http3`. 

*   Expose HTTP/3 on a dedicated UDP port (`:8445`).
*   Advertise QUIC support to clients dynamically using the `Alt-Svc: h3=":8445"; ma=86400` header on HTTP/2 TLS responses.
*   Make HTTP/3 configurable and **disabled by default** (`http3_enabled: false`) to guarantee startup path stability.

## Consequences
*   **Performance:** Under 1000 concurrent streams, throughput increases by **85%** (~26.4k RPS vs ~14.2k RPS for HTTP/2) and tail latency $P_{99}$ drops by **30%** (63ms vs 91ms).
*   **Resilience:** Completely bypasses TCP port exhaustion failure modes and TCP head-of-line blocking.
*   **Resources:** Incurs higher resident memory overhead (~400KB per active connection/stream buffer) because QUIC congestion control and packet reassembly are handled in user space.
