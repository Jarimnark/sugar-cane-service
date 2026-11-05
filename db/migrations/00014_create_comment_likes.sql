-- +goose Up
CREATE TABLE comment_likes (
    user_id    UUID        NOT NULL REFERENCES users (id)         ON DELETE CASCADE,
    comment_id UUID        NOT NULL REFERENCES game_comments (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, comment_id)
);

CREATE INDEX idx_comment_likes_comment_id ON comment_likes (comment_id);

GRANT SELECT, INSERT, DELETE ON comment_likes TO app;

-- +goose Down
DROP TABLE comment_likes;
