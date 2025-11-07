-- name: UpdateModel :one
UPDATE "model"
SET
  name = COALESCE(SQLC.NARG(name)::VARCHAR, name),
  description = COALESCE(SQLC.NARG(description)::TEXT, description),
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
