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
    u.username ILIKE '%' || sqlc.arg (search)::TEXT || '%'
  )
  AND NOT EXISTS (
    SELECT
      1
    FROM
      user_to_project up
    WHERE
      up.user_id = u.id
      AND up.project_id = sqlc.arg (project_id)
  )
ORDER BY
  u.created_at DESC
LIMIT
  sqlc.arg (page_size)::INT
OFFSET
  sqlc.arg (page_offset)::INT;
