-- name: CreateModel :one
INSERT INTO
  "model" (
    id,
    project_id,
    name,
    description,
    file_path,
    image_path,
    image_extension,
    model_extension
  )
VALUES
  (
    SQLC.ARG(id)::UUID,
    SQLC.ARG(project_id)::UUID,
    SQLC.ARG(name),
    SQLC.ARG(description),
    SQLC.ARG(file_path),
    SQLC.ARG(image_path),
    SQLC.ARG(image_extension),
    SQLC.ARG(model_extension)
  )
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
