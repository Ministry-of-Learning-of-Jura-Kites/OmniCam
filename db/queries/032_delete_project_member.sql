-- name: DeleteProjectMember :exec
DELETE FROM user_to_project
WHERE
  user_id = sqlc.arg (user_id)
  AND project_id = sqlc.arg (project_id);
