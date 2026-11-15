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

Option A: Vue 3 + Vite. Details: [`../../spikes/sqlite_vue/spike-vue-sqlite.md`](../../spikes/sqlite_vue/spike-vue-sqlite.md)

Option B: React 19 + Vite. Details: [`../../spikes/nestjs/SPIKE.md`](../../spikes/nestjs/SPIKE.md#3-front-end-react--vite-strengths-and-weaknesses)

Option C: Flutter Web. Details: [`../../spikes/flutter/SPIKE.md`](../../spikes/flutter/SPIKE.md) (scores are estimates: no measurements were kept)

| Criterion                              | Weight | Option A: Vue | Option B: React | Option C: Flutter |
|----------------------------------------|:------:|:-------------:|:---------------:|:-----------------:|
| Security (XSS, CSRF & token storage)   |   3    |       4       |        4        |         3         |
| Dynamic form rendering (`/about.json`) |   3    |       5       |        4        |         3         |
| Accessibility (WCAG AA) & i18n         |   3    |       4       |        3        |         2         |
| Team learning curve                    |   2    |       4       |        4        |         2         |
| OpenAPI / TypeScript integration       |   2    |       4       |        4        |         3         |
| Docker build & bundle footprint        |   1    |       5       |        3        |         2         |
| **Weighted total**                     |        |    **60**     |     **52**      |      **36**       |

**What has and has not been spiked**

- React (Vite) and Vue (Vite) have a web spike each; Flutter has a recovered web spike without measurements.
- **React Native has no spike.** It is chosen for the mobile client because it shares TypeScript and the OpenAPI-generated client with the React web client (see [ADR 0005](../adr/0005-react-react-native.md)), not because it was tested.

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

* **Selected Technology:** React (web) + React Native (mobile)
* **Justification (250–300 words):**

The matrix ranks Vue first (60), React second (52) and Flutter third (36), so this choice is not driven by the totals. It is driven by the mobile client, which the matrix does not measure. The subject asks for a mobile client that lists workflows, enables and disables them and shows what they did, while business logic stays on the server. With React for the web, the mobile client can be React Native: one language (TypeScript), the same API client generated from the server's OpenAPI description, and the same hooks for state. The screens are written twice, but the part that carries the logic is shared. Vue has no equivalent path and would add a second technology for mobile.

The two points React lost against Vue are dynamic forms (4 against 5) and accessibility (3 against 4). The accessibility score is an estimate for both: each produces semantic HTML, and `eslint-plugin-jsx-a11y` and `react-i18next` are mature. The team also has more React experience (learning curve 4, and 5 for the member who built the spike), and the React spike already renders a list from the API and shows server validation errors. Flutter was rejected for the web: it draws on a canvas, which hurts accessibility (scored 2), it needs Dart, and its bundle is heavier.

Not tested yet: WCAG AA checks with automated tools, token storage (HttpOnly cookies on web, Keychain / Keystore on mobile) and any React Native code. We propose adding a "mobile client" criterion to this matrix so the decision is documented.

* **Practical Evidence / Spike Reference:** [`spikes/nestjs/SPIKE.md`](../../spikes/nestjs/SPIKE.md#3-front-end-react--vite-strengths-and-weaknesses), [`spikes/sqlite_vue/spike-vue-sqlite.md`](../../spikes/sqlite_vue/spike-vue-sqlite.md), [`spikes/flutter/SPIKE.md`](../../spikes/flutter/SPIKE.md)
