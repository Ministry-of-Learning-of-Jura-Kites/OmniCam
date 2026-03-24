-- name: CreateUser :one
INSERT INTO
  "user" (email, first_name, last_name, username, password)
VALUES
  (
    SQLC.ARG(email),
    SQLC.ARG(first_name),
    SQLC.ARG(last_name),
    SQLC.ARG(username),
    SQLC.ARG(password)
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
