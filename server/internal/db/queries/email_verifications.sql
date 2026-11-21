-- name: CreateEmailVerification :exec
INSERT INTO email_verifications (user_id, token_hash, expires_at)
VALUES ($1, $2, $3);

-- name: InvalidateEmailVerifications :exec
UPDATE email_verifications SET used_at = now()
WHERE user_id = $1 AND used_at IS NULL;

-- name: ConsumeEmailVerification :one
UPDATE email_verifications SET used_at = now()
WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now()
RETURNING user_id;
