-- name: GameExists :one
SELECT EXISTS(SELECT 1 FROM games WHERE id = $1);

-- name: CountCommentsForGame :one
SELECT COUNT(*)::bigint FROM game_comments WHERE game_id = $1;

-- name: ListCommentsForGameNewest :many
SELECT
    c.id,
    c.game_id,
    c.user_id,
    c.content,
    c.anonymous,
    c.like_count,
    c.created_at,
    COALESCE(u.display_name, '')::text AS author_name,
    u.photo_url    AS author_photo_url,
    g.owner_id     AS game_owner_id,
    (cl.user_id IS NOT NULL)::bool AS liked
FROM game_comments c
JOIN games g      ON g.id = c.game_id
LEFT JOIN users u ON u.id = c.user_id
LEFT JOIN comment_likes cl ON cl.comment_id = c.id AND cl.user_id = $2::uuid
WHERE c.game_id = $1
ORDER BY c.created_at DESC
LIMIT $3::int OFFSET $4::int;

-- name: ListCommentsForGameTop :many
SELECT
    c.id,
    c.game_id,
    c.user_id,
    c.content,
    c.anonymous,
    c.like_count,
    c.created_at,
    COALESCE(u.display_name, '')::text AS author_name,
    u.photo_url    AS author_photo_url,
    g.owner_id     AS game_owner_id,
    (cl.user_id IS NOT NULL)::bool AS liked
FROM game_comments c
JOIN games g      ON g.id = c.game_id
LEFT JOIN users u ON u.id = c.user_id
LEFT JOIN comment_likes cl ON cl.comment_id = c.id AND cl.user_id = $2::uuid
WHERE c.game_id = $1
ORDER BY c.like_count DESC, c.created_at DESC
LIMIT $3::int OFFSET $4::int;

-- name: GetCommentByID :one
SELECT * FROM game_comments WHERE id = $1;

-- name: GetCommentWithAuthor :one
SELECT
    c.id,
    c.game_id,
    c.user_id,
    c.content,
    c.anonymous,
    c.like_count,
    c.created_at,
    COALESCE(u.display_name, '')::text AS author_name,
    u.photo_url    AS author_photo_url,
    g.owner_id     AS game_owner_id
FROM game_comments c
JOIN games g      ON g.id = c.game_id
LEFT JOIN users u ON u.id = c.user_id
WHERE c.id = $1;

-- name: InsertComment :one
INSERT INTO game_comments (game_id, user_id, content, anonymous)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM game_comments WHERE id = $1 AND user_id = $2;
