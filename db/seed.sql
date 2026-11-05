-- db/seed.sql
--
-- Local-dev seed data. Idempotent: safe to re-run.
-- Load with:
--   psql "$DATABASE_URL" -f db/seed.sql
--
-- Fixed UUIDs so curl examples can hardcode IDs.
-- Dev session token lets you hit authed routes without going through OAuth:
--   -H 'Cookie: session_id=dev_session_token_1234567890abcdef0000000000000000'

BEGIN;

-- ─────────────────────────────────────────────────────────────
-- Genres
-- ─────────────────────────────────────────────────────────────
INSERT INTO genres (id, name, slug) VALUES
    ('a1111111-0000-0000-0000-000000000001', 'Action',   'action'),
    ('a1111111-0000-0000-0000-000000000002', 'Puzzle',   'puzzle'),
    ('a1111111-0000-0000-0000-000000000003', 'Strategy', 'strategy'),
    ('a1111111-0000-0000-0000-000000000004', 'Arcade',   'arcade'),
    ('a1111111-0000-0000-0000-000000000005', 'RPG',      'rpg')
ON CONFLICT (slug) DO NOTHING;

-- ─────────────────────────────────────────────────────────────
-- Platforms
-- ─────────────────────────────────────────────────────────────
INSERT INTO platforms (id, name, slug) VALUES
    ('b2222222-0000-0000-0000-000000000001', 'Web',     'web'),
    ('b2222222-0000-0000-0000-000000000002', 'Android', 'android'),
    ('b2222222-0000-0000-0000-000000000003', 'iOS',     'ios')
ON CONFLICT (slug) DO NOTHING;

-- ─────────────────────────────────────────────────────────────
-- Users (OAuth-style: password_hash = NULL)
--   u1 = the test user you'll log in as
--   u2 = a second owner so you can test cross-user ownership checks
-- ─────────────────────────────────────────────────────────────
INSERT INTO users (id, email, password_hash, display_name, bio, email_verified) VALUES
    ('11111111-1111-1111-1111-111111111111', 'test@local.dev',  NULL, 'Test User',  'Local dev tester',  TRUE),
    ('22222222-2222-2222-2222-222222222222', 'other@local.dev', NULL, 'Other Dev',  'Second owner',      TRUE)
ON CONFLICT (email) DO NOTHING;

-- OAuth link for u1 (so RemoveConnection has something to delete)
INSERT INTO oauth_accounts (user_id, provider, provider_id) VALUES
    ('11111111-1111-1111-1111-111111111111', 'google', 'google-sub-u1-local')
ON CONFLICT (provider, provider_id) DO NOTHING;

-- ─────────────────────────────────────────────────────────────
-- Session for u1 — plug this into the session_id cookie.
-- expires_at = NOW() + 30 days.
-- ─────────────────────────────────────────────────────────────
INSERT INTO sessions (id, user_id, ip_address, user_agent, expires_at) VALUES
    ('dev_session_token_1234567890abcdef0000000000000000',
     '11111111-1111-1111-1111-111111111111',
     '127.0.0.1',
     'curl/local',
     NOW() + INTERVAL '30 days')
ON CONFLICT (id) DO UPDATE
    SET expires_at = EXCLUDED.expires_at,
        last_used_at = NOW();

-- ─────────────────────────────────────────────────────────────
-- Games
-- Mix of published + drafted, two owners.
-- like_count / visit_count seeded so /home trending/popular ordering is visible.
-- ─────────────────────────────────────────────────────────────
INSERT INTO games (id, owner_id, slug, name, dev, pitch, description, bg, players, tags, status, like_count, visit_count, link_web) VALUES
    ('c3333333-0000-0000-0000-000000000001',
     '11111111-1111-1111-1111-111111111111',
     'neon-runner', 'Neon Runner', 'Test User',
     'A fast arcade runner through neon streets.',
     '{"v":1,"blocks":[{"type":"p","children":[{"text":"Dash through neon streets."}]}]}',
     '#FF5733', '1', ARRAY['arcade','runner','neon'],
     'published', 42, 500, 'https://example.com/neon-runner'),

    ('c3333333-0000-0000-0000-000000000002',
     '11111111-1111-1111-1111-111111111111',
     'cube-logic', 'Cube Logic', 'Test User',
     'A chill puzzle game about rotating cubes.',
     '{"v":1,"blocks":[{"type":"p","children":[{"text":"Rotate cubes to match the target."}]}]}',
     '#3498DB', '1', ARRAY['puzzle','relaxing'],
     'published', 120, 300, 'https://example.com/cube-logic'),

    ('c3333333-0000-0000-0000-000000000003',
     '22222222-2222-2222-2222-222222222222',
     'tactic-grid', 'Tactic Grid', 'Other Dev',
     'Turn-based tactics on a 6x6 grid.',
     '{"v":1,"blocks":[{"type":"p","children":[{"text":"Outsmart the AI on a 6x6 grid."}]}]}',
     '#2ECC71', '1-2', ARRAY['strategy','tactics','turn-based'],
     'published', 85, 700, 'https://example.com/tactic-grid'),

    ('c3333333-0000-0000-0000-000000000004',
     '22222222-2222-2222-2222-222222222222',
     'pixel-quest', 'Pixel Quest', 'Other Dev',
     'A retro pixel-art RPG with a tiny party of three.',
     '{"v":1,"blocks":[{"type":"p","children":[{"text":"A short retro RPG."}]}]}',
     '#9B59B6', '1', ARRAY['rpg','retro','pixel'],
     'published', 60, 200, 'https://example.com/pixel-quest'),

    ('c3333333-0000-0000-0000-000000000005',
     '11111111-1111-1111-1111-111111111111',
     'ninja-jump', 'Ninja Jump', 'Test User',
     'Tap to jump across rooftops. Beat your best run.',
     '{"v":1,"blocks":[{"type":"p","children":[{"text":"One tap, infinite rooftops."}]}]}',
     '#E67E22', '1', ARRAY['action','arcade','runner'],
     'published', 15, 900, 'https://example.com/ninja-jump'),

    -- A drafted game for u1 — tests /me/game listing and publish flow.
    ('c3333333-0000-0000-0000-000000000006',
     '11111111-1111-1111-1111-111111111111',
     'secret-wip', 'Secret WIP', 'Test User',
     'Work in progress — not published yet.',
     '{"v":1,"blocks":[{"type":"p","children":[{"text":"Still cooking."}]}]}',
     '#95A5A6', '1', ARRAY['wip'],
     'drafted', 0, 0, NULL)
