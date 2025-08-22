-- name: GetAllProjects :many
SELECT
  name,
  description,
  created_at
FROM
  "project";
