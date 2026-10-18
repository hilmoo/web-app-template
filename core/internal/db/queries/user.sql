-- name: GetUser :one
SELECT "id",
    "email",
    "password_hash",
    "full_name",
    "created_at"
FROM users
WHERE "id" = $1;

-- name: GetUserByEmail :one
SELECT "id",
    "email",
    "password_hash",
    "full_name",
    "created_at"
FROM users
WHERE "email" = $1;

-- name: CreateUser :one
INSERT INTO users("email", "full_name", "password_hash")
    VALUES ($1, $2, $3)
RETURNING "id", "email", "password_hash", "full_name", "created_at";

-- name: UpdateUser :one
UPDATE
    users
SET "email" = COALESCE(sqlc.narg('email'), "email"),
    "full_name" = COALESCE(sqlc.narg('full_name'), "full_name"),
    "password_hash" = COALESCE(sqlc.narg('password_hash'), "password_hash")
WHERE "id" = $1
RETURNING "id",
    "email",
    "password_hash",
    "full_name",
    "created_at";

