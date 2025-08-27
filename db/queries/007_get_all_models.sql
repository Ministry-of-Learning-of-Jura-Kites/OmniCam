-- name: GetAllModels :many
SELECT
  id,
  project_id,
  name,
  description,
  version,
  created_at,
  updated_at
FROM
  "model"
WHERE
  project_id = sqlc.arg (project_id)::UUID
ORDER BY
  created_at DESC;
