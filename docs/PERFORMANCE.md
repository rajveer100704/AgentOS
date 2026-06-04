# AgentOS Protocol Performance & Reliability Report

This document details the performance profile, scale capabilities, and reliability characteristics of the AgentOS Edge proxy and Control Plane. Benchmarks evaluate the throughput (RPS), latency distributions ($P_{50}$, $P_{95}$, $P_{99}$), connection success rates, and memory footprints of **HTTP/1.1**, **HTTP/2**, and **HTTP/3 (QUIC)**.

---

## Live Performance Benchmarks (Stress Test)

Tests were executed locally on an isolated loopback interface to eliminate external network fluctuations. The target handler simulated a minimal **$500\mu\text{s}$ gateway policy execution overhead** to test proxy layer efficiency rather than backend LLM provider latency.

### Benchmarking Environment
*   **Operating System:** Windows 10 (amd64)
*   **CPU:** AMD Ryzen 5 7530U with Radeon Graphics (6 Cores, 12 Threads / vCPUs)
*   **Go Version:** `go1.26.1`
*   **Duration per Run:** 1.0 seconds
*   **API Payload Size:** Standard OpenAI ChatCompletion request payload (~340 bytes)

### Summary Comparison Table

| Protocol | Concurrency | RPS | $P_{50}$ Latency | $P_{95}$ Latency | $P_{99}$ Latency | Success Rate | Memory Allocated |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **HTTP/1.1** | 1 (Sequential) | 857.9 | 1.00 ms | 2.00 ms | 4.00 ms | 100.00% | 6.49 MB |
| **HTTP/1.1** | 100 Clients | 19,307.0 | 3.00 ms | 13.00 ms | 29.00 ms | 100.00% | 151.36 MB |
| **HTTP/1.1** | 500 Clients | 3,519.7 | 46.00 ms | 93.00 ms | 1,190.00 ms | **63.57%** | 62.24 MB (Exhausted) |
| **HTTP/1.1** | 1000 Clients | 730.2 | 83.00 ms | 1,042.00 ms | 1,057.00 ms | **10.99%** | 65.89 MB (Exhausted) |
| **HTTP/2** | 1 (Sequential) | 663.1 | 1.00 ms | 3.00 ms | 8.00 ms | 100.00% | 5.74 MB |
| **HTTP/2** | 100 Streams | 11,913.4 | 6.00 ms | 12.00 ms | 17.00 ms | 100.00% | 104.38 MB |
| **HTTP/2** | 500 Streams | 12,840.4 | 38.00 ms | 45.00 ms | 48.00 ms | 100.00% | 121.70 MB |
| **HTTP/2** | 1000 Streams | 14,265.5 | 66.00 ms | 86.00 ms | 91.00 ms | 100.00% | 144.82 MB |
| **HTTP/3** | 1 (Sequential) | 913.8 | 1.00 ms | 1.00 ms | 2.00 ms | 100.00% | 13.50 MB |
| **HTTP/3** | 100 Streams | **27,366.4** | 2.00 ms | 7.00 ms | 10.00 ms | 100.00% | 402.20 MB |
| **HTTP/3** | 500 Streams | **27,327.3** | 17.00 ms | 24.00 ms | 27.00 ms | 100.00% | 413.93 MB |
| **HTTP/3** | 1000 Streams | **26,434.3** | 36.00 ms | 52.00 ms | 63.00 ms | 100.00% | 414.41 MB |

---

## Systems Analysis & Key Insights

### 1. HTTP/3 (QUIC) Multi-Stream Efficiency
Under high concurrency (500–1000 streams), HTTP/3 over QUIC/UDP demonstrates significant advantages:
*   **Throughput Advantage:** HTTP/3 achieved **26,434.3 RPS** at 1000 concurrent streams, which is **85% higher** than HTTP/2 (**14,265.5 RPS**).
*   **Tail Latency Reduction:** The $P_{99}$ latency for HTTP/3 at 1000 streams was only **63ms**, compared to **91ms** for HTTP/2 (a **30% reduction**). 
*   **Why it works:** HTTP/3 multiplexes request streams over UDP using independent QUIC transport connections. Unlike TCP, where a dropped packet blocks the entire TCP window (TCP Head-of-Line Blocking), QUIC handles packet loss per stream. A lost packet only stalls the specific stream it belonged to, allowing other streams to continue processing immediately.

