-- name: CreateProject :one
INSERT INTO
  "project" (id, name, description, image_path, file_extension)
VALUES
  (
    SQLC.ARG(id)::UUID,
    SQLC.ARG(name)::VARCHAR,
    SQLC.ARG(description)::TEXT,
    SQLC.ARG(image_path)::TEXT,
    SQLC.ARG(file_extension)::TEXT
  )
RETURNING
  id,
  name,
  description,
  image_path,
  file_extension,
  created_at,
  updated_at;
