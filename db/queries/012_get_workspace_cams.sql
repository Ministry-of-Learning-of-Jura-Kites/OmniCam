-- name: GetModelWorkspaceCamsByID :one
SELECT
  name,
  file_path,
  description,
  cameras,
  version,
  created_at,
  updated_at
FROM
  "user_model_workspace"
WHERE
  user_id = sqlc.arg (user_id)::UUID
  AND model_id = sqlc.arg (model_id)::UUID;
