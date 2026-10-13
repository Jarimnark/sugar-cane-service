-- +goose Up
CREATE TABLE game_genres (
    game_id  UUID NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    genre_id UUID NOT NULL REFERENCES genres (id) ON DELETE CASCADE,
    PRIMARY KEY (game_id, genre_id)
);

CREATE INDEX idx_game_genres_genre_id ON game_genres (genre_id);

GRANT SELECT, INSERT, DELETE ON game_genres TO app;

-- +goose Down
DROP TABLE game_genres;
