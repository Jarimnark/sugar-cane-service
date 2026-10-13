-- name: CreateGame :one
INSERT INTO games (owner_id, slug, name, dev, pitch, description, bg, players, tags)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
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
SET name        = $2,
    pitch       = $3,
    description = $4,
    link_web    = $5,
    link_android= $6,
    link_ios    = $7,
    bg          = $8,
    players     = $9,
    tags        = $10,
    updated_at  = NOW()
WHERE id = $1 AND owner_id = $11
RETURNING *;

-- name: PublishGame :one
UPDATE games SET status = 'published', updated_at = NOW()
WHERE id = $1 AND owner_id = $2
RETURNING *;

-- name: DeleteGame :exec
DELETE FROM games WHERE id = $1 AND owner_id = $2;
