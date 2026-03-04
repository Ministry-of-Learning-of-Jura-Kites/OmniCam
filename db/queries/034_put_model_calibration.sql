-- name: UpdateModelCalibration :one
UPDATE "model"
SET
  scale_factor = SQLC.ARG(scale_factor)::FLOAT,
  model_height = SQLC.ARG(model_height)::FLOAT,
  updated_at = NOW()
WHERE
  id = SQLC.ARG(model_id)::UUID
RETURNING
  scale_factor,
  model_height;
