-- name: CreateWorkspace :one
INSERT INTO
  "user_model_workspace" (
    user_id,
    model_id,
    cameras,
    base_cameras,
    version,
    base_version,
    created_at,
    updated_at
  )
SELECT
  sqlc.arg (user_id)::UUID,
  sqlc.arg (model_id)::UUID,
  cameras,
  cameras,
  version,
  version,
  NOW(),
  NOW()
FROM
  "model"
WHERE
  id = sqlc.arg (model_id)::UUID
RETURNING
  user_id,
  model_id,
  cameras,
  base_cameras,
  version,
  base_version,
  created_at,
  updated_at;
