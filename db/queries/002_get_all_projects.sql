-- name: GetAllProjects :many
SELECT
  id,
  name,
  description,
  created_at,
  updated_at
FROM
  "project";
