-- +goose Up
CREATE TABLE saved_games (
    user_id  UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    game_id  UUID        NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    saved_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, game_id)
);

CREATE INDEX idx_saved_games_game_id ON saved_games (game_id);

GRANT SELECT, INSERT, DELETE ON saved_games TO app;

-- +goose Down
DROP TABLE saved_games;
