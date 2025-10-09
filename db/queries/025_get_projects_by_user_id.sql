-- name: GetProjectsByUserId :many
SELECT
  p.id,
  p.name,
  p.description,
  p.image_path,
  p.created_at,
  p.updated_at
FROM
  project p
  INNER JOIN user_to_project up ON p.id = up.project_id
WHERE
  up.user_id = sqlc.arg (user_id)
ORDER BY
  p.created_at DESC
LIMIT
  sqlc.arg (page_size)::INT
OFFSET
  sqlc.arg (page_offset)::INT;
