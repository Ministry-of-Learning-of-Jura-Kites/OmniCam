-- name: CountProjects :one
SELECT
  COUNT(*) AS total
FROM
  "project";
