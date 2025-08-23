-- name: CreateProject :one
INSERT INTO
  "project" (name, description)
VALUES
  (
    sqlc.arg (name)::VARCHAR,
    sqlc.arg (description)::TEXT
  )
RETURNING
  id,
  name,
  description,
  created_at,
  updated_at;
