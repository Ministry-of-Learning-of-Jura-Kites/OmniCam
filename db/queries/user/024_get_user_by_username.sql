-- name: GetUserByUsername :one
SELECT
  *
FROM
  "user"
WHERE
  username = $1
LIMIT
  1;
