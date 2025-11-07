-- name: UpdateWorkspaceCams :one
UPDATE "user_model_workspace"
SET
  cameras = CASE
    WHEN SQLC.NARG(value)::JSONB IS NULL THEN cameras - SQLC.ARG(key)::TEXT[] -- delete key if value is NULL
    ELSE JSONB_SET(
      cameras,
      SQLC.ARG(key)::TEXT[],
      SQLC.NARG(value)::JSONB,
      TRUE
    ) -- upsert key
  END,
  version = version + 1,
  updated_at = NOW()
WHERE
  user_id = SQLC.ARG(user_id)::UUID
  AND model_id = SQLC.ARG(model_id)::UUID
RETURNING
  version;
