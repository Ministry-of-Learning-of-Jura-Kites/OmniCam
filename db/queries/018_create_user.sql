-- name: CreateUser :one
INSERT INTO
  "user" (email, name, password)
VALUES
  (
    sqlc.arg (email),
    sqlc.arg (name),
    sqlc.arg (password)
  )
RETURNING
  id,
  email,
  name,
  password,
  created_at,
  updated_at;
