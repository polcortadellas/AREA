CREATE TABLE users (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email             text NOT NULL UNIQUE CHECK (email = lower(email)),
    password_hash     text,
    google_sub        text UNIQUE,
    email_verified_at timestamptz,
    created_at        timestamptz NOT NULL DEFAULT now(),
    CHECK (password_hash IS NOT NULL OR google_sub IS NOT NULL)
);
