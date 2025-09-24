-- name: GetAllModels :many
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
  project_id = $1::UUID
ORDER BY
  created_at DESC
LIMIT
  $2::INT
OFFSET
  $3::INT;
