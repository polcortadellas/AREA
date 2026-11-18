# Architecture

Milestone 1 Deliverable | Stack: Go, PostgreSQL, Redis Streams, React (Web) & React Native (Android) | Team Size: 4

## 1. Context & Architectural Drivers

AREA is an event-driven automation platform (inspired by IFTTT and Zapier) composed of an Application Server (Go), a Web Client (React), and an Android Mobile Client (React Native). Our architecture is governed by four technical drivers:

- **Non-Blocking, Fault-Tolerant Execution:** Hooks (polling and webhooks) must detect external events and execute multi-step workflows (including conditional branching and >= 4 sequential reactions) asynchronously without blocking HTTP traffic or losing queued runs if a worker stops.

- **Modular Service Extensibility:** Adding a new third-party Service, Action, or REAction must not require modifying the core workflow execution engine, and must automatically expose its parameter and output schemas via /about.json.

- **Transport-Agnostic Hexagonal Core:** Domain logic (graph validation, template resolution, execution engine) is isolated from HTTP presentation so the OpenAPI 3.0 REST API and future adapters invoke identical domain operations.

- **Team & Delivery Constraints:** Built by a 4-member team under fixed milestones (Plan, MVP, Final Defense), requiring single-command Docker Compose reproducibility and clear vertical module ownership.

## 2. Evaluated Architecture Options

The core system-level problem for AREA is how asynchronous workflow execution is decoupled from HTTP request handling and how state consistency is preserved across components under worker crashes or restarts.

### Option 1: Modular Monolith + Transactional Outbox & Redis Streams Broker (Selected)

- **How It Works:** A single Go codebase is organized into isolated domain modules (auth, workflows, connectors, engine) compiled into one binary deployed in three runtime roles: api (REST API, OAuth/OIDC, webhooks), scheduler (single-leader polling loop), and worker (horizontally scalable DAG executor pool). Triggered runs are persisted in PostgreSQL with status PENDING and published to Redis Streams (AOF enabled), where workers consume messages via Consumer Groups.

- **Engineering Advantages:** True fault isolation between ingestion and execution (workers can stop, crash, or scale via WORKER_CONCURRENCY while Redis Streams retains unacknowledged messages on disk); shared compile-time Go structs and connector interfaces without RPC serialization overhead or schema drift; and native Consumer Group semantics (XREADGROUP, XACK, XAUTOCLAIM).

- **Disadvantages & Accepted Trade-offs:** Dual-write consistency window between PostgreSQL and Redis Streams if the API crashes before publishing; an additional stateful container (redis) to monitor in /healthz; and compile-time-only module boundaries enforced via internal/ packages and PR reviews.

- **Reliability, Deduplication & Side-Effect Controls (At-Least-Once Delivery):** (1) Transactional Outbox & Recovery Sweep: Runs are committed to PostgreSQL as PENDING before stream publication; a recovery sweep re-enqueues stale PENDING runs (> 30s) and reclaims crashed worker messages via Redis XAUTOCLAIM. (2) Atomic Worker Claim: To prevent duplicate execution if both the sweep and stream deliver a job, workers execute an atomic SQL state transition (UPDATE runs SET status='RUNNING', started_at=NOW() WHERE id=$1 AND status='PENDING'); if 0 rows match, the duplicate message is immediately ACKed and skipped. (3) Single-Leader Polling & DB Deduplication: The polling scheduler acquires a PostgreSQL session advisory lock (pg_try_advisory_lock) so scaling worker replicas never causes parallel polling, and ingests events via INSERT ... ON CONFLICT (workflow_id, external_event_id) DO NOTHING.

### Option 2: Modular Monolith + PostgreSQL SKIP LOCKED Queue (Considered & Rejected)

- **How It Works:** Eliminates Redis entirely; background jobs live in a PostgreSQL table and workers poll using SELECT ... FOR UPDATE SKIP LOCKED.

- **Honest Rejection Rationale:** At student-project traffic volume, PostgreSQL row-locking would not fail from I/O saturation, and our Transactional Outbox recovery sweep already queries PostgreSQL periodically. A PostgreSQL-backed job queue (e.g. River) would also satisfy ENG-ASYNC, which accepts 'an existing broker or job queue'. Option 2 was therefore rejected for two project-specific reasons rather than as a universally worse design: (1) consumer groups, pending-entry reclamation (XAUTOCLAIM) and acknowledgement semantics are native in Redis Streams, whereas on PostgreSQL we would either adopt an additional queue library or write visibility-timeout and redelivery logic ourselves; and (2) the team explicitly aimed to gain practical experience operating a dedicated streaming broker alongside relational storage. Accepted cost: one more stateful container and a dual-write window, mitigated by the outbox sweep and atomic claim described above. Option 2 remains a viable fallback if Redis operation proves too costly.

### Option 3: Microservices Architecture (Considered & Rejected)

- **How It Works:** Decomposes the platform into separate API Gateway, Auth Service, Workflow Engine, and per-connector microservices over gRPC/HTTP.

- **Why It Was Rejected:** Distributed tracing, inter-service auth, multiple Dockerfiles, and distributed sagas create disproportionate operational overhead for a 4-person team delivering Server, React Web, and React Native Android within fixed milestones.

## 3. Decision Summary Matrix

