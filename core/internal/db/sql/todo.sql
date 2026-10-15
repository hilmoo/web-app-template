-- name: TodoCreate :one
INSERT INTO todos("title", "description", "owner_id")
    VALUES (sqlc.arg('title'), sqlc.narg('description'), sqlc.arg('owner_id'))
RETURNING *;

-- name: TodoGetById :one
SELECT *
FROM todos
WHERE "id" = sqlc.arg('id')
    AND "owner_id" = sqlc.arg('owner_id');

-- name: TodoList :many
SELECT *
FROM todos
WHERE "owner_id" = sqlc.arg('owner_id')
ORDER BY "created_at" DESC;

-- name: TodoUpdate :one
UPDATE
    todos
SET "title" = COALESCE(sqlc.narg('title'), "title"),
    "description" = COALESCE(sqlc.narg('description'), "description"),
    "done" = COALESCE(sqlc.narg('done'), "done")
WHERE "id" = sqlc.arg('id')
    AND "owner_id" = sqlc.arg('owner_id')
RETURNING *;

-- name: TodoDelete :execrows
DELETE FROM todos
WHERE "id" = sqlc.arg('id')
    AND "owner_id" = sqlc.arg('owner_id');

