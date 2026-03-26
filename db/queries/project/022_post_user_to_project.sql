-- name: AddUserToProject :one
INSERT INTO
  user_to_project (user_id, project_id, role)
VALUES
  (
    SQLC.ARG(user_id),
    SQLC.ARG(project_id),
    SQLC.ARG(role)
  )
RETURNING
  user_id,
  project_id,
  role;
