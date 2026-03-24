-- name: GetUser :one
SELECT
  id,
  username,
  email,
  first_name,
  last_name
FROM
  "user"
WHERE
  id = $1;
