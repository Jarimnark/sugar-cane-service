-- +goose Up
CREATE TABLE users (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    email          VARCHAR(255) NOT NULL,
    password_hash  TEXT,
    display_name   VARCHAR(100) NOT NULL,
    bio            TEXT,
    photo_url      TEXT,
    email_verified BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT users_email_key UNIQUE (email)
);

CREATE INDEX idx_users_email ON users (email);

GRANT SELECT, INSERT, UPDATE, DELETE ON users TO app;

-- +goose Down
DROP TABLE users;
