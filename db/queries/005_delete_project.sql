-- name: DeleteProject :one
DELETE FROM "project"
WHERE
  id = sqlc.arg (id)::UUID
RETURNING
  id;
