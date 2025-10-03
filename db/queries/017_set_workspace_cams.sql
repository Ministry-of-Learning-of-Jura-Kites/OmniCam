-- name: UpdateSetWorkspaceCams :exec
UPDATE "user_model_workspace"
SET
  cameras = sqlc.arg (cameras)::JSONB,
  base_cameras = sqlc.arg (base_cameras)::JSONB,
  version = sqlc.arg (base_version)::INT,
  base_version = sqlc.arg (base_version)::INT,
  updated_at = NOW()
WHERE
  model_id = sqlc.arg (model_id)::UUID
  AND user_id = sqlc.arg (user_id)::UUID;
