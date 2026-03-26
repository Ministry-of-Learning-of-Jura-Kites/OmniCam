-- name: DeleteProject :one
DELETE FROM "project"
WHERE
  id = SQLC.ARG(id)::UUID
RETURNING
  id;
