-- name: UpdateProject :one
UPDATE "project"
SET
  name = COALESCE(NULLIF(sqlc.arg (name)::VARCHAR, ''), name),
  description = COALESCE(
    NULLIF(sqlc.arg (description)::TEXT, ''),
    description
  )
WHERE
  id = sqlc.arg (id)::UUID
RETURNING
  id,
  name,
  description,
  created_at,
  updated_at;
