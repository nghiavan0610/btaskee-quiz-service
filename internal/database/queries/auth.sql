-- name: FindEmailSignUp :one
SELECT id, email
FROM users
WHERE email = $1
LIMIT 1;

-- name: FindUsernameSignUp :one
SELECT id, username
FROM users
WHERE username = $1
LIMIT 1;

-- name: CheckEmailOrUsernameExists :one
SELECT 
    id,
    CASE 
        WHEN email = $1 THEN 'email'
        WHEN username = $2 THEN 'username'
    END as conflict_field
FROM users
WHERE email = $1 OR username = $2
LIMIT 1;

-- name: RegisterAccount :one
INSERT INTO users (
  username,
  email,
  password,
  is_active
) VALUES (
  $1, $2, $3, true
)
RETURNING id, username, email, is_active, created_at, updated_at, last_login_at;

-- name: GetUserByEmailIncludePassword :one
SELECT *
FROM users
WHERE email = $1
LIMIT 1;