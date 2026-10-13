-- name: InsertSavedGame :exec
INSERT INTO saved_games (user_id, game_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: DeleteSavedGame :exec
DELETE FROM saved_games WHERE user_id = $1 AND game_id = $2;

-- name: ListSavedGamesByUser :many
SELECT g.* FROM saved_games sg
JOIN games g ON g.id = sg.game_id
WHERE sg.user_id = $1
ORDER BY sg.saved_at DESC;
