-- +goose Up
CREATE TABLE sessions (
    id           VARCHAR(64)  PRIMARY KEY,
    user_id      UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    ip_address   INET,
    user_agent   TEXT,
    last_used_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ  NOT NULL
);

CREATE INDEX idx_sessions_user_id    ON sessions (user_id);
CREATE INDEX idx_sessions_expires_at ON sessions (expires_at);

GRANT SELECT, INSERT, UPDATE, DELETE ON sessions TO app;

-- +goose Down
DROP TABLE sessions;
