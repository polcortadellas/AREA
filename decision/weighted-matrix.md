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

| Criterion                       | Weight | Option A: Go | Option B | Option C | Option D |
|---------------------------------|:------:|:------------:|:--------:|:--------:|:--------:|
| Security                        |   3    |      4       |          |          |          |
| OAuth2 / OIDC support           |   3    |      4       |          |          |          |
| OpenAPI generation              |   3    |      4       |          |          |          |
| Concurrency (hooks)             |   2    |      5       |          |          |          |
| Team learning curve             |   2    |      2       |          |          |          |
| Ecosystem and libraries         |   2    |      3       |          |          |          |
| Docker deployment               |   2    |      5       |          |          |          |
| Typing and maintainability      |   1    |      5       |          |          |          |
| **Weighted total**              |        |    **71**    |          |          |          |

## 3. Team comparison

Flag every criterion where two members differ by **more than 2 points**. Each divergence is a finding, not a rounding error.

| Criterion | Member 1 | Member 2 | Member 3 | Difference > 2 | Cause of divergence |
|-----------|:--------:|:--------:|:--------:|:--------------:|---------------------|
|           |          |          |          |                |                     |
    
> Summary for `decision/RECOMMENDATION.md` (250-300 words).