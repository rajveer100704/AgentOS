# AgentOS Architecture Design Document

![AgentOS Architecture Flow Diagram](assets/ARCHITECTURE.png)

AgentOS (formerly AgentOS) is a production-grade AI agent governance and high-performance edge infrastructure platform. It sits between AI agents (e.g. Claude, GPT, Custom Agents) and the tools/LLM providers they interact with. It acts as an intercepting gateway to enforce access control policies, rotate LLM keys, log tamper-proof evidence, and distribute traces.

---

## High-Level Architecture Flow

The following diagram details the flow of a request from an AI Agent through the Edge Proxy, Control Plane, Policy Engine, Credential Broker, and Evidence Chain. It showcases the integration of HTTP/3, OpenTelemetry (OTel), Prometheus, Grafana, and Jaeger.

```mermaid
graph TB
    Agent["AI Agent (Claude / GPT / Client)"]

    subgraph EdgeLayer["AgentOS Edge Proxy (:8443/:8444/:8445)"]
        H3["HTTP/3 Listener (QUIC/UDP :8445)"]
        H2["HTTP/2 TLS Listener (TCP :8443)"]
        H2C["HTTP/2 Cleartext (TCP :8444)"]
        AltSvc["Alt-Svc Advertising"]
        
        PathRouter["Path Router\n/api/* → Gateway (:8080)\n/admin/* → Observatory (:8081)"]
        EdgeRL["Token-Bucket Rate Limiter"]
        ALS["Adaptive Load Shedder\n(CPU- & Queue-depth aware)"]
        CB["Upstream Circuit Breakers"]
        ConnPool["HTTP Connection Pool"]
        
        EdgeMetrics["Prometheus Counters:\nAgentOS_edge_requests_total\nAgentOS_edge_active_connections\nAgentOS_edge_loadshed_total"]
    end

    subgraph ControlPlane["AgentOS Control Plane (:8080)"]
        Auth["Auth & Tenant Validation"]
        RL["Tenant Rate Limiter (RPM/TPM)"]
        BehavKill["Behavioral Risk Assessment"]
        PolicyIn["Input Policy Engine\n(Keyword / Regex / PII / WASM)"]
        Router["Model Router"]
        Creds["Credential Broker\n(Vault / AWS / static keys)"]
        LLMCalls["LLM Adapters\n(OpenAI / Anthropic / Ollama)"]
        PolicyOut["Output Policy Engine"]
        Evidence["Evidence Chain\n(SHA-256 block chain)"]
    end

    subgraph Observatory["AgentOS Observatory & Monitoring"]
        AdminAPI["Admin API (:8081)"]
        TraceStore["TraceStore (In-process ring buffer)"]
        Dashboard["Observatory Dashboard UI"]
        
        Prometheus["Prometheus Server\n(Scrapes :8081/metrics)"]
        Grafana["Grafana Dashboard\n(Served vs Rate-limited vs Shed)"]
        Jaeger["Jaeger OTLP Backend\n(Distributed Traces)"]
    end

    %% Routing Flow
    Agent -->|QUIC / UDP| H3
    Agent -->|TLS / TCP| H2
    Agent -->|Cleartext| H2C
    H2 -.->|Advertises H3 via| AltSvc
    
    H3 --> PathRouter
    H2 --> PathRouter
    H2C --> PathRouter
    
    PathRouter --> EdgeRL
    EdgeRL --> ALS
    ALS --> CB
    CB --> ConnPool
    ConnPool -->|Forwarded Request| Auth

    Auth --> RL --> BehavKill --> PolicyIn
    PolicyIn --> Router
    Router --> Creds
    Creds --> LLMCalls
    LLMCalls --> PolicyOut
    PolicyOut --> Evidence

    %% Telemetry & Metrics Flow
    EdgeLayer -.->|Metrics| EdgeMetrics
    EdgeMetrics -.->|Scraped by| Prometheus
    ControlPlane -.->|Prometheus Counters/Histograms| Prometheus
    Prometheus -.->|Visualized by| Grafana
    
    EdgeLayer -.->|OTel Spans: edge.request| Jaeger
    ControlPlane -.->|OTel Spans: gateway, policy, provider, evidence| Jaeger
    Jaeger -.->|Fed into| TraceStore
    TraceStore -.->|Renders in| Dashboard
    AdminAPI -.->|Exposes| TraceStore
```

---

## 3-Layer Infrastructure Architecture

AgentOS is organized into three decoupled logical layers to optimize performance, scalability, and security:

### 1. AgentOS Edge Proxy Layer
*   **Purpose:** Highly-optimized traffic management, protocol termination, and boundary reliability.
*   **Protocols:** Supports HTTP/1.1, HTTP/2 (via ALPN and H2C), and HTTP/3 (over QUIC/UDP).
*   **Key Responsibilities:**
    *   **TLS 1.3 Termination:** Modern, secure cryptography using ECDSA P-256 / X25519.
    *   **Alt-Svc Advertisement:** Automatically injects `Alt-Svc: h3=":8445"; ma=86400` headers on HTTP/2 responses to allow seamless client upgrades to QUIC.
    *   **Adaptive Load Shedding:** CPU-aware and Queue-depth-aware request rejection. Drops low-priority requests during spikes while guaranteeing throughput for critical requests (marked by `X-Priority: critical`).
    *   **Circuit Breaking:** Independent circuit breakers per upstream target (e.g. gateway, admin API). Trips when error rates spike, protecting downstreams from cascading failure.
    *   **OTel Span Injection:** Generates the outer `edge.request` span, mapping client handshake overhead and upstream response latency.

