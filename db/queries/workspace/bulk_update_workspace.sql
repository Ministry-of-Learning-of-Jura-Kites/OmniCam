-- name: BulkUpdateWorkspace :one
UPDATE "user_model_workspace"
SET
  -- Handle Cameras
  cameras = (cameras - SQLC.ARG(cam_deletes)::TEXT[]) || SQLC.ARG(cam_upserts)::JSONB,
  -- Handle Trapezoids
  target_area_trapezoids = (
    target_area_trapezoids - SQLC.ARG(trap_deletes)::TEXT[]
  ) || SQLC.ARG(trap_upserts)::JSONB,
  -- Calibration (Always updated or passed back as current values)
  scale_factor = SQLC.ARG(scale_factor)::FLOAT,
  model_height = SQLC.ARG(model_height)::FLOAT,
  version = version + 1,
  updated_at = NOW()
WHERE
  user_id = SQLC.ARG(user_id)::UUID
  AND model_id = SQLC.ARG(model_id)::UUID
RETURNING
  version;
