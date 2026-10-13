-- name: CreateUser :one
INSERT INTO users (email, password_hash, display_name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET display_name = $2,
    bio          = $3,
    photo_url    = $4,
    updated_at   = NOW()
WHERE id = $1
RETURNING *;

-- name: VerifyUserEmail :exec
UPDATE users SET email_verified = TRUE, updated_at = NOW() WHERE id = $1;
