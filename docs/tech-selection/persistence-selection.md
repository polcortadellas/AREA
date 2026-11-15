# Weighted decision matrix: Persistence stack

## 1. Criteria and weights
(locked before scoring)

Weight scale: **3** = decisive, **2** = important, **1** = nice to have.  
Score per option: **1** = very poor, **5** = excellent.

| # | Criterion | Weight (1-3) | What "5" means for us |
|---|-----------|:------------:|-----------------------|
| 1 | Security & Encryption at Rest (`SEC-CRYPTO`) | 3 | Native binary types for AES-GCM ciphertexts, role isolation, TLS connections, strong CVE record |
| 2 | Hybrid Data Model (Relational + Dynamic JSON) | 3 | Strict foreign keys for users/areas plus indexed JSON queries for dynamic service configurations |
| 3 | ACID Transactions & Concurrency | 3 | Full transactional guarantees for workflow state updates, deduplication, and idempotency keys |
| 4 | Backup & Restore Reliability (`SYS-BACKUP`) | 2 | Fast, single-command consistent dumps and automated restore scripts |
| 5 | Backend Ecosystem & Driver Maturity | 2 | High-performance, battle-tested drivers and migration tools |
| 6 | Operational Footprint in Docker | 1 | Official lightweight image, fast startup, straightforward health checks (`SYS-HEALTH`) |

Maximum possible weighted score: Σweights × 5 = **70**

## 2. Individual scoring (member: ______)

Done without discussing with anyone. Total = Σ(weight × score).

Option A: PostgreSQL 16/17. Details: [`../../spikes/nestjs/SPIKE.md`](../../spikes/nestjs/SPIKE.md#4-persistence-postgresql--typeorm-strengths-and-weaknesses)

Option B: SQLite. Details: [`../../spikes/sqlite_vue/spike-vue-sqlite.md`](../../spikes/sqlite_vue/spike-vue-sqlite.md)

Option C: MongoDB 8.0. Details: [`../../spikes/mongodb/SPIKE.md`](../../spikes/mongodb/SPIKE.md)

| Criterion                                      | Weight | Option A: PostgreSQL | Option B: SQLite | Option C: MongoDB |
|------------------------------------------------|:------:|:--------------------:|:----------------:|:-----------------:|
| Security & Encryption at Rest (`SEC-CRYPTO`)   |   3    |          4           |        2         |         3         |
| Hybrid Data Model (Relational + Dynamic JSON)  |   3    |          5           |        3         |         2         |
| ACID Transactions & Concurrency                |   3    |          5           |        3         |         3         |
| Backup & Restore Reliability (`SYS-BACKUP`)    |   2    |          5           |        4         |         4         |
| Backend Ecosystem & Driver Maturity            |   2    |          4           |        3         |         4         |
| Operational Footprint in Docker                |   1    |          4           |        5         |         2         |
| **Weighted total**                             |        |        **64**        |      **43**      |      **42**       |

**What has and has not been spiked**

- PostgreSQL: tested through TypeORM in the NestJS spike (see its SPIKE.md) and a Bun driver in the Hono/Bun compose file (connection not verified). It has **not** been tested with a Go driver.
- SQLite: tested with Go in the SQLite + Vue spike. MongoDB: tested with pymongo.
- Redis Streams, the message broker chosen in [ADR 0003](../adr/0003-workflow-engine-broker.md), is not a persistence option in this matrix and **has not been validated in a spike yet**. The validation (stop and restart a consumer with `XAUTOCLAIM`) is part of task M-11.

## 3. Team comparison

Flag every criterion where two members differ by **more than 2 points**. Each divergence is a finding, not a rounding error.

| Criterion | Pol | Achille | Keigan | Edu | Difference > 2 | Cause of divergence |
|-----------|:---:|:-------:|:------:|:---:|:--------------:|---------------------|
| Security & Encryption at Rest (`SEC-CRYPTO`) | | | | | | |
| Hybrid Data Model (Relational + Dynamic JSON) | | | | | | |
| ACID Transactions & Concurrency | | | | | | |
| Backup & Restore Reliability (`SYS-BACKUP`) | | | | | | |
| Backend Ecosystem & Driver Maturity | | | | | | |
| Operational Footprint in Docker | | | | | | |

## 4. Final Choice & Practical Evidence

* **Selected Technology:** PostgreSQL
* **Justification (250–300 words):**

PostgreSQL scored 64 out of 70, against 43 for SQLite and 42 for MongoDB, and the gap comes from the three criteria with weight 3. The hybrid model: AREA needs foreign keys between users, linked identities and workflows, plus flexible JSON for each service's configuration, and PostgreSQL gives both (`jsonb`). MongoDB accepted a workflow owned by a user that does not exist, because it has no foreign keys. ACID: PostgreSQL has real transactions, unique constraints for hook deduplication and `SELECT ... FOR UPDATE SKIP LOCKED` for job queues. MongoDB refused multi-document transactions on a default container (it needs a replica set), and SQLite allows a single writer, so the SQLite spike limited the pool to one connection. SQLite is also a file on one machine, which blocks running several server instances. Security: no option encrypts at rest by default, so SEC-CRYPTO will be done in the application with AES-GCM ciphertexts stored in `bytea` columns.

Operationally, `pg_dump` and `pg_restore` are single commands, and the official `postgres:alpine` image starts quickly with a `pg_isready` healthcheck and a named volume, as shown in the NestJS and Hono compose files. Data survived restarts in the NestJS spike, and the same volume will serve backups.

Limits of the evidence: PostgreSQL was tested through TypeORM (NestJS) and a Bun driver, not yet with the Go driver, and `jsonb` columns were not exercised in any spike. Schema changes will need a migration tool, since TypeORM's `synchronize` is only for spikes. These are the first things to verify with the Go server.

* **Practical Evidence / Spike Reference:** [`spikes/nestjs/SPIKE.md`](../../spikes/nestjs/SPIKE.md#4-persistence-postgresql--typeorm-strengths-and-weaknesses), [`spikes/hono_bun/docker-compose.yml`](../../spikes/hono_bun/docker-compose.yml), [`spikes/sqlite_vue/spike-vue-sqlite.md`](../../spikes/sqlite_vue/spike-vue-sqlite.md), [`spikes/mongodb/SPIKE.md`](../../spikes/mongodb/SPIKE.md)
