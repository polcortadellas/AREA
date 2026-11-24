# Architecture decision records

| ADR | Decision | Status | Date |
|---|---|---|---|
| [0001](0001-backend-stack-go.md) | Go for the server | Accepted | 2026-11-15 |
| [0002](0002-persistence-postgresql.md) | PostgreSQL for persistence | Accepted | 2026-11-15 |
| [0003](0003-workflow-engine-broker.md) | Modular monolith, transactional outbox and Redis Streams | Accepted | 2026-11-15 |
| [0004](0004-typed-nodes.md) | Typed connector nodes and a start-up registry | Accepted | 2026-11-15 |
| [0005](0005-react-react-native.md) | React for the web, React Native for Android | Accepted | 2026-11-15 |

Each record says why it deserved a record, then gives the context, the decision, the real alternatives and why they were rejected for this project, the accepted consequences, the evidence in the code, and what the product would look like under the main alternative.

A new decision takes the next number. A record is not rewritten when its decision changes: the new record says which one it supersedes, and the old one gets `Status: Superseded by ADR 00XX`.
