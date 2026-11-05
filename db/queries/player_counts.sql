-- name: ListPlayerCounts :many
SELECT id, name, slug FROM player_counts ORDER BY sort ASC;

-- name: GetPlayerCountBySlug :one
SELECT id, name, slug FROM player_counts WHERE slug = $1;
