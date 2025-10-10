-- name: CountUsersForAddMembers :one
SELECT
  COUNT(*)
FROM
  "user" u
WHERE
  (
    u.username ILIKE '%' || sqlc.arg (search)::TEXT || '%'
  )
  AND NOT EXISTS (
    SELECT
      1
    FROM
      user_to_project up
    WHERE
      up.user_id = u.id
      AND up.project_id = sqlc.arg (project_id)
  );
