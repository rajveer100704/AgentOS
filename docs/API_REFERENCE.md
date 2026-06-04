# API Reference

AgentOS exposes two separate server ports:
1. **Gateway API (Port 8080)**: Handles proxying and routing requests from AI agents to underlying providers.
2. **Admin API (Port 8081)**: Manages configuration, observability metrics, human-in-the-loop approvals, and audit/evidence trails.

---

## Gateway API (port 8080)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Gateway health check |
| `POST` | `/v1/chat/completions` | Chat completions (supports streaming and non-streaming) |
| `GET` | `/v1/models` | List available models |
| `WS` | `/v1/ws` | WebSocket endpoint for persistent connection proxying |

---

## Admin API (port 8081)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Admin server health check |
| `GET` | `/metrics` | Prometheus metrics scrape target |
| `GET` | `/admin/v1/usage` | Usage metrics tracking per tenant |
| `GET` | `/admin/v1/providers` | Dynamic provider status and health check state |
| `GET` | `/admin/v1/tenants` | Tenant configuration summary |
| `GET` | `/admin/v1/policies` | Active runtime policy rules |
| `GET` | `/admin/v1/whoami` | Identifies caller role and tenant via client API key |
| `GET` | `/admin/v1/analytics` | Real-time traffic analytics and anomaly metrics |
| `GET` | `/admin/v1/alerts` | History of recent anomaly alerts |
| `POST` | `/admin/v1/alerts/{id}/acknowledge` | Mark an anomaly alert as acknowledged |
| `GET` | `/admin/v1/budgets` | Budget consumption statuses and forecasting |
| `GET` | `/admin/v1/cost-recommendations` | Cost optimization recommendations |
| `GET` | `/admin/v1/audit` | Query audit log (filter by actor, action type, tenant, etc.) |
| `POST` | `/admin/v1/audit/verify` | Verify overall audit chain block hash integrity |
| `POST` | `/admin/v1/graphql` | GraphQL admin interface |
| `GET` | `/admin/v1/approvals` | List all pending human-in-the-loop approvals |
| `POST` | `/admin/v1/approvals/{id}/approve` | Approve a pending action request |
| `POST` | `/admin/v1/approvals/{id}/deny` | Deny a pending action request |
| `GET` | `/admin/v1/evidence/sessions` | List active and archived evidence sessions |
| `GET` | `/admin/v1/evidence/sessions/{id}/export` | Export raw session evidence chain (JSON) |
| `GET` | `/admin/v1/evidence/sessions/{id}/report` | Generate human-readable Markdown evidence report |
| `GET` | `/admin/v1/evidence/sessions/{id}/report.html` | Generate HTML format evidence report |
| `POST` | `/admin/v1/evidence/sessions/{id}/verify` | Verify block hash integrity of a specific session's chain |
| `GET` | `/admin/v1/credentials` | List active dynamic, task-scoped credentials |
| `POST` | `/admin/v1/credentials/{id}/revoke` | Revoke a dynamic credential immediately |
| `GET` | `/admin/v1/manifests` | List active task manifests |
| `POST` | `/admin/v1/manifests` | Register a new task manifest |
| `GET` | `/admin/v1/manifests/{id}/drift` | Query task execution drift events for a manifest |
| `GET` | `/admin/v1/tickets` | List outstanding capability tickets |
| `GET` | `/admin/v1/sessions/{id}/risk` | Retrieve session cumulative behavioral risk score |
| `POST` | `/admin/v1/test-action` | Dry-run policy checks against a mock ActionEnvelope |
| `POST` | `/admin/v1/simulate` | Run a simulated policy validation with traces |
| `GET` | `/admin/v1/rollouts` | List canary rollouts |
| `GET` | `/admin/v1/health/detailed` | Detailed cluster health containing provider check results |
| `GET` | `/admin/v1/supply-chain` | Supply chain trust checks and code signature metrics |
