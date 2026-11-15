# Features (MoSCoW)

Milestone 1 Deliverable | Stack: Go, PostgreSQL, Redis Streams, React (Web) & React Native (Android) | Team Size: M = 4

## 1. Quantitative Project Rules & Balanced Scoring Strategy

- **Services (S):** S >= 1 + M = 5 services (3 OAuth2 services: GitHub, Google Workspace, Spotify; 2 API/Webhook services: Discord, OpenWeather = 60% OAuth2, exceeding the >= 50% requirement). Plus a pre-scoped 6th OAuth2 fallback (GitLab, owned by Edu, task F-08c) if Spotify's dev-mode allowlist restricts external testing.

- **Actions + REActions (A + R):** A + R >= 2M + 2 = 10 minimum. Every service provides at least 2 components, totaling 13 distinct Actions/REActions (GitHub: 4, Google Workspace: 3, Spotify: 2, Discord: 2 REActions, OpenWeather: 2). Discord exposes two REActions rather than an Action: Discord channel webhooks only accept outgoing messages, and its app-level Webhook Events (app authorisation, entitlements, Social SDK lobbies) carry no channel activity that could serve as a meaningful trigger.

- **Additional Catalogue Difficulty (D):** D >= 2M = 8 points. We commit to 9 official catalogue features totaling exactly 15 points according to the official rubric index: 4 x Category 1 (1 pt) + 4 x Category 2 (2 pts) + 1 x Category 3 (3 pts) = 15 pts (7-point safety buffer).

> **Fair Catalogue Point Distribution & Cross-Stack Exposure Policy**
>
> Catalogue points follow ownership of the code they verify: Pol (6 pts: ENG-ASYNC [3] + ENG-RETRY [2] + SYS-HEALTH [1], since the health probe reports his worker and scheduler heartbeats), Edu (4 pts: ENG-EXT [2] + SEC-DEPS [1] + ENG-ADR [1]), Achille (3 pts: SEC-CRYPTO [2] + SYS-SECRETS [1]) and Keigan (2 pts: SYS-BACKUP [2]). Keigan's lower point count is deliberate: he carries the largest share of the common base (both clients, the builder, EN/ES i18n and the WCAG AA audit). Measured in scheduled working days, every member carries between roughly 14 and 18 days in the Final track. Every member also owns cross-stack work (Achille: React auth & email-confirmation UI M-18 and the React Native Google OIDC login F-02c, so that a second member works on the mobile client; Edu: React Connected Services UI M-19; Pol: React DLQ Replay UI F-06; Keigan: Go pg_dump backup/restore scripts F-11a) and participates in mandatory cross-layer PR reviews.

## 2. Priority 1: Must Have (Mandatory Common Base)

All mandatory common-base requirements from the project specification—including account email confirmation, automated deployment, schema-driven dynamic forms, and cross-client execution history—are prioritized first:

| Feature | Concrete Implementation Scope | Layers | Effort | Origin | Primary Owner(s) |
|---|---|---|---|---|---|
| Local Auth, Email Confirmation & OIDC | Register with Argon2id + email confirmation link (tested via Mailpit SMTP in Docker Compose), JWT login, and verified-email Google OIDC SSO (GitHub is OAuth2-only and is used for delegated service access, not identity). | Back, Web, Mobile | Medium | Subject PDF | Achille (Back) & Keigan/Achille (Clients) |
| Delegated Service Auth & Auto-Refresh | Connect third-party services via OAuth2, persist encrypted tokens, and proactively refresh expiring access tokens in background. | Back, Web | Medium | Subject PDF | Achille (Back) & Edu (Web UI M-19) |
| Service Discovery (/about.json) | Dynamic endpoint generated from Go TriggerNode/ReactionNode registries exposing Services, Actions, REActions, inputs, and output keys. | Back | Low | Subject PDF | Edu |
| Dynamic Schema-Driven Workflow Builder | React builder dynamically generating parameter forms and condition selectors from /about.json (1 trigger, <=1 if/else with 2 non-empty branches, >=4 reactions). | Back, Web | High | Subject PDF | Keigan (React UI) & Pol (Graph API) |
| Deterministic Conditional Logic | Condition node evaluating comparison operators (>, <, ==, contains) on upstream outputs to route execution into one of two non-merging branches. | Back, Web | Medium | Subject PDF | Pol (Engine) & Keigan (React UI) |
| Typed Key-Path Data Mapping | Pass declared upstream keys ({{ trigger.field }}, {{ node_1.id }}) with save-time OutputSchema validation (HTTP 422) and defined missing-value semantics. | Back, Web | Medium | Subject PDF | Pol (Engine) & Edu (Node Schemas) |
| Autonomous Hook Engine | Single-leader polling scheduler (pg_try_advisory_lock) and webhook receiver with DB-level event deduplication running without active clients. | Back | High | Subject PDF | Pol |
| Service & Action Quota (5 S, 13 A+R) | Deliver 5 services (>=2 components each, 3 OAuth2): GitHub (4), Google (3), Spotify (2), Discord (2 REActions: send message, create forum thread; SSRF-validated), OpenWeather (2). | Back | Medium | Subject PDF | Edu |
| Workflow Execution History UI | Detailed run history across React Web and React Native Android showing timestamps, executed branch, per-step status, and sanitized errors. | Back, Web, Mobile | Medium | Subject PDF | Keigan (UI) & Pol (API) |
| Android Mobile Client (React Native) | Installable React Native Android APK with runtime server URL config to list, enable/disable, and inspect workflow execution history (builder excluded). | Mobile | High | Subject PDF | Keigan |
| OpenAPI 3.0 Specification | Browsable OpenAPI 3.0 contract (openapi.yaml at /docs) covering all endpoints and error responses, verified against live routes in CI. | Back | Low | Subject PDF | Edu |
| WCAG AA & EN/ES i18n Compliance | English and Spanish i18n built in from Day 2 and audited against WCAG AA via axe-core, keyboard testing, and Android Accessibility Scanner. | Web, Mobile | Medium | Subject PDF | Keigan |
| Docker Compose, CI & Auto-Deployment | Single-command local build (including APK) + GitHub Actions CI + automated CD pipeline deploying the stack without manual steps. | DevOps | Medium | Subject PDF | Achille |

