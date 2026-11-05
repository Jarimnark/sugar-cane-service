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
WHERE status = 'published'
  AND $1::text <> ''
  AND name ILIKE '%' || $1::text || '%'
ORDER BY like_count DESC LIMIT 10;

-- name: ListGamesFiltered :many
SELECT g.* FROM games g
WHERE g.status = 'published'
  AND ($1::uuid IS NULL OR EXISTS (
    SELECT 1 FROM game_genres gg WHERE gg.game_id = g.id AND gg.genre_id = $1::uuid
  ))
  AND ($2::uuid IS NULL OR EXISTS (
    SELECT 1 FROM game_platforms gp WHERE gp.game_id = g.id AND gp.platform_id = $2::uuid
  ))
  AND ($6::text = '' OR g.name ILIKE '%' || $6::text || '%')
ORDER BY
  CASE WHEN $3::text = 'popular'    THEN g.visit_count END DESC NULLS LAST,
  CASE WHEN $3::text = 'most_liked' THEN g.like_count  END DESC NULLS LAST,
  CASE WHEN $3::text NOT IN ('popular','most_liked') THEN g.created_at END DESC NULLS LAST
LIMIT $4::int OFFSET $5::int;

-- name: GetRelatedGames :many
SELECT * FROM games
WHERE status = 'published' AND id != $1::uuid AND tags && $2::text[]
ORDER BY like_count DESC LIMIT 6;
