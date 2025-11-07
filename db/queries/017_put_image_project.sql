-- name: UpdateProjectImage :one
UPDATE "project"
SET
  image_path = SQLC.ARG(image_path)::TEXT,
  file_extension = SQLC.ARG(file_extension)::TEXT,
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
