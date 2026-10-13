-- +goose Up
CREATE TABLE game_comments (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id    UUID        NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    content    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_game_comments_game_id ON game_comments (game_id, created_at DESC);
CREATE INDEX idx_game_comments_user_id ON game_comments (user_id);

GRANT SELECT, INSERT, UPDATE, DELETE ON game_comments TO app;

-- +goose Down
DROP TABLE game_comments;
