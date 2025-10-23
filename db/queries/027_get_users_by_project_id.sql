-- name: GetUsersByProjectId :many
SELECT
  u.id,
  u.username,
  u.email,
  u.first_name,
  u.last_name,
  u.created_at,
  u.updated_at,
  up.role
FROM
  "user" u
  INNER JOIN user_to_project up ON u.id = up.user_id
WHERE
  up.project_id = SQLC.ARG(project_id)::UUID
ORDER BY
  u.created_at DESC;
