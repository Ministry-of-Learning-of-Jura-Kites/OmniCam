-- name: PostProjectMembers :exec
INSERT INTO
  user_to_project (project_id, user_id, role)
SELECT
  SQLC.ARG(project_id)::UUID,
  UNNEST(SQLC.ARG(user_ids)::UUID[]),
  UNNEST(SQLC.ARG(roles)::TEXT[])::role
ON CONFLICT (project_id, user_id) DO UPDATE
SET ROLE = excluded.role;
