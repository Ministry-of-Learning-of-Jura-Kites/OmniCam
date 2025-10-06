-- name: GetAllUser :many
SELECT
  id,
  name,
  email,
  username,
  surname,
  created_at,
  updated_at
FROM
  "user";
