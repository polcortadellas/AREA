-- The client's node ID ("start", "true_1"): edges and {{ key.field }} references in the API use it.
ALTER TABLE nodes ADD COLUMN key text;
UPDATE nodes SET key = id::text;
ALTER TABLE nodes ALTER COLUMN key SET NOT NULL;
ALTER TABLE nodes ADD CONSTRAINT nodes_workflow_id_key_key UNIQUE (workflow_id, key);
