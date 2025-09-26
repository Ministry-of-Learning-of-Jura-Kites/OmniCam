-- name: GetAllUser :many
SELECT
  id,
  name,
  email,
  created_at,
  updated_at
FROM
  "user";
