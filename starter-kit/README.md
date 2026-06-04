# AgentOS Governed Coding Agent Starter Kit

[![CI](https://github.com/saivedant169/AgentOS/actions/workflows/ci.yaml/badge.svg)](https://github.com/saivedant169/AgentOS/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/saivedant169/AgentOS)](https://goreportcard.com/report/github.com/saivedant169/AgentOS)
[![Go Reference](https://pkg.go.dev/badge/github.com/saivedant169/AgentOS.svg)](https://pkg.go.dev/github.com/saivedant169/AgentOS)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](../LICENSE)
[![Docker](https://img.shields.io/docker/pulls/saivedant169/AgentOS)](https://hub.docker.com/r/saivedant169/AgentOS)

AgentOS sits between your coding agent and the tools it uses. Every tool call passes through AgentOS, which decides: **allow**, **review** (human approval), or **block**. You get least-privilege enforcement, an approval workflow, and a tamper-evident audit trail -- without changing your agent or your tools.

This starter kit gets you from zero to governed agents in 15 minutes.

---

## Prerequisites

You need one of:
- **Docker** (recommended) -- Docker 20.10+ and Docker Compose v2
- **Go 1.26.2+** -- if you want to build from source
- **Node.js 18+** -- only if you want to run the mock MCP server for testing

## 15-Minute Quickstart

### Step 1: Install

```bash
cd starter-kit
./install.sh
```

This builds AgentOS, copies your chosen policy pack, starts the service, and verifies it's healthy.

### Step 2: Pick a policy pack

Four policy packs are included in `policies/`. Copy one to configure your governance posture:

| Pack | Use when... |
|------|------------|
| `readonly.yaml` | Agent should only read. No writes, no deletes. Audits and investigations. |
| `pr-writer.yaml` | Agent writes code and opens PRs. Destructive ops blocked, writes reviewed. |
| `infra-review.yaml` | Agent does infrastructure work. Everything destructive needs human approval. |
| `sql-explorer.yaml` | BI/data agent explores a database read-first. SELECT allowed, writes reviewed, destructive SQL blocked. See `sql-explorer-tuning-notes.md`. |

```bash
# Example: use the pr-writer policy
cp policies/pr-writer.yaml ../configs/AgentOS-policy.yaml
```

### Step 3: Connect your editor

See the `editors/` directory for setup guides:
- [Claude Code](editors/claude-code.md)
- [Cursor](editors/cursor.md)
- [MCP JSON template](editors/mcp.json.template)

### Step 4: Test it works

```bash
# Health check
curl http://localhost:8080/health

# Should return: allow
curl -s -X POST http://localhost:8081/admin/v1/test-action \
  -H "Content-Type: application/json" \
  -H "X-API-Key: starter-key-001" \
  -d '{"protocol":"git","tool":"github.list_repos","target":"myorg/myrepo","capability":"read"}' | jq .decision

# Should return: block
curl -s -X POST http://localhost:8081/admin/v1/test-action \
  -H "Content-Type: application/json" \
  -H "X-API-Key: starter-key-001" \
  -d '{"protocol":"shell","tool":"shell.rm","target":"/","capability":"delete"}' | jq .decision
```

### Step 5: Run the efficacy tests

```bash
./tests/run-efficacy-tests.sh
```

This sends 20 attack scenarios and verifies AgentOS blocks or flags each one.

---

## Architecture

```
+------------------+       +----------------------------------+       +----------------+
|                  |       |           AgentOS              |       |                |
|  Coding Agent    |       |                                  |       |  GitHub API    |
|  (Claude Code,   | ----> |  Policy Engine                   | ----> |  Shell / CLI   |
|   Cursor, etc.)  |       |    allow / review / block        |       |  PostgreSQL    |
|                  | <---- |  Credential Broker               | <---- |  HTTP APIs     |
|  MCP Client      |       |    (task-scoped, short-lived)    |       |  Cloud APIs    |
|                  |       |  Evidence Chain                   |       |                |
|                  |       |    (SHA-256, hash-linked)         |       |                |
+------------------+       +----------------------------------+       +----------------+
```

1. Agent sends a tool call (MCP, shell, SQL, HTTP)
2. AgentOS normalizes it into an `ActionEnvelope`
3. Policy engine evaluates rules and returns allow/review/block
4. If review: action enters the approval queue; operator approves or denies
5. If allowed: AgentOS issues task-scoped credentials and executes
6. Result recorded in the tamper-evident evidence chain

---

## Policy Customization

### Policy file structure

Each policy file defines rules evaluated top-to-bottom. First match wins.

```yaml
tool_policies:
  enabled: true
  default_decision: "block"    # block | review | allow
  rules:
    - protocol: "git"          # git | shell | sql | http | mcp | *
      tool: "github.list_*"    # tool name, supports * wildcards
      target: "myorg/*"        # optional: scope to specific targets
      capability: "read"       # optional: read | write | delete | execute
      decision: "allow"        # allow | review | block
```

### Writing your own rules

Start from the closest policy pack and modify:

```bash
cp policies/pr-writer.yaml my-policy.yaml
```

Common customizations:
- **Allow a specific tool**: Add a rule with `decision: "allow"` above the default block
- **Block a tool for a specific target**: Add a rule with both `tool` and `target`
- **Review all writes to production**: Add `target: "production*"` with `decision: "review"`
- **Change the default**: Set `default_decision` to `review` instead of `block` for a more permissive baseline

### Testing policy changes

After editing, restart AgentOS and test:

```bash
# Restart
docker compose restart AgentOS

# Test your specific rule
curl -s -X POST http://localhost:8081/admin/v1/test-action \
  -H "Content-Type: application/json" \
  -H "X-API-Key: starter-key-001" \
  -d '{"protocol":"shell","tool":"shell.kubectl","target":"production","capability":"execute"}' | jq .
```

---

## Approval Workflow

When a policy evaluates to `review`, the action is queued for human approval.

### Listing pending approvals

```bash
curl -s http://localhost:8081/admin/v1/approvals \
  -H "X-API-Key: starter-key-001" | jq .
```

### Approving an action

```bash
curl -s -X POST http://localhost:8081/admin/v1/approvals/{envelope_id}/approve \
  -H "Content-Type: application/json" \
  -H "X-API-Key: starter-key-001" \
  -d '{"reviewer":"your-name","comment":"Looks good"}' | jq .
```

### Denying an action

```bash
curl -s -X POST http://localhost:8081/admin/v1/approvals/{envelope_id}/deny \
  -H "Content-Type: application/json" \
  -H "X-API-Key: starter-key-001" \
  -d '{"reviewer":"your-name","comment":"Too broad, scope it down"}' | jq .
```

### CLI alternative

```bash
agentctl approve <envelope_id> --reviewer your-name --comment "Approved"
agentctl deny <envelope_id> --reviewer your-name --comment "Denied"
```

---

## Evidence Export

Every action -- allowed, reviewed, or blocked -- is recorded in a tamper-evident hash chain.

### Export a session's evidence

```bash
# List sessions
curl -s http://localhost:8081/admin/v1/evidence/sessions \
  -H "X-API-Key: starter-key-001" | jq .

# Export a session as a JSON bundle
curl -s http://localhost:8081/admin/v1/evidence/sessions/{session_id}/export \
  -H "X-API-Key: starter-key-001" > evidence-bundle.json
```

### Verify chain integrity

```bash
# Via API
curl -s -X POST http://localhost:8081/admin/v1/evidence/sessions/{session_id}/verify \
  -H "X-API-Key: starter-key-001" | jq .

# Via CLI
agentctl evidence verify --session {session_id}
```

### What's in an evidence bundle

See `examples/sample-evidence-bundle.json` for a complete example. Each bundle contains:
- Session metadata (ID, start/end time, agent identity)
- Ordered list of ActionEnvelopes with policy decisions
- Approval records (reviewer, timestamp, comment)
- SHA-256 hash chain linking each record to the previous
- Session manifest hash for tamper detection

---

## Troubleshooting

### AgentOS won't start

```bash
# Check logs
docker compose logs AgentOS

# Validate config syntax
docker compose exec AgentOS ./AgentOS --config configs/AgentOS.yaml --validate
```

### Agent can't connect

1. Verify AgentOS is running: `curl http://localhost:8080/health`
2. Verify MCP gateway is running: `curl http://localhost:8082/health`
3. Check the MCP bridge script path in `.mcp.json` is absolute
4. Check the agent's MCP configuration points to the right bridge script

### Policy not taking effect

1. Rules are evaluated top-to-bottom, first match wins
2. Check your rule's `protocol` and `tool` fields match exactly (wildcards use `*`)
3. Restart AgentOS after config changes -- policies are loaded at startup
4. Use the test-action endpoint to debug: `POST /admin/v1/test-action`

### Approval queue not working

1. Verify the action's policy evaluates to `review` (not `allow` or `block`)
2. Check pending approvals: `GET /admin/v1/approvals`
3. Ensure you're using an API key with `admin` or `operator` role to approve/deny

### Evidence chain verification fails

This means the chain has been tampered with. Possible causes:
- Database was manually edited
- AgentOS was restarted with a different evidence store
- Bug report: file an issue with the verification output

---

## Production Deployment

See the `deploy/` directory for production-ready configurations:
- `deploy/docker-compose.yaml` -- Docker Compose with PostgreSQL and Redis
- `deploy/helm/` -- Helm chart for Kubernetes
- `deploy/terraform/` -- Terraform module for AWS ECS

---

## Further Reading

- [Main AgentOS README](../README.md) -- full feature documentation
- [Configuration reference](../configs/AgentOS.example.yaml) -- all config options
- [Policy packs](../configs/policy-packs/) -- additional policy templates
