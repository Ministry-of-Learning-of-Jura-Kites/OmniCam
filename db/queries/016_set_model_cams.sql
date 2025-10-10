-- name: UpdateModelCams :one
UPDATE "model"
SET
  cameras = SQLC.ARG(value)::JSONB,
  version = version + 1,
  updated_at = NOW()
WHERE
  id = SQLC.ARG(model_id)::UUID
RETURNING
  version;
