-- +goose Up
ALTER TABLE games ADD COLUMN screenshots TEXT[] NOT NULL DEFAULT '{}';

GRANT UPDATE ON games TO app;

-- +goose Down
ALTER TABLE games DROP COLUMN screenshots;
