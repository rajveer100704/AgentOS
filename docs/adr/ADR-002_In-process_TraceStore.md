# ADR-002: In-Process TraceStore for Zero-Dependency Observability

## Status
Accepted

## Context
Exposing distributed traces is crucial to understand Agent-to-Tool policy and provider execution overhead. However, running a standard external trace backend (like Jaeger with Elasticsearch or Cassandra) in local development or small-scale deployments introduces high setup friction and excessive resource consumption. We need a way to serve real distributed traces directly to the developer dashboard out-of-the-box with zero configuration or external dependencies.

## Decision
Build a thread-safe, in-process ring buffer (called `TraceStore`) that implements the OpenTelemetry `SpanProcessor` interface:

*   Expose trace collections via a REST endpoint `/admin/v1/traces` served on the Admin port (`:8081`).
*   Fix the buffer size to 500 entries (older traces are evicted automatically via FIFO).
*   Structure the buffer to group spans dynamically by their shared `TraceID`.
*   Maintain Jaeger OTLP capability in parallel (via feature flag `telemetry.exporter: "jaeger"`) so that traces can be exported to standard enterprise telemetry systems in production.

## Consequences
*   **Observability:** Developers can explore detailed span lifecycles, latency allocations, and policy decisions immediately on launch.
*   **Frictionless Setup:** Zero external infrastructure is required to explore traces locally.
*   **Memory:** Telemetry buffer consumes a static, bounded footprint (~15MB), avoiding unbounded memory leakage.
