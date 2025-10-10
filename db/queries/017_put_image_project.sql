-- name: UpdateProjectImage :one
UPDATE "project"
SET
  image_path = SQLC.ARG(image_path)::TEXT,
  updated_at = NOW()
WHERE
  id = SQLC.ARG(id)::UUID
RETURNING
  id,
  name,
  description,
  image_path,
  created_at,
  updated_at;