## 3. Priority 2: Selected Official Catalogue Features (15-Point Target)

Categories and point values match the official rubric index (Cat. 1 = 1 pt, Cat. 2 = 2 pts, Cat. 3 = 3 pts; Total = 15 pts):

| Feature & ID | Official Cat / Pts | What It Adds | Learning Rationale | Main Risk / Verification Criterion | Owner (Pts) |
|---|---|---|---|---|---|
| Asynchronous Worker Engine<br>[ENG-ASYNC] | Cat. 3<br>(3 pts) | Decouples trigger ingestion from execution via Transactional Outbox + Redis Streams Consumer Groups, XAUTOCLAIM recovery, and scalable Go workers. | Master event-driven streaming, consumer group offsets, atomic claims, and zero-loss crash recovery in distributed systems. | Prove live during defense that stopping workers queues jobs and restarting processes all pending runs with zero duplicate runs (atomic claim). | Pol<br>(3 pts) |
| Extensible Go Connectors<br>[ENG-EXT] | Cat. 2<br>(2 pts) | Typed TriggerNode/ReactionNode interfaces with startup plugin registry and auto-generated /about.json; adding a service requires zero engine edits. | Apply clean plugin architecture and runtime metadata reflection in Go so third-party integrations scale safely. | Follow HOWTOCONTRIBUTE.md to add and expose a trivial connector live during the final defense. | Edu<br>(2 pts) |
| Retries & Dead-Letter Replay<br>[ENG-RETRY] | Cat. 2<br>(2 pts) | Bounded exponential backoff retries with jitter, per-REAction idempotency keys, Dead-Letter Queue (DLQ) store, and step-resuming replay endpoint + UI. | Implement resilient failure recovery patterns that prevent duplicate side effects (e.g., duplicate emails/issues) when replaying failed steps. | When REAction 2 of 4 fails and is replayed after fixing the fault, Step 1 (recorded as SUCCEEDED) is not re-executed and Steps 2–4 complete; the residual duplicate window for non-idempotent providers is documented. | Pol<br>(2 pts) |
| Encryption at Rest<br>[SEC-CRYPTO] | Cat. 2<br>(2 pts) | Encrypts stored OAuth2 access/refresh tokens in PostgreSQL using AES-GCM 256 with env-injected master key and documented rotation script. | Gain hands-on experience with authenticated symmetric encryption, unique nonces, tamper detection, and key rotation. | Verify before/after SQL dumps, single-byte ciphertext tampering rejection, key rotation, and zero tokens in logs. | Achille<br>(2 pts) |
| Automated Backup & Restore<br>[SYS-BACKUP] | Cat. 2<br>(2 pts) | Scheduled timestamped pg_dump backups stored outside the DB container with retention rotation and one-command restore.sh (tested post-encryption). | Engineer automated PostgreSQL backup pipelines and disaster-recovery procedures verified under live data destruction. | Demonstrate live database destruction and 100% recovery of users, encrypted service tokens, and workflows during evaluation; worst-case data loss (RPO) equals the stated backup interval. | Keigan<br>(2 pts) |
| Architecture Decision Records<br>[ENG-ADR] | Cat. 1<br>(1 pt) | Committed Markdown ADRs (docs/adr/) documenting context, rejected alternatives, decisions, consequences, and traceability to code. | Practice structured architectural trade-off evaluation and defend system behavior under rejected design alternatives. | Every member traces their architectural decision from problem and rejected options to concrete code evidence. | Edu (lead) & All<br>(1 pt) |
| Dependency Vulnerability Scanning<br>[SEC-DEPS] | Cat. 1<br>(1 pt) | Automated govulncheck, npm audit (React & React Native), and Trivy container checks in GitHub Actions blocking high/critical CVEs. | Integrate automated software supply-chain security gates into CI/CD and document compensating controls for open findings. | CI blocks PRs with unmitigated critical CVEs; any accepted open findings have documented compensating controls. Evidence note co-authored by Edu and Achille (F-18). | Edu<br>(1 pt) |
| Dependency Health Checks<br>[SYS-HEALTH] | Cat. 1<br>(1 pt) | Strictly bounded (< 1s, 800ms timeout) GET /healthz probe checking PostgreSQL, Redis, worker heartbeat, and OIDC providers without leaking secrets. | Design multi-dependency readiness probes that fail fast under network stalls and expose backend observability. | Stopping PostgreSQL, Redis, or workers returns HTTP 503 naming the failing component in < 1s; restarting restores HTTP 200. | Pol<br>(1 pt) |
| Secrets Management & Scanning<br>[SYS-SECRETS] | Cat. 1<br>(1 pt) | Clean Git history enforced by Gitleaks, complete .env.example, environment-based secret injection, and automated log redaction. | Enforce zero-trust credential hygiene across Git history, runtime configuration, and structured application logs. | Full execution log grep and Gitleaks history scan during defense confirm zero leaked secrets. | Achille<br>(1 pt) |

