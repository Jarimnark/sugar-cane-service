-- name: DeleteOAuthAccountByUserAndProvider :exec
DELETE FROM oauth_accounts WHERE user_id = $1 AND provider = $2;

-- name: SoftDeleteUser :exec
UPDATE users SET deleted_at = NOW() WHERE id = $1;

-- name: ListOAuthAccountsByUser :many
SELECT * FROM oauth_accounts WHERE user_id = $1 ORDER BY created_at;

-- name: ListUserSessions :many
SELECT * FROM sessions
WHERE user_id = $1 AND expires_at > NOW()
ORDER BY last_used_at DESC;

-- name: DeleteUserSessionsExcept :exec
DELETE FROM sessions WHERE user_id = $1 AND id <> $2;
