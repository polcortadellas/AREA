CREATE TABLE user_services (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    service           text NOT NULL,
    access_token_enc  bytea NOT NULL, -- AES-GCM ciphertext
    refresh_token_enc bytea,
    expires_at        timestamptz,
    scopes            text NOT NULL DEFAULT '',
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, service)
);
