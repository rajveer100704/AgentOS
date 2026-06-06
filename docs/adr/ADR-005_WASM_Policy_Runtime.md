# ADR-005: WebAssembly (WASM) Policy Runtime

## Status
Accepted

## Context
Autonomous coding agents execute complex actions that must be governed at a fine-grained level. Basic rule engines (regex patterns, keyword lists, PII filters) are fast but lack the flexibility to evaluate complex stateful constraints (e.g., parsing a SQL AST to block table drops, or analyzing file content changes). 

Embedding raw Go code within the gateway binary requires custom recompilation, while executing dynamic scripts (e.g., JavaScript or Lua) inside the Go process introduces security sandboxing risks, garbage collection pauses, and performance degradation. We need a performant, language-agnostic, and secure sandboxing technology to execute custom, user-authored policy extensions.

## Decision
Adopt WebAssembly (WASM) as the core extensibility mechanism for custom, user-defined runtime policies.

*   **Runtime Engine:** Use `github.com/tetratelabs/wazero` as the WebAssembly runtime engine. Because it is a zero-dependency, pure-Go WASM compiler and interpreter, it integrates seamlessly into the AgentOS binary and supports frictionless compilation across Windows, macOS, and Linux without requiring CGO.
*   **Host-Guest SDK:** Define a strict guest-host boundary (ABI). AgentOS serializes the execution request (`ActionEnvelope`) to guest memory, invokes the exported `evaluate` function, and retrieves a structured decision (Allow, Block, or Flag for Human-in-the-Loop review).
*   **Zero-Trust Isolation:** Disable all WASI capabilities (no filesystem access, no system environment variables, and no raw network sockets) to prevent guest policies from executing arbitrary actions or leaking secrets.

## Consequences
*   **Safety:** Guarantees absolute isolation. Even if a custom policy is malicious or contains bugs, it cannot crash the AgentOS host process or access host resources.
*   **Performance:** Near-native execution speed. Using `wazero`'s JIT compiler, the latency overhead added by WASM policy evaluation is minimal (~1.2–1.5 μs in microbenchmarks).
*   **Developer Experience:** Authors must compile their policy logic to `.wasm` files. To simplify this, AgentOS provides Go and Rust SDK templates in the `examples/` directory.
