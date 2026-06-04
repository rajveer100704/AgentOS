# AgentOS System Limitations & Architectural Constraints

This document outlines the architectural boundaries, system constraints, and resource limits of the AgentOS platform. Reviewing these constraints is critical for capacity planning and production deployment design.

---

## 1. TraceStore Capacity & Memory Bound
*   **Constraint:** The in-process `TraceStore` uses an in-memory ring buffer (default: last 500 traces).
*   **Implication:** Traces are not persistent. Restarting the AgentOS gateway binary immediately clears all trace history from the Observatory dashboard.
*   **Mitigation:** For production setups, enable the Jaeger exporter (`telemetry.exporter: "jaeger"`) to write traces to an external persistent database (Elasticsearch, Cassandra, or Badger).

## 2. UDP Buffer Limits under Stress
*   **Constraint:** HTTP/3 over QUIC utilizes UDP packets. Operating systems (especially Windows and default Linux configurations) enforce tight limits on UDP socket receive/send buffers.
*   **Implication:** Under extreme concurrency stress tests (1000+ streams), UDP packet drops can occur at the OS kernel boundary, triggering high QUIC packet retransmissions or skips.
*   **Mitigation:** Increase the kernel UDP read/write buffer sizes before running high-throughput HTTP/3 servers. For Linux:
    ```bash
    sysctl -w net.core.rmem_max=2500000
    sysctl -w net.core.wmem_max=2500000
    ```

## 3. Software-Level Tenant Isolation
*   **Constraint:** Multi-tenancy is enforced logically via middleware validation using API keys.
*   **Implication:** Tenants share the same Go runtime memory space and CPU cycles. AgentOS does not provide virtualized process-level or container-level tenant isolation.
*   **Mitigation:** If strict tenant isolation is required (e.g. multi-tenant hosting), deploy separate AgentOS replicas in isolated Kubernetes namespaces rather than relying solely on logical tenant keys.

## 4. WASM Policy Latency Overhead
*   **Constraint:** Running custom policy plugins in WASM requires WebAssembly runtime invocation (`wazero`).
*   **Implication:** While Go-native policies execute in under $2\mu\text{s}$, WASM policy invocation incurs compiling/JIT overhead and memory boundary crossing delays (typically 10–50$\mu\text{s}$ per check).
*   **Mitigation:** Keep WASM modules lightweight and use native Go policies (regex/keyword) for high-frequency or simple rule checks.

## 5. Single Binary Fail-Closed Policy
*   **Constraint:** In governance mode, the platform defaults to **fail-closed**.
*   **Implication:** If a policy engine or an upstream dependency (like an external webhook notifier) experiences a deadlock or timeout, the request is blocked.
*   **Mitigation:** Ensure webhook notifier timeouts are tightly bound, and enable break-glass overrides only in development environments.
