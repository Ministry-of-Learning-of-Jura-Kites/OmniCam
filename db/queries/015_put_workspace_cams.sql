-- name: UpdateWorkspaceCams :exec
UPDATE "user_model_workspace"
SET
  cameras = CASE
    WHEN sqlc.narg (value)::JSONB IS NULL THEN cameras - sqlc.arg (key)::TEXT[] -- delete key if value is NULL
    ELSE JSONB_SET(
      cameras,
      sqlc.arg (key)::TEXT[],
      sqlc.narg (value)::JSONB,
      TRUE
    ) -- upsert key
  END,
  version = version + 1,
  updated_at = NOW()
WHERE
  user_id = sqlc.arg (user_id)::UUID
  AND model_id = sqlc.arg (model_id)::UUID;
