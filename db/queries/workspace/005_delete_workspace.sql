-- name: DeleteWorkspace :exec
DELETE FROM "user_model_workspace"
WHERE
  user_id = SQLC.ARG(user_id)::UUID
  AND model_id = SQLC.ARG(model_id)::UUID;
