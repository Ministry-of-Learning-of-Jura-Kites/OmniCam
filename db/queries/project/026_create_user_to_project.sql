-- name: CreateUserToProject :exec
INSERT INTO
  user_to_project (project_id, user_id, role)
VALUES
  ($1, $2, $3);
