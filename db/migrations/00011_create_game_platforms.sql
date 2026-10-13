-- +goose Up
CREATE TABLE game_platforms (
    game_id     UUID NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    platform_id UUID NOT NULL REFERENCES platforms (id) ON DELETE CASCADE,
    PRIMARY KEY (game_id, platform_id)
);

CREATE INDEX idx_game_platforms_platform_id ON game_platforms (platform_id);

GRANT SELECT, INSERT, DELETE ON game_platforms TO app;

-- +goose Down
DROP TABLE game_platforms;
