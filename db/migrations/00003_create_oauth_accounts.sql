-- +goose Up
CREATE TABLE oauth_accounts (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    provider    VARCHAR(50)  NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT oauth_accounts_provider_key UNIQUE (provider, provider_id)
);

CREATE INDEX idx_oauth_accounts_user_id ON oauth_accounts (user_id);

GRANT SELECT, INSERT, DELETE ON oauth_accounts TO app;

-- +goose Down
DROP TABLE oauth_accounts;
