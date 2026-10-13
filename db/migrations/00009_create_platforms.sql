-- +goose Up
CREATE TABLE platforms (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(100) NOT NULL,
    slug       VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT platforms_name_key UNIQUE (name),
    CONSTRAINT platforms_slug_key UNIQUE (slug)
);

GRANT SELECT, INSERT, UPDATE, DELETE ON platforms TO app;

-- +goose Down
DROP TABLE platforms;
