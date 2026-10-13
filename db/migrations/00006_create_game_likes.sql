-- +goose Up
CREATE TABLE game_likes (
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    game_id    UUID        NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, game_id)
);

CREATE INDEX idx_game_likes_game_id ON game_likes (game_id);

GRANT SELECT, INSERT, DELETE ON game_likes TO app;

-- +goose Down
DROP TABLE game_likes;
