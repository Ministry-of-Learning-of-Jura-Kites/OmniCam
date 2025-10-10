-- name: GetUserByIdentifier :one
SELECT
  id,
  email,
  username,
  first_name,
  last_name,
  password,
  created_at,
  updated_at
FROM
  "user"
WHERE
  email = SQLC.ARG(identifier)
  OR username = SQLC.ARG(identifier)
LIMIT
  1;
