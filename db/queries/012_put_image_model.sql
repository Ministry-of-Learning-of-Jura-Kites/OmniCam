-- name: UpdateModelImage :one
UPDATE "model"
SET
  image_path = SQLC.ARG(image_path)::TEXT,
  image_extension = SQLC.ARG(image_extension)::TEXT,
  version = version + 1,
  updated_at = NOW()
WHERE
  id = SQLC.ARG(id)::UUID
RETURNING
  id,
  project_id,
  name,
  image_path,
  file_path,
  image_extension,
  model_extension,
  description,
  version,
  created_at,
  updated_at;
