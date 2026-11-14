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

## 2. Individual scoring (member: Pol)

Done without discussing with anyone. Total = Σ(weight × score).

* **Option A:** ______________________ (Details: `spikes/...`)
* **Option B:** SQLite (Details: `../../spikes/sqlite_vue/spike-vue-sqlite.md`)
* **Option C:** ______________________ (Details: `spikes/...`)
* **Option D:** ______________________ (Details: `spikes/...`)

| Criterion | Weight | Option A | Option B | Option C | Option D |
|-----------|:------:|:--------:|:--------:|:--------:|:--------:|
| Security & Encryption at Rest (`SEC-CRYPTO`) | 3 |  | 2 |  |  |
| Hybrid Data Model (Relational + Dynamic JSON) | 3 |  | 3 |  |  |
| ACID Transactions & Concurrency | 3 |  | 3 |  |  |
| Backup & Restore Reliability (`SYS-BACKUP`) | 2 |  | 4 |  |  |
| Backend Ecosystem & Driver Maturity | 2 |  | 3 |  |  |
| Operational Footprint in Docker | 1 |  | 5 |  |  |
| **Weighted total** | |  | **43** |  |  |

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

* **Selected Technology:** ______________________
* **Justification (250–300 words):**
* **Practical Evidence / Spike Reference:**