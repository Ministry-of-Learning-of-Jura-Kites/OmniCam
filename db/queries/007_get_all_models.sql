-- name: GetAllfdfModels :many
SELECT
  id,
  project_id,
  name,
  description,
  version,
  image_path,
  file_path,
  created_at,
  updated_at
FROM
  "model"
WHERE
  project_id = sqlc.arg (project_id)::UUID
ORDER BY
  created_at DESC
LIMIT
  sqlc.arg (page_size)::INT
OFFSET
  sqlc.arg (page_offset)::INT;
