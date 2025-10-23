-- name: DeleteProjectMember :exec
DELETE FROM user_to_project
WHERE
  user_id = SQLC.ARG(user_id)
  AND project_id = SQLC.ARG(project_id);
