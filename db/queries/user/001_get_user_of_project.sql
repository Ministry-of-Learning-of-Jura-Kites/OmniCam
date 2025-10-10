-- name: GetUserOfProject :one
SELECT
  id,
  email,
  first_name,
  last_name,
  username,
  created_at,
  updated_at,
  utp.role
FROM
  "user" u
  LEFT JOIN "user_to_project" utp ON u.id = utp.user_id
WHERE
  (
    COALESCE(u.username = SQLC.NARG(username), FALSE)
    OR COALESCE(u.email = SQLC.NARG(email), FALSE)
    OR COALESCE(utp.user_id = SQLC.NARG(user_id), FALSE)
  )
  AND project_id = SQLC.ARG(projectid);
