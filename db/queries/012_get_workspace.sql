-- name: GetWorkspaceByID :one
SELECT
  name,
  file_path,
  description,
  CASE
    WHEN 'cameras' = ANY (
      COALESCE(sqlc.narg (fields)::TEXT[], '{}'::TEXT[])
    ) THEN cameras::JSONB
    ELSE NULL::JSONB
  END AS cameras,
  version,
  created_at,
  updated_at
FROM
  "user_model_workspace"
WHERE
  user_id = sqlc.arg (user_id)::UUID
  AND model_id = sqlc.arg (model_id)::UUID;
