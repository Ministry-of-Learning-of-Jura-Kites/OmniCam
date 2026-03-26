-- name: GetWorkspaceByID :one
SELECT
  SQLC.EMBED(m),
  CASE
    WHEN 'cameras' = ANY (COALESCE(SQLC.NARG(fields)::TEXT[], '{}'::TEXT[])) THEN umw.cameras::JSONB
    ELSE NULL::JSONB
  END AS cameras,
  CASE
    WHEN 'base_cameras' = ANY (COALESCE(SQLC.NARG(fields)::TEXT[], '{}'::TEXT[])) THEN base_cameras::JSONB
    ELSE NULL::JSONB
  END AS base_cameras,
  CASE
    WHEN 'target_area_trapezoids' = ANY (COALESCE(SQLC.NARG(fields)::TEXT[], '{}'::TEXT[])) THEN umw.target_area_trapezoids::JSONB
    ELSE NULL::JSONB
  END AS target_area_trapezoids,
  umw.scale_factor,
  umw.model_height,
  umw.version,
  base_version,
  umw.created_at,
  umw.updated_at
FROM
  "user_model_workspace" AS umw
  LEFT JOIN "model" AS m ON m.id = umw.model_id
WHERE
  user_id = SQLC.ARG(user_id)::UUID
  AND model_id = SQLC.ARG(model_id)::UUID;
