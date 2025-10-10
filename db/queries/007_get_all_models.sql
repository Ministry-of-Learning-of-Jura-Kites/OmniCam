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
  project_id = SQLC.ARG(project_id)::UUID
ORDER BY
  created_at DESC
LIMIT
  SQLC.ARG(page_size)::INT
OFFSET
  SQLC.ARG(page_offset)::INT;
