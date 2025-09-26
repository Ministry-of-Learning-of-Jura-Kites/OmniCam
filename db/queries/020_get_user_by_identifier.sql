-- name: GetUserByIdentifier :one
SELECT
  id,
  email,
  name,
  password,
  created_at,
  updated_at
FROM
  "user"
WHERE
  email = sqlc.arg (identifier)
  OR name = sqlc.arg (identifier)
LIMIT
  1;
