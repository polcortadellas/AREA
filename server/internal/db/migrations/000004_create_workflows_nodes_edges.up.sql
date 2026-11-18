CREATE TABLE workflows (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name        text NOT NULL,
    enabled     boolean NOT NULL DEFAULT true,
    poll_cursor jsonb,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX workflows_user_id_idx ON workflows (user_id);

CREATE TABLE nodes (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id uuid NOT NULL REFERENCES workflows (id) ON DELETE CASCADE,
    kind        text NOT NULL CHECK (kind IN ('trigger', 'reaction', 'condition')),
    type        text NOT NULL,
    config      jsonb NOT NULL DEFAULT '{}',
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX nodes_workflow_id_idx ON nodes (workflow_id);

CREATE TABLE edges (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id  uuid NOT NULL REFERENCES workflows (id) ON DELETE CASCADE,
    from_node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
    to_node_id   uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
    branch       text CHECK (branch IN ('true', 'false')),
    UNIQUE (from_node_id, to_node_id)
);

CREATE INDEX edges_workflow_id_idx ON edges (workflow_id);
