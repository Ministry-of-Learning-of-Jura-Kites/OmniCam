-- name: CreateWorkspace :one
INSERT INTO
  "user_model_workspace" (
    user_id,
    model_id,
    cameras,
    base_cameras,
    scale_factor,
    model_height,
    version,
    base_version,
    created_at,
    updated_at
  )
SELECT
  SQLC.ARG(user_id)::UUID,
  SQLC.ARG(model_id)::UUID,
  cameras,
  cameras,
  1.0,
  0.0,
  version,
  version,
  NOW(),
  NOW()
FROM
  "model"
WHERE
  id = SQLC.ARG(model_id)::UUID
RETURNING
  user_id,
  model_id,
  cameras,
  base_cameras,
  scale_factor,
  model_height,
  version,
  base_version,
  created_at,
  updated_at;
