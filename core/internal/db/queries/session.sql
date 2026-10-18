-- name: SessionGetByTokenHash :one
SELECT *
FROM sessions
WHERE "token_hash" = $1
    AND "expires_at" > NOW();

-- name: SessionCreateNew :exec
INSERT INTO sessions("user_id", "token_hash", "expires_at", "ip_address", "user_agent")
    VALUES ($1, $2, $3, $4, $5);

-- name: SessionDeleteByTokenHash :exec
DELETE FROM sessions
WHERE "token_hash" = $1;

-- name: SessionListByUserId :many
SELECT *
FROM sessions
WHERE "user_id" = $1
    AND "expires_at" > NOW()
    AND (sqlc.narg(CURSOR)::uuid IS NULL
        OR id < sqlc.narg(CURSOR)::uuid)
ORDER BY "id" DESC
LIMIT $2;

-- name: SessionGetById :one
SELECT *
FROM sessions
WHERE "id" = $1
    AND "expires_at" > NOW()
    AND "user_id" = $2;

-- name: SessionDeleteById :exec
DELETE FROM sessions
WHERE "id" = $1
    AND "user_id" = $2;

