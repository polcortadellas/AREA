-- name: UpsertUserService :exec
INSERT INTO user_services (user_id, service, access_token_enc, refresh_token_enc, expires_at, scopes)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id, service) DO UPDATE SET
    access_token_enc  = EXCLUDED.access_token_enc,
    refresh_token_enc = EXCLUDED.refresh_token_enc,
    expires_at        = EXCLUDED.expires_at,
    scopes            = EXCLUDED.scopes,
    created_at        = now(),
    updated_at        = now();

-- name: ListUserServices :many
SELECT service, scopes, created_at FROM user_services
WHERE user_id = $1
ORDER BY service;

-- name: DeleteUserService :execrows
DELETE FROM user_services WHERE user_id = $1 AND service = $2;

-- name: LockUserService :one
SELECT * FROM user_services
WHERE user_id = $1 AND service = $2
FOR UPDATE;

-- name: LockNextExpiringUserService :one
SELECT * FROM user_services
WHERE refresh_token_enc IS NOT NULL
  AND expires_at < sqlc.arg(before)
  AND service = ANY(sqlc.arg(services)::text[])
  AND NOT (id = ANY(sqlc.arg(skip)::uuid[]))
ORDER BY expires_at
LIMIT 1
FOR UPDATE SKIP LOCKED;

-- name: UpdateUserServiceTokens :exec
UPDATE user_services
SET access_token_enc = $2, refresh_token_enc = $3, expires_at = $4, updated_at = now()
WHERE id = $1;
