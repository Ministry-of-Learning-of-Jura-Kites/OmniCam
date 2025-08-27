-- name: CreateModel :one
INSERT INTO
  "model" (project_id, name, description, file_path)
VALUES
  (
    sqlc.arg (project_id)::UUID,
    sqlc.arg (name),
    sqlc.arg (description),
    sqlc.arg (file_path)
  )
RETURNING
  id,
  project_id,
  name,
  description,
  version,
  created_at,
  updated_at;
