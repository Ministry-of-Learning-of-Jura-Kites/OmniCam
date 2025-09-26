-- name: UpdateModelCams :one
UPDATE "model"
SET
  cameras = sqlc.arg (value)::JSONB,
  version = version + 1,
  updated_at = NOW()
WHERE
  id = sqlc.arg (model_id)::UUID
RETURNING
  version;
