-- +goose Up
ALTER TABLE game_comments ALTER COLUMN user_id DROP NOT NULL;

-- +goose Down
ALTER TABLE game_comments ALTER COLUMN user_id SET NOT NULL;
