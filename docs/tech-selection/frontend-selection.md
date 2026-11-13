# Weighted decision matrix: Front-end stack

## 1. Criteria and weights
(locked before scoring)

Weight scale: **3** = decisive, **2** = important, **1** = nice to have.  
Score per option: **1** = very poor, **5** = excellent.

| # | Criterion | Weight (1-3) | What "5" means for us |
|---|-----------|:------------:|-----------------------|
| 1 | Security (XSS, CSRF & token storage) | 3 | Auto-escaping templates, safe HTTP-only cookie/token handling, minimal dependency attack surface |
| 2 | Dynamic form rendering (`/about.json`) | 3 | Effortless reactive binding to generate inputs and condition builders from dynamic JSON schemas |
| 3 | Accessibility (WCAG AA) & i18n | 3 | Semantic HTML control, mature i18n plugin, easy keyboard navigation and screen-reader support |
| 4 | Team learning curve | 2 | Everyone productive in under 1 week with clear conventions |
| 5 | OpenAPI / TypeScript integration | 2 | Clean typed API client generation and strict compile-time checks |
| 6 | Docker build & bundle footprint | 1 | Fast static build served cleanly via lightweight container |

Maximum possible weighted score: Σweights × 5 = **70**

## 2. Individual scoring (member: ______)

Done without discussing with anyone. Total = Σ(weight × score).

* **Option A:** ______________________ (Details: `spikes/...`)
* **Option B:** ______________________ (Details: `spikes/...`)
* **Option C:** ______________________ (Details: `spikes/...`)
* **Option D:** ______________________ (Details: `spikes/...`)

| Criterion | Weight | Option A | Option B | Option C | Option D |
|-----------|:------:|:--------:|:--------:|:--------:|:--------:|
| Security (XSS, CSRF & token storage) | 3 | | | | |
| Dynamic form rendering (`/about.json`) | 3 | | | | |
| Accessibility (WCAG AA) & i18n | 3 | | | | |
| Team learning curve | 2 | | | | |
| OpenAPI / TypeScript integration | 2 | | | | |
| Docker build & bundle footprint | 1 | | | | |
| **Weighted total** | | | | | |

## 3. Team comparison

Flag every criterion where two members differ by **more than 2 points**. Each divergence is a finding, not a rounding error.

| Criterion | Pol | Achille | Keigan | Edu | Difference > 2 | Cause of divergence |
|-----------|:---:|:-------:|:------:|:---:|:--------------:|---------------------|
| Security (XSS, CSRF & token storage) | | | | | | |
| Dynamic form rendering (`/about.json`) | | | | | | |
| Accessibility (WCAG AA) & i18n | | | | | | |
| Team learning curve | | | | | | |
| OpenAPI / TypeScript integration | | | | | | |
| Docker build & bundle footprint | | | | | | |

## 4. Final Choice & Practical Evidence

* **Selected Technology:** ______________________
* **Justification (250–300 words):**
* **Practical Evidence / Spike Reference:**