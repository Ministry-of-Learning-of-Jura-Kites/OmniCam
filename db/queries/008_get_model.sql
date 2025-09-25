-- name: GetModelByID :one
SELECT
  id,
  project_id,
  name,
  file_path,
  image_path,
  description,
  version,
  created_at,
  updated_at
FROM
  "model"
WHERE
  id = sqlc.arg (id)::UUID;
