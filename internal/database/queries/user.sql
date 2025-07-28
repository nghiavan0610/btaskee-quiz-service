-- name: GetUserDetail :one
SELECT id, username, email, is_active, created_at, updated_at, last_login_at
FROM users
WHERE id = $1;

-- name: FindValidateName :one
SELECT id, username
FROM users
WHERE username = $1
  AND (sqlc.narg('exclude_id')::bigint IS NULL OR id != sqlc.narg('exclude_id'))
LIMIT 1;

-- name: UpdateUser :one
UPDATE users 
SET username = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, username, email, is_active, created_at, updated_at, last_login_at;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: UpdateUserLastLogin :one
UPDATE users 
SET last_login_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING id, username, email, is_active, created_at, updated_at, last_login_at;
