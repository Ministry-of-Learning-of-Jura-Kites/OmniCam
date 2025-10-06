-- name: GetAllUser :many
SELECT
  id,
  email,
  username,
  first_name,
  last_name,
  created_at,
  updated_at
FROM
  "user";
