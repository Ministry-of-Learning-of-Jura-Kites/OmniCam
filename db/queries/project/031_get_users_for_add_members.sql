-- name: GetUsersForAddMembers :many
SELECT
  u.id,
  u.username,
  u.email,
  u.first_name,
  u.last_name,
  u.created_at,
  u.updated_at
FROM
  "user" u
WHERE
  (
    u.username ILIKE '%' || SQLC.ARG(search)::TEXT || '%'
  )
  AND NOT EXISTS (
    SELECT
      1
    FROM
      user_to_project up
    WHERE
      up.user_id = u.id
      AND up.project_id = SQLC.ARG(project_id)
  )
ORDER BY
  u.created_at DESC
LIMIT
  SQLC.ARG(page_size)::INT
OFFSET
  SQLC.ARG(page_offset)::INT;
