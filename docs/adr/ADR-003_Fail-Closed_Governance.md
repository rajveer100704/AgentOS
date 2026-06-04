# ADR-003: Fail-Closed Policy Boundary

## Status
Accepted

## Context
When securing autonomous AI agent pipelines, system components (such as database connections, external policy validation webhooks, or the WASM JIT compiler) can fail or time out. Under such exceptions, we must decide whether to let the agent transaction pass unchecked (fail-open, prioritizing availability) or block the transaction (fail-closed, prioritizing security).

## Decision
Enforce a strict **fail-closed** governance architecture. 

*   If any input/output policy evaluation fails to execute, compiles incorrectly, or experiences a timeout, the request is immediately aborted.
*   Return a structured HTTP `503 Service Unavailable` or `500 Internal Server Error` to the client.
*   Log the exact stack trace and compilation error details in the tamper-evident audit ledger to prevent silent failures.
*   Provide a local developer override flag (`fail_open: true`) strictly restricted to offline development environments.

## Consequences
*   **Security:** Guarantees that an agent can never execute a transaction without undergoing full policy checks, preventing bypass attacks during system faults.
*   **Availability:** Availability is directly tied to policy engine stability. Webhook deadlocks or database connection failures will manifest as blocked agent operations, requiring tight bounds on upstream timeout parameters.
