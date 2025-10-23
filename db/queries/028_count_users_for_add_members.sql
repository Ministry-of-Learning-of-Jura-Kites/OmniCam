-- name: CountUsersForAddMembers :one
SELECT
  COUNT(*)
FROM
  "user" u
WHERE
  (
    u.username ILIKE '%' || SQLC.ARG(search)::TEXT || '%'
  )
  AND NOT EXISTS (
    SELECT
      1
    FROM
      user_to_project up
    WHERE
      up.user_id = u.id
      AND up.project_id = SQLC.ARG(project_id)
  );
