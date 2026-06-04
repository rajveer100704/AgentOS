# AgentOS Features

This document provides a comprehensive list of capabilities supported by the AgentOS platform, categorized into core execution governance and supporting infrastructure.

## Execution Governance (the core)

### Protocol-Boundary Enforcement
- Normalize agent actions into `ActionEnvelope` objects
- Evaluate per-tool and per-action policies
- Support for MCP, HTTP, shell, Git, and SQL action types

### Policy Engine
- **Input policies**: block prompt injection, detect PII before it reaches providers
- **Output policies**: filter harmful content in responses
- Keyword blocklist, regex patterns, PII detection (email, SSN, credit card)
- Per-policy actions: `allow`, `review`, `block`
- WASM policy plugins for custom filters (any language that compiles to WebAssembly)
- Fail-closed governance mode (configurable break-glass for development)

### Tamper-Evident Evidence
- SHA-256 hash-chained audit log with append-only writes
- Session manifest with ordered action records
- Policy decisions, approval records, credential issuance records
- Human-readable Markdown and HTML evidence reports for auditors
- Exportable evidence bundles with `agentctl evidence export`
- Tamper detection and verification via `agentctl evidence verify`

### Human-in-the-Loop Approvals
- Review queue for risky actions (human approves or denies before execution)
- Slack notifications with Block Kit messages and approve/deny deep links
- GitHub PR comment notifications with risk-level indicators
- Configurable auto-deny timeout for unreviewed actions
- `agentctl approve` / `agentctl deny` CLI commands

### Behavioral Session Policy
- 6 built-in detection rules: exfiltration, privilege escalation, credential abuse, destructive sequences, suspicious fan-out, repeated escalation
- Cumulative risk scoring per session (0-100)
- Kill switch: auto-block sessions that exceed a configurable risk threshold
- Session-level anomaly detection catches patterns that individual action checks miss

### Task Manifests & Drift Enforcement
- Declare agent intent: allowed tools, protocols, verbs, resources, action limits, budgets
- Drift detection compares declared intent vs actual execution
- Configurable enforcement mode: `warn` (log only) or `enforce` (block violations)
- 7 drift types: unexpected tool, resource, protocol, verb, exceeded actions, exceeded budget, manifest expired

### Enterprise RBAC
- Three-role hierarchy: admin, operator, viewer
- Per-API-key role assignment
- Org/team/project/environment identity hierarchy
- Separation-of-duties rules (policy author cannot approve, admin cannot operate sessions)
- Backward-compatible tenant config

## Supporting Infrastructure

These features support the governance plane and remain fully functional:

### AI Gateway
- OpenAI-compatible API for 10+ providers (OpenAI, Anthropic, Ollama, Gemini, Azure, Groq, Mistral, Together, Bedrock)
- Streaming (SSE) and non-streaming support
- WebSocket support for long-lived connections at `/v1/ws`
- GraphQL admin API alongside REST

### Intelligent Routing
- Route by model name with fallback chains
- Circuit breaker, retry with exponential backoff
- Priority, round-robin, and least-latency strategies
- Canary rollouts with auto-promotion/rollback based on error rate and p95 latency
- Multi-region routing with cross-region fallback

### Rate Limiting & Load Shedding
- Per-tenant sliding window rate limits (requests/min, tokens/min)
- In-memory or Redis-backed for distributed deployments
- Load shedding with 3 priority tiers (high bypasses queue, low shed first at 80%)

### Caching & Cost
- Exact-match response caching with TTL and LRU eviction
- Semantic caching via embedding similarity (cosine threshold configurable)
- Cost optimization engine with model downgrade recommendations
- Budget enforcement (global, per-tenant, per-model) with alert/warn/block thresholds

### Request/Response Transformation
- PII stripping from responses (email, phone, SSN, credit card)
- Per-tenant system prompt injection and overrides
- Model aliasing (map friendly names to provider models)

### Observability
- OpenTelemetry traces with per-request spans
- Prometheus metrics at `/metrics`
- Real-time analytics with anomaly detection (static + statistical baseline)
- Structured JSON logging via Zap

### Kubernetes Operator
- 5 CRDs: Gateway, Provider, Route, Tenant, Policy
- Validation webhooks for all CRDs
- Multi-cluster federation (control plane + data plane)
