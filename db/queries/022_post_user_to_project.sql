-- name: AddUserToProject :one
INSERT INTO
  user_to_project (user_id, project_id, role)
VALUES
  (
    sqlc.arg (user_id),
    sqlc.arg (project_id),
    sqlc.arg (role)
  )
RETURNING
  user_id,
  project_id,
  role;
