-- name: CreateProject :one
INSERT INTO
  "project" (id, name, description, image_path)
VALUES
  (
    SQLC.ARG(id)::UUID,
    SQLC.ARG(name)::VARCHAR,
    SQLC.ARG(description)::TEXT,
    SQLC.ARG(image_path)::TEXT
  )
RETURNING
  id,
  name,
  description,
  image_path,
  created_at,
  updated_at;
