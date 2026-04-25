-- name: GetUserWithModel :one
SELECT
  u.id,
  u.email,
  u.first_name,
  u.last_name,
  u.username,
  u.created_at,
  u.updated_at,
  utp.role
FROM
  "user" u
  LEFT JOIN "user_to_project" utp ON u.id = utp.user_id
  LEFT JOIN "model" m ON m.project_id = utp.project_id
WHERE
  (
    COALESCE(u.username = SQLC.NARG(username), FALSE)
    OR COALESCE(u.email = SQLC.NARG(email), FALSE)
    OR COALESCE(utp.user_id = SQLC.NARG(user_id), FALSE)
  )
  AND m.id = SQLC.ARG(model_id);
