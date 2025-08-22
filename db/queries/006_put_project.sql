-- name: UpdateProject :one
UPDATE "project"
SET
  name = sqlc.arg (name)::VARCHAR,
  description = sqlc.arg (description)::TEXT
WHERE
  id = sqlc.arg (id)::UUID
RETURNING
  id,
  name,
  description,
  created_at;
