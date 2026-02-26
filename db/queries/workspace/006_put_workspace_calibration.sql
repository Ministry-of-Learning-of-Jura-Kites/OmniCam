-- name: UpdateWorkspaceCalibration :one
UPDATE "user_model_workspace"
SET
  scale_factor = SQLC.ARG(scale_factor)::FLOAT,
  model_height = SQLC.ARG(model_height)::FLOAT,
  version = version + 1,
  updated_at = NOW()
WHERE
  user_id = SQLC.ARG(user_id)::UUID
  AND model_id = SQLC.ARG(model_id)::UUID
RETURNING
  version,
  scale_factor,
  model_height;
