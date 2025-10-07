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
  SQLC.ARG(page_size)::INT
OFFSET
  SQLC.ARG(page_offset)::INT;
