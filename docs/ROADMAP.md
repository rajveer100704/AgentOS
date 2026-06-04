# AgentOS Project Roadmap

## Completed Milestones

- [x] **Phase 1-4**: Full AI gateway implementation featuring model routing, response caching, policy filtering, RBAC, multi-cluster federation, and a Kubernetes operator.
- [x] **Phase 5**: Semantic caching using embedding vector similarity, cost optimization suggestions, custom request/response text transforms, priority-queue-based load shedding, WebSocket proxying, GraphQL admin console, and WebAssembly (WASM) policy SDK.
- [x] **Phase 6**: MCP remote gateway supporting JSON-RPC routing, dynamic tool allowlisting/denylisting, operator review execution path, and queue storage.
- [x] **Phase 7**: Task-scoped credential broker enabling temporary Github App JWTs, dynamic AWS STS SigV4 credentials, Vault database credentials, and dynamic credential metadata injection into the evidence chain.
- [x] **Phase 8**: Evidence export and verification tooling CLI (`agentctl verify`, `agentctl evidence`) with three tuned policy packs for coding agents.

## Enterprise-Grade Capabilities

- [x] **Tier 1 (Core Integrity)**: Typed resource model checks, TaskManifest definition, manifest execution drift detection, cryptographic capability tickets, policy dry-runs and diffs, sandbox execution helpers, and formatted human-auditable evidence.
- [x] **Tier 2 (Integrations & Compliance)**: Behavioral session rules, Slack Block Kit & GitHub PR approval webhooks, enterprise identity tracking, separation of duties compliance, and signed policy supply-chain keys.
- [x] **Tier 3 (Reliability)**: HA setup, automated db recovery/retention strategies, threat modeling docs, and security guides.

## Runtime & Framework Integrations

- [x] **Notifications**: Alert systems triggering automatically on approval states or Slack/GitHub reviews.
- [x] **Kill Switch**: Sessions automatically blocked when risk thresholds are exceeded.
- [x] **Drift Enforcement**: Option to switch from `warn` to `enforce` mode to actively block out-of-scope actions.
- [x] **Formatted Reports**: Auto-rendering Markdown and HTML reports suitable for security officers and auditors.

## Adoption & Tools

- [x] **Phase 9**: Governed Coding Agent Starter Kit including 3 tuned policy packs, developer editor integrations, Docker/Helm/Terraform templates, test harness, and sample outputs.
- [x] **Phase 10**: Fully documented PR-writer scenario walkthrough, quick installation scripts, and design-partner onboarding material.
