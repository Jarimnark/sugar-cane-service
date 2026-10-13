-- name: GetHomeTrending :many
SELECT * FROM games WHERE status = 'published' ORDER BY visit_count DESC LIMIT 5;

-- name: GetHomePopular :many
SELECT * FROM games WHERE status = 'published' ORDER BY like_count DESC LIMIT 6;

-- name: GetHomeNew :many
SELECT * FROM games WHERE status = 'published' ORDER BY created_at DESC LIMIT 5;

-- name: IncrementVisitCount :exec
UPDATE games SET visit_count = visit_count + 1 WHERE id = $1;

-- name: SuggestGames :many
SELECT id, slug, name, dev, bg FROM games
WHERE status = 'published' AND name ILIKE '%' || $1::text || '%'
ORDER BY like_count DESC LIMIT 10;

-- name: ListGamesFiltered :many
SELECT DISTINCT g.* FROM games g
LEFT JOIN game_genres gg ON gg.game_id = g.id
LEFT JOIN game_platforms gp ON gp.game_id = g.id
WHERE g.status = 'published'
  AND ($1::uuid IS NULL OR gg.genre_id = $1::uuid)
  AND ($2::uuid IS NULL OR gp.platform_id = $2::uuid)
ORDER BY
  CASE WHEN $3::text = 'popular' THEN g.like_count END DESC,
  CASE WHEN $3::text != 'popular' THEN g.created_at END DESC
LIMIT $4::int OFFSET $5::int;

-- name: GetRelatedGames :many
SELECT * FROM games
WHERE status = 'published' AND id != $1::uuid AND tags && $2::text[]
ORDER BY like_count DESC LIMIT 6;
