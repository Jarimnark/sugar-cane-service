-- name: DeleteOAuthAccountByUserAndProvider :exec
DELETE FROM oauth_accounts WHERE user_id = $1 AND provider = $2;

-- name: SoftDeleteUser :exec
UPDATE users SET deleted_at = NOW() WHERE id = $1;
