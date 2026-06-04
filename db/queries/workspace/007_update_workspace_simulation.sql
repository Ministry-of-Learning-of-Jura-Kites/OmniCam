-- name: UpdateWorkspaceSimulation :one
UPDATE user_model_workspace
SET
  simulation = SQLC.ARG(simulation)::JSONB,
  updated_at = NOW()
WHERE
  user_id = SQLC.ARG(user_id)::UUID
  AND model_id = SQLC.ARG(model_id)::UUID
RETURNING
  version;
