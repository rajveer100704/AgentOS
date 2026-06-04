# AgentOS — End-to-End Sequence Flow

## Full Request Lifecycle

This document traces every request from the AI agent through to evidence recording, showing which OTel spans are emitted at each stage.

```mermaid
sequenceDiagram
    participant A as AI Agent
    participant E as AgentOS Edge<br/>:8443 TLS
    participant MW as Middleware Chain<br/>(Auth/RL/CORS)
    participant GW as Gateway Handler
    participant PI as Policy Engine<br/>(Input)
    participant RT as Router
    participant OAI as OpenAI / Provider
    participant PO as Policy Engine<br/>(Output)
    participant CR as Credential Broker
    participant AP as Approval Queue
    participant EV as Evidence Chain
    participant JS as Jaeger + TraceStore

    Note over A,E: PHASE 1 — Edge Layer (AgentOS Edge)
    A->>E: HTTPS/2 POST /api/v1/chat/completions
    activate E
    Note over E: [span: edge.request start]<br/>TLS 1.3 decrypt<br/>Check CPU% → adaptive load shed<br/>Token bucket rate limit<br/>Circuit breaker allow()?<br/>Select upstream: /api/* → :8080

    E->>MW: HTTP/2 (internal h2c)
    deactivate E

    Note over A,MW: PHASE 2 — Middleware Chain
    activate MW
    MW->>MW: Auth — extract API key → tenant
    MW->>MW: RateLimit — Allow(tenantID)?
    MW->>MW: CORS headers
    MW->>MW: BudgetCheck (model-level)
    MW->>MW: Logging + Prometheus counter
    MW->>GW: pass request
    deactivate MW

    Note over GW,PI: PHASE 3 — Gateway + Policy Input
    activate GW
    Note over GW: [span: gateway.chat_completion start]<br/>Behavioral kill-switch check<br/>Apply model aliases<br/>Apply transforms

    GW->>PI: CheckInput(messages)
    activate PI
    Note over PI: [span: policy.check_input]<br/>keyword → regex → PII → WASM<br/>Attr: decision, policy_name
    PI-->>GW: Violation (allow/block/warn)
    deactivate PI

    alt Violation == block
        GW-->>A: 403 policy_violation
        Note over GW: span ended with error
    end

    Note over GW,RT: PHASE 4 — Provider Call (most expensive stage)
    GW->>RT: RouteWithProvider(ctx, req)
    activate RT
    Note over RT: [span: provider.call]<br/>Select provider by model/strategy<br/>Check circuit breaker
    RT->>OAI: POST /v1/chat/completions
    activate OAI
    Note over OAI: [span: provider.openai]<br/>Attrs: tokens_prompt, tokens_completion<br/>Latency: typically 500ms–3000ms
    OAI-->>RT: ChatCompletionResponse
    deactivate OAI
    RT-->>GW: RoutedResponse{Response, Provider}
    deactivate RT

    Note over GW,PO: PHASE 5 — Policy Output
    GW->>PO: CheckOutput(response.choices[0].content)
    activate PO
    Note over PO: [span: policy.check_output]<br/>keyword → regex → WASM
    PO-->>GW: Violation (allow/block)
    deactivate PO

    Note over GW,EV: PHASE 6 — Evidence + Side Effects
    GW->>EV: Record(ActionEnvelope)
    activate EV
    Note over EV: [span: evidence.record]<br/>Attrs: hash, session_id<br/>SHA-256 chain append
    EV-->>GW: ok
    deactivate EV

    GW->>GW: Usage.Record(tenant, provider, model, tokens)
    GW->>GW: Analytics.Record(DataPoint)
    GW->>GW: AuditLog(event)

    GW-->>A: 200 ChatCompletionResponse
    Note over GW: [span: gateway.chat_completion end]<br/>Total duration recorded
    deactivate GW

    Note over GW,JS: PHASE 7 — Telemetry Export (async)
    GW--)JS: OTLP gRPC export (batched, ~5s)<br/>Spans: edge + gateway + policy×2 + provider + evidence
    JS--)JS: Jaeger UI shows full trace tree
    JS--)JS: TraceStore.Add(TraceSummary) → dashboard
```

---

## MCP Tool Call Flow

```mermaid
sequenceDiagram
    participant A as AI Agent
    participant E as AgentOS Edge :8443
    participant MCP as MCP Gateway :9090
    participant TP as Tool Policy Engine
    participant AP as Approval Queue
    participant CR as Credential Broker
    participant Tool as Upstream Tool Server

    A->>E: POST /mcp/* (HTTP/2)
    E->>MCP: forward (circuit breaker checked)

    activate MCP
    Note over MCP: [span: mcpgw.tool_call]<br/>Parse MCP tool call envelope

    MCP->>TP: Check(protocol, tool, target, capability)
    TP-->>MCP: Decision: allow | review | block

    alt Decision == review
        MCP->>AP: Enqueue(ApprovalItem)
        Note over AP: [span: approval.enqueue]<br/>Notify GitHub/Slack<br/>Wait for human decision
        AP-->>MCP: approved / denied
    end

    alt Decision == allow
        Note over MCP: [span: credential.issue] (call-site wrapper)
        MCP->>CR: Issue(credentialRef)
        CR-->>MCP: Token (scoped, TTL-bound)
        MCP->>Tool: Execute with credential
        Tool-->>MCP: Result
    end

    MCP-->>A: MCP Response
    deactivate MCP
```

---

## Span Inventory

| Span Name | Emitted By | Key Attributes |
|-----------|------------|----------------|
| `edge.request` | edgeproxy/proxy.go | `http.method`, `http.url`, `upstream`, `status_code` |
| `gateway.chat_completion` | gateway/handler.go | `tenant_id`, `model`, `stream` |
| `policy.check_input` | gateway/handler.go (wrapper) | `decision`, `policy_name`, `action` |
| `provider.call` | gateway/handler.go (wrapper) | `provider`, `model`, `latency_ms` |
| `provider.openai` | gateway/handler.go (wrapper) | `tokens_prompt`, `tokens_completion`, `tokens_total` |
| `provider.anthropic` | gateway/handler.go (wrapper) | same as openai |
| `provider.ollama` | gateway/handler.go (wrapper) | same as openai |
| `policy.check_output` | gateway/handler.go (wrapper) | `decision` |
| `evidence.record` | gateway/handler.go (wrapper) | `hash`, `session_id` |
| `mcpgw.tool_call` | mcpgw (call-site) | `tool`, `target`, `capability`, `decision` |
| `approval.enqueue` | mcpgw (call-site) | `approval_id`, `tool`, `timeout` |
| `credential.issue` | mcpgw (call-site) | `credential_name`, `ttl`, `scope` |

---

## Latency Budget (Typical)

```
edge.request              total: ~1350ms
├── edge overhead              5ms   (TLS + routing + metrics)
├── gateway.chat_completion  1345ms
│   ├── policy.check_input      3ms
│   ├── provider.call         1330ms
│   │   └── provider.openai   1330ms  ← usually 95%+ of total
│   ├── policy.check_output     4ms
│   └── evidence.record         2ms
```

P95 latency is dominated entirely by the upstream LLM provider.
