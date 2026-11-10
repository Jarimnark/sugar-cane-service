-- +goose Up
-- +goose StatementBegin

-- Seed the 3 canonical platforms from the Hopbu design. Idempotent so this
-- migration can rerun alongside any manual rows already present (e.g. on
-- staging where someone may have inserted ad-hoc values).
INSERT INTO platforms (name, slug) VALUES
    ('Web game',    'web'),
    ('Mobile game', 'mobile'),
    ('PC game',     'pc')
ON CONFLICT (slug) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- We don't DELETE the seeded platforms on rollback: any games linked to them
-- in game_platforms would get cascade-deleted if their platform row went away,
-- and the platforms table itself still exists. Restore from a backup if you
-- need a truly clean rollback.

-- +goose StatementEnd
