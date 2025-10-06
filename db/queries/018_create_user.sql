-- name: CreateUser :one
INSERT INTO
  "user" (email, name, surname, username, password)
VALUES
  (
    sqlc.arg (email),
    sqlc.arg (name),
    sqlc.arg (surname),
    sqlc.arg (username),
    sqlc.arg (password)
  )
RETURNING
  id,
  email,
  name,
  password,
  username,
  surname,
  created_at,
  updated_at;
