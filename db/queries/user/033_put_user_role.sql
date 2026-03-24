-- name: PutUserRole :exec
UPDATE user_to_project
SET ROLE = $1
WHERE
  project_id = $2
  AND user_id = $3;
