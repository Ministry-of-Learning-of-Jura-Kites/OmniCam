-- name: GetProjectById :one
SELECT
  name,
  description,
  created_at
FROM
  "project"
WHERE
  id = sqlc.arg (id)::UUID;
