-- name: DeleteModel :one
DELETE FROM "model"
WHERE
  id = sqlc.arg (id)::UUID
RETURNING
  id;