### 2. Control Plane Layer
*   **Purpose:** Fine-grained API key validation, security policy enforcement, routing, and trust verification.
*   **Key Responsibilities:**
    *   **Tenant Authentication & Rate Limiting:** Exposes custom rate limit ceilings per tenant, enforcing Request-Per-Minute (RPM) and Token-Per-Minute (TPM) limits early in the request chain.
    *   **Input/Output Policy Engine:** Evaluates prompts and completions against keyword lists, regular expressions, PII patterns, and custom WASM sandbox modules.
    *   **Model-to-Provider Router:** Decouples requested models (e.g., `smart`, `fast`) from actual backend providers (OpenAI, Anthropic, Ollama, Mock).
    *   **Credential Broker:** Resolves, rotates, and injects LLM API credentials securely at request time (using Vault, AWS STS, or static environments), ensuring agents never touch raw API keys.
    *   **Evidence Chain:** Creates a cryptographically linked ledger (SHA-256 chain) of request/response envelopes, providing non-repudiation and immutable logging for audit.

### 3. Observatory & Analytics Layer
*   **Purpose:** Real-time system monitoring, distributed trace visualization, and manual governance controls.
*   **Key Responsibilities:**
    *   **TraceStore:** In-process 500-trace ring buffer implementing the OpenTelemetry `SpanProcessor` interface. Enables immediate, zero-dependency trace explorer lookup inside the dashboard without requiring a Jaeger database.
    *   **Jaeger Integration:** Long-term OTLP trace backend integration for structural trace analysis.
    *   **Grafana Dashboards:** Visual representation of requests served, rate-limited, and shed, coupled with active connection gauges, circuit breaker status timelines, and P95 latency buckets.
    *   **Human-in-the-Loop Approvals:** Holds high-risk requests (e.g. commands to delete tables or execute code) in a suspended state until manually approved by an operator via the Admin API.

---

## Detailed Request Lifecycle

```mermaid
sequenceDiagram
    autonumber
    participant Agent as AI Agent
    participant Edge as AgentOS Edge (:8443)
    participant GW as Gateway (:8080)
    participant Policy as Policy Engine
    participant Creds as Credential Broker
    participant LLM as LLM Provider
    participant Ev as Evidence Chain
    participant Observatory as Observatory (Jaeger/TraceStore)

    Agent->>Edge: HTTPS POST /api/v1/chat/completions (HTTP/2 or HTTP/3)
    activate Edge
    Note over Edge: 1. TLS Termination / UDP Receive<br/>2. Connection Count / Load Shed Check<br/>3. Start OTel Span: edge.request
    
    Edge->>GW: Forwarded request over internal loopback H2C/HTTP
    activate GW
    Note over GW: 4. Tenant Auth & rate limit check<br/>5. Start OTel Span: gateway.chat_completion
    
    GW->>Policy: Evaluate Input Prompt
    activate Policy
    Note over Policy: OTel Span: policy.check_input
    Policy-->>GW: Policy Decision (Allow / Block / Review)
    deactivate Policy

    alt Decision is Block (403)
        GW-->>Edge: Error Response (policy_violation)
        Edge-->>Agent: 403 Forbidden
    else Decision is Review (202)
        Note over GW: Push request to Approval Queue<br/>Suspend execution
        GW-->>Edge: HTTP 202 Accepted (queued)
        Edge-->>Agent: 202 Accepted (queued)
    else Decision is Allow (200)
        GW->>Creds: Retrieve Provider API Key
        activate Creds
        Creds-->>GW: Decrypted LLM Key (rotated)
        deactivate Creds
        
        GW->>LLM: Secure POST /v1/chat/completions
        activate LLM
        Note over LLM: OTel Span: provider.call
        LLM-->>GW: ChatCompletion Response
        deactivate LLM
        
        GW->>Policy: Evaluate Output Completion
        activate Policy
        Note over Policy: OTel Span: policy.check_output
        Policy-->>GW: Policy Decision (Allow / Redact)
        deactivate Policy
        
        GW->>Ev: Commit Request & Response
        activate Ev
        Note over Ev: OTel Span: evidence.record<br/>Compute SHA-256 hash link
        Ev-->>GW: Block committed
        deactivate Ev

        GW-->>Edge: Response Envelope
        deactivate GW
        Edge-->>Agent: Response with Alt-Svc Headers
        deactivate Edge
    end

    %% Async Telemetry Export
    GW-)Observatory: Export OTel Spans (OTLP)
    Edge-)Observatory: Export OTel Spans (OTLP)
    Observatory-)Observatory: Process & store traces
```

## Trace Span Topology

Each transaction is tracked end-to-end under a single trace ID, propagating the following hierarchy:

*   `edge.request` [Edge Layer]
    *   `gateway.chat_completion` [Control Plane]
        *   `policy.check_input` [Policy Engine]
        *   `provider.call` [LLM Layer]
            *   `provider.openai` / `provider.anthropic` / `provider.ollama` (dependent adapter span)
        *   `policy.check_output` [Policy Engine]
        *   `evidence.record` [Evidence Chain]
