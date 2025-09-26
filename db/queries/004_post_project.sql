-- name: CreateProject :one
INSERT INTO
  "project" (id, name, description, image_path)
VALUES
  (
    sqlc.arg (id)::UUID,
    sqlc.arg (name)::VARCHAR,
    sqlc.arg (description)::TEXT,
    sqlc.arg (image_path)::TEXT
  )
RETURNING
  id,
  name,
  description,
  image_path,
  created_at,
  updated_at;
