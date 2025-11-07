-- name: UpdateProject :one
UPDATE "project"
SET
  name = COALESCE(SQLC.NARG(name)::VARCHAR, name),
  description = COALESCE(SQLC.NARG(description)::TEXT, description),
  updated_at = NOW()
WHERE
  id = SQLC.ARG(id)::UUID
RETURNING
  id,
  name,
  description,
  image_path,
  file_extension,
  created_at,
  updated_at;
