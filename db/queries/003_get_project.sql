-- name: GetProjectById :one
SELECT
  name,
  description,
  created_at,
  updated_at
FROM
  "project"
WHERE
  id = sqlc.arg (id)::UUID;
