-- name: DeleteModel :one
DELETE FROM "model"
WHERE
  id = SQLC.ARG(id)::UUID
RETURNING
  id;
