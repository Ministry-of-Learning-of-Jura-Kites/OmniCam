-- name: GetUserByIdentifier :one
SELECT
  id,
  email,
  name,
  surname,
  username,
  password,
  created_at,
  updated_at
FROM
  "user"
WHERE
  email = sqlc.arg (identifier)
  OR username = sqlc.arg (identifier)
LIMIT
  1;
