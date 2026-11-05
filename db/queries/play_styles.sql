-- name: ListPlayStyles :many
SELECT id, name, slug FROM play_styles ORDER BY sort ASC;

-- name: GetPlayStyleBySlug :one
SELECT id, name, slug FROM play_styles WHERE slug = $1;