| Evaluation Criterion | Option 1: Modular Monolith + Broker | Option 2: Modular Monolith + PG Queue | Option 3: Microservices | Verdict |
|---|---|---|---|---|
| Async Decoupling & Stream Semantics | High (Native Redis Streams & Consumer Groups) | Medium (Custom SQL polling & visibility logic) | High (Full process isolation) | Option 1 |
| State Consistency | High (Transactional Outbox + atomic SQL claim) | High (Single ACID database transaction) | Low (Requires distributed sagas) | Option 1 |
| Operational Complexity | Low-Medium (api, scheduler, worker, postgres, redis + Mailpit) | Low (3 backend containers) | High (6+ containers & internal RPC) | Option 1 |
| Development Velocity & Rubric Fit | High (Shared Go types; native stream semantics) | High (Shared Go types; ENG-ASYNC met via PG job queue) | Low (High infrastructure boilerplate) | Option 1 |
| Final Verdict | Selected | Rejected (Custom queue machinery; learning goal) | Rejected (Excessive ops overhead) | Selected |

## 4. Target System Architecture & Core Domain Specifications

```
+-----------------------+       +-----------------------+
|      Web Client       |       |    Android Client     |
|        (React)        |       |    (React Native)     |
+-----------+-----------+       +-----------+-----------+
            |                               |
            +---------------+---------------+
                            | HTTPS / REST (OpenAPI 3.0 Contract)
                            v
+-----------------------------------------------------------+
|                Application Server (Go - API)              |
|  - Auth (Argon2id/JWT + Email Verify + OIDC) & Token Vault|
|  - DAG & Key-Path Validator + /about.json Schema Discovery|
|  - SSRF-Safe Webhook Receiver & /healthz (<1s Probe)      |
+-----------------------------+-----------------------------+
                              |
              +---------------+---------------+
              |                               |
              v (Outbox & State)              v (Publish Event)
+-----------------------+       +---------------------------+
|      PostgreSQL       |       |  Message Broker (Redis)   |
| - Users & AES Tokens  |       | - Persistent Streams (AOF)|
| - Workflows (DAG)     |       | - Consumer Groups         |
| - Runs & Step Logs    |       | - XAUTOCLAIM Recovery     |
| - Dead-Letter Queue   |       +-------------+-------------+
+-----------+-----------+                     | (Consume & Atomic Claim)
            ^                                 v
            | (Persist Steps)   +---------------------------+
            +-------------------+  Go Scheduler & Workers   |
                                | - Advisory-Locked Poller  |
                                | - DAG Executor & Retries  |
                                | - Connector Registry      |
                                +-------------+-------------+
                                              |
                                              v (HTTPS)
                                +---------------------------+
                                |    External Services      |
                                | (GitHub, Google, etc.)    |
                                +---------------------------+
```

### Explicit Core Specifications & Edge-Case Semantics

- **Typed Connector & Node Contracts:** Services implement ServiceConnector (Metadata()). Actions (triggers) implement TriggerNode (ValidateConfig(), InputSchema(), OutputSchema(), Poll(ctx, config, cursor, auth), HandleWebhook(ctx, config, req)) because triggers detect events rather than execute side effects; their InputSchema() lists the parameters the user sets, such as a repository. REActions implement ReactionNode (ValidateConfig(), InputSchema(), OutputSchema(), Execute(ctx, inputs, auth)).

- **Save-Time Key-Path Validation & Runtime Missing-Value Semantics:** Because every node declares its OutputSchema(), the workflow validator rejects any {{ node.key }} reference pointing to an undeclared key or non-ancestor node with HTTP 422 at save time. At runtime, if an optional declared key resolves to null/missing, string interpolation substitutes an empty string ''; if a required input parameter receives a missing value, the step fails with the non-retryable error MISSING_TEMPLATE_KEY and goes straight to the DLQ without retries, since retrying cannot make a missing key appear (raw {{...}} text is never rendered).

- **Empty Branch Policy:** A conditional node (if/else) must contain at least one REAction on both its true and false branches; workflows with an empty branch are rejected with HTTP 422 at save time (single-path workflows omit the condition node entirely).

- **Multi-Step Failure & Replay Semantics (e.g., REAction 2 of 4 Fails):** Execution halts immediately at the failing step (Step 1 = SUCCEEDED, Step 2 = FAILED, Steps 3–4 = SKIPPED). Bounded exponential backoff retries Step 2; if retries exhaust, the run enters the Dead-Letter Queue (DLQ). On manual DLQ replay, previously SUCCEEDED steps are skipped via their persisted run_steps record and per-step idempotency key, resuming execution cleanly from Step 2. Known limit: delivery is at-least-once. Idempotency keys prevent duplicates where the provider supports them or where a pre-check is possible (e.g. looking up an issue already created for this run before creating another); for providers where neither applies (Discord webhooks, Gmail send), a crash between the external call and the run_steps commit can repeat that single side effect on replay. This residual window is documented rather than claimed away.

- **OIDC Account Linking & SSRF Guardrails:** Google is the OIDC identity provider; GitHub, which offers OAuth2 but not OIDC for user sign-in, is used only for delegated service access. External OIDC identities auto-link to an existing local account strictly when the provider asserts email_verified == true (otherwise HTTP 409 requires explicit linking after local login). User-supplied Discord webhook URLs are validated against an explicit host/path allowlist (https://discord.com/api/webhooks/* and https://discordapp.com/api/webhooks/*) to block SSRF.
