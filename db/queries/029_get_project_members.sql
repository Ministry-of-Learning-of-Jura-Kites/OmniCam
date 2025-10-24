-- name: GetProjectMembers :many
SELECT
  up.project_id,
  up.user_id,
  up.role,
  u.username,
  u.email,
  u.first_name,
  u.last_name,
  u.created_at
FROM
  user_to_project AS up
  JOIN "user" AS u ON up.user_id = u.id
WHERE
  up.project_id = $1
ORDER BY
  u.created_at ASC;
