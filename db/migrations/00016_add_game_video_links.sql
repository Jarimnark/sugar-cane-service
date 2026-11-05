-- +goose Up
ALTER TABLE games ADD COLUMN video_links TEXT[] NOT NULL DEFAULT '{}';

GRANT UPDATE ON games TO app;

-- +goose Down
ALTER TABLE games DROP COLUMN video_links;
