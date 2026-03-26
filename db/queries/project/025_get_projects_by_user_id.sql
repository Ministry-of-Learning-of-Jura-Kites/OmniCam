-- name: GetProjectsByUserId :many
SELECT
  p.id,
  p.name,
  p.description,
  p.image_path,
  p.image_extension,
  p.created_at,
  p.updated_at
FROM
  project p
  INNER JOIN user_to_project up ON p.id = up.project_id
WHERE
  up.user_id = SQLC.ARG(user_id)
ORDER BY
  p.created_at DESC
LIMIT
  SQLC.ARG(page_size)::INT
OFFSET
  SQLC.ARG(page_offset)::INT;
