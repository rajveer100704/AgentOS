# ADR-006: Evidence Chain Storage Model

## Status
Accepted

## Context
AgentOS generates a continuous stream of logs, policy decisions, and approval manifests during agent sessions. In high-security environments, these records serve as legal and operational "evidence" of what the agent did. 

If the storage layer is compromised, an attacker (or a compromised agent with shell access) could attempt to delete, alter, or inject entries into these logs to cover their tracks. Standard relational databases (like PostgreSQL) or flat filesystems are mutable by default and do not provide built-in cryptographic proof of integrity. We need a storage model that guarantees immutability, tamper-detection, and historical integrity for the audit logs.

## Decision
Implement an append-only **Evidence Chain** utilizing cryptographic hash-chaining:

*   **Cryptographic Linkage:** Each audit record is serialized into a standardized JSON format (`EvidenceEnvelope`) containing the payload, a timestamp, and a `PrevHash` field. The `PrevHash` is the SHA-256 hash of the immediately preceding `EvidenceEnvelope` in the session.
*   **Tamper-Evidence:** The head of the chain (the latest envelope's hash) is tracked in-memory and exposed to clients. To verify that no logs were modified, deleted, or inserted retrospectively, the verifier recalculates the hashes starting from the genesis envelope. Any mutation breaks the chain's hash matching.
*   **Storage Independence:** Serve evidence logs as flat, append-only JSONL files. These files are optimized for streaming to high-durability, immutable storage targets such as AWS S3 Object Lock or Google Cloud Storage WORM (Write-Once-Read-Many) buckets. Relational databases (e.g. PostgreSQL) are utilized purely for indexing and search queries, never as the source of truth for audit integrity.

## Consequences
*   **Immutability:** Provides mathematical guarantees of log integrity. If a log file is modified on disk, the next verification run will fail immediately.
*   **Efficiency:** Calculating SHA-256 hashes adds negligible overhead (~1.3 μs per envelope) compared to database write transactions.
*   **Query Overhead:** Finding specific past actions in a raw hash chain requires scanning. To optimize this, AgentOS maintains asynchronous write-ahead indices to a relational database, separating the *storage integrity layer* from the *query interface layer*.
