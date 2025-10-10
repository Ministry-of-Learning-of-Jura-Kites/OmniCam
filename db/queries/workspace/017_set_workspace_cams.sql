-- name: UpdateSetWorkspaceCams :exec
UPDATE "user_model_workspace"
SET
  cameras = SQLC.ARG(cameras)::JSONB,
  base_cameras = SQLC.ARG(base_cameras)::JSONB,
  version = SQLC.ARG(base_version)::INT,
  base_version = SQLC.ARG(base_version)::INT,
  updated_at = NOW()
WHERE
  model_id = SQLC.ARG(model_id)::UUID
  AND user_id = SQLC.ARG(user_id)::UUID;
