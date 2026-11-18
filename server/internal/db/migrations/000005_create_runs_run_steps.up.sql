CREATE TABLE runs (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id       uuid NOT NULL REFERENCES workflows (id) ON DELETE CASCADE,
    external_event_id text NOT NULL,
    status            text NOT NULL DEFAULT 'PENDING'
                      CHECK (status IN ('PENDING', 'RUNNING', 'SUCCEEDED', 'FAILED', 'DEAD')),
    trigger_payload   jsonb NOT NULL DEFAULT '{}',
    error             text,
    created_at        timestamptz NOT NULL DEFAULT now(),
    started_at        timestamptz,
    finished_at       timestamptz,
    UNIQUE (workflow_id, external_event_id)
);

-- recovery sweep of stale PENDING runs
CREATE INDEX runs_status_created_at_idx ON runs (status, created_at);

CREATE TABLE run_steps (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id          uuid NOT NULL REFERENCES runs (id) ON DELETE CASCADE,
    node_id         uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
    status          text NOT NULL DEFAULT 'PENDING'
                    CHECK (status IN ('PENDING', 'RUNNING', 'SUCCEEDED', 'FAILED', 'SKIPPED')),
    attempt         integer NOT NULL DEFAULT 0,
    idempotency_key text NOT NULL,
    output          jsonb,
    error           text,
    started_at      timestamptz,
    finished_at     timestamptz,
    UNIQUE (run_id, node_id)
);