## 4. Definition of Done & Feature Acceptance Criteria

We separate two levels of 'done'. The team-wide Definition of Done (DoD) is a single quality checklist that every issue, of any kind, must meet before it moves to Done on the board. Feature acceptance criteria are specific to each catalogue feature: they are taken from the catalogue's 'What your boss defined as done' and are not reinvented. A catalogue feature counts as delivered only when every one of its criteria can be observed.

- **Review & CI:** Merged to main through a pull request reviewed by a member working on another layer; lint, type-check, tests, build and security scans are green.

- **Tests:** The change is covered by at least one success-path and one failure-path test.

- **API contract:** If the HTTP API changed, openapi.yaml is updated and the contract tests pass; any new node appears in /about.json.

- **Clients:** All user-facing text uses i18n keys with EN and ES translations; no new critical axe-core violations on web or Accessibility Scanner issues on Android.

- **Secrets:** No secret, token or personal data appears in code, fixtures or logs.

- **Documentation:** README.md or HOWTOCONTRIBUTE.md is updated if setup or extension changed; an ADR is added or superseded if an architectural decision changed.

- **Reproducibility:** The stack still starts from a clean clone with docker compose up --build.

Feature acceptance criteria (condensed from the catalogue; the full wording is copied into each feature's issue). The 'Where observed' column feeds the rubric traceability matrix in the final documentation (F-17).

| Feature | Where Observed | Acceptance Criteria (Definition of Done) | Owner |
|---|---|---|---|
| ENG-ASYNC | Live: docker compose stop worker; ADR 0003; config docs | (1) Execution continues outside the triggering request, which answers before the work finishes. (2) An existing broker carries pending work and the choice is justified. (3) A separate worker can be stopped, work queued, then restarted without losing that work. (4) Third-party waits are bounded by documented timeouts. (5) Worker capacity is changed through configuration (WORKER_CONCURRENCY), not source edits. | Pol |
| ENG-EXT | Git history of last connector; HOWTOCONTRIBUTE.md; /about.json | (1) Nodes respect a common, documented interface covering validation and execution. (2) Node types register at start-up, isolated from execution logic, with no dynamic loading of third-party code. (3) The commit adding the last node type touches only that node and its registration. (4) HOWTOCONTRIBUTE.md has a reproducible procedure and a trivial node is added from it during the defense. (5) A new Action/REAction appears in about.json without a second hard-coded catalogue or a payload-shape change. | Edu |
| ENG-RETRY | DLQ view (F-06); retry config; ADR on retries | (1) Failed nodes are retried a bounded number of times with growing delays, stated before the evaluator provokes a failure. (2) Retries do not require a broker. (3) Side-effecting REActions are protected against duplicate effects during retry and the mechanism (and its limit) is explained. (4) Exhausted work lands in an inspectable, replayable dead-letter store. (5) The evaluator removes the cause and replays a dead-lettered run successfully. (6) Policy, attempts and delays are configuration. (7) The team explains how retry, persistence and idempotence depend on each other. | Pol |
| SEC-CRYPTO | docs/security/SEC-CRYPTO.md; psql on user_services; run logs | (1) A dump taken before encryption shows what was exposed and identifies sensitive data. (2) Third-party tokens are encrypted with an authenticated cipher (AES-GCM) and the evaluator reads ciphertext. (3) The key lives outside source, repository and database, and its provisioning is documented. (4) No token, clear or encrypted, appears in the logs of a full AREA run. (5) Key rotation is documented, including data encrypted with the previous key. (6) Altering one byte of a stored token is detected and refused. | Achille |
| SYS-BACKUP | restore.sh; backup volume; README | (1) Timestamped backups run on a schedule and are stored outside the database container. (2) A single documented command restores a backup. (3) The evaluator destroys data and the product works again on the restored data. (4) Worst-case data loss is stated and follows from the schedule. (5) Tokens in backups are encrypted (via SEC-CRYPTO). | Keigan |
| ENG-ADR | docs/adr/ | (1) At least three significant ADRs, one on the workflow engine, each saying why it deserved a record. (2) Each gives context, real alternatives, decision and accepted consequences. (3) Rejected options are rejected for this project, not declared universally bad. (4) Records are dated, carry a status, and superseded ones point to their replacement. (5) Any member can describe the product under a rejected option the evaluator picks. | Edu (lead) & All |
| SEC-DEPS | GitHub Actions history; docs/security/SEC-DEPS.md | (1) Scanning runs automatically on every contribution across server and both clients. (2) A stated severity threshold blocks contributions, shown in history or provoked live with a known-vulnerable dependency. (3) The current report is reachable and the most severe open finding is explained. (4) Accepted findings record rationale and compensating measures. (5) The team explains how a newly disclosed vulnerability reaches its attention (scheduled scan and repository alerts). | Edu (evidence with Achille) |
| SYS-HEALTH | GET /healthz | (1) A machine-readable endpoint reports each dependency separately: PostgreSQL, Redis, the OIDC provider and the hook runner (scheduler/worker heartbeat). (2) A required dependency failure turns the answer into a failing status. (3) The evaluator stops the database container and sees it identified. (4) Checks are bounded so the endpoint answers within one second. (5) No secret, connection string or stack trace is exposed. | Pol |
| SYS-SECRETS | Gitleaks CI; .env.example; log grep of a full run | (1) Secret scanning over current content and history is clean, or every historical secret is documented as revoked with proof. (2) A complete .env.example documents every variable with a description and harmless placeholder. (3) A fresh clone configured from it starts the product, done by the evaluator. (4) Secrets reach containers through the environment, never committed files or baked images. (5) No secret appears in any service log. | Achille |

## 5. Priority 3: Should Have & Priority 4: Could Have / Won't Have

| Item | Priority Tier | Concrete Scope & Rationale | Owner |
|---|---|---|---|
| Webhook Signature Verification | Priority 3: Should Have | Validate HMAC signatures and replay timestamps on incoming webhooks to prevent spoofed trigger invocations. | Pol |
| Ready-to-Use Workflow Presets | Priority 3: Should Have | Curated gallery of one-click starter workflow templates in React to accelerate live demonstrations. | Keigan |
| Per-User Scoped API Tokens | Priority 4: Could Have (Bonus) | Revocable SHA-256 hashed personal access tokens for authenticating external automation scripts. | Achille |
| Failure Alert Notifications & Dark Mode | Priority 4: Could Have (Bonus) | In-app banner on repeated workflow failure and persistent light/dark theme toggle in React. | Keigan |
| Declarative User-Uploaded JSON Connectors | Won't Have (Out of Scope) | Replaced by native Go connectors via ENG-EXT to avoid building a custom HTTP/OAuth DSL and complex SSRF defenses. | N/A |
| Visual Drag-and-Drop Node Graph Canvas | Won't Have (Out of Scope) | Replaced by a clean, WCAG AA-compliant step-by-step form wizard in React. | N/A |
| iOS Mobile Client & Heavy AI Guardrails | Won't Have (Out of Scope) | iOS is an optional side quest; mobile scope is React Native Android. Heavy AI-NODE/EVAL excluded to protect core schedule. | N/A |
