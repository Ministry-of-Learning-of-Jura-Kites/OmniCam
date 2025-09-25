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
    sqlc.arg (id)::UUID,
    sqlc.arg (project_id)::UUID,
    sqlc.arg (name),
    sqlc.arg (description),
    sqlc.arg (file_path),
    sqlc.arg (image_path)
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
