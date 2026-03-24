-- name: GetUserWithProjects :many
SELECT
  u.id AS user_id,
  u.email,
  u.username,
  u.first_name,
  u.last_name,
  u.password,
  u.created_at,
  u.updated_at,
  p.id AS project_id,
  p.name AS project_name,
  p.description AS project_description,
  p.image_path AS project_image_path,
  p.image_extension AS project_image_extension,
  utp.role AS role
FROM
  "user" u
  LEFT JOIN user_to_project utp ON u.id = utp.user_id
  LEFT JOIN project p ON p.id = utp.project_id
WHERE
  u.email = SQLC.ARG(email);
