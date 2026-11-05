-- name: CreateGame :one
INSERT INTO games (owner_id, slug, name, dev, pitch, description, bg, players, tags, video_links, screenshots, banner_url, play_style)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetGameByID :one
SELECT * FROM games WHERE id = $1;

-- name: GetGameBySlug :one
SELECT * FROM games WHERE slug = $1;

-- name: ListPublishedGames :many
SELECT * FROM games WHERE status = 'published' ORDER BY created_at DESC;

-- name: ListGamesByOwner :many
SELECT * FROM games WHERE owner_id = $1 ORDER BY created_at DESC;

-- name: UpdateGame :one
UPDATE games
SET name         = $2,
    pitch        = $3,
    description  = $4,
    link_web     = $5,
    link_android = $6,
    link_ios     = $7,
    bg           = $8,
    players      = $9,
    tags         = $10,
    video_links  = $11,
    screenshots  = $12,
    banner_url   = $13,
    play_style   = $14,
    dev          = $15,
    updated_at   = NOW()
WHERE id = $1 AND owner_id = $16
RETURNING *;

-- name: PublishGame :one
UPDATE games SET status = 'published', updated_at = NOW()
WHERE id = $1 AND owner_id = $2
RETURNING *;

-- name: DeleteGame :exec
DELETE FROM games WHERE id = $1 AND owner_id = $2;

-- name: DeleteGameGenres :exec
DELETE FROM game_genres WHERE game_id = $1;

-- name: InsertGameGenres :exec
INSERT INTO game_genres (game_id, genre_id)
SELECT $1, unnest($2::uuid[]);

-- name: DeleteGamePlatforms :exec
DELETE FROM game_platforms WHERE game_id = $1;

-- name: InsertGamePlatforms :exec
INSERT INTO game_platforms (game_id, platform_id)
SELECT $1, unnest($2::uuid[]);

-- name: GetGameGenres :many
SELECT g.id, g.name, g.slug FROM genres g
JOIN game_genres gg ON gg.genre_id = g.id
WHERE gg.game_id = $1;

-- name: GetGamePlatforms :many
SELECT p.id, p.name, p.slug FROM platforms p
JOIN game_platforms gp ON gp.platform_id = p.id
WHERE gp.game_id = $1;
