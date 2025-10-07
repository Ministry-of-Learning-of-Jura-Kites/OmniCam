-- name: GetModelByID :one
SELECT
  id,
  project_id,
  m.name,
  m.file_path,
  m.image_path,
  m.description,
  CASE
    WHEN 'cameras' = ANY (
      COALESCE(sqlc.narg (fields)::TEXT[], '{}'::TEXT[])
    ) THEN m.cameras::JSONB
    ELSE NULL::JSONB
  END AS cameras,
  (umw.model_id IS NOT NULL)::BOOLEAN AS workspace_exists,
  m.version,
  m.created_at,
  m.updated_at
FROM
  "model" AS m
  LEFT JOIN "user_model_workspace" AS umw ON m.id = umw.model_id
WHERE
  id = sqlc.arg (id)::UUID
  AND COALESCE(umw.user_id = sqlc.narg (user_id)::UUID, TRUE);
