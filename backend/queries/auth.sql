-- name: CreateActionToken :one
INSERT INTO auth_action_tokens (
    user_id,
    purpose,
    secret_hash,
    expires_at,
    requested_ip
)
VALUES (
    sqlc.arg(user_id),
    sqlc.arg(purpose),
    sqlc.arg(secret_hash),
    sqlc.arg(expires_at),
    sqlc.narg(requested_ip)
)
RETURNING
    id,
    user_id,
    purpose,
    secret_hash,
    expires_at,
    consumed_at,
    requested_ip,
    created_at;

-- name: GetActionTokenForUpdate :one
SELECT
    id,
    user_id,
    purpose,
    secret_hash,
    expires_at,
    consumed_at,
    requested_ip,
    created_at
FROM auth_action_tokens
WHERE id = sqlc.arg(id)
  AND purpose = sqlc.arg(purpose)
FOR UPDATE;

-- name: ConsumeActionToken :one
UPDATE auth_action_tokens
SET consumed_at = NOW()
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
  AND purpose = sqlc.arg(purpose)
  AND consumed_at IS NULL
  AND expires_at > NOW()
RETURNING
    id,
    user_id,
    purpose,
    consumed_at;

-- name: InvalidatePreviousActionTokens :many
UPDATE auth_action_tokens
SET consumed_at = NOW()
WHERE user_id = sqlc.arg(user_id)
  AND purpose = sqlc.arg(purpose)
  AND consumed_at IS NULL
RETURNING
    id,
    user_id,
    purpose,
    consumed_at;

-- name: LockUserForActionTokenChange :one
SELECT id
FROM users
WHERE id = sqlc.arg(user_id)
FOR UPDATE;
