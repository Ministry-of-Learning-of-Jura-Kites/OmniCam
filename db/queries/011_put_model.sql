-- name: UpdateModel :one
UPDATE "model"
SET
  name = COALESCE(sqlc.narg (name)::VARCHAR, name),
  description = COALESCE(sqlc.narg (description)::TEXT, description),
  version = version + 1,
  updated_at = NOW()
WHERE
  id = sqlc.arg (id)::UUID
RETURNING
  id,
  project_id,
  name,
  description,
  version,
  created_at,
  updated_at;