ON CONFLICT (slug) DO NOTHING;

-- ─────────────────────────────────────────────────────────────
-- game_genres (many-to-many)
-- ─────────────────────────────────────────────────────────────
INSERT INTO game_genres (game_id, genre_id) VALUES
    ('c3333333-0000-0000-0000-000000000001', 'a1111111-0000-0000-0000-000000000004'), -- Neon Runner → Arcade
    ('c3333333-0000-0000-0000-000000000001', 'a1111111-0000-0000-0000-000000000001'), -- Neon Runner → Action
    ('c3333333-0000-0000-0000-000000000002', 'a1111111-0000-0000-0000-000000000002'), -- Cube Logic → Puzzle
    ('c3333333-0000-0000-0000-000000000003', 'a1111111-0000-0000-0000-000000000003'), -- Tactic Grid → Strategy
    ('c3333333-0000-0000-0000-000000000004', 'a1111111-0000-0000-0000-000000000005'), -- Pixel Quest → RPG
    ('c3333333-0000-0000-0000-000000000005', 'a1111111-0000-0000-0000-000000000001'), -- Ninja Jump → Action
    ('c3333333-0000-0000-0000-000000000005', 'a1111111-0000-0000-0000-000000000004')  -- Ninja Jump → Arcade
ON CONFLICT DO NOTHING;

-- ─────────────────────────────────────────────────────────────
-- game_platforms (many-to-many)
-- ─────────────────────────────────────────────────────────────
INSERT INTO game_platforms (game_id, platform_id) VALUES
    ('c3333333-0000-0000-0000-000000000001', 'b2222222-0000-0000-0000-000000000001'), -- Neon Runner → Web
    ('c3333333-0000-0000-0000-000000000001', 'b2222222-0000-0000-0000-000000000002'), -- Neon Runner → Android
    ('c3333333-0000-0000-0000-000000000002', 'b2222222-0000-0000-0000-000000000001'), -- Cube Logic → Web
    ('c3333333-0000-0000-0000-000000000003', 'b2222222-0000-0000-0000-000000000001'), -- Tactic Grid → Web
    ('c3333333-0000-0000-0000-000000000004', 'b2222222-0000-0000-0000-000000000003'), -- Pixel Quest → iOS
    ('c3333333-0000-0000-0000-000000000005', 'b2222222-0000-0000-0000-000000000002'), -- Ninja Jump → Android
    ('c3333333-0000-0000-0000-000000000005', 'b2222222-0000-0000-0000-000000000003')  -- Ninja Jump → iOS
ON CONFLICT DO NOTHING;

-- ─────────────────────────────────────────────────────────────
-- u1 likes two games owned by u2 (so /me/list and like counts have data)
-- ─────────────────────────────────────────────────────────────
INSERT INTO game_likes (user_id, game_id) VALUES
    ('11111111-1111-1111-1111-111111111111', 'c3333333-0000-0000-0000-000000000003'),
    ('11111111-1111-1111-1111-111111111111', 'c3333333-0000-0000-0000-000000000004')
ON CONFLICT DO NOTHING;

-- ─────────────────────────────────────────────────────────────
-- u1's saved list
-- ─────────────────────────────────────────────────────────────
INSERT INTO saved_games (user_id, game_id) VALUES
    ('11111111-1111-1111-1111-111111111111', 'c3333333-0000-0000-0000-000000000003'),
    ('11111111-1111-1111-1111-111111111111', 'c3333333-0000-0000-0000-000000000002')
ON CONFLICT DO NOTHING;

COMMIT;
