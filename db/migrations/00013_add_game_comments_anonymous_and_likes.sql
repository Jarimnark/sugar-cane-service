-- +goose Up
ALTER TABLE game_comments
    ADD COLUMN anonymous  BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN like_count INT     NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE game_comments
    DROP COLUMN anonymous,
    DROP COLUMN like_count;
