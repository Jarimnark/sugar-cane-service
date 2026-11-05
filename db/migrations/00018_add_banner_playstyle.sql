-- +goose Up
-- +goose StatementBegin

-- 1. play_styles master (7 rows).
CREATE TABLE play_styles (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(50)  NOT NULL,
    slug       VARCHAR(50)  NOT NULL,
    sort       INTEGER      NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT play_styles_name_key UNIQUE (name),
    CONSTRAINT play_styles_slug_key UNIQUE (slug)
);

INSERT INTO play_styles (name, slug, sort) VALUES
    ('Casual',       'casual',       1),
    ('Co-op',        'co-op',        2),
    ('PvP',          'pvp',          3),
    ('PvE',          'pve',          4),
    ('Story-Driven', 'story-driven', 5),
    ('Fast Reflex',  'fast-reflex',  6),
    ('Brain',        'brain',        7);

GRANT SELECT, INSERT, UPDATE, DELETE ON play_styles TO app;

-- 2. player_counts master (3 rows).
CREATE TABLE player_counts (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(50)  NOT NULL,
    slug       VARCHAR(50)  NOT NULL,
    sort       INTEGER      NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT player_counts_name_key UNIQUE (name),
    CONSTRAINT player_counts_slug_key UNIQUE (slug)
);

INSERT INTO player_counts (name, slug, sort) VALUES
    ('1 Player',                 'single-player',     1),
    ('2 Players (Same Device)',  'local-2p',          2),
    ('Online Multiplayer',       'online-multiplayer', 3);

GRANT SELECT, INSERT, UPDATE, DELETE ON player_counts TO app;

-- 3. Genre seed (13 rows). Idempotent upsert by slug so any existing rows stay.
INSERT INTO genres (name, slug) VALUES
    ('Action',            'action'),
    ('Adventure',         'adventure'),
    ('Puzzle',            'puzzle'),
    ('Strategy',          'strategy'),
    ('MMORPG',            'mmorpg'),
    ('Simulation',        'simulation'),
    ('Racing & Driving',  'racing-driving'),
    ('Shooting',          'shooting'),
    ('Platformer',        'platformer'),
    ('Idle & Clicker',    'idle-clicker'),
    ('Horror',            'horror'),
    ('Visual Novel',      'visual-novel'),
    ('Rhythm & Music',    'rhythm-music')
ON CONFLICT (slug) DO NOTHING;

-- 4. games: new columns + relax NOT NULL on previously required fields.
ALTER TABLE games
    ADD COLUMN banner_url TEXT,
    ADD COLUMN play_style VARCHAR(50);

ALTER TABLE games
    ALTER COLUMN dev     DROP NOT NULL,
    ALTER COLUMN bg      DROP NOT NULL,
    ALTER COLUMN players DROP NOT NULL;

-- 5. Backfill existing rows so the new publish rules + FKs are satisfied.
--    - banner_url: Mesa-palette placeholder so game cards render without art.
--    - play_style: 'casual' is the safest default (lowest commitment taxonomy).
--    - players: normalize legacy free-text ("1-4", "100+ players", etc.) to the
--      3 canonical slugs. Everything multi-player maps to online; single stays
--      single; anything unrecognizable falls back to single-player.
UPDATE games SET banner_url = 'https://placehold.co/1600x900/F7E3C9/433F34?text=Hopbu'
 WHERE banner_url IS NULL;

UPDATE games SET play_style = 'casual' WHERE play_style IS NULL;

UPDATE games SET players = CASE
    WHEN players ILIKE '%single%' OR players ILIKE '1 player%' THEN 'single-player'
    WHEN players ILIKE '%same device%' OR players ILIKE '2 player%' OR players = '2 players' THEN 'local-2p'
    WHEN players ILIKE '%online%' OR players ILIKE '%+%' OR players ~ '[-–]'
         OR players ILIKE '%multiplayer%' OR players ILIKE '%massive%' THEN 'online-multiplayer'
    ELSE 'single-player'
END
WHERE players IS NULL OR players NOT IN ('single-player', 'local-2p', 'online-multiplayer');

-- 6. FKs. Point play_style at play_styles.slug and players at player_counts.slug.
ALTER TABLE games
    ADD CONSTRAINT games_play_style_fk
        FOREIGN KEY (play_style) REFERENCES play_styles (slug),
    ADD CONSTRAINT games_player_count_fk
        FOREIGN KEY (players) REFERENCES player_counts (slug);

CREATE INDEX idx_games_play_style ON games (play_style);
CREATE INDEX idx_games_players    ON games (players);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_games_players;
DROP INDEX IF EXISTS idx_games_play_style;

ALTER TABLE games DROP CONSTRAINT IF EXISTS games_player_count_fk;
ALTER TABLE games DROP CONSTRAINT IF EXISTS games_play_style_fk;

-- We leave backfilled values in place on down-migrate — they're still valid
-- strings. Callers that need to roll all the way back should restore from a
-- backup rather than rely on the down path.
ALTER TABLE games
    ALTER COLUMN players SET NOT NULL,
    ALTER COLUMN bg      SET NOT NULL,
    ALTER COLUMN dev     SET NOT NULL;

ALTER TABLE games
    DROP COLUMN play_style,
    DROP COLUMN banner_url;

DROP TABLE player_counts;
DROP TABLE play_styles;

-- Genre rows are intentionally NOT removed on rollback — removing them would
-- cascade-delete every game_genres link for games that used them.

-- +goose StatementEnd
