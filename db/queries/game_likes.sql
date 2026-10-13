-- name: InsertGameLike :exec
INSERT INTO game_likes (user_id, game_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: DeleteGameLike :exec
DELETE FROM game_likes WHERE user_id = $1 AND game_id = $2;

-- name: IncrementLikeCount :exec
UPDATE games SET like_count = like_count + 1 WHERE id = $1;

-- name: DecrementLikeCount :exec
UPDATE games SET like_count = GREATEST(like_count - 1, 0) WHERE id = $1;
