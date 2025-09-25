-- name: GetProjectById :one
SELECT
  name,
  description,
  created_at,
  image_path,
  updated_at
FROM
  "project"
WHERE
  id = sqlc.arg (id)::UUID;
