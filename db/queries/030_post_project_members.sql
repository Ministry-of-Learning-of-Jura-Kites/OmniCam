-- name: PostProjectMembers :exec
INSERT INTO
  user_to_project (project_id, user_id, role)
VALUES
  ($1, $2, $3)
ON CONFLICT (project_id, user_id) DO UPDATE
SET ROLE = excluded.role;