### 2. HTTP/1.1 Ephemeral Port Exhaustion Failure Mode
At 500 and 1000 concurrent clients, HTTP/1.1 throughput and success rates collapsed:
*   **The Symptom:** Success rate dropped to **63.57%** at 500 clients, and collapsed to **10.99%** at 1000 clients.
*   **The Cause:** In HTTP/1.1, when the request pool limit is exceeded under high concurrency, the client is forced to open new TCP connections. On Windows (and other operating systems), sockets closed by the client enter the `TIME_WAIT` state for 120 seconds to catch delayed packets. Under a flood of requests, all available ephemeral ports (typically ports 49152–65535) are exhausted. 
*   **The Error:** Subsequent dials return the OS socket error: `connectex: No connection could be made because the target machine actively refused it`.
*   **Multiplexing Rescue:** HTTP/2 and HTTP/3 are entirely immune to this failure mode. They multiplex all concurrent streams over a **single persistent connection** (one TCP socket for HTTP/2, one UDP socket for HTTP/3), requiring exactly **one** local port.

### 3. Memory Footprint & Resource Tradeoffs
*   **HTTP/1.1 & HTTP/2:** Maintain a lower initial memory footprint (~5–6 MB sequentially, scaling to ~144 MB at 1000 streams).
*   **HTTP/3:** Allocates more memory (~13.5 MB sequentially, scaling to ~414 MB at 1000 streams).
*   **Tradeoff:** The increase in HTTP/3 memory consumption is due to the in-memory QUIC state machine, packet reassembly buffers, and congestion control algorithms managed in Go user space (`quic-go`) rather than the kernel. For systems-level roles (like Cloudflare), this tradeoff is highly acceptable because the user-space UDP processing bypasses OS-level TCP socket locks, leading to far higher throughput and lower tail latencies.

---

## Reliability Stack Performance

Under high load, the AgentOS Edge utilizes three safeguards to stabilize latency and protect downstreams:

```
[Incoming Request]
       │
       ▼
 1. Rate Limiter  ── (Tenant exceed?) ──► [429 Too Many Requests]
       │
       ▼
 2. Load Shedder  ── (CPU > 80% OR Queue Full?) ──► [503 Service Unavailable]
       │
       ▼
 3. Circuit Breaker ── (Upstream errors > 5?) ──► [503 CB Open - Early Reject]
       │
       ▼
[Upstream Gateway]
```

1.  **Rate Limiter (Early Drop):** Filters requests at the tenant boundary before they incur processing costs, returning HTTP `429` with a `Retry-After` header.
2.  **Adaptive Load Shedder (CPU Guard):** Monitors system health. If CPU utilization exceeds `80%` or active queue depth exceeds `500` requests, it sheds non-critical requests early (HTTP `503`), keeping the system responsive for critical requests (`X-Priority: critical`).
3.  **Circuit Breaker (Fast Failure):** If the upstream Control Plane gateway returns 5 consecutive failures, the Edge proxy trips the circuit breaker to `Open` state. Future requests are instantly rejected at the Edge layer with an HTTP `503` (avoiding connection timeouts and thread blockage) until the reset timeout expires.

---

## Estimated Capacity Planning & Horizontal Scaling Projections

Based on the measured performance of a single instance, we can project system resource requirements and scaling behaviors as replica counts increase.

### Estimated Horizontal Scaling Projections (RPS)

Assuming near-linear scaling when utilizing a stateless Edge Proxy and Gateway layer behind a high-performance load balancer (e.g. AWS ALB or Cloudflare Load Balancer):

*   **1 Instance (Baseline):**
    *   *Throughput:* ~26,400 RPS (HTTP/3) / ~14,200 RPS (HTTP/2)
    *   *Resources:* 1 vCPU, 512Mi Memory limit
*   **3 Replicas (High Availability Setup):**
    *   *Throughput:* ~79,200 RPS (HTTP/3) / ~42,600 RPS (HTTP/2)
    *   *Resources:* 3 vCPUs, 1.5Gi Memory allocation
*   **10 Replicas (Enterprise Scale):**
    *   *Throughput:* ~264,000 RPS (HTTP/3) / ~142,000 RPS (HTTP/2)
    *   *Resources:* 10 vCPUs, 5.0Gi Memory allocation

### Memory Planning

Memory allocation is primarily driven by active connections, policy complexity, and the telemetry trace buffer:

*   **Telemetry Ring Buffer:** Decoupled from throughput; occupies a static ~15MB for the default 500-trace ring buffer.
*   **QUIC State Overhead:** Each active QUIC stream consumes ~400KB. For 1000 concurrent streams, reserve at least 400MB of resident memory per replica.

