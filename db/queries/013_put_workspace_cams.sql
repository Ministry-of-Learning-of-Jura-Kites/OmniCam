-- name: UpdateWorkspaceCams :exec
UPDATE "user_model_workspace"
SET
  cameras = sqlc.arg (cameras)::JSONB,
  version = version + 1,
  updated_at = NOW()
WHERE
  user_id = sqlc.arg (user_id)::UUID
  AND model_id = sqlc.arg (model_id)::UUID;
