-- name: CreateUser :one
INSERT INTO
  "user" (email, first_name, last_name, username, password)
VALUES
  (
    sqlc.arg (email),
    sqlc.arg (first_name),
    sqlc.arg (last_name),
    sqlc.arg (username),
    sqlc.arg (password)
  )
RETURNING
  id,
  email,
  first_name,
  last_name,
  username,
  password,
  created_at,
  updated_at;
