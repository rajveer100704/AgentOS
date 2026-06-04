# ADR-004: Task-Scoped Dynamic Credentials

## Status
Accepted

## Context
Providing autonomous AI agents with long-lived, high-privilege static credentials (such as GitHub OAuth tokens, primary AWS access keys, or database root passwords) is a severe security liability. If the agent's environment is compromised via prompt injection, or if the agent executes out-of-scope code, these long-lived credentials can be exfiltrated or abused, leading to unauthorized resource modification or data leaks. We need a way to restrict the agent's credentials to the exact, minimum permissions required for a specific task and ensure they expire automatically.

## Decision
Adopt runtime credential brokering at the execution boundary. Instead of passing long-lived credentials directly to the agent runtime, the Control Plane integrates with external identity providers and credential stores (e.g., GitHub Apps, AWS STS, HashiCorp Vault) to dynamically mint short-lived, task-scoped credentials:

*   **GitHub Integration:** Generate short-lived GitHub App Installation Access Tokens (valid for maximum 1 hour) restricted to specific repositories and permissions required for the PR.
*   **Cloud Infrastructure:** Issue temporary AWS STS session tokens or Google Cloud IAM Service Account credentials bound to a specific session.
*   **Database Interception:** Inject dynamic database roles (via HashiCorp Vault) that automatically expire or get revoked after the task completes.
*   **Evidence Binding:** Bind each issued credential's cryptographic fingerprint or reference ID directly to the `ActionEnvelope` and the tamper-evident evidence chain for audit purposes.

## Consequences
*   **Security:** Drastically reduces the attack surface. An exfiltrated credential is only valid for a specific task, restricted to minimal resources, and expires within a short window (e.g., 15–60 minutes).
*   **Auditability:** Guarantees a clear audit trail that connects tool execution, policy decisions, and credential usage to a specific agent session.
*   **Complexity:** Adds runtime dependency overhead. The Control Plane must manage connection parameters, OAuth applications, and API keys for the broker providers, and handle renewal and revocation lifecycles.
