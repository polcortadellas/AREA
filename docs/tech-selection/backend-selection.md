# Weighted decision matrix: <decision, e.g. back-end>

## 1. Criteria and weights (locked before scoring)

Weight scale: **3** = decisive, **2** = important, **1** = nice to have.
Score per option: **1** = very poor, **5** = excellent.

| # | Criterion                       | Weight (1-3) | What "5" means for us                    |
|---|---------------------------------|:------------:|------------------------------------------|
| 1 | Security                        |      3       | Safe defaults, mature CVE track record   |
| 2 | OAuth2 / OIDC support           |      3       | Audited libraries, no hand-rolled auth   |
| 3 | OpenAPI generation              |      3       | Spec from code, testable against the implementation |
| 4 | Concurrency (hooks)             |      2       | Many tasks in parallel without blocking  |
| 5 | Team learning curve             |      2       | Everyone productive in under 1 week      |
| 6 | Ecosystem and libraries         |      2       | Queues, cron and ORM covered without inventing them |
| 7 | Docker deployment               |      2       | Small image, reproducible build          |
| 8 | Typing and maintainability      |      1       | Errors caught before runtime             |

Maximum possible weighted score: Σweights × 5 = **___**

## 2. Individual scoring (member: ______)

Done without discussing with anyone. Total = Σ(weight × score).


Option A: Go, ~20 mins/ 0 experience before. Details: [`../../spikes/go/SPIKE-ping.md`](../../spikes/go/SPIKE-ping.md)

Option B: NestJS ~15 mins (AI-assisted) / 0 NestJS experience before, some JS / Next.js. Details: [`../../spikes/nestjs/SPIKE-ping.md`](../../spikes/nestjs/SPIKE-ping.md)

Option C: Hono ~15 mins / 0 Hono/Bun experience before, some JS / TS. Details: [`../../spikes/hono_bun/SPIKE.md`](../../spikes/hono_bun/SPIKE.md)

Option D: FastAPI ~3 mins (AI-assisted) / 0 FastAPI experience before, some Python. Details: [`../../spikes/fastapi/SPIKE.md`](../../spikes/fastapi/SPIKE.md)

| Criterion                       | Weight | Option A: Go | Option B: NestJS | Option C: Hono/Bun | Option D: FastAPI |
|---------------------------------|:------:|:------------:|:----------------:|:------------------:|:-----------------:|
| Security                        |   3    |      4       |        3         |         3          |         4         |
| OAuth2 / OIDC support           |   3    |      4       |        4         |         4          |         4         |
| OpenAPI generation              |   3    |      4       |        5         |         4          |         5         |
| Concurrency (hooks)             |   2    |      5       |        3         |         4          |         3         |
| Team learning curve             |   2    |      2       |        4         |         5          |         4         |
| Ecosystem and libraries         |   2    |      3       |        5         |         4          |         5         |
| Docker deployment               |   2    |      5       |        4         |         5          |         3         |
| Typing and maintainability      |   1    |      5       |        4         |         4          |         3         |
| **Weighted total**              |        |    **71**    |      **72**      |       **73**       |      **72**       |

**What has and has not been spiked:** the Go spike only covers `/ping`: it has no Dockerfile and no test, while NestJS, Hono/Bun and FastAPI do. [`spikes/DISTANCE.md`](../../spikes/DISTANCE.md) explains how this affects the comparison.

## 3. Team comparison

Flag every criterion where two members differ by **more than 2 points**. Each divergence is a finding, not a rounding error.

| Criterion | Member 1 | Member 2 | Member 3 | Difference > 2 | Cause of divergence |
|-----------|:--------:|:--------:|:--------:|:--------------:|---------------------|
|           |          |          |          |                |                     |

> Summary for `RECOMMENDATION.md` (250-300 words).

## 4. Final Choice & Practical Evidence

* **Selected Technology:** Go (standard library)
* **Justification (250–300 words):**

Go's matrix total (71) is within two points of NestJS (72), FastAPI (72) and Hono/Bun (73), so the totals alone do not decide. We choose Go because of what AREA's server mostly does: run many hooks in parallel and stay small and predictable. Go scored 5 on concurrency (goroutines suit polling many services at once, while NestJS's in-process scheduler needed a manual overlap guard), 5 on Docker (the Go image in the SQLite spike is 21 MB, non-root, static binary, against 401 MB for NestJS and 214 MB for FastAPI) and 5 on typing. The standard library covers the HTTP server, and few dependencies mean a smaller supply-chain surface than the ~740 npm packages of the NestJS spike. The server keeps all business logic, so the web and mobile clients only depend on its OpenAPI description, not on its language.

The weak point is the learning curve (2): nobody on the team knew Go before. The ping service took about 20 minutes and compiled at the first attempt, which is encouraging but is not proof for a whole project. Two risks were not tested in the spikes: generating the OpenAPI description from Go code (we would probably write it spec-first and generate types from it) and OAuth2 / OIDC with an audited library. They must be checked first in the next milestone, and we keep FastAPI as the fallback if either fails.

Expected cost: more time to learn the language than the TypeScript options, accepted in exchange for the concurrency and deployment benefits.

* **Practical Evidence / Spike Reference:** [`spikes/go/SPIKE-ping.md`](../../spikes/go/SPIKE-ping.md), [`spikes/sqlite_vue/backend`](../../spikes/sqlite_vue/backend) (Go API with tests, 21 MB image), [`spikes/nestjs/SPIKE.md`](../../spikes/nestjs/SPIKE.md), [`spikes/fastapi/SPIKE.md`](../../spikes/fastapi/SPIKE.md), [`spikes/hono_bun/SPIKE.md`](../../spikes/hono_bun/SPIKE.md)
