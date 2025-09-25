-- name: CountModels :one
SELECT
  COUNT(*)::BIGINT
FROM
  "model"
WHERE
  project_id = sqlc.arg (project_id)::UUID;
