-- name: GetModelsByProjectID :many
SELECT
  *
FROM
  "model"
WHERE
  project_id = $1
ORDER BY
  created_at DESC;
