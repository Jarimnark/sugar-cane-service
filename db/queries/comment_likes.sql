-- name: InsertCommentLike :exec
INSERT INTO comment_likes (user_id, comment_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: DeleteCommentLike :exec
DELETE FROM comment_likes WHERE user_id = $1 AND comment_id = $2;

-- name: IncrementCommentLikeCount :exec
UPDATE game_comments SET like_count = like_count + 1 WHERE id = $1;

-- name: DecrementCommentLikeCount :exec
UPDATE game_comments SET like_count = GREATEST(like_count - 1, 0) WHERE id = $1;
