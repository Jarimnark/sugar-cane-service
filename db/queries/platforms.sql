-- name: ListPlatforms :many
SELECT * FROM platforms ORDER BY name;

-- name: CountGamesByPlatform :many
SELECT p.id, p.name, p.slug, COUNT(gp.game_id)::int AS game_count
FROM platforms p
LEFT JOIN game_platforms gp ON gp.platform_id = p.id
GROUP BY p.id ORDER BY game_count DESC;
