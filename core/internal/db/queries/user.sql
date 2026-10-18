-- name: UserGetById :one
SELECT *
FROM users
WHERE "id" = $1;

-- name: UserGetByEmail :one
SELECT *
FROM users
WHERE "email" = $1;

-- name: UserCreatePassword :one
INSERT INTO users("email", "full_name", "password_hash")
    VALUES ($1, $2, $3)
RETURNING *;

-- name: UserUpdate :one
UPDATE
    users
SET "email" = COALESCE(sqlc.narg('email'), "email"),
    "full_name" = COALESCE(sqlc.narg('full_name'), "full_name"),
    "password_hash" = COALESCE(sqlc.narg('password_hash'), "password_hash")
WHERE "id" = $1
RETURNING *;

