-- name: GetModelByID :one
SELECT
  id,
  project_id,
  name,
  file_path,
  image_path,
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
  "model"
WHERE
  id = sqlc.arg (id)::UUID;
