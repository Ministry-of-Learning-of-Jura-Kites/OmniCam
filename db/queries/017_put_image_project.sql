-- name: UpdateProjectImage :one
UPDATE "project"
SET
  image_path = sqlc.arg (image_path)::TEXT,
  updated_at = NOW()
WHERE
  id = sqlc.arg (id)::UUID
RETURNING
  id,
  name,
  description,
  image_path,
  created_at,
  updated_at;
