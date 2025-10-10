-- name: CreateModel :one
INSERT INTO
  "model" (
    id,
    project_id,
    name,
    description,
    file_path,
    image_path
  )
VALUES
  (
    SQLC.ARG(id)::UUID,
    SQLC.ARG(project_id)::UUID,
    SQLC.ARG(name),
    SQLC.ARG(description),
    SQLC.ARG(file_path),
    SQLC.ARG(image_path)
  )
RETURNING
  id,
  project_id,
  name,
  image_path,
  file_path,
  description,
  version,
  created_at,
  updated_at;
