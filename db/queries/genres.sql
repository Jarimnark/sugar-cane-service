-- name: ListGenres :many
SELECT * FROM genres ORDER BY name;

-- name: CountGamesByGenre :many
SELECT g.id, g.name, g.slug, COUNT(gg.game_id)::int AS game_count
FROM genres g
LEFT JOIN game_genres gg ON gg.genre_id = g.id
GROUP BY g.id ORDER BY game_count DESC;
