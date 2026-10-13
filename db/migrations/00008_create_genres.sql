-- +goose Up
CREATE TABLE genres (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(100) NOT NULL,
    slug       VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT genres_name_key UNIQUE (name),
    CONSTRAINT genres_slug_key UNIQUE (slug)
);

GRANT SELECT, INSERT, UPDATE, DELETE ON genres TO app;

-- +goose Down
DROP TABLE genres;
