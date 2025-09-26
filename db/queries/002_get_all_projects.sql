-- name: GetAllProjects :many
SELECT
  id,
  name,
  description,
  image_path,
  created_at,
  updated_at
FROM
  "project"
ORDER BY
  created_at ASC
LIMIT
  sqlc.arg (page_size)::INT
OFFSET
  sqlc.arg (page_offset)::INT;
