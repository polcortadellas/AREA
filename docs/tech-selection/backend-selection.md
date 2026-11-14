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


Option A: Go, ~20 mins/ 0 experience before

Option B: NestJS ~15 mins (AI-assisted) / 0 NestJS experience before, some JS / Next.js. Details: [`../spikes/nestjs/SPIKE-ping.md`](../spikes/nestjs/SPIKE-ping.md)

Option C: Hono ~15 mins / 0 Hono/Bun experience before, some JS / TS. Details: [`spikes/hono_bun/README.md`](../spikes/hono_bun/README.md)

| Criterion                       | Weight | Option A: Go | Option B: NestJS | Option C: Hono/Bun | Option D |
|---------------------------------|:------:|:------------:|:----------------:|:------------------:|:--------:|
| Security                        |   3    |      4       |        3         |         3          |          |
| OAuth2 / OIDC support           |   3    |      4       |        4         |         4          |          |
| OpenAPI generation              |   3    |      4       |        5         |         4          |          |
| Concurrency (hooks)             |   2    |      5       |        3         |         4          |          |
| Team learning curve             |   2    |      2       |        4         |         5          |          |
| Ecosystem and libraries         |   2    |      3       |        5         |         4          |          |
| Docker deployment               |   2    |      5       |        4         |         5          |          |
| Typing and maintainability      |   1    |      5       |        4         |         4          |          |
| **Weighted total**              |        |    **71**    |      **72**      |       **73**       |          |

## 3. Team comparison

Flag every criterion where two members differ by **more than 2 points**. Each divergence is a finding, not a rounding error.

| Criterion | Member 1 | Member 2 | Member 3 | Difference > 2 | Cause of divergence |
|-----------|:--------:|:--------:|:--------:|:--------------:|---------------------|
|           |          |          |          |                |                     |

> Summary for `RECOMMENDATION.md` (250-300 words).