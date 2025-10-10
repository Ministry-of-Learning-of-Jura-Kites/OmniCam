-- name: CountModels :one
SELECT
  COUNT(*)::BIGINT
FROM
  "model"
WHERE
  project_id = SQLC.ARG(project_id)::UUID;
