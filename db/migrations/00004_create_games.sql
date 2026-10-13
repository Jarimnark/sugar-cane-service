-- +goose Up
CREATE TABLE games (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    slug        VARCHAR(255) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    link_web     TEXT,
    link_android TEXT,
    link_ios     TEXT,
    domain       VARCHAR(255),
    dev         VARCHAR(255) NOT NULL,
    pitch       TEXT         NOT NULL,
    description JSONB        NOT NULL DEFAULT '{"v":1,"blocks":[]}', -- custom Doc format: {v, blocks:[{type,level?,ordered?,children?,items?}]}
    bg          VARCHAR(7)   NOT NULL,
    players     VARCHAR(20)  NOT NULL,
    tags        TEXT[]       NOT NULL DEFAULT '{}',
    status      VARCHAR(20)  NOT NULL DEFAULT 'drafted' CHECK (status IN ('drafted', 'published')),
    like_count  INTEGER      NOT NULL DEFAULT 0,
    visit_count INTEGER      NOT NULL DEFAULT 0,
    badge       VARCHAR(50),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT games_slug_key UNIQUE (slug)
);

CREATE INDEX idx_games_owner_id    ON games (owner_id);
CREATE INDEX idx_games_status      ON games (status);
CREATE INDEX idx_games_created_at  ON games (created_at DESC);
CREATE INDEX idx_games_like_count  ON games (like_count DESC);
CREATE INDEX idx_games_visit_count ON games (visit_count DESC);
CREATE INDEX idx_games_tags ON games USING GIN (tags);

GRANT SELECT, INSERT, UPDATE, DELETE ON games TO app;

-- +goose Down
DROP TABLE games;
